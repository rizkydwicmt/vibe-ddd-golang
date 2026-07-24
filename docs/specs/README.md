# Feature Specifications

This directory holds implementation-ready feature specifications. A feature spec is one
reviewable package containing requirements, an optional transport contract, design, and
executable tasks:

```text
docs/specs/<jira-key>-<feature-slug>/
├── requirements.md
├── api-contract.md       # required when HTTP/gRPC/event behavior changes
├── design.md
└── tasks.md
```

Jira remains the delivery tracker for ownership, sprint, and status. The repository spec is
the versioned source of truth that agents and reviewers can read with the code.

## Quick Start

Humans and agents start from the index at the bottom of this file, then read only the artifact
needed for the current decision:

| Goal | Read |
|---|---|
| Understand scope, rules, or acceptance | `requirements.md` |
| Review an endpoint, payload, status, or error | `requirements.md` + relevant `api-contract.md` operation |
| Implement repository behavior | assigned requirements + contract operation + relevant `design.md` component |
| Execute or verify work | assigned entry in `tasks.md` |
| Understand a durable decision or flow | linked [ADR](../adr/README.md) or [diagram](../diagram/README.md) from `design.md` |

Create or update a package with [`/write-spec`](../../.claude/commands/write-spec.md), which follows
the authoritative [`writing-spec` skill](../../.claude/skills/writing-spec/SKILL.md):

```text
/write-spec PD-1234 add scorecard template management
```

1. Copy `0000-feature-name/` to `<jira-key>-<feature-slug>/` and update the index below.
2. Draft and approve `requirements.md`; unresolved behavior stays blocking.
3. Draft and approve `api-contract.md` when HTTP, gRPC, or event behavior changes. Otherwise mark
   it `Not required` and remove unused operation sections.
4. Draft and approve `design.md` against the approved requirements and contract.
5. Draft `tasks.md`, implement one assigned slice, and record actual verification evidence.
6. Generate Swagger/OpenAPI after implementation and verify it against the approved contract;
   do not use generated output as the default specification context.

Do not read every artifact by default. Follow the links and stable IDs for the current task:
`REQ-*` → `AC-*` → `API-*` when applicable → design component → `T-*` → evidence.

## Artifact responsibilities

| Artifact | Owns | Must not own |
|---|---|---|
| `requirements.md` | observable behavior, rules, scope, acceptance criteria, open questions | code layout, implementation steps |
| `api-contract.md` | exact HTTP/gRPC/event operations, payloads, statuses, stable codes, compatibility | business rationale, internal implementation mapping |
| `design.md` | repository-specific solution, internal mapping, data model, failure handling, verification strategy | product requirements, duplicated transport schemas, task progress |
| `tasks.md` | ordered implementation slices, Jira mapping, requirement coverage, completion evidence | new behavior or architecture decisions |
| [`docs/adr/`](../adr/README.md) | durable cross-feature decisions and tradeoffs | feature task lists |
| [`docs/diagram/`](../diagram/README.md) | durable visual views of flows, boundaries, topology, and data models | duplicated prose specifications |

## Lifecycle

Each file has its own approval gate:

```text
requirements: Draft → Approved
api-contract: Draft → Approved (when applicable)
design:       Draft → Approved
tasks:        Draft → Approved → In Progress → Completed
```

Work proceeds in order: approve requirements, then the API contract when applicable, then design,
then tasks. Implementation must not start while a blocking open question remains or while any
required predecessor is unapproved.
The agent must ask the user every blocking question and wait for the answer; listing an open
question does not permit design or implementation to continue. Questions should invite relevant
Jira/PRD links, screenshots, payload examples, existing implementations, or preferred direction.

Architecture artifacts are routed between requirements and design:

```text
Jira/PRD → requirements Approved → contract Approved → ADR/diagram scan → design Approved → tasks Approved → code
```

An existing Accepted ADR or current diagram is an input to requirements and design. After
requirements establish the behavioral boundary, every design completes a Visual Architecture
Check. Required ADRs must be Accepted before design approval. Promote a feature-local diagram to
`docs/diagram/` when the view becomes reusable; do not maintain two divergent sources.

## Traceability

- Requirement IDs use `REQ-<CAPABILITY>-NNN`.
- Acceptance criteria use `AC-<CAPABILITY>-NNN`.
- Spec tasks use `T-NNN` and record the real Jira task key when available.
- API operations use `API-<CAPABILITY>-NNN` and link to `REQ-*` and `AC-*`.
- Every acceptance criterion has exactly one evidence owner responsible for final verification.
- Every completed task records test or runtime evidence.
- Jira descriptions link the spec directory and list the assigned requirement/acceptance IDs.
- Requirements, contract when applicable, and design have explicit revisions before tasks are approved.

## Non-invention rule

Agents must not invent business rules, API fields, defaults, permissions, state transitions,
error/result codes, retry semantics, idempotency behavior, or compatibility guarantees. Missing
or contradictory behavior is added to `requirements.md` under Blocking Open Questions and stops
design or implementation.

The API contract may specify exact transport representation only after requirements define the
behavior. Generated Swagger/OpenAPI is derived verification; it is not a reason to copy a full
generated document into a feature package.

Agents may choose the simplest repository-conforming private implementation detail when it is
reversible and does not change observable behavior.

## Relationship to ADRs and diagrams

- Create an ADR only for a cross-cutting, contractual, or hard-to-reverse decision.
- Every design completes a Visual Architecture Check.
- Require a diagram for added or changed boundaries, actors/external systems, multi-component
  flows, lifecycle/state transitions, entity relationships, concurrency/retry/failure flows,
  dependency direction, or runtime topology.
- Do not require a diagram for fields, validations, filters, labels, local refactors, or
  single-step behavior already clear from acceptance criteria.
- Create or update a diagram under `docs/diagram/` when a durable flow, state model, boundary,
  topology, or data model is added or changed.
- `design.md` links owning ADRs and diagrams; it does not copy their rationale or full visual.
- A feature-local diagram stays in `design.md` when it only explains that feature.
- Reject decorative diagrams: a useful diagram has at least two meaningful nodes/states/entities
  and one relationship, transition, or message.

## Index

| Jira | Feature | Requirements | API Contract | Design | Tasks |
|---|---|---|---|---|---|
| — | Template package | [Template](0000-feature-name/requirements.md) | [Template](0000-feature-name/api-contract.md) | [Template](0000-feature-name/design.md) | [Template](0000-feature-name/tasks.md) |
