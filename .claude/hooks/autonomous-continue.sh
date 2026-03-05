#!/bin/bash
# Autonomous Execution Stop Hook (v4.5 - Auto Reset)
# Checks for pending work and blocks Claude from stopping if more work exists
#
# v4.5 features:
# - Auto-reset iteration counter on new sprint detection
# - Auto-reset after completion acknowledged (AUTONOMOUS_COMPLETE removed)
# - Tracks last_sprint_id to detect sprint changes
#
# v4.4 features:
# - Rapid loop detection: stops when iterations happen faster than 5 seconds
# - Prevents runaway auto_decide_fallback loops that waste iterations
# - Tracks last_iteration_time and fast_loop_count in workflow state
# - Notifies on stuck detection before stopping
#
# v4.3 features:
# - Distinct notifications for sprint completion vs handoff completion
# - Sprint name included in handoff notifications for cross-referencing
# - Project name extracted from sprint ID (not launch directory)
# - Improved notification summaries and descriptions
#
# v4.2 features:
# - Validation checks before accepting <comware:done>
# - Runs tests, build, typecheck, lint before completion
# - Blocks completion if must_pass checks fail
# - Retry logic with configurable max_retries
# - Validation notifications to Slack
#
# v4.1 features:
# - Per-handoff notifications: <comware:handoff_complete id="...">
# - Handoff tracking in workflow.yaml
#
# v4 features:
# - Slack + Desktop notifications via notify_all()
# - Explicit signal detection: <comware:done>, <comware:blocked>, <comware:paused>
# - Stuck loop detection from transcript analysis
# - Notifications on: completion, blocked, paused, limits, errors, stuck
#
# v3 features:
# - Goal-aware execution: Uses .claude/state/goal-state.yaml for completion detection
# - Completion criteria: Continues until ALL goal criteria are met
# - Session spanning: Goals persist across sessions via checkpoints
#
# v2 features:
# - Response-aware: Parses Claude's response to detect continuation questions
# - Explicit YES/NO decisions for "Shall I continue with Phase X?"
#
# Configuration: .claude/autonomous.yaml
# Goal State: .claude/state/goal-state.yaml (optional, for goal-driven execution)
# Scope: .claude/autonomous-scope.yaml (optional, for multi-phase work)
# State: .claude/state/workflow.yaml
# Pause: touch .claude/STOP

# Determine project root from hook location (not pwd, which may be a subdirectory)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
CONFIG_FILE="${PROJECT_ROOT}/.claude/autonomous.yaml"
SCOPE_FILE="${PROJECT_ROOT}/.claude/autonomous-scope.yaml"
STATE_FILE="${PROJECT_ROOT}/.claude/state/workflow.yaml"
STOP_FILE="${PROJECT_ROOT}/.claude/STOP"
LOG_FILE="${PROJECT_ROOT}/.claude/state/execution.log"
GOAL_STATE_FILE="${PROJECT_ROOT}/.claude/state/goal-state.yaml"
GOAL_FUNCTIONS_FILE="${SCRIPT_DIR}/goal-aware-functions.sh"

# Source goal-aware functions if available
if [ -f "$GOAL_FUNCTIONS_FILE" ]; then
    source "$GOAL_FUNCTIONS_FILE"
    GOAL_AWARE=true
else
    GOAL_AWARE=false
fi

# -----------------------------------------------------------------------------
# CLI logging helper (stderr output visible in Claude Code CLI)
# -----------------------------------------------------------------------------
log_cli() {
    echo "[autonomous] $1" >&2
}

log_cli "Hook called at $(date '+%H:%M:%S') [project: $PROJECT_ROOT]"

# -----------------------------------------------------------------------------
# Helper functions for JSON output
# -----------------------------------------------------------------------------
output_block() {
    local reason="$1"
    log_cli "BLOCKING - $reason"
    # Escape quotes and newlines in reason for JSON
    reason=$(echo "$reason" | sed 's/"/\\"/g' | tr '\n' ' ')
    echo "{\"decision\": \"block\", \"reason\": \"$reason\"}"
    exit 0
}

output_allow_stop() {
    local reason="${1:-No pending work}"
    log_cli "ALLOWING STOP - $reason"
    # No output needed - just exit 0 to allow Claude to stop
    exit 0
}

# -----------------------------------------------------------------------------
# Notification Functions (v4 feature)
# Sends alerts to Slack and macOS desktop when significant events occur
# -----------------------------------------------------------------------------

# Get project name for notifications
# Priority: project.yaml > directory name (no longer uses sprint ID)
get_project_name() {
    local project_name=""

    # Try to read from project.yaml (most reliable source)
    local project_yaml="${PROJECT_ROOT}/.claude/project.yaml"
    if [ -f "$project_yaml" ]; then
        project_name=$(yq '.project.name // ""' "$project_yaml" 2>/dev/null || echo "")
    fi

    # Fall back to directory name if not found in project.yaml
    if [ -z "$project_name" ]; then
        project_name=$(basename "$PROJECT_ROOT")
    fi

    echo "$project_name"
}

# Get sprint ID for notifications
get_sprint_id() {
    if [ -f "$STATE_FILE" ]; then
        yq '.workflow.id // ""' "$STATE_FILE" 2>/dev/null || echo ""
    else
        echo ""
    fi
}

# Get sprint name (human readable) from sprint ID
get_sprint_name() {
    local sprint_id=$(get_sprint_id)
    if [ -n "$sprint_id" ]; then
        # Convert sprint-2026-01-25-marketplace-production-ready to "Marketplace Production Ready"
        echo "$sprint_id" | sed -E 's/^sprint-[0-9]{4}-[0-9]{2}-[0-9]{2}-//' | tr '-' ' ' | awk '{for(i=1;i<=NF;i++) $i=toupper(substr($i,1,1)) tolower(substr($i,2))}1'
    else
        echo "Unknown Sprint"
    fi
}

# Get sprint/goal title for notifications
get_execution_title() {
    local title=""

    # Try goal-state.yaml first
    if [ -f "$GOAL_STATE_FILE" ]; then
        title=$(yq '.goal.title // ""' "$GOAL_STATE_FILE" 2>/dev/null || echo "")
    fi

    # Try autonomous-scope.yaml if no goal title
    if [ -z "$title" ] && [ -f "$SCOPE_FILE" ]; then
        title=$(yq '.title // .description // ""' "$SCOPE_FILE" 2>/dev/null || echo "")
    fi

    # Try workflow.yaml for current phase
    if [ -z "$title" ] && [ -f "$STATE_FILE" ]; then
        local phase=$(yq '.workflow.current_phase // ""' "$STATE_FILE" 2>/dev/null || echo "")
        if [ -n "$phase" ]; then
            title="Phase: $phase"
        fi
    fi

    # Default if nothing found
    if [ -z "$title" ]; then
        title="Autonomous Execution"
    fi

    echo "$title"
}

# Calculate duration from workflow start time
get_execution_duration() {
    if [ ! -f "$STATE_FILE" ]; then
        echo "unknown"
        return
    fi

    local started=$(yq '.workflow.started // ""' "$STATE_FILE" 2>/dev/null || echo "")

    if [ -z "$started" ]; then
        echo "unknown"
        return
    fi

    # Convert ISO timestamp to epoch seconds
    local start_epoch=$(date -j -f "%Y-%m-%dT%H:%M:%S" "${started%%.*}" +%s 2>/dev/null || echo "")

    if [ -z "$start_epoch" ]; then
        # Try alternative format
        start_epoch=$(date -j -f "%Y-%m-%d %H:%M:%S" "$started" +%s 2>/dev/null || echo "")
    fi

    if [ -z "$start_epoch" ]; then
        echo "unknown"
        return
    fi

    local now_epoch=$(date +%s)
    local duration_secs=$((now_epoch - start_epoch))

    # Format as human-readable
    if [ "$duration_secs" -lt 60 ]; then
        echo "${duration_secs}s"
    elif [ "$duration_secs" -lt 3600 ]; then
        local mins=$((duration_secs / 60))
        local secs=$((duration_secs % 60))
        echo "${mins}m ${secs}s"
    else
        local hours=$((duration_secs / 3600))
        local mins=$(((duration_secs % 3600) / 60))
        echo "${hours}h ${mins}m"
    fi
}

