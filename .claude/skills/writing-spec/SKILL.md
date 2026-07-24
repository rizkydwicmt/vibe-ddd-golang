---
name: writing-spec
description: Protocol for converting Jira or product requests into repository feature specs under docs/specs/. Use when drafting requirements or API contracts, designing a feature, decomposing Jira-linked tasks, reviewing spec completeness or contract traceability, or implementing from an approved spec.
---

# Feature Specification Protocol

## Purpose

A feature spec is a repository-native package with three core files and one conditional contract
file:

```text
docs/specs/<jira-key>-<feature-slug>/
├── requirements.md
├── api-contract.md       # required when HTTP/gRPC/event behavior changes
├── design.md
└── tasks.md
```

Jira tracks ownership and delivery. The package makes behavior, design, task boundaries, and
evidence versioned beside the code.

## Workflow

1. **Draft requirements** from Jira and linked PRD material. Preserve source references, split
   behavior into stable requirement and acceptance IDs, and expose contradictions as Blocking
   Open Questions.
2. **Clarify requirements** by asking the user every blocking question before approval. Explain
   which behavior each answer controls and invite PRD/Jira links, screenshots, payload examples,
   existing screens/services, edge cases, or preferred direction. Wait for answers; never resolve
   a question only from likely intent.
3. **Approve requirements** only when scope, state transitions, authorization, failures, data
   integrity, concurrency semantics, and externally observable defaults are explicit. Blocking
   Open Questions must be `None`.
4. **Draft the API contract when applicable** from approved requirements. Give every new or
   changed operation a stable `API-<CAPABILITY>-NNN` ID and specify its exact request, success,
   failure, authorization, idempotency, and compatibility behavior. Do not copy generated
   Swagger/OpenAPI into the package.
5. **Clarify the API contract** by asking before selecting any client-visible field, status, result
   code, header, retry rule, or compatibility behavior not governed by requirements, an Accepted
   ADR, or an existing contract.
6. **Approve the API contract** only when every operation maps to requirements and acceptance
   criteria, contract questions are resolved, and the artifact is revision-pinned. Mark it
   `Not required` when no transport contract changes.
7. **Draft design** against approved requirements, the approved contract when applicable, and
   repository conventions. Complete the Architecture Decision Scan and Visual Architecture Check
   first. Add the smallest useful
   feature-local Mermaid view when the feature has meaningful structure or flow; promote it to
   `docs/diagram/` when it becomes a reusable service-template view. Link durable ADRs and
   diagrams instead of duplicating them, then map every requirement to a design component and
   verification level.
8. **Clarify design** by asking before selecting any architecture, dependency, storage,
   migration, concurrency, or compatibility behavior not governed by requirements, an Accepted
   ADR, or repository convention. Invite technical references or an existing implementation.
9. **Approve design** only when the contract, transaction boundaries, failures, migration impact,
   security, and test strategy are implementable without inventing behavior. Required ADRs must
   be Accepted and required durable diagrams must be current.
10. **Draft tasks** from the approved design. Give each acceptance criterion one evidence owner,
   record Jira keys when available, make dependencies explicit, and require concrete evidence.
11. **Implement** one bounded task at a time. Read only its requirements, applicable contract
   operation, design components, linked ADRs/diagrams, and the smallest code reference needed.
12. **Verify both directions**: every acceptance criterion is implemented, every contract
   operation is tested or probed, and no observable behavior exists without a requirement basis.

## Jira Relationship

- One Jira feature normally owns one spec package.
- One spec package may map to multiple Jira implementation tasks.
- `T-NNN` is the stable repository task ID; Jira keys may change with planning.
- Jira task descriptions link the spec directory and list assigned requirement and acceptance
  IDs.
- Never invent Jira keys. Use `Unassigned` until the task exists.

## Artifact Boundaries

- `requirements.md`: what and why for the user; no implementation plan.
- `api-contract.md`: exact transport representation for new or changed HTTP/gRPC/event
  operations; no new business behavior or internal implementation plan.
