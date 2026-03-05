# Sprint Dispatch: {Sprint Name}

## Mission

{One sentence describing what this sprint delivers}
**Exit Criteria:** {Specific measurable outcome}

**Project:** `/Users/jima/oztam/workspace/oztam-data-analyst`
**PRD:** `{path to PRD}`
**Duration:** ~{X} hours

---

## Handoffs ({N} total)

Execute in dependency order. P0 handoffs are blocking.

---

### Handoff 1: `{handoff-id}` (P0, {X}h)

**{Handoff Title}**

{Description of what to do}

**Tasks:**
1. Task one
2. Task two
3. Task three

**Acceptance:**
- [ ] Criterion one
- [ ] Criterion two

**Signals:**
- When starting: `<comware:handoff_start id="{handoff-id}">`
- When done: `<comware:handoff_complete id="{handoff-id}">`
- If blocked/failed: `<comware:handoff_failed id="{handoff-id}">`

---

### Handoff 2: `{handoff-id}` (P0, {X}h)

**{Handoff Title}**

{Description}

**Tasks:**
1. Task one

**Acceptance:**
- [ ] Criterion one

Blocked by: {dependency-handoff-id}

**Signals:**
- When starting: `<comware:handoff_start id="{handoff-id}">`
- When done: `<comware:handoff_complete id="{handoff-id}">`
- If blocked/failed: `<comware:handoff_failed id="{handoff-id}">`

---

### FINAL Handoff: `{sprint-prefix}-999-commit-verify` (P0, 0.5h)

**Commit, Verify CI, Update Goal Progress**

This handoff is MANDATORY. Do NOT skip it. Do NOT signal `<comware:done>` until CI is green and goal progress is updated.

**Tasks:**
1. Review all changes made during the sprint: `git status`
2. Stage all relevant files (avoid committing secrets or generated files)
3. Create commit with descriptive message summarizing sprint work
4. Push to remote branch
5. Monitor CI pipeline until it completes
6. If CI fails: fix issues, commit fix, push again, repeat until green
7. **Update goal progress in `.claude/project.yaml`:**
   - Find the goal this sprint advances
   - Mark any newly satisfied `success_criteria` as `completed: true`
   - Recalculate `progress`: (completed / total) × 100
   - If all criteria complete, set goal `status: completed` and add `completed_date`

**Acceptance:**
- [ ] All sprint changes committed
- [ ] Commit message describes what was accomplished
- [ ] Changes pushed to remote
- [ ] CI pipeline passed (all checks green)
- [ ] No uncommitted work remains
- [ ] Goal success_criteria updated for any newly completed items
- [ ] Goal progress percentage recalculated

Blocked by: ALL preceding handoffs

**Signals:**
- When starting: `<comware:handoff_start id="{sprint-prefix}-999-commit-verify">`
- When done: `<comware:handoff_complete id="{sprint-prefix}-999-commit-verify">`
- If blocked: `<comware:handoff_failed id="{sprint-prefix}-999-commit-verify">`

---

## Completion

After ALL handoffs complete (including commit-verify):

1. Verify all acceptance criteria met
2. Verify CI is green
3. Signal completion:

```
<comware:done>
```

---

## Notes

- {Any additional context}
- {Constraints or considerations}