# Get summary of completed work from todos
get_execution_summary() {
    local summary=""

    # Try to get completion stats from todos
    if [ -n "$TODO_FILE" ] && [ -f "$TODO_FILE" ]; then
        local completed=$(jq -r '[.todos[] | select(.status == "completed")] | length' "$TODO_FILE" 2>/dev/null || echo "0")
        local total=$(jq -r '.todos | length' "$TODO_FILE" 2>/dev/null || echo "0")

        if [ "$total" -gt 0 ]; then
            summary="$completed/$total tasks completed"

            # Get last completed task name
            local last_task=$(jq -r '[.todos[] | select(.status == "completed")] | last | .content // ""' "$TODO_FILE" 2>/dev/null || echo "")
            if [ -n "$last_task" ]; then
                # Truncate if too long
                if [ ${#last_task} -gt 40 ]; then
                    last_task="${last_task:0:37}..."
                fi
                summary="$summary • Last: $last_task"
            fi
        fi
    fi

    # If no todo summary, try goal progress
    if [ -z "$summary" ] && [ -f "$GOAL_STATE_FILE" ]; then
        local progress=$(yq '.goal.progress.overall_percent // 0' "$GOAL_STATE_FILE" 2>/dev/null || echo "0")
        local criteria_met=$(yq '[.goal.completion_criteria[] | select(.met == true)] | length' "$GOAL_STATE_FILE" 2>/dev/null || echo "0")
        local criteria_total=$(yq '.goal.completion_criteria | length' "$GOAL_STATE_FILE" 2>/dev/null || echo "0")

        if [ "$criteria_total" -gt 0 ]; then
            summary="${progress}% complete ($criteria_met/$criteria_total criteria met)"
        fi
    fi

    # Default
    if [ -z "$summary" ]; then
        summary="Execution finished"
    fi

    echo "$summary"
}

# Send Slack notification using blocks format
notify_slack() {
    local emoji="$1"
    local title="$2"
    local message="$3"
    local color="${4:-#6366f1}"  # Default indigo

    # Get webhook URL from config (project-level, then global fallback)
    local webhook_url=$(yq '.notifications.webhook_url // ""' "$CONFIG_FILE" 2>/dev/null || echo "")

    # Fallback to global config if project-level not set
    if [ -z "$webhook_url" ] || [ "$webhook_url" = "null" ]; then
        webhook_url=$(yq '.notifications.slack_webhook // ""' ~/.claude/config.yaml 2>/dev/null || echo "")
    fi

    if [ -z "$webhook_url" ] || [ "$webhook_url" = "null" ]; then
        log_cli "No Slack webhook configured (checked project and global) - skipping notification"
        return 0
    fi

    local project_name=$(get_project_name)
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')

    # Get enriched context for more informative notifications
    local exec_title=$(get_execution_title)
    local duration=$(get_execution_duration)
    local summary=$(get_execution_summary)

    # Escape special characters for JSON
    exec_title=$(echo "$exec_title" | sed 's/"/\\"/g' | tr '\n' ' ')
    summary=$(echo "$summary" | sed 's/"/\\"/g' | tr '\n' ' ')
    message=$(echo "$message" | sed 's/"/\\"/g' | tr '\n' ' ')

    # Build JSON payload with blocks format (required by Slack)
    # Using Claude/Anthropic branding for the bot identity
    local payload=$(cat << EOF
{
    "username": "Claude Code",
    "icon_url": "https://www.anthropic.com/images/icons/apple-touch-icon.png",
    "blocks": [
        {
            "type": "header",
            "text": {
                "type": "plain_text",
                "text": "${emoji} ${title}",
                "emoji": true
            }
        },
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": "*📋 Task:* ${exec_title}\n*📁 Project:* ${project_name}"
            }
        },
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": "*📊 Summary:* ${summary}\n*⏱️ Duration:* ${duration}"
            }
        },
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": "*Status:* ${message}"
            }
        },
        {
            "type": "context",
            "elements": [
                {
                    "type": "mrkdwn",
                    "text": "🤖 Autonomous execution • Iteration ${CURRENT_ITER:-0}/${MAX_ITER:-50} • ${timestamp}"
                }
            ]
        }
    ]
}
EOF
)

    # Send notification (background, don't block hook)
    curl -s -X POST -H 'Content-type: application/json' \
        -d "$payload" \
        "$webhook_url" > /dev/null 2>&1 &

    log_cli "Slack notification sent: $title - $exec_title"
}

# Send macOS desktop notification
notify_desktop() {
    local title="$1"
    local message="$2"
    local sound="${3:-Glass}"  # Default notification sound

    # Only works on macOS
    if [ "$(uname)" != "Darwin" ]; then
        return 0
    fi

    local project_name=$(get_project_name)

    # Use osascript for notification
    osascript -e "display notification \"${message}\" with title \"${title}\" subtitle \"${project_name}\" sound name \"${sound}\"" 2>/dev/null &

    log_cli "Desktop notification sent: $title"
}

# Send to all notification channels
notify_all() {
    local type="$1"      # done, blocked, paused, limit, errors, stuck
    local message="$2"

    case "$type" in
        "done")
            # Use specialized sprint complete notification
            notify_sprint_complete "$message"
            notify_desktop "🏁 Sprint Complete" "$message" "Hero"
            ;;
        "blocked")
            notify_slack "🚫" "Execution Blocked" "$message" "#ef4444"
            notify_desktop "🚫 Blocked" "$message" "Basso"
            ;;
        "paused")
            notify_slack "⏸️" "Execution Paused" "$message" "#f59e0b"
            notify_desktop "⏸️ Paused" "$message" "Pop"
            ;;
        "limit")
            notify_slack "⚠️" "Iteration Limit Reached" "$message" "#f59e0b"
            notify_desktop "⚠️ Limit" "$message" "Sosumi"
            ;;
        "errors")
            notify_slack "❌" "Max Errors Reached" "$message" "#ef4444"
            notify_desktop "❌ Errors" "$message" "Basso"
            ;;
        "stuck")
            notify_slack "🔄" "Stuck Loop Detected" "$message" "#f59e0b"
            notify_desktop "🔄 Stuck" "$message" "Funk"
            ;;
        "goal_complete")
            notify_slack "🎯" "Goal Complete" "$message" "#22c55e"
            notify_desktop "🎯 Goal Complete" "$message" "Hero"
            ;;
        "handoff_complete")
            # Handled by specialized notify_handoff_complete function
            # This is just a fallback
            notify_slack "✓" "Handoff Complete" "$message" "#10b981"
            notify_desktop "✓ Handoff" "$message" "Pop"
            ;;
        *)
            notify_slack "ℹ️" "Notification" "$message" "#6366f1"
            notify_desktop "ℹ️ Update" "$message" "Pop"
            ;;
    esac
}

