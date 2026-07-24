# Requirements: <Feature Name>

- **Jira feature**: `<PROJECT-0000>`
- **Status**: Draft
- **Revision**: <integer, starting at 1>
- **Owner**: <team or role>
- **PRD references**: <references or `—`>
- **API contract**: [api-contract.md](api-contract.md) — Required / Not required
- **Related ADRs**: [ADR index](../../adr/README.md) — <links or `—`>
- **Related diagrams**: [Diagram index](../../diagram/README.md) — <links or `—`>

## Source Material

| Source | Reference | Used for |
|---|---|---|
| Jira | `<PROJECT-0000>` | user story, scope, acceptance input |
| PRD | <section or `—`> | business rules and terminology |
| Existing contract | <path/link or `—`> | compatibility constraints |

Anything not supported by source material or an explicit approval belongs under Blocking Open
Questions. Do not present agent assumptions as requirements.

## Purpose

One paragraph describing the user outcome and business value. Do not describe the solution.

## Actors

- **<Actor>** — <goal and relevant authorization boundary>.

## Definitions

- **<Term>** — <one stable meaning used by this feature>.

## Scope

### Included

- <observable capability>

### Excluded

- <explicitly excluded behavior or follow-up feature>

## Assumptions

- <confirmed assumption> — Source: <Jira/PRD/approved decision/reference contract>.

Unconfirmed assumptions belong under Blocking Open Questions, not here.

## Business Rules

### REQ-<CAPABILITY>-001: <Rule title>

The system must <observable behavior>.

#### Acceptance criteria

- `AC-<CAPABILITY>-001`: Given <state>, when <action>, then <measurable result>.
- `AC-<CAPABILITY>-002`: Given <failure state>, when <action>, then <stable failure result>.

## Contract Commitments

Delete this section when the feature has no public or internal transport change. Record only
client-visible commitments here; exact fields and wire representation belong in `api-contract.md`.

| Requirement | Observable commitment | Contract operation |
|---|---|---|
| `REQ-...` | <operation behavior, stable status/code, or compatibility guarantee> | `API-...` |

## State and Operation Matrix

Delete this section when the feature has no meaningful state-dependent behavior.

| Current state | Operation | Result | Allowed |
|---|---|---|---:|
| <state> | <operation> | <new state or effect> | Yes/No |

## Failure Cases

| Condition | Expected behavior | Result code |
|---|---|---|
| <condition> | <safe observable result> | <stable code or blocking question> |

## Authorization

- <who may read or mutate what; state server-side enforcement>

## Data Integrity and Concurrency

- <invariant that must survive retries, races, or partial failure>

## Accessibility

- <keyboard, focus, labels, non-color communication, or `Not applicable`>

## Scenarios

### SC-001: <Scenario title>

**Given** <initial state>
**When** <action>
**Then** <observable result>

## Blocking Open Questions

| ID | Question | Controls | Requested reference/example | Owner | Status | Resolution |
|---|---|---|---|---|---|---|
| `Q-001` | <decision required> | <behavior/contract affected> | <PRD, screenshot, payload, example, direction> | <person/role> | Open | — |

Use `None` only when every behavior-affecting question has been resolved.
The agent must ask every Open question and wait for an answer before approval or design.

## Requirement Coverage

| Requirement | Acceptance criteria | Source |
|---|---|---|
| `REQ-<CAPABILITY>-001` | `AC-<CAPABILITY>-001`, `AC-<CAPABILITY>-002` | <Jira/PRD/approval> |

Every requirement and acceptance criterion must appear exactly once. This table proves source
coverage; design and task traceability are added in their respective files.

## Approval

- [ ] Scope and exclusions are explicit.
- [ ] Success and failure behavior are measurable.
- [ ] Authorization is explicit or confirmed not applicable.
- [ ] State transitions and destructive effects are explicit or not applicable.
- [ ] Defaults, ordering, pagination, retry, and idempotency are explicit where applicable.
- [ ] Data integrity and concurrency invariants are explicit where applicable.
- [ ] Every requirement and acceptance criterion has stable IDs and source coverage.
- [ ] API contract applicability is explicit; required contract operations are linked to requirements.
- [ ] Blocking Open Questions is `None`.
- **Approved by**: —
- **Approved at**: —
