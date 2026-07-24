# Write Docs

Create or update architecture documentation using an ADR, Mermaid diagram, or both.

## Topic

$ARGUMENTS

## Protocols

Read the relevant authoritative skill before writing:

- `.claude/skills/writing-adr/SKILL.md` for architectural decisions.
- `.claude/skills/writing-diagram/SKILL.md` for visual architecture documentation.

## Routing

- If the decision, scope, alternatives, owning system, or intended flow is unclear, ask the user
  concise questions and invite existing ADRs, diagrams, incidents, contracts, or examples before
  drafting. Wait for the answer instead of inventing context.
- Use `writing-adr` when the request is a cross-cutting, hard-to-reverse decision or contract.
- Use `writing-diagram` when the request is primarily a flow, boundary, topology, or data model.
- Use both when the ADR skill requires a visual companion.
- If the request warrants neither artifact, explain why and do not create placeholder docs.

Execute through `docs-writer` when delegation is useful. Pass the selected skill path and topic;
do not restate either skill's policy in the prompt. Report every created or modified path.