# Send handoff completion notification (distinct from sprint complete)
notify_handoff_complete() {
    local handoff_id="$1"
    local handoff_count="$2"

    local webhook_url=$(yq '.notifications.webhook_url // ""' "$CONFIG_FILE" 2>/dev/null || echo "")
    if [ -z "$webhook_url" ] || [ "$webhook_url" = "null" ]; then
        webhook_url=$(yq '.notifications.slack_webhook // ""' ~/.claude/config.yaml 2>/dev/null || echo "")
    fi

    if [ -z "$webhook_url" ] || [ "$webhook_url" = "null" ]; then
        log_cli "No Slack webhook configured (checked project and global) - skipping handoff notification"
        return 0
    fi

    local project_name=$(get_project_name)
    local sprint_name=$(get_sprint_name)
    local sprint_id=$(get_sprint_id)
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    local duration=$(get_execution_duration)

    # Escape special characters
    handoff_id=$(echo "$handoff_id" | sed 's/"/\\"/g')
    sprint_name=$(echo "$sprint_name" | sed 's/"/\\"/g')

    # Build payload - distinct format from sprint complete
    local payload=$(cat << EOF
{
    "username": "Claude Code",
    "icon_url": "https://www.anthropic.com/images/icons/apple-touch-icon.png",
    "blocks": [
        {
            "type": "header",
            "text": {
                "type": "plain_text",
                "text": "✓ Handoff Complete",
                "emoji": true
            }
        },
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": "*📋 Sprint:* ${sprint_name}\n*📁 Project:* ${project_name}"
            }
        },
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": "*✓ Completed:* \`${handoff_id}\`\n*📊 Progress:* ${handoff_count} handoffs complete"
            }
        },
        {
            "type": "context",
            "elements": [
                {
                    "type": "mrkdwn",
                    "text": "⏱️ ${duration} • ${timestamp}"
                }
            ]
        }
    ]
}
EOF
)

    curl -s -X POST -H 'Content-type: application/json' \
        -d "$payload" \
        "$webhook_url" > /dev/null 2>&1 &

    log_cli "Handoff notification sent: $handoff_id"
}

# Send sprint start notification
notify_sprint_start() {
    local webhook_url=$(yq '.notifications.webhook_url // ""' "$CONFIG_FILE" 2>/dev/null || echo "")
    if [ -z "$webhook_url" ] || [ "$webhook_url" = "null" ]; then
        webhook_url=$(yq '.notifications.slack_webhook // ""' ~/.claude/config.yaml 2>/dev/null || echo "")
    fi

    if [ -z "$webhook_url" ] || [ "$webhook_url" = "null" ]; then
        log_cli "No Slack webhook configured (checked project and global) - skipping sprint start notification"
        return 0
    fi

    local project_name=$(get_project_name)
    local sprint_name=$(get_sprint_name)
    local sprint_id=$(get_sprint_id)
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')

    # Escape special characters
    sprint_name=$(echo "$sprint_name" | sed 's/"/\\"/g')

    # Build payload
    local payload=$(cat << EOF
{
    "username": "Claude Code",
    "icon_url": "https://www.anthropic.com/images/icons/apple-touch-icon.png",
    "blocks": [
        {
            "type": "header",
            "text": {
                "type": "plain_text",
                "text": "🚀 Sprint Started",
                "emoji": true
            }
        },
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": "*📋 Sprint:* ${sprint_name}\n*📁 Project:* ${project_name}"
            }
        },
        {
            "type": "context",
            "elements": [
                {
                    "type": "mrkdwn",
                    "text": "🚀 Sprint started • ${timestamp}"
                }
            ]
        }
    ]
}
EOF
)

    curl -s -X POST -H 'Content-type: application/json' \
        -d "$payload" \
        "$webhook_url" > /dev/null 2>&1 &

    log_cli "Sprint start notification sent: $sprint_name ($project_name)"
}

# Send sprint completion notification (distinct from handoff)
notify_sprint_complete() {
    local message="$1"

    local webhook_url=$(yq '.notifications.webhook_url // ""' "$CONFIG_FILE" 2>/dev/null || echo "")
    if [ -z "$webhook_url" ] || [ "$webhook_url" = "null" ]; then
        webhook_url=$(yq '.notifications.slack_webhook // ""' ~/.claude/config.yaml 2>/dev/null || echo "")
    fi

    if [ -z "$webhook_url" ] || [ "$webhook_url" = "null" ]; then
        log_cli "No Slack webhook configured (checked project and global) - skipping sprint notification"
        return 0
    fi

    local project_name=$(get_project_name)
    local sprint_name=$(get_sprint_name)
    local sprint_id=$(get_sprint_id)
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    local duration=$(get_execution_duration)
    local handoff_count=$(get_handoff_count)

    # Get summary from todos or other sources
    local summary=$(get_execution_summary)

    # Escape special characters
    sprint_name=$(echo "$sprint_name" | sed 's/"/\\"/g')
    summary=$(echo "$summary" | sed 's/"/\\"/g' | tr '\n' ' ')
    message=$(echo "$message" | sed 's/"/\\"/g' | tr '\n' ' ')

    # Build payload - distinct format with sprint focus
    local payload=$(cat << EOF
{
    "username": "Claude Code",
    "icon_url": "https://www.anthropic.com/images/icons/apple-touch-icon.png",
    "blocks": [
        {
            "type": "header",
            "text": {
                "type": "plain_text",
                "text": "🏁 Sprint Complete",
                "emoji": true
            }
        },
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": "*📋 Sprint:* ${sprint_name}\n*📁 Project:* ${project_name}"
            }
        },
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": "*📊 Summary:* ${summary}\n*✓ Handoffs:* ${handoff_count} completed"
            }
        },
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": "*⏱️ Duration:* ${duration}\n*Status:* ${message}"
            }
        },
        {
            "type": "context",
            "elements": [
                {
                    "type": "mrkdwn",
                    "text": "🏁 Sprint finished • Iteration ${CURRENT_ITER:-0}/${MAX_ITER:-50} • ${timestamp}"
                }
            ]
        }
    ]
}
EOF
)

    curl -s -X POST -H 'Content-type: application/json' \
        -d "$payload" \
        "$webhook_url" > /dev/null 2>&1 &

    log_cli "Sprint completion notification sent: $sprint_name"
}

# -----------------------------------------------------------------------------
# Track completed handoff in workflow.yaml
# -----------------------------------------------------------------------------
record_handoff_completion() {
    local handoff_id="$1"
    local timestamp=$(date -Iseconds)

    if [ ! -f "$STATE_FILE" ]; then
        return 1
    fi

    # Initialize handoffs_completed array if it doesn't exist
    yq -i '.workflow.handoffs_completed //= []' "$STATE_FILE" 2>/dev/null || true

    # Check if handoff already recorded (avoid duplicates)
    local exists=$(yq ".workflow.handoffs_completed[] | select(.id == \"$handoff_id\")" "$STATE_FILE" 2>/dev/null || echo "")
    if [ -n "$exists" ]; then
        log_cli "Handoff $handoff_id already recorded"
        return 0
    fi

    # Add new handoff record
    yq -i ".workflow.handoffs_completed += [{\"id\": \"$handoff_id\", \"completed_at\": \"$timestamp\"}]" "$STATE_FILE" 2>/dev/null || true

    # Get count of completed handoffs
    local count=$(yq '.workflow.handoffs_completed | length' "$STATE_FILE" 2>/dev/null || echo "0")

    log_cli "Recorded handoff completion: $handoff_id (total: $count)"
    return 0
}

# Get count of completed handoffs
get_handoff_count() {
    if [ ! -f "$STATE_FILE" ]; then
        echo "0"
        return
    fi
    yq '.workflow.handoffs_completed | length // 0' "$STATE_FILE" 2>/dev/null || echo "0"
}

# -----------------------------------------------------------------------------
# Validation Functions (v4.2)
# Run checks before accepting completion signal
# -----------------------------------------------------------------------------

# Check if validation is enabled and get mode
get_validation_mode() {
    if [ ! -f "$CONFIG_FILE" ]; then
        echo "none"
        return
    fi

    local enabled=$(yq '.validation.enabled // false' "$CONFIG_FILE" 2>/dev/null || echo "false")
    if [ "$enabled" != "true" ]; then
        echo "none"
        return
    fi

    yq '.validation.mode // "none"' "$CONFIG_FILE" 2>/dev/null || echo "none"
}

# Run a single validation check
# Returns: "pass", "fail", or "warn" followed by output
run_single_check() {
    local name="$1"
    local command="$2"
    local timeout="$3"
    local must_pass="$4"

    log_cli "Running validation check: $name"

    # Run command with timeout, capture output and exit code
    local output
    local exit_code

    # Use timeout command if available
    if command -v timeout &> /dev/null; then
        output=$(timeout "${timeout}s" bash -c "$command" 2>&1)
        exit_code=$?
    elif command -v gtimeout &> /dev/null; then
        output=$(gtimeout "${timeout}s" bash -c "$command" 2>&1)
        exit_code=$?
    else
        # Fallback without timeout
        output=$(bash -c "$command" 2>&1)
        exit_code=$?
    fi

    # Handle timeout (exit code 124)
    if [ $exit_code -eq 124 ]; then
        echo "timeout|Check timed out after ${timeout}s"
        return
    fi

    # Determine result based on exit code and must_pass
    if [ $exit_code -eq 0 ]; then
        echo "pass|$output"
    elif [ "$must_pass" = "true" ]; then
        echo "fail|$output"
    else
        echo "warn|$output"
    fi
}

# Run all validation checks
# Returns JSON-like structure with results
run_validation_checks() {
    local mode=$(get_validation_mode)

    if [ "$mode" = "none" ]; then
        log_cli "Validation disabled (mode=none)"
        echo "skip|Validation disabled"
        return
    fi

    if [ "$mode" = "ci" ]; then
        log_cli "CI validation not implemented yet"
        echo "skip|CI validation not implemented"
        return
    fi

    # Local validation mode
    log_cli "Running local validation checks..."

    # Change to project directory
    cd "$PROJECT_ROOT" || {
        echo "fail|Cannot change to project directory"
        return
    }

    # Get number of checks
    local check_count=$(yq '.validation.local.checks | length // 0' "$CONFIG_FILE" 2>/dev/null || echo "0")

    if [ "$check_count" -eq 0 ]; then
        log_cli "No validation checks configured"
        echo "skip|No checks configured"
        return
    fi

    local all_pass=true
    local any_fail=false
    local results=""
    local failed_checks=""
    local passed_count=0
    local failed_count=0
    local warn_count=0

    # Iterate through checks
    for i in $(seq 0 $((check_count - 1))); do
        local name=$(yq ".validation.local.checks[$i].name" "$CONFIG_FILE" 2>/dev/null)
        local command=$(yq ".validation.local.checks[$i].command" "$CONFIG_FILE" 2>/dev/null)
        local timeout=$(yq ".validation.local.checks[$i].timeout // 60" "$CONFIG_FILE" 2>/dev/null)
        local must_pass=$(yq ".validation.local.checks[$i].must_pass // true" "$CONFIG_FILE" 2>/dev/null)

        # Skip if command not found
        if [ -z "$command" ] || [ "$command" = "null" ]; then
            continue
        fi

        local result=$(run_single_check "$name" "$command" "$timeout" "$must_pass")
        local status=$(echo "$result" | cut -d'|' -f1)
        local output=$(echo "$result" | cut -d'|' -f2-)

        # Truncate output to first 500 chars for notifications
        local short_output="${output:0:500}"

        case "$status" in
            "pass")
                results="${results}✓ $name: passed\n"
                ((passed_count++))
                ;;
            "fail")
                results="${results}✗ $name: FAILED\n"
                failed_checks="${failed_checks}\n✗ $name: $short_output"
                any_fail=true
                all_pass=false
                ((failed_count++))
                ;;
            "warn")
                results="${results}⚠ $name: warnings\n"
                ((warn_count++))
                ;;
            "timeout")
                results="${results}⏱ $name: timed out\n"
                if [ "$must_pass" = "true" ]; then
                    failed_checks="${failed_checks}\n⏱ $name: timed out after ${timeout}s"
                    any_fail=true
                    all_pass=false
                    ((failed_count++))
                else
                    ((warn_count++))
                fi
                ;;
        esac
    done

    # Return summary
    if [ "$any_fail" = "true" ]; then
        echo "fail|${passed_count} passed, ${failed_count} failed, ${warn_count} warnings${failed_checks}"
    elif [ "$all_pass" = "true" ]; then
        echo "pass|${passed_count} passed, ${warn_count} warnings"
    else
        echo "warn|${passed_count} passed, ${warn_count} warnings"
    fi
}

# Format validation results for display
format_validation_results() {
    local status="$1"
    local details="$2"

    case "$status" in
        "pass")
            echo "✅ All validation checks passed: $details"
            ;;
        "fail")
            echo "❌ Validation failed: $details"
            ;;
        "warn")
            echo "⚠️ Validation passed with warnings: $details"
            ;;
        "skip")
            echo "⏭️ Validation skipped: $details"
            ;;
    esac
}

# Send validation notification to Slack
notify_validation_result() {
    local status="$1"
    local details="$2"
    local attempts="$3"
    local max_retries="$4"

    local webhook_url=$(yq '.notifications.webhook_url // ""' "$CONFIG_FILE" 2>/dev/null || echo "")
    if [ -z "$webhook_url" ] || [ "$webhook_url" = "null" ]; then
        webhook_url=$(yq '.notifications.slack_webhook // ""' ~/.claude/config.yaml 2>/dev/null || echo "")
    fi

    if [ -z "$webhook_url" ] || [ "$webhook_url" = "null" ]; then
        log_cli "No Slack webhook configured (checked project and global) - skipping validation notification"
        return 0
    fi

    local project_name=$(get_project_name)
    local exec_title=$(get_execution_title)
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')

    # Escape special characters
    exec_title=$(echo "$exec_title" | sed 's/"/\\"/g' | tr '\n' ' ')
    details=$(echo "$details" | sed 's/"/\\"/g' | tr '\n' ' ')

    local emoji color header
    case "$status" in
        "pass")
            emoji="✅"
            color="#22c55e"
            header="Sprint Validation Passed"
            ;;
        "fail")
            emoji="❌"
            color="#ef4444"
            header="Sprint Validation Failed (attempt $attempts/$max_retries)"
            ;;
        "warn")
            emoji="⚠️"
            color="#f59e0b"
            header="Sprint Validation Passed (with warnings)"
            ;;
        "skip")
            emoji="⏭️"
            color="#6b7280"
            header="Validation Skipped"
            ;;
    esac

    local payload=$(cat << EOF
{
    "username": "Claude Code",
    "icon_url": "https://www.anthropic.com/images/icons/apple-touch-icon.png",
    "blocks": [
        {
            "type": "header",
            "text": {
                "type": "plain_text",
                "text": "${emoji} ${header}",
                "emoji": true
            }
        },
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": "*📋 Task:* ${exec_title}\n*📁 Project:* ${project_name}"
            }
        },
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": "*Checks:* ${details}"
            }
        },
        {
            "type": "context",
            "elements": [
                {
                    "type": "mrkdwn",
                    "text": "🔍 Validation • ${timestamp}"
                }
            ]
        }
    ]
}
EOF
)

    curl -s -X POST -H 'Content-type: application/json' \
        -d "$payload" \
        "$webhook_url" > /dev/null 2>&1 &

    log_cli "Validation notification sent: $header"
}

# Track validation attempts in workflow.yaml
update_validation_state() {
    local status="$1"
    local error="$2"
    local timestamp=$(date -Iseconds)

    if [ ! -f "$STATE_FILE" ]; then
        return 1
    fi

    # Initialize validation section if needed
    yq -i '.workflow.validation //= {}' "$STATE_FILE" 2>/dev/null || true

    # Update validation state
    yq -i ".workflow.validation.last_result = \"$status\"" "$STATE_FILE" 2>/dev/null || true
    yq -i ".workflow.validation.last_checked = \"$timestamp\"" "$STATE_FILE" 2>/dev/null || true

    if [ "$status" = "pass" ]; then
        yq -i ".workflow.validation.passed_at = \"$timestamp\"" "$STATE_FILE" 2>/dev/null || true
        yq -i ".workflow.validation.attempts = 0" "$STATE_FILE" 2>/dev/null || true
    else
        # Increment attempts
        local attempts=$(yq '.workflow.validation.attempts // 0' "$STATE_FILE" 2>/dev/null || echo "0")
        attempts=$((attempts + 1))
        yq -i ".workflow.validation.attempts = $attempts" "$STATE_FILE" 2>/dev/null || true

        # Store error (truncated)
        local short_error="${error:0:200}"
        short_error=$(echo "$short_error" | sed 's/"/\\"/g' | tr '\n' ' ')
        yq -i ".workflow.validation.last_error = \"$short_error\"" "$STATE_FILE" 2>/dev/null || true
    fi

    log_cli "Updated validation state: $status"
}

# Get current validation attempt count
get_validation_attempts() {
    if [ ! -f "$STATE_FILE" ]; then
        echo "0"
        return
    fi
    yq '.workflow.validation.attempts // 0' "$STATE_FILE" 2>/dev/null || echo "0"
}

# -----------------------------------------------------------------------------
# Detect explicit signals in Claude's response
# <comware:done>, <comware:blocked>, <comware:paused>, <comware:handoff_complete>
# -----------------------------------------------------------------------------
detect_explicit_signal() {
    local response="$1"

    if echo "$response" | grep -q "<comware:done>"; then
        echo "done"
        return 0
    elif echo "$response" | grep -q "<comware:blocked>"; then
        echo "blocked"
        return 0
    elif echo "$response" | grep -q "<comware:paused>"; then
        echo "paused"
        return 0
    elif echo "$response" | grep -q "<comware:handoff_complete"; then
        echo "handoff_complete"
        return 0
    fi

    return 1
}

# Extract handoff ID from handoff_complete signal
# macOS compatible (BSD grep doesn't support -P, so use sed instead)
extract_handoff_id() {
    local response="$1"
    # Match: <comware:handoff_complete id="quality-001"> or similar
    echo "$response" | sed -n 's/.*<comware:handoff_complete[^>]*id="\([^"]*\)".*/\1/p' | head -1
}

# -----------------------------------------------------------------------------
# Check dependencies
# -----------------------------------------------------------------------------
if ! command -v yq &> /dev/null; then
    echo "Error: yq is required but not installed. Install with: brew install yq" >&2
    exit 1
fi

if ! command -v jq &> /dev/null; then
    echo "Error: jq is required but not installed. Install with: brew install jq" >&2
    exit 1
fi

# -----------------------------------------------------------------------------
# Read stdin (contains context from Claude Code)
# -----------------------------------------------------------------------------
STDIN_DATA=$(cat)

# -----------------------------------------------------------------------------
# Extract Claude's last response for analysis
# The stdin contains transcript_path pointing to a JSONL file
# -----------------------------------------------------------------------------
TRANSCRIPT_PATH=$(echo "$STDIN_DATA" | jq -r '.transcript_path // ""' 2>/dev/null || echo "")
LAST_RESPONSE=""

if [ -n "$TRANSCRIPT_PATH" ] && [ -f "$TRANSCRIPT_PATH" ]; then
    # Read last assistant message from transcript (JSONL format - one JSON per line)
    if grep -q '"role":"assistant"' "$TRANSCRIPT_PATH" 2>/dev/null; then
        LAST_LINE=$(grep '"role":"assistant"' "$TRANSCRIPT_PATH" | tail -1)
        if [ -n "$LAST_LINE" ]; then
            # Extract text content from the message
            LAST_RESPONSE=$(echo "$LAST_LINE" | jq -r '
                .message.content |
                if type == "array" then
                    map(select(.type == "text")) | map(.text) | join("\n")
                elif type == "string" then
                    .
                else
                    ""
                end
            ' 2>/dev/null || echo "")
        fi
    fi
else
    log_cli "No transcript path found in stdin"
fi

log_cli "Last response length: ${#LAST_RESPONSE} chars"

# -----------------------------------------------------------------------------
# Check for explicit signals in Claude's response
# These take highest priority and trigger immediate notifications
# -----------------------------------------------------------------------------
if [ -n "$LAST_RESPONSE" ]; then
    EXPLICIT_SIGNAL=$(detect_explicit_signal "$LAST_RESPONSE")

    if [ -n "$EXPLICIT_SIGNAL" ]; then
        log_cli "Detected explicit signal: <comware:$EXPLICIT_SIGNAL>"

        case "$EXPLICIT_SIGNAL" in
            "done")
                # v4.2: Run validation checks before accepting completion
                VALIDATION_MODE=$(get_validation_mode)
                log_cli "Validation mode: $VALIDATION_MODE"

                if [ "$VALIDATION_MODE" != "none" ]; then
                    # Run validation checks
                    VALIDATION_RESULT=$(run_validation_checks)
                    VALIDATION_STATUS=$(echo "$VALIDATION_RESULT" | cut -d'|' -f1)
                    VALIDATION_DETAILS=$(echo "$VALIDATION_RESULT" | cut -d'|' -f2-)

                    log_cli "Validation result: $VALIDATION_STATUS"

                    # Get retry config
                    MAX_RETRIES=$(yq '.validation.max_retries // 3' "$CONFIG_FILE" 2>/dev/null || echo "3")
                    CURRENT_ATTEMPTS=$(get_validation_attempts)

                    case "$VALIDATION_STATUS" in
                        "pass"|"warn")
                            # Validation passed - allow completion
                            update_validation_state "pass" ""
                            notify_validation_result "$VALIDATION_STATUS" "$VALIDATION_DETAILS" "$CURRENT_ATTEMPTS" "$MAX_RETRIES"
                            notify_all "done" "Task completed with validation passed"
                            output_allow_stop "Explicit <comware:done> signal received - validation passed"
                            ;;
                        "skip")
                            # Validation disabled - allow completion
                            notify_all "done" "Task completed (validation skipped)"
                            output_allow_stop "Explicit <comware:done> signal received - validation skipped"
                            ;;
                        "fail")
                            # Validation failed - check retry count
                            NEW_ATTEMPTS=$((CURRENT_ATTEMPTS + 1))
                            update_validation_state "fail" "$VALIDATION_DETAILS"
                            notify_validation_result "fail" "$VALIDATION_DETAILS" "$NEW_ATTEMPTS" "$MAX_RETRIES"

                            if [ "$NEW_ATTEMPTS" -ge "$MAX_RETRIES" ]; then
                                # Max retries exceeded - force stop with failure
                                log_cli "Max validation retries ($MAX_RETRIES) exceeded - forcing stop"
                                yq -i ".workflow.validation.forced_stop = true" "$STATE_FILE" 2>/dev/null || true
                                notify_all "errors" "Validation failed after $MAX_RETRIES attempts - stopping"
                                output_allow_stop "Validation failed after $MAX_RETRIES attempts - VALIDATION_FAILED"
                            else
                                # Block and ask Claude to fix
                                log_cli "Validation failed (attempt $NEW_ATTEMPTS/$MAX_RETRIES) - blocking completion"

                                # Create actionable fix message
                                FIX_MESSAGE="VALIDATION FAILED - Fix these issues before signaling completion again:

$VALIDATION_DETAILS

Attempt $NEW_ATTEMPTS/$MAX_RETRIES. Run the failing command(s) to see full errors, fix them, then signal <comware:done> again."

                                output_block "$FIX_MESSAGE"
                            fi
                            ;;
                    esac
                else
                    # No validation configured - allow completion as before
                    notify_all "done" "Task completed (signal received) after $CURRENT_ITER iterations"
                    output_allow_stop "Explicit <comware:done> signal received"
                fi
                ;;
            "blocked")
                # Extract reason if provided: <comware:blocked reason="...">
                # macOS compatible (use sed instead of grep -P)
                BLOCK_REASON=$(echo "$LAST_RESPONSE" | sed -n 's/.*<comware:blocked[^>]*reason="\([^"]*\)".*/\1/p' | head -1)
                [ -z "$BLOCK_REASON" ] && BLOCK_REASON="Requires user intervention"
                notify_all "blocked" "$BLOCK_REASON"
                output_allow_stop "Explicit <comware:blocked> signal: $BLOCK_REASON"
                ;;
            "paused")
                # Extract reason if provided
                # macOS compatible (use sed instead of grep -P)
                PAUSE_REASON=$(echo "$LAST_RESPONSE" | sed -n 's/.*<comware:paused[^>]*reason="\([^"]*\)".*/\1/p' | head -1)
                [ -z "$PAUSE_REASON" ] && PAUSE_REASON="Waiting for user input"
                notify_all "paused" "$PAUSE_REASON"
                output_allow_stop "Explicit <comware:paused> signal: $PAUSE_REASON"
                ;;
            "handoff_complete")
                # Extract handoff ID and record completion
                HANDOFF_ID=$(extract_handoff_id "$LAST_RESPONSE")
                if [ -n "$HANDOFF_ID" ]; then
                    record_handoff_completion "$HANDOFF_ID"
                    HANDOFF_COUNT=$(get_handoff_count)
                    # Use specialized handoff notification (includes sprint name)
                    notify_handoff_complete "$HANDOFF_ID" "$HANDOFF_COUNT"
                    notify_desktop "✓ Handoff: $HANDOFF_ID" "$(get_sprint_name)" "Pop"
                    log_cli "Handoff $HANDOFF_ID complete, continuing execution"
                    # Don't stop - continue to next handoff
                    # Fall through to normal continuation logic
                else
                    log_cli "Handoff complete signal received but no ID found"
                fi
                ;;
        esac
    fi
fi

# -----------------------------------------------------------------------------
# NEW: Detect continuation questions in Claude's response
# These patterns indicate Claude is asking to continue, NOT that work is done
# -----------------------------------------------------------------------------
detect_continuation_question() {
    local response="$1"

    # Patterns that indicate a continuation question (case insensitive)
    local patterns=(
        "shall i continue"
        "should i continue"
        "do you want me to continue"
        "ready for phase"
        "shall i proceed"
        "should i proceed"
        "do you want me to proceed"
        "shall i move on"
        "should i move on"
        "ready to continue"
        "continue with phase"
        "proceed with phase"
        "move on to phase"
        "shall i start phase"
        "want me to start"
        "shall i implement"
        "should i implement"
        "shall i build"
        "should i build"
        "ready when you are"
    )

    local response_lower=$(echo "$response" | tr '[:upper:]' '[:lower:]')

    for pattern in "${patterns[@]}"; do
        if echo "$response_lower" | grep -q "$pattern"; then
            # Extract what they want to continue with (look for "Phase X" or similar)
            local continuation_target=""

            # Try to extract "Phase N" or similar context
            continuation_target=$(echo "$response" | grep -oiE "(phase\s*[0-9]+|step\s*[0-9]+|next\s+\w+|[A-Z][a-z]+\s+(feature|system|module|implementation))" | head -1)

            if [ -z "$continuation_target" ]; then
                # Try to get context from the line containing the question
                continuation_target=$(echo "$response" | grep -i "$pattern" | head -1 | sed 's/.*continue with\s*//i; s/\?.*//; s/shall i.*//i' | head -c 50)
            fi

            echo "$continuation_target"
            return 0
        fi
    done

    return 1
}

# -----------------------------------------------------------------------------
# NEW: Check if a continuation target is in scope
# Uses autonomous-scope.yaml if present, otherwise defaults to IN SCOPE
# -----------------------------------------------------------------------------
check_in_scope() {
    local target="$1"

    # If no scope file, default to in scope (permissive)
    if [ ! -f "$SCOPE_FILE" ]; then
        log_cli "No scope file - defaulting to in-scope"
        echo "in_scope"
        return 0
    fi

    # Check if auto_continue is enabled
    local auto_continue=$(yq '.auto_continue // true' "$SCOPE_FILE" 2>/dev/null || echo "true")
    if [ "$auto_continue" = "false" ]; then
        log_cli "auto_continue is false - requires user approval"
        echo "requires_approval"
        return 0
    fi

    # Check explicit exclusions
    local excluded=$(yq '.out_of_scope[]' "$SCOPE_FILE" 2>/dev/null || echo "")
    local target_lower=$(echo "$target" | tr '[:upper:]' '[:lower:]')

    while IFS= read -r exclusion; do
        local exclusion_lower=$(echo "$exclusion" | tr '[:upper:]' '[:lower:]')
        if echo "$target_lower" | grep -qi "$exclusion_lower"; then
            log_cli "Target '$target' matches exclusion '$exclusion'"
            echo "out_of_scope"
            return 0
        fi
    done <<< "$excluded"

    # Check phase limits
    local max_phase=$(yq '.limits.max_phase // 999' "$SCOPE_FILE" 2>/dev/null || echo "999")
    local phase_num=$(echo "$target" | grep -oE '[0-9]+' | head -1)

    if [ -n "$phase_num" ] && [ "$phase_num" -gt "$max_phase" ]; then
        log_cli "Phase $phase_num exceeds max_phase $max_phase"
        echo "out_of_scope"
        return 0
    fi

    echo "in_scope"
    return 0
}

# -----------------------------------------------------------------------------
# Check for context low signal in Claude's response
# If Claude says "Context low" or similar, run /compact and continue
# -----------------------------------------------------------------------------
if echo "$LAST_RESPONSE" | grep -qi "context low\|run /compact\|running low on context\|context.*running.*low"; then
    log_cli "Context low detected - instructing to run /compact and continue"
    output_block "Run /compact to free up context, then continue with the current task."
fi

# -----------------------------------------------------------------------------
# Exit silently if autonomous mode not configured
# -----------------------------------------------------------------------------
[ -f "$CONFIG_FILE" ] || exit 0

# Check if enabled
ENABLED=$(yq '.enabled // false' "$CONFIG_FILE" 2>/dev/null || echo "false")
[ "$ENABLED" = "true" ] || exit 0

# -----------------------------------------------------------------------------
# Check for user interrupt
# -----------------------------------------------------------------------------
if [ -f "$STOP_FILE" ]; then
    output_allow_stop "User interrupt (.claude/STOP exists)"
fi

# -----------------------------------------------------------------------------
# GOAL-AWARE CHECKING (v3 feature)
# If there's an active goal, use goal completion as primary signal
# -----------------------------------------------------------------------------
# Initialize goal completion flag
GOAL_COMPLETE=false

# Read current iteration for goal check (need state file first)
ensure_state_file() {
    if [ ! -f "$STATE_FILE" ]; then
        mkdir -p "$(dirname "$STATE_FILE")"
        cat > "$STATE_FILE" << 'YAML'
workflow:
  id: null
  started: null
  iteration: 0
  paused: false
  current_phase: null
  last_action: null
  last_action_status: null
  last_error: null
  error_count: 0
  checkpoints: []
YAML
    fi
}
ensure_state_file

# -----------------------------------------------------------------------------
# Auto-reset iteration counter (v4.5)
# Reset when: new sprint detected, or completion file cleaned up
# -----------------------------------------------------------------------------
auto_reset_iterations() {
    local current_sprint=$(yq '.workflow.id // "none"' "$STATE_FILE" 2>/dev/null || echo "none")
    local last_sprint=$(yq '.workflow.last_sprint_id // "none"' "$STATE_FILE" 2>/dev/null || echo "none")
    local current_iter=$(yq '.workflow.iteration // 0' "$STATE_FILE" 2>/dev/null || echo "0")

    # Check for new sprint
    if [ "$current_sprint" != "none" ] && [ "$current_sprint" != "$last_sprint" ]; then
        log_cli "New sprint detected ($current_sprint) - resetting iteration counter"
        yq -i ".workflow.iteration = 0" "$STATE_FILE" 2>/dev/null || true
        yq -i ".workflow.last_sprint_id = \"$current_sprint\"" "$STATE_FILE" 2>/dev/null || true
        yq -i ".workflow.fast_loop_count = 0" "$STATE_FILE" 2>/dev/null || true
        yq -i ".workflow.last_iteration_time = 0" "$STATE_FILE" 2>/dev/null || true

        # NOTE: Sprint start notification is now sent by task-start-notification.sh (PreToolUse hook)
        # when the Task tool is invoked. No need to duplicate here.
        log_cli "Sprint start notification delegated to task-start-notification.sh hook"

        return 0
    fi

    # Check if completion was acknowledged (file removed after being created)
    local complete_marker="${PROJECT_ROOT}/.claude/state/.last_complete_time"
    local complete_file="${PROJECT_ROOT}/.claude/AUTONOMOUS_COMPLETE"

    if [ ! -f "$complete_file" ] && [ -f "$complete_marker" ]; then
        # Completion file was removed (user acknowledged) - reset for next task
        log_cli "Previous task completed and acknowledged - resetting iteration counter"
        rm -f "$complete_marker"
        yq -i ".workflow.iteration = 0" "$STATE_FILE" 2>/dev/null || true
        yq -i ".workflow.fast_loop_count = 0" "$STATE_FILE" 2>/dev/null || true
        yq -i ".workflow.last_iteration_time = 0" "$STATE_FILE" 2>/dev/null || true
        return 0
    fi

    # Mark completion time if complete file exists
    if [ -f "$complete_file" ] && [ ! -f "$complete_marker" ]; then
        date +%s > "$complete_marker"
    fi

    return 1
}
auto_reset_iterations

CURRENT_ITER=$(yq '.workflow.iteration // 0' "$STATE_FILE" 2>/dev/null || echo "0")

if [ "$GOAL_AWARE" = "true" ] && [ -f "$GOAL_STATE_FILE" ]; then
    log_cli "Goal-aware mode: checking goal state"

    # Check if goal is complete
    GOAL_RESULT=$(check_goal_completion)
    GOAL_ACTION=$(echo "$GOAL_RESULT" | cut -d'|' -f1)
    GOAL_MESSAGE=$(echo "$GOAL_RESULT" | cut -d'|' -f2-)

    log_cli "Goal check result: $GOAL_ACTION"

    if [ "$GOAL_ACTION" = "continue" ]; then
        # Goal is not complete - block stop and continue

        # Update iteration counter
        NEW_ITER=$((CURRENT_ITER + 1))
        yq -i ".workflow.iteration = $NEW_ITER" "$STATE_FILE" 2>/dev/null || true
        yq -i ".workflow.error_count = 0" "$STATE_FILE" 2>/dev/null || true
        yq -i ".workflow.last_action = \"goal_continuation\"" "$STATE_FILE" 2>/dev/null || true

        # Update goal activity timestamp
        update_goal_activity 2>/dev/null || true

        # Log
        LOG_ENABLED=$(yq '.logging.enabled // true' "$CONFIG_FILE" 2>/dev/null || echo "true")
        if [ "$LOG_ENABLED" = "true" ]; then
            mkdir -p "$(dirname "$LOG_FILE")"
            GOAL_TITLE=$(yq '.goal.title // "Unknown"' "$GOAL_STATE_FILE" 2>/dev/null || echo "Unknown")
            GOAL_PROGRESS=$(yq '.goal.progress.overall_percent // 0' "$GOAL_STATE_FILE" 2>/dev/null || echo "0")
            echo "[$(date -Iseconds)] iter=$NEW_ITER action=goal_continuation goal=\"$GOAL_TITLE\" progress=${GOAL_PROGRESS}%" >> "$LOG_FILE"
        fi

        # Block with goal-specific message
        output_block "$GOAL_MESSAGE"
    elif [ "$GOAL_ACTION" = "stop" ]; then
        # Goal is complete - allow stop but log completion
        LOG_ENABLED=$(yq '.logging.enabled // true' "$CONFIG_FILE" 2>/dev/null || echo "true")
        if [ "$LOG_ENABLED" = "true" ]; then
            mkdir -p "$(dirname "$LOG_FILE")"
            GOAL_TITLE=$(yq '.goal.title // "Unknown"' "$GOAL_STATE_FILE" 2>/dev/null || echo "Unknown")
            echo "[$(date -Iseconds)] action=goal_complete goal=\"$GOAL_TITLE\" reason=\"$GOAL_MESSAGE\"" >> "$LOG_FILE"
        fi

        # Don't exit yet - let it fall through to allow final cleanup
        # But set a flag so we know goal is done
        GOAL_COMPLETE=true
        log_cli "Goal complete: $GOAL_MESSAGE"
    fi
fi

# -----------------------------------------------------------------------------
# Read configuration
# -----------------------------------------------------------------------------
MAX_ITER=$(yq '.limits.max_iterations // 50' "$CONFIG_FILE" 2>/dev/null || echo "50")
MAX_ERRORS=$(yq '.limits.max_consecutive_errors // 3' "$CONFIG_FILE" 2>/dev/null || echo "3")
AUTONOMY_LEVEL=$(yq '.autonomy_level // "semi"' "$CONFIG_FILE" 2>/dev/null || echo "semi")
AUTO_DECIDE=$(yq '.behavior.auto_decide // false' "$CONFIG_FILE" 2>/dev/null || echo "false")

log_cli "Config: max_iter=$MAX_ITER, max_errors=$MAX_ERRORS, auto_decide=$AUTO_DECIDE"

# -----------------------------------------------------------------------------
# Read current state (ensure_state_file already called in goal-aware section)
# -----------------------------------------------------------------------------
# CURRENT_ITER already set above
ERROR_COUNT=$(yq '.workflow.error_count // 0' "$STATE_FILE" 2>/dev/null || echo "0")
PAUSED=$(yq '.workflow.paused // false' "$STATE_FILE" 2>/dev/null || echo "false")

# Check if workflow is paused
if [ "$PAUSED" = "true" ]; then
    output_allow_stop "Workflow paused"
fi

# Check for completion marker (Claude signals done by touching this file)
COMPLETE_FILE="${PROJECT_ROOT}/.claude/AUTONOMOUS_COMPLETE"
if [ -f "$COMPLETE_FILE" ]; then
    if [ -f "$LOG_FILE" ]; then
        echo "[$(date -Iseconds)] Autonomous complete signal detected" >> "$LOG_FILE"
    fi
    # DO NOT delete the file here. If the allow-stop doesn't take effect
    # (e.g., race condition), the file must persist so subsequent hook
    # invocations also detect it. Cleanup happens in auto_reset_iterations
    # when a new sprint/task starts.
    output_allow_stop "Completion signal received"
fi

# -----------------------------------------------------------------------------
# Safety checks
# -----------------------------------------------------------------------------

# Stuck loop detection - check for repeated patterns in recent history
detect_stuck_loop() {
    if [ -z "$TRANSCRIPT_PATH" ] || [ ! -f "$TRANSCRIPT_PATH" ]; then
        return 1
    fi

    # Get last 10 assistant messages
    local recent_messages=$(grep '"role":"assistant"' "$TRANSCRIPT_PATH" 2>/dev/null | tail -10)
    local message_count=$(echo "$recent_messages" | wc -l | tr -d ' ')

    if [ "$message_count" -lt 5 ]; then
        return 1  # Not enough history
    fi

    # Check for repeated tool calls (same tool called 5+ times in a row)
    local tool_pattern=$(grep '"role":"assistant"' "$TRANSCRIPT_PATH" 2>/dev/null | tail -5 | \
        jq -r '.message.content[]? | select(.type == "tool_use") | .name' 2>/dev/null | \
        sort | uniq -c | sort -rn | head -1)

    local repeated_count=$(echo "$tool_pattern" | awk '{print $1}')
    local repeated_tool=$(echo "$tool_pattern" | awk '{print $2}')

    if [ -n "$repeated_count" ] && [ "$repeated_count" -ge 5 ]; then
        echo "Same tool ($repeated_tool) called $repeated_count times"
        return 0
    fi

    # Check for similar error messages repeated
    local error_pattern=$(grep -i "error\|failed\|exception" "$TRANSCRIPT_PATH" 2>/dev/null | tail -10 | \
        sort | uniq -c | sort -rn | head -1)

    local error_count=$(echo "$error_pattern" | awk '{print $1}')
    if [ -n "$error_count" ] && [ "$error_count" -ge 4 ]; then
        echo "Same error repeated $error_count times"
        return 0
    fi

    # Check for short/no-tool responses (Claude responding with just "Done." etc.)
    # This catches the case where Claude wants to stop but the hook keeps blocking.
    local last_5_text=$(grep '"role":"assistant"' "$TRANSCRIPT_PATH" 2>/dev/null | tail -5 | \
        jq -r '[.message.content[]? | select(.type == "text") | .text] | join("")' 2>/dev/null)
    local has_tool_use=$(grep '"role":"assistant"' "$TRANSCRIPT_PATH" 2>/dev/null | tail -5 | \
        jq -r '.message.content[]? | select(.type == "tool_use") | .name' 2>/dev/null | head -1)

    if [ -z "$has_tool_use" ] && [ -n "$last_5_text" ]; then
        # Last 5 responses had no tool calls - likely spinning with minimal text
        local short_count=0
        while IFS= read -r line; do
            local text_len=${#line}
            if [ "$text_len" -lt 100 ]; then
                short_count=$((short_count + 1))
            fi
        done <<< "$(grep '"role":"assistant"' "$TRANSCRIPT_PATH" 2>/dev/null | tail -5 | \
            jq -r '[.message.content[]? | select(.type == "text") | .text] | join("")' 2>/dev/null)"

        if [ "$short_count" -ge 4 ]; then
            echo "5 consecutive short responses with no tool calls (likely completion loop)"
            return 0
        fi
    fi

    return 1
}

# Check for stuck loop
STUCK_REASON=$(detect_stuck_loop)
if [ -n "$STUCK_REASON" ]; then
    log_cli "Stuck loop detected: $STUCK_REASON"
    if [ -f "$LOG_FILE" ]; then
        echo "[$(date -Iseconds)] Stuck loop detected: $STUCK_REASON" >> "$LOG_FILE"
    fi
    notify_all "stuck" "$STUCK_REASON - may need intervention"
    output_allow_stop "Stuck loop detected: $STUCK_REASON"
fi

# Check iteration limit
if [ "$CURRENT_ITER" -ge "$MAX_ITER" ]; then
    # Log and allow stop
    if [ -f "$LOG_FILE" ]; then
        echo "[$(date -Iseconds)] Max iterations reached ($CURRENT_ITER/$MAX_ITER)" >> "$LOG_FILE"
    fi
    notify_all "limit" "Reached $CURRENT_ITER/$MAX_ITER iterations"
    output_allow_stop "Max iterations reached ($CURRENT_ITER/$MAX_ITER)"
fi

# Check error threshold
if [ "$ERROR_COUNT" -ge "$MAX_ERRORS" ]; then
    if [ -f "$LOG_FILE" ]; then
        echo "[$(date -Iseconds)] Max errors reached ($ERROR_COUNT/$MAX_ERRORS)" >> "$LOG_FILE"
    fi
    notify_all "errors" "Hit $ERROR_COUNT consecutive errors"
    output_allow_stop "Max errors reached ($ERROR_COUNT/$MAX_ERRORS)"
fi

# -----------------------------------------------------------------------------
# Find pending todos
# NOTE: If goal-aware mode handled continuation, we won't reach here (output_block exits)
# This section is fallback for non-goal autonomous mode or when no goal exists
# -----------------------------------------------------------------------------
find_todo_file() {
    local candidates=(
        "${PROJECT_ROOT}/.claude/todos/current.json"
        "${HOME}/.claude/todos/current.json"
    )

    for candidate in "${candidates[@]}"; do
        if [ -f "$candidate" ]; then
            echo "$candidate"
            return 0
        fi
    done

    return 1
}

TODO_FILE=$(find_todo_file) || TODO_FILE=""

if [ -n "$TODO_FILE" ] && [ -f "$TODO_FILE" ]; then
    IN_PROGRESS=$(jq -r '[.todos[] | select(.status == "in_progress")] | length' "$TODO_FILE" 2>/dev/null || echo "0")
    PENDING=$(jq -r '[.todos[] | select(.status == "pending")] | length' "$TODO_FILE" 2>/dev/null || echo "0")
    COMPLETED=$(jq -r '[.todos[] | select(.status == "completed")] | length' "$TODO_FILE" 2>/dev/null || echo "0")
    TOTAL=$((IN_PROGRESS + PENDING + COMPLETED))

    log_cli "Todos: $IN_PROGRESS in_progress, $PENDING pending, $COMPLETED completed"

    if [ "$IN_PROGRESS" -gt 0 ] || [ "$PENDING" -gt 0 ]; then
        if [ "$IN_PROGRESS" -gt 0 ]; then
            NEXT_TASK=$(jq -r '.todos[] | select(.status == "in_progress") | .content' "$TODO_FILE" 2>/dev/null | head -1)
            TASK_STATUS="in_progress"
        else
            NEXT_TASK=$(jq -r '.todos[] | select(.status == "pending") | .content' "$TODO_FILE" 2>/dev/null | head -1)
            TASK_STATUS="pending"
        fi

        # Update iteration counter
        NEW_ITER=$((CURRENT_ITER + 1))
        yq -i ".workflow.iteration = $NEW_ITER" "$STATE_FILE" 2>/dev/null || true
        yq -i ".workflow.error_count = 0" "$STATE_FILE" 2>/dev/null || true
        yq -i ".workflow.last_action = \"continuation\"" "$STATE_FILE" 2>/dev/null || true

        # Log
        LOG_ENABLED=$(yq '.logging.enabled // true' "$CONFIG_FILE" 2>/dev/null || echo "true")
        if [ "$LOG_ENABLED" = "true" ]; then
            mkdir -p "$(dirname "$LOG_FILE")"
            echo "[$(date -Iseconds)] iter=$NEW_ITER status=$TASK_STATUS task=\"$NEXT_TASK\"" >> "$LOG_FILE"
        fi

        # Block Claude from stopping with reason
        REASON="Continue with: $NEXT_TASK [Autonomous mode: iteration $NEW_ITER/$MAX_ITER, $COMPLETED/$TOTAL completed]"
        output_block "$REASON"
    fi
else
    log_cli "No todo file found"
fi

# -----------------------------------------------------------------------------
# NEW: Response-aware continuation (BEFORE generic auto_decide)
# Check if Claude is asking a continuation question
# -----------------------------------------------------------------------------
if [ "$AUTO_DECIDE" = "true" ] && [ -n "$LAST_RESPONSE" ]; then
    log_cli "Checking for continuation question in response..."

    CONTINUATION_TARGET=$(detect_continuation_question "$LAST_RESPONSE")

    if [ -n "$CONTINUATION_TARGET" ]; then
        log_cli "Detected continuation question for: $CONTINUATION_TARGET"

        # Check scope
        SCOPE_STATUS=$(check_in_scope "$CONTINUATION_TARGET")
        log_cli "Scope status: $SCOPE_STATUS"

        # Update iteration counter
        NEW_ITER=$((CURRENT_ITER + 1))
        yq -i ".workflow.iteration = $NEW_ITER" "$STATE_FILE" 2>/dev/null || true
        yq -i ".workflow.error_count = 0" "$STATE_FILE" 2>/dev/null || true
        yq -i ".workflow.last_action = \"continuation_approved\"" "$STATE_FILE" 2>/dev/null || true

        case "$SCOPE_STATUS" in
            "in_scope")
                # Log
                LOG_ENABLED=$(yq '.logging.enabled // true' "$CONFIG_FILE" 2>/dev/null || echo "true")
                if [ "$LOG_ENABLED" = "true" ]; then
                    mkdir -p "$(dirname "$LOG_FILE")"
                    echo "[$(date -Iseconds)] iter=$NEW_ITER action=continuation_approved target=\"$CONTINUATION_TARGET\"" >> "$LOG_FILE"
                fi

                # EXPLICIT YES - This is the key fix!
                REASON="YES - Continue with ${CONTINUATION_TARGET}. [Autonomous mode: iteration $NEW_ITER/$MAX_ITER]"
                output_block "$REASON"
                ;;

            "out_of_scope")
                if [ -f "$LOG_FILE" ]; then
                    echo "[$(date -Iseconds)] iter=$NEW_ITER action=out_of_scope target=\"$CONTINUATION_TARGET\"" >> "$LOG_FILE"
                fi
                # Allow stop - out of scope
                output_allow_stop "Requested continuation ($CONTINUATION_TARGET) is out of scope"
                ;;

            "requires_approval")
                if [ -f "$LOG_FILE" ]; then
                    echo "[$(date -Iseconds)] iter=$NEW_ITER action=requires_approval target=\"$CONTINUATION_TARGET\"" >> "$LOG_FILE"
                fi
                # Allow stop - needs user approval
                output_allow_stop "Continuation requires user approval (auto_continue is false)"
                ;;
        esac
    fi
fi

# -----------------------------------------------------------------------------
# ENGINEER DECISION ENGINE (v5.1 - uses built-in agent hook)
# The engineer evaluator now runs as a separate type:agent Stop hook
# configured in .claude/settings.json. This avoids subprocess issues
# with claude -p. The command hook focuses on signals, safety, and logging.
# The agent hook focuses on intelligent evaluation.
# See settings.json for the agent hook configuration.
# -----------------------------------------------------------------------------

# -----------------------------------------------------------------------------
# Auto-decide mode: Generic fallback when no specific patterns detected
# This is the FALLBACK - continuation questions handled above
# Also falls back here if engineer decision is disabled or fails
# -----------------------------------------------------------------------------
if [ "$AUTO_DECIDE" = "true" ]; then
    # -------------------------------------------------------------------------
    # RATE LIMITING: Detect rapid iteration loops (v4.4 fix)
    # If iterations happen faster than MIN_ITERATION_SECONDS, it's likely a loop
    # where Claude isn't doing meaningful work between hook invocations.
    # -------------------------------------------------------------------------
    MIN_ITERATION_SECONDS=5
    MAX_FAST_ITERATIONS=5

    CURRENT_TIME=$(date +%s)
    LAST_ITER_TIME=$(yq '.workflow.last_iteration_time // 0' "$STATE_FILE" 2>/dev/null || echo "0")

    if [ "$LAST_ITER_TIME" -gt 0 ]; then
        ELAPSED=$((CURRENT_TIME - LAST_ITER_TIME))
        log_cli "Time since last iteration: ${ELAPSED}s"

        if [ "$ELAPSED" -lt "$MIN_ITERATION_SECONDS" ]; then
            # Iteration happened too fast - likely a loop
            FAST_LOOP_COUNT=$(yq '.workflow.fast_loop_count // 0' "$STATE_FILE" 2>/dev/null || echo "0")
            FAST_LOOP_COUNT=$((FAST_LOOP_COUNT + 1))
            yq -i ".workflow.fast_loop_count = $FAST_LOOP_COUNT" "$STATE_FILE" 2>/dev/null || true

            log_cli "Fast iteration detected: $FAST_LOOP_COUNT consecutive (threshold: $MAX_FAST_ITERATIONS)"

            if [ "$FAST_LOOP_COUNT" -ge "$MAX_FAST_ITERATIONS" ]; then
                # Too many fast iterations - this is a stuck loop
                log_cli "RAPID LOOP DETECTED: $FAST_LOOP_COUNT fast iterations - stopping"

                # Log the loop detection
                if [ -f "$LOG_FILE" ]; then
                    echo "[$(date -Iseconds)] RAPID LOOP DETECTED: $FAST_LOOP_COUNT iterations in <${MIN_ITERATION_SECONDS}s each" >> "$LOG_FILE"
                fi

                # Notify and stop
                notify_all "stuck" "Rapid iteration loop: $FAST_LOOP_COUNT iterations without meaningful work"

                # Reset counter for next run
                yq -i ".workflow.fast_loop_count = 0" "$STATE_FILE" 2>/dev/null || true

                output_allow_stop "Rapid loop detected: $FAST_LOOP_COUNT consecutive fast iterations"
            fi
        else
            # Normal iteration speed - reset the counter
            yq -i ".workflow.fast_loop_count = 0" "$STATE_FILE" 2>/dev/null || true
        fi
    fi

    # Store current time for next iteration check
    yq -i ".workflow.last_iteration_time = $CURRENT_TIME" "$STATE_FILE" 2>/dev/null || true

    # Update iteration counter
    NEW_ITER=$((CURRENT_ITER + 1))
    yq -i ".workflow.iteration = $NEW_ITER" "$STATE_FILE" 2>/dev/null || true
    yq -i ".workflow.error_count = 0" "$STATE_FILE" 2>/dev/null || true
    yq -i ".workflow.last_action = \"auto_decide\"" "$STATE_FILE" 2>/dev/null || true

    # Log
    LOG_ENABLED=$(yq '.logging.enabled // true' "$CONFIG_FILE" 2>/dev/null || echo "true")
    if [ "$LOG_ENABLED" = "true" ]; then
        mkdir -p "$(dirname "$LOG_FILE")"
        echo "[$(date -Iseconds)] iter=$NEW_ITER action=auto_decide_fallback" >> "$LOG_FILE"
    fi

    # Get custom prompt or use default
    AUTO_DECIDE_PROMPT=$(yq '.behavior.auto_decide_prompt // ""' "$CONFIG_FILE" 2>/dev/null || echo "")

    if [ -n "$AUTO_DECIDE_PROMPT" ]; then
        REASON="$AUTO_DECIDE_PROMPT"
    else
        # Enhanced decision prompt - more actionable
        read -r -d '' REASON << 'PROMPT' || true
AUTONOMOUS MODE [Iteration NEW_ITER/MAX_ITER]

ACTION REQUIRED: Execute one focused action. Do NOT ask questions.

DECISION PRIORITY (pick first that applies):
1. ERRORS/FAILURES? → Fix them immediately
2. ISSUES IDENTIFIED? → Address them (findings are work, not reports)
3. OPTIONS PRESENTED? → Select highest-impact, lowest-risk choice and execute
4. NEXT STEPS MENTIONED? → Execute the most important one
5. WORK READY TO CONTINUE? → Continue with the logical next step
6. TRULY COMPLETE? → Signal done: touch .claude/AUTONOMOUS_COMPLETE

STOP CONDITIONS (signal done ONLY if ALL apply):
• Original goal achieved AND verified working
• No actionable items remain
• No logical next steps exist
• Would significantly expand beyond original request

ANTI-PATTERNS:
• Asking "shall I continue?" - just continue
• Summarizing without acting
• Reporting issues without fixing

Execute now.
PROMPT
        REASON="${REASON//NEW_ITER/$NEW_ITER}"
        REASON="${REASON//MAX_ITER/$MAX_ITER}"
    fi

    output_block "$REASON"
fi

# -----------------------------------------------------------------------------
# Final sweep: Before stopping, ask Claude to verify nothing is outstanding
# This catches edge cases where work remains but wasn't captured in todos
# -----------------------------------------------------------------------------
FINAL_SWEEP_ENABLED=$(yq '.behavior.final_sweep // true' "$CONFIG_FILE" 2>/dev/null || echo "true")
FINAL_SWEEP_DONE=$(yq '.workflow.final_sweep_done // false' "$STATE_FILE" 2>/dev/null || echo "false")

if [ "$FINAL_SWEEP_ENABLED" = "true" ] && [ "$FINAL_SWEEP_DONE" = "false" ]; then
    log_cli "Running final sweep before stopping..."

    # Mark that we're doing the final sweep (prevents infinite loop)
    yq -i ".workflow.final_sweep_done = true" "$STATE_FILE" 2>/dev/null || true

    # Update iteration counter
    NEW_ITER=$((CURRENT_ITER + 1))
    yq -i ".workflow.iteration = $NEW_ITER" "$STATE_FILE" 2>/dev/null || true

    # Log
    LOG_ENABLED=$(yq '.logging.enabled // true' "$CONFIG_FILE" 2>/dev/null || echo "true")
    if [ "$LOG_ENABLED" = "true" ]; then
        mkdir -p "$(dirname "$LOG_FILE")"
        echo "[$(date -Iseconds)] iter=$NEW_ITER action=final_sweep" >> "$LOG_FILE"
    fi

    # Get custom final sweep prompt or use default
    FINAL_SWEEP_PROMPT=$(yq '.behavior.final_sweep_prompt // ""' "$CONFIG_FILE" 2>/dev/null || echo "")

    if [ -n "$FINAL_SWEEP_PROMPT" ]; then
        REASON="$FINAL_SWEEP_PROMPT"
    else
        # Check if there was a multi-phase plan in the session
        PHASE_CONTEXT=""
        if [ -n "$TRANSCRIPT_PATH" ] && [ -f "$TRANSCRIPT_PATH" ]; then
            # Look for phase mentions in the transcript
            if grep -qi "phase 2\|phase 3\|phase 4\|phase 5" "$TRANSCRIPT_PATH" 2>/dev/null; then
                PHASE_CONTEXT="IMPORTANT: A multi-phase plan was discussed. Check if later phases should be executed."
            fi
        fi

        read -r -d '' REASON << PROMPT || true
FINAL CHECK before completing autonomous session.

${PHASE_CONTEXT}

Review what was accomplished and verify:
1. Was a multi-phase plan proposed? If so, should we continue with the next phase?
2. Is there anything outstanding from the ORIGINAL user request?
3. Did any errors or issues get mentioned but not resolved?

If YES to any: Execute the most important remaining item (or ask about continuing to next phase).
If NO to all: Signal completion with: touch .claude/AUTONOMOUS_COMPLETE

Be thorough but don't invent new work. Only flag genuinely outstanding items.
PROMPT
    fi

    output_block "$REASON"
fi

# -----------------------------------------------------------------------------
# No pending work and final sweep done - allow Claude to stop
# -----------------------------------------------------------------------------
yq -i ".workflow.last_action = \"completed\"" "$STATE_FILE" 2>/dev/null || true
yq -i ".workflow.last_action_status = \"success\"" "$STATE_FILE" 2>/dev/null || true
yq -i ".workflow.final_sweep_done = false" "$STATE_FILE" 2>/dev/null || true  # Reset for next run

LOG_ENABLED=$(yq '.logging.enabled // true' "$CONFIG_FILE" 2>/dev/null || echo "true")
if [ "$LOG_ENABLED" = "true" ] && [ -f "$LOG_FILE" ]; then
    if [ "$GOAL_COMPLETE" = "true" ]; then
        GOAL_TITLE=$(yq '.goal.title // "Unknown"' "$GOAL_STATE_FILE" 2>/dev/null || echo "Unknown")
        echo "[$(date -Iseconds)] GOAL COMPLETE: \"$GOAL_TITLE\" after $CURRENT_ITER iterations" >> "$LOG_FILE"
    else
        echo "[$(date -Iseconds)] Workflow complete after $CURRENT_ITER iterations" >> "$LOG_FILE"
    fi
fi

if [ "$GOAL_COMPLETE" = "true" ]; then
    GOAL_TITLE=$(yq '.goal.title // "Unknown"' "$GOAL_STATE_FILE" 2>/dev/null || echo "Unknown")
    notify_all "goal_complete" "Goal achieved: $GOAL_TITLE"
    output_allow_stop "Goal complete - all criteria met"
else
    # Sprint completion notifications are now handled by task-notification.sh
    # (PostToolUse hook for Task tool) which includes handoff count.
    # This hook just handles session end logging.
    local sprint_id=$(get_sprint_id)
    local handoff_count=$(get_handoff_count)

    if [ -n "$sprint_id" ] && [ "$sprint_id" != "null" ] && [ "$handoff_count" -gt 0 ]; then
        # Sprint was executed - task-notification.sh already sent completion notice
        log_cli "Sprint $sprint_id completed with $handoff_count handoffs (notification sent by task hook)"
    else
        # Normal session end (no sprint execution)
        log_cli "Session ending without sprint execution"
    fi
    output_allow_stop "All work complete ($CURRENT_ITER iterations)"
fi
