---
description: Create implementation plan and draft PR for a feature
argument-hint: TITLE DESCRIPTION
---

# Plan Command

Creates a detailed implementation plan document and draft PR to track feature implementation progress.

## Parameters

- **$1**: Feature title (e.g., "GET /api/tags", "User authentication")
- **$2**: Brief description (e.g., "endpoint to fetch all tags", "JWT-based auth")

## Usage

### With Arguments
```
/plan "GET /api/tags" "endpoint to fetch all tags"
```

### Called from Other Commands
This command is designed to be reused by other commands like `/api`:
```
See @.claude/commands/plan.md for plan creation workflow
```

## Workflow

### Step 1: Create Branch and Plan Document

1. **Generate plan file name**: `docs/prompts/YYYY-MM-DD-{feature-slug}.md`
2. **Create feature branch**: `git checkout -b {branch-name}`
3. **Create plan document** with the following structure:

```markdown
# {Feature Title}

## Status & Links

**Status**: [ ] Not Started
**PR**: (to be added after PR creation)

## Context

{Description of what this feature does and why}

### Original Request
{User's original request or command}

### Key Requirements
- Requirement 1
- Requirement 2
- Requirement 3

## Methodology

Following TDD principles as defined in `@docs/prompts/TDD.md`:
- Red → Green → Refactor cycle
- Write failing test first
- Implement minimum code to pass
- Refactor only when tests pass

## Feature-Specific Requirements

### Request/Response Format
{If applicable - API endpoint details}

### Authentication
{If applicable - auth requirements}

### Validation Rules
{If applicable - validation logic}

### Database Operations
{If applicable - queries/migrations needed}

## Implementation Steps

{Detailed checklist of steps to complete the feature}

- Step 1
- Step 2
- Step 3
...

## Verification Commands

```bash
# Commands to test and verify the implementation
make test
make lint
```

## Notes

- Additional context
- Gotchas or special considerations
- Links to relevant documentation
```

### Step 2: Commit and Push Plan

```bash
git add docs/prompts/YYYY-MM-DD-{feature}.md
git commit -m "docs: add implementation plan for {feature}"
git push -u origin {branch-name}
```

### Step 3: Create Draft PR

```bash
gh pr create --draft \
  --title "feat: {feature title}" \
  --body "Implementation plan: docs/prompts/YYYY-MM-DD-{feature}.md

## Summary
{Brief description}

## Plan
See the implementation plan document for detailed steps and requirements.

## Status
🚧 Draft - Work in progress"
```

### Step 4: Update Plan with PR Link

1. Get PR number from the created PR
2. Update plan document:
   ```markdown
   **Status**: [ ] In Progress
   **PR**: https://github.com/{owner}/{repo}/pull/{number}
   ```
3. Commit and push:
   ```bash
   git add docs/prompts/YYYY-MM-DD-{feature}.md
   git commit -m "docs: add PR link to plan"
   git push
   ```

## Output

The command should output:
1. ✅ Plan document created at `docs/prompts/YYYY-MM-DD-{feature}.md`
2. ✅ Branch created and pushed: `{branch-name}`
3. ✅ Draft PR created: `#{pr-number}`
4. ✅ Plan updated with PR link

## Benefits

- **Reusable**: Can be called by any command that needs planning
- **Consistent**: All plans follow the same structure
- **Tracked**: Plan document is first commit in PR
- **Discoverable**: Plans are organized in `docs/prompts/`
- **Linked**: Plan references PR, PR references plan

## Examples

### Example 1: API Endpoint
```
/plan "GET /api/tags" "endpoint to fetch all tags without authentication"
```

### Example 2: Feature
```
/plan "User authentication" "JWT-based authentication with refresh tokens"
```

### Example 3: Refactoring
```
/plan "Extract user service" "refactor user-related code into dedicated service"
```

## Integration with Other Commands

The `/api` command uses this workflow:

```markdown
### Phase 0: Create Plan and Draft PR

Follow the plan creation workflow from `@.claude/commands/plan.md`:
1. Create branch and plan document
2. Commit and push plan
3. Create draft PR
4. Update plan with PR link

This makes the plan document the natural first commit that creates the PR.
```

## Notes

- Plan file naming: `YYYY-MM-DD-{feature-slug}.md` (e.g., `2025-10-23-get-api-tags.md`)
- Branch naming: Match your project conventions (e.g., `feat/api-tags`, `feature/get-tags`)
- Status values: "Not Started", "In Progress", "Completed"
- Always mark PR as draft initially
- Update PR description as implementation progresses
