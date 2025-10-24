# Workflow Enforcement Fixes

## Issues Identified

During the implementation of `GET /api/profiles/:username`, two critical workflow violations occurred:

### Issue 1: Missing Draft PR Creation
**Problem**: Implementation proceeded directly on `main` branch without creating a draft PR first.

**Impact**:
- No PR history showing TDD phases (RED → GREEN)
- No CI status visibility for each phase
- No draft status to prevent premature review
- Missing complete implementation history in PR

**Expected Behavior**:
1. Create feature branch
2. Create and commit plan document (first commit)
3. Push branch
4. Create DRAFT PR (plan document is first commit in PR)
5. Update plan with PR link
6. Begin TDD implementation with each commit triggering CI in PR

### Issue 2: No Auto-Selection in /api Command
**Problem**: When `/api` was called without arguments, it asked user to specify which endpoint instead of autonomously selecting the next logical endpoint.

**Expected Behavior**:
- Analyze `docs/spec.md` for all endpoints
- Check `main.go` for implemented endpoints
- Determine next logical endpoint based on:
  - Dependencies satisfied
  - Complexity (prefer simpler)
  - Feature grouping
- Present selection with rationale
- Ask ONLY for confirmation (not which endpoint to implement)

## Fixes Applied

### Fix 1: Updated CLAUDE.md with Critical Workflow Requirements

Added new section "Implementation Plans & History" with clear requirements:

**Key Additions**:
1. Plan document MUST be created first (before any code)
2. Draft PR MUST be created after plan (strict order enforced)
3. Never implement without draft PR (stop and create if missing)
4. `/api` command behavior clarified for both modes

**Why this matters**:
- Plan document is natural first commit that creates PR
- Each TDD phase commit (RED/GREEN) triggers CI with visible status
- Draft PR prevents premature review
- Complete history visible from plan to completion
- CI failures caught immediately

### Fix 2: Enhanced /api Command with Explicit Instructions

Updated `.claude/commands/api.md` with:

**Critical Warning at Top**:
```markdown
⚠️ **CRITICAL**: This command MUST follow the plan → draft PR → implement workflow:
1. Create plan document first (never write code before plan exists)
2. Create draft PR immediately after plan (enables CI visibility)
3. Then begin TDD implementation (RED → GREEN visible in PR)
```

**Auto-Selection Mode Strengthened**:
- Explicit instruction: "YOU MUST autonomously find and select"
- Detailed steps for analysis and selection
- Clear output format for presenting choice
- Only ask for confirmation to proceed, not which endpoint

**Phase 0 Made Mandatory**:
- Renamed to "Phase 0: Create Plan and Draft PR (MANDATORY FIRST STEP)"
- Added warning: "DO NOT PROCEED TO PHASE 1 WITHOUT COMPLETING THIS PHASE"
- Detailed step-by-step commands with exact syntax
- Verification checklist before proceeding
- Clear explanation of why this phase is mandatory

### Fix 3: Enhanced plan.md Command Documentation

Already had good structure, reinforced the workflow benefits:
- Plan document as first commit
- Draft PR creation process
- PR and plan linking
- Reusable by other commands

## Testing the Fixes

To verify the fixes work correctly, try:

### Test 1: Auto-Selection Mode
```bash
/api
```

Expected:
1. Claude reads spec.md and main.go
2. Analyzes unimplemented endpoints
3. Presents clear selection: "Next endpoint: GET /api/tags"
4. Explains rationale
5. Asks: "Ready to implement GET /api/tags?"
6. On confirmation: creates branch → plan → PR → implements

### Test 2: Manual Selection Mode
```bash
/api GET /api/tags
```

Expected:
1. Creates branch: `feat/api-tags`
2. Creates plan: `docs/prompts/YYYY-MM-DD-get-api-tags.md`
3. Commits and pushes plan
4. Creates draft PR with plan as first commit
5. Updates plan with PR link
6. Begins TDD: RED commit → GREEN commit (each visible in PR with CI)

### Test 3: Stopping Mid-Implementation
If Claude starts writing code without a draft PR:

Expected:
1. Claude should catch this (CLAUDE.md says "STOP immediately")
2. Create plan document and draft PR first
3. Then resume implementation

## Benefits of Fixed Workflow

### For Development Process
- ✅ Clear implementation plan before coding starts
- ✅ TDD phases visible in PR commit history
- ✅ CI runs on each phase showing RED → GREEN progression
- ✅ Draft status prevents premature review
- ✅ Easy to see implementation progress

### For Code Review
- ✅ Reviewer sees complete history from plan to implementation
- ✅ Can review plan first, then implementation
- ✅ Each commit is atomic and testable
- ✅ CI status shows which commits pass/fail
- ✅ Clear separation of test commits vs implementation commits

### For Team Collaboration
- ✅ Consistent workflow across all features
- ✅ Plan documents serve as feature documentation
- ✅ Easy to pick up abandoned work (plan shows what's left)
- ✅ Historical record of implementation decisions

## Verification

Run these commands to verify the fixes are in place:

```bash
# Check CLAUDE.md has critical workflow section
grep -A 10 "CRITICAL WORKFLOW REQUIREMENTS" CLAUDE.md

# Check /api command has warning
grep -A 5 "CRITICAL" .claude/commands/api.md

# Check Phase 0 is mandatory
grep "MANDATORY" .claude/commands/api.md

# Check auto-selection instructions
grep "YOU MUST autonomously" .claude/commands/api.md
```

All should return results showing the enforced requirements.

## Next Steps

When implementing the next endpoint:
1. Use `/api` without arguments to test auto-selection
2. Verify Claude creates branch → plan → draft PR → then implements
3. Check PR shows plan as first commit
4. Verify CI runs on each TDD phase commit
5. Confirm draft PR status is maintained until completion

## Related Documents

- `CLAUDE.md` - Project-level workflow requirements
- `.claude/commands/api.md` - /api command implementation
- `.claude/commands/plan.md` - Plan creation workflow
- `docs/prompts/TDD.md` - TDD methodology reference
