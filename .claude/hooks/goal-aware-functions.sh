#!/bin/bash
# Goal-Aware Functions for Autonomous Hook
# Source this file from autonomous-continue.sh to add goal persistence support
#
# These functions check .claude/state/goal-state.yaml for:
# - Active goals with incomplete completion criteria
# - Phase progress and remaining phases
# - Session checkpointing needs
#
# Usage: source goal-aware-functions.sh
#        Then call check_goal_completion() in your hook

# Goal state file location
GOAL_STATE_FILE="${PROJECT_ROOT}/.claude/state/goal-state.yaml"
CHECKPOINTS_DIR="${PROJECT_ROOT}/.claude/state/checkpoints"

# -----------------------------------------------------------------------------
# Check if there's an active goal
# Returns: "true" if active goal exists, "false" otherwise
# -----------------------------------------------------------------------------
has_active_goal() {
    if [ ! -f "$GOAL_STATE_FILE" ]; then
        echo "false"
        return 1
    fi

    local status=$(yq '.goal.status // "none"' "$GOAL_STATE_FILE" 2>/dev/null || echo "none")

    if [ "$status" = "active" ] || [ "$status" = "paused" ]; then
        echo "true"
        return 0
    fi

    echo "false"
    return 1
}

# -----------------------------------------------------------------------------
# Get goal completion status
# Returns: JSON object with completion info
# -----------------------------------------------------------------------------
get_goal_completion_status() {
    if [ ! -f "$GOAL_STATE_FILE" ]; then
        echo '{"has_goal": false}'
        return 1
    fi

    # Count criteria met/total
    local total_criteria=$(yq '.goal.completion_criteria | length' "$GOAL_STATE_FILE" 2>/dev/null || echo "0")
    local met_criteria=$(yq '[.goal.completion_criteria[] | select(.met == true)] | length' "$GOAL_STATE_FILE" 2>/dev/null || echo "0")

    # Get phase info
    local total_phases=$(yq '.goal.phases | length' "$GOAL_STATE_FILE" 2>/dev/null || echo "0")
    local completed_phases=$(yq '[.goal.phases[] | select(.status == "completed")] | length' "$GOAL_STATE_FILE" 2>/dev/null || echo "0")
    local current_phase=$(yq '.goal.phases[] | select(.status == "in_progress") | .name' "$GOAL_STATE_FILE" 2>/dev/null | head -1 || echo "")

    # Get progress
    local progress=$(yq '.goal.progress.overall_percent // 0' "$GOAL_STATE_FILE" 2>/dev/null || echo "0")

    # Get goal title
    local title=$(yq '.goal.title // "Unknown Goal"' "$GOAL_STATE_FILE" 2>/dev/null || echo "Unknown Goal")

    cat << EOF
{
    "has_goal": true,
    "title": "$title",
    "criteria_met": $met_criteria,
    "criteria_total": $total_criteria,
    "phases_completed": $completed_phases,
    "phases_total": $total_phases,
    "current_phase": "$current_phase",
    "progress_percent": $progress,
    "is_complete": $([ "$met_criteria" -eq "$total_criteria" ] && [ "$total_criteria" -gt 0 ] && echo "true" || echo "false")
}
EOF
}

# -----------------------------------------------------------------------------
# Check if goal completion criteria are all met
# Returns: 0 if all met, 1 if not
# -----------------------------------------------------------------------------
is_goal_complete() {
    if [ ! -f "$GOAL_STATE_FILE" ]; then
        # No goal = consider complete (nothing to do)
        return 0
    fi

    local total=$(yq '.goal.completion_criteria | length' "$GOAL_STATE_FILE" 2>/dev/null || echo "0")
    local met=$(yq '[.goal.completion_criteria[] | select(.met == true)] | length' "$GOAL_STATE_FILE" 2>/dev/null || echo "0")

    if [ "$total" -eq 0 ]; then
        # No criteria defined = can't determine completion
        return 1
    fi

    if [ "$met" -eq "$total" ]; then
        return 0
    fi

    return 1
}

# -----------------------------------------------------------------------------
# Get next incomplete criterion
# Returns: The criterion text, or empty if all complete
# -----------------------------------------------------------------------------
get_next_criterion() {
    if [ ! -f "$GOAL_STATE_FILE" ]; then
        echo ""
        return
    fi

    yq '.goal.completion_criteria[] | select(.met != true) | .criterion' "$GOAL_STATE_FILE" 2>/dev/null | head -1
}

