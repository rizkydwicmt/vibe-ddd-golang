# Deliver Domain

Create or extend a domain and run every applicable delivery check.

## Request

$ARGUMENTS

The argument must identify the domain and requested behavior. A missing domain means create it;
an existing domain means extend it. Before editing, run the clarification gate from
`.claude/skills/agentic-delivery/SKILL.md`. If any material behavior, contract, or direction is
missing, ask the user and explicitly invite references or examples; wait for the answer.

## Protocols

Read these authoritative skills before editing:

- `.claude/skills/agentic-delivery/SKILL.md` for orchestration and safety boundaries.
- `.claude/skills/golang-ddd-domain/SKILL.md` for domain structure and correctness.
- `.claude/skills/writing-spec/SKILL.md` when the request references a spec package, Jira task,
  requirement ID, acceptance ID, API operation ID, contract file, or spec task ID.

## Workflow

Execute the complete loop from `.claude/skills/agentic-delivery/SKILL.md`, applying
`.claude/skills/golang-ddd-domain/SKILL.md` as the domain policy. Use `ddd-reviewer` for the
independent review. When schema or runtime proof applies, also load
`.claude/skills/atlas-migration/SKILL.md` or `.claude/skills/api-runtime-verification/SKILL.md`;
do not copy their rules into prompts.

When a spec package is supplied, require approved `requirements.md`, an approved applicable
`api-contract.md`, and `design.md`, resolve the requested slice through `tasks.md`, implement
only its assigned IDs, and record real completion evidence. Stop on missing mappings, blocking
questions, or artifact conflicts.
Ask the user to resolve them and wait; do not continue from a guessed interpretation.