- `design.md`: how this repository satisfies approved requirements; no new product behavior.
- `tasks.md`: execution order and evidence; no new requirements or architectural rationale.
- ADR: durable cross-feature decision and tradeoffs.
- Feature-local diagram: visual explanation needed only by this feature; keep it in `design.md`.
- Durable diagram: reusable visual architecture under `docs/diagram/`.

## Stop Conditions

Stop before design or implementation when any of these is true:

- a business rule, default, status transition, permission, error/result code, retry,
  idempotency, compatibility behavior, or destructive effect is unspecified;
- Jira, PRD, requirements, accepted ADR, diagram, or existing public contract contradicts;
- a blocking question remains;
- a predecessor artifact is not approved;
- an applicable API contract is missing, contradictory, or unapproved;
- a contract operation has no requirement or acceptance-criteria mapping;
- an acceptance criterion has no evidence owner or verification method.

When stopped, ask the user the blocking questions and wait. Include the affected requirement or
decision, why it matters, and the kind of reference/example that would resolve it. Reporting an
Open Question without asking it is insufficient.

## Hard Rules

- Directory name: `<jira-key-lowercase>-<feature-slug>`; use `untracked-<feature-slug>` only when
  no Jira feature exists.
- `requirements.md`, `design.md`, and `tasks.md` are required. `api-contract.md` is required when
  the feature adds or changes HTTP, gRPC, or event behavior; otherwise mark applicability
  `Not required` and do not invent an operation.
- Requirements and acceptance criteria have stable IDs and are never silently renumbered.
- API operations have stable `API-<CAPABILITY>-NNN` IDs and are never silently repurposed.
- Every requirement appears in the design traceability table.
- Every applicable API operation appears once in the contract index and maps to requirements,
  acceptance criteria, design, and verification.
- Every acceptance criterion appears once with an evidence owner in the task coverage table.
- Requirements, applicable contract, and design have an explicit revision before tasks are
  approved.
- No implementation starts from Draft requirements, an applicable Draft contract, or Draft design.
- Agents never invent observable behavior. Reversible private details follow the simplest
  established repository pattern.
- Material ambiguity requires a user answer or explicit authorization of a stated assumption;
  recording an unanswered question does not authorize continued execution.
- A design that introduces a durable decision must add or link an ADR.
- Every design must complete the Visual Architecture Check.
- A diagram is required when the feature adds or changes a domain/service boundary, actor or
  external-system interaction, multi-component request/event flow, lifecycle/state transition,
  entity relationship, concurrency/retry/failure flow, dependency direction, or runtime topology.
- A diagram is not required for a field, validation, filter, label, local refactor, or single-step
  behavior already clear from acceptance criteria.
- When triggered, a feature-local diagram must show at least two meaningful components, states,
  or entities plus a relationship, transition, or message. A decorative one-node diagram fails
  the check.
- A design that changes or adds a durable flow, state model, boundary, topology, or data model
  must add or update a diagram under `docs/diagram/`.
- Required verification cannot be marked optional when it proves an approved acceptance
  criterion.
- The contract artifact documents only new or changed operations and does not duplicate shared
  envelope, auth, versioning, or error conventions.
- Generated Swagger/OpenAPI is derived verification; it is not loaded as the default spec context.
- Update `docs/specs/README.md` in the same change. No orphan spec packages.

## Review Checklist

- Requirements are testable and describe success, failure, authorization, and concurrency where
  applicable.
- Applicable API operations have exact request, success, failure, authorization, and compatibility
  behavior with stable IDs.
- Design covers every requirement without changing its meaning.
- Design maps each contract operation without duplicating its wire schema.
- Tasks cover every acceptance criterion without overlap or unowned work.
- Jira links are real or explicitly `Unassigned`.
- ADR and diagram links exist and use the correct lifecycle.
- Completion claims cite actual tests or runtime probes.
