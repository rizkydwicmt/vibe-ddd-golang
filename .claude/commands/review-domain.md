# Review Domain

Review a domain (or the current diff) for boilerplate and approved-contract compliance using the
`ddd-reviewer` agent (`.claude/agents/ddd-reviewer.md`).

## Scope

$ARGUMENTS

Interpret the argument as a domain name (`internal/application/<domain>/`), a file path,
or empty — empty means review the current working diff (`git diff` + staged).

## Protocol

Use `.claude/skills/golang-ddd-domain/SKILL.md` for domain compliance. Also use
`.claude/skills/writing-spec/SKILL.md` when the supplied scope includes `docs/specs/` files or an
approved feature spec governs the implementation. Do not reinterpret product intent; review only
traceability and implementation alignment.

## Sequence

1. Resolve the scope: domain name → all files under `internal/application/<domain>/` plus
   its registrations (`internal/server/api/providers.go`, `internal/server/api/module.go`,
   `internal/server/migration/migration.server.go`); empty → changed files from
   `git diff --name-only HEAD`.
2. Include changed `docs/specs/` files in diff-based review. Spawn the `ddd-reviewer` agent with
   the resolved file list and applicable skill paths. Do not restate or reinterpret the rubrics.
3. Relay the agent's findings verbatim (`path:line: severity: problem. fix.`),
   most severe first. If the agent returns `No findings.`, say so and stop.
4. Only apply fixes when the user asks.

## Verify

- [ ] Every finding has file:line, severity, and a concrete fix
- [ ] Registrations were checked, not just domain files
- [ ] Approved contract operations were checked when a spec governs the change
