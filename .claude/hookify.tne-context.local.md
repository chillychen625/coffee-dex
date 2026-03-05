---
name: tne-context-memory-reminder
enabled: true
event: all
action: info
conditions:
  - field: user_prompt
    operator: regex_match
    pattern: .*
---

**TNE-CONTEXT:** Update `progress.md` after tasks, `decisionLog.md` after decisions, `activeContext.md` on focus changes. Log sessions to `devPrompts.md`.
