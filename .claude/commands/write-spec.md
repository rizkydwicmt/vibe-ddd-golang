# Write Spec

Create or update a repository feature specification from a Jira task, PRD excerpt, or requested
behavior.

## Request

$ARGUMENTS

## Protocol

Read `.claude/skills/writing-spec/SKILL.md` and follow its artifact order and approval gates.
Use `.claude/skills/writing-adr/SKILL.md` for required durable decisions. Use
`.claude/skills/writing-diagram/SKILL.md` only when the Visual Architecture Check identifies a
reusable or durable view for `docs/diagram/`; feature-local views stay in `design.md`.

## Sequence

1. Resolve the Jira feature key and feature slug; use `untracked-<slug>` when no Jira key exists.
2. Create the package from `docs/specs/0000-feature-name/` or update the existing package.
3. Draft only the earliest unapproved artifact unless the user explicitly supplies approvals;
   for transport changes, draft `api-contract.md` after approved requirements and before design.
4. Add contradictions and missing behavior as Blocking Open Questions; never decide them silently.
5. Ask the user those blocking questions immediately. State what each answer controls and invite
   Jira/PRD links, screenshots, payloads, examples, existing implementations, or preferred
   direction. Wait before drafting the next artifact or approving the current one.
6. Update the `docs/specs/README.md` index in the same change.
7. Report created paths, current gate, answers still required, and the next action after approval.