# -----------------------------------------------------------------------------
# Get current phase info
# Returns: Phase name and status
# -----------------------------------------------------------------------------
get_current_phase() {
    if [ ! -f "$GOAL_STATE_FILE" ]; then
        echo ""
        return
    fi

    local in_progress=$(yq '.goal.phases[] | select(.status == "in_progress") | .name' "$GOAL_STATE_FILE" 2>/dev/null | head -1)

    if [ -n "$in_progress" ]; then
        echo "$in_progress"
        return
    fi

    # If no in_progress, get first pending
    yq '.goal.phases[] | select(.status == "pending") | .name' "$GOAL_STATE_FILE" 2>/dev/null | head -1
}

# -----------------------------------------------------------------------------
# Check if checkpoint needed (based on time or phase change)
# Returns: "true" if checkpoint should be created, "false" otherwise
# -----------------------------------------------------------------------------
needs_checkpoint() {
    if [ ! -f "$GOAL_STATE_FILE" ]; then
        echo "false"
        return
    fi

    # Get last activity timestamp
    local last_activity=$(yq '.goal.last_activity // "1970-01-01T00:00:00Z"' "$GOAL_STATE_FILE" 2>/dev/null)

    # Get checkpoint frequency (default 15 minutes)
    local frequency_minutes=$(yq '.goal.runtime.checkpoint_frequency_minutes // 15' "$GOAL_STATE_FILE" 2>/dev/null || echo "15")

    # Calculate time difference (requires GNU date or compatible)
    local now=$(date +%s)
    local last=$(date -d "$last_activity" +%s 2>/dev/null || date -j -f "%Y-%m-%dT%H:%M:%S" "${last_activity%%Z*}" +%s 2>/dev/null || echo "0")

    local diff_minutes=$(( (now - last) / 60 ))

    if [ "$diff_minutes" -ge "$frequency_minutes" ]; then
        echo "true"
    else
        echo "false"
    fi
}

# -----------------------------------------------------------------------------
# Update goal progress (called by hook after confirming continuation)
# -----------------------------------------------------------------------------
update_goal_activity() {
    if [ ! -f "$GOAL_STATE_FILE" ]; then
        return
    fi

    local timestamp=$(date -Iseconds)
    yq -i ".goal.last_activity = \"$timestamp\"" "$GOAL_STATE_FILE" 2>/dev/null || true
}

# -----------------------------------------------------------------------------
# Generate continuation message for goal-driven execution
# More specific than generic auto_decide, uses goal context
# -----------------------------------------------------------------------------
generate_goal_continuation_message() {
    local status_json=$(get_goal_completion_status)

    local title=$(echo "$status_json" | jq -r '.title')
    local progress=$(echo "$status_json" | jq -r '.progress_percent')
    local current_phase=$(echo "$status_json" | jq -r '.current_phase')
    local criteria_met=$(echo "$status_json" | jq -r '.criteria_met')
    local criteria_total=$(echo "$status_json" | jq -r '.criteria_total')
    local phases_done=$(echo "$status_json" | jq -r '.phases_completed')
    local phases_total=$(echo "$status_json" | jq -r '.phases_total')

    local next_criterion=$(get_next_criterion)

    cat << EOF
GOAL-DRIVEN EXECUTION [Progress: ${progress}%]

Goal: ${title}
Phase: ${current_phase} (${phases_done}/${phases_total} complete)
Criteria: ${criteria_met}/${criteria_total} met

NEXT CRITERION TO SATISFY:
${next_criterion}

INSTRUCTIONS:
1. Continue working on current phase: ${current_phase}
2. Focus on satisfying: ${next_criterion}
3. Update progress as you complete work
4. Signal phase complete when all phase work is done
5. Do NOT stop until all ${criteria_total} criteria are met

Execute now.
EOF
}

# -----------------------------------------------------------------------------
# Main check function - call this from the hook
# Returns: "continue" with reason, or "stop" with reason
# -----------------------------------------------------------------------------
check_goal_completion() {
    # Check if goal exists
    if [ "$(has_active_goal)" = "false" ]; then
        echo "stop|No active goal"
        return
    fi

    # Check if goal is complete
    if is_goal_complete; then
        echo "stop|Goal complete - all criteria met"
        return
    fi

    # Goal exists and is not complete - should continue
    local message=$(generate_goal_continuation_message)
    echo "continue|$message"
}

# -----------------------------------------------------------------------------
# Export functions for use in main hook
# -----------------------------------------------------------------------------
export -f has_active_goal
export -f get_goal_completion_status
export -f is_goal_complete
export -f get_next_criterion
export -f get_current_phase
export -f needs_checkpoint
export -f update_goal_activity
export -f generate_goal_continuation_message
export -f check_goal_completion
