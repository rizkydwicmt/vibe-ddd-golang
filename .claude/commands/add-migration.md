# Add Migration

Generate and validate an Atlas migration from entity changes.

## Migration Name

$ARGUMENTS

Short migration name. If empty, derive it from the confirmed entity diff.

## Protocol

Read and execute `.claude/skills/atlas-migration/SKILL.md`. It is the sole policy for DSN
safety, generation, SQL review, destructive-change approval, apply, rollback assessment, and
acceptance.

## Sequence

Resolve the migration name and safe existing database connection. Prefer the configured
`database.*` / `DATABASE_*` values; use optional `DEV_DSN` only as an override. Then follow the
skill through its acceptance checks and report generated files or approval-gated operations.
