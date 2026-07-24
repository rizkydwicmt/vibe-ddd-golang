# API Contract: <Feature Name>

- **Jira feature**: `<PROJECT-0000>`
- **Requirements**: [requirements.md](requirements.md)
- **Requirements revision**: <integer, starting at 1>
- **Requirements status**: Draft
- **Contract revision**: <integer, starting at 1>
- **Applicability**: HTTP / gRPC / event / Not required
- **Status**: Draft
- **Generated contract**: <path or `Not generated until implementation`>

Do not define a new observable behavior here. Requirements own behavior; this file owns the
exact transport representation. Do not design against unapproved requirements.

If this feature adds or changes no public or internal transport contract, set **Applicability**
to `Not required`, record the reason below, and remove the operation sections.

## Contract Rules

- Document only new or changed operations; link shared conventions instead of copying them.
- Give every operation a stable `API-<CAPABILITY>-NNN` ID.
- Link every operation to at least one `REQ-*` and `AC-*` ID.
- Specify request fields, success data, failure codes, authorization, and compatibility behavior.
- Keep business rationale in `requirements.md` and internal mapping in `design.md`.
- Use examples only when the shape or edge case is not clear from the tables.
- Generated Swagger/OpenAPI is a derived verification artifact, not a second design source.

## Source Material

| Source | Reference | Used for |
|---|---|---|
| Requirements | [requirements.md](requirements.md) | observable behavior and acceptance criteria |
| Existing contract | <path/link or `—`> | compatibility constraints |
| ADR | <path/link or `—`> | shared wire or versioning decision |
| Generated contract | <path or `Not generated until implementation`> | implementation verification |

## Shared Conventions

| Concern | Decision or authority |
|---|---|
| Response envelope | <ADR/path/convention or `Not applicable`> |
| Authentication | <middleware/header/guard or `Not applicable`> |
| Tenant or resource scope | <scope rule or `Not applicable`> |
| API version | <header/path/registry or `Not applicable`> |
| Result/error codes | <enum/path or `Not applicable`> |
| Pagination | <cursor/limit/defaults or `Not applicable`> |
| Idempotency | <key/retry behavior or `Not applicable`> |
| Serialization | <case/null/time/identifier rules or `Repository convention`> |

## Operation Index

| Operation ID | Transport | Method/procedure | Path/topic | Requirements | Acceptance criteria |
|---|---|---|---|---|---|
| `API-<CAPABILITY>-001` | HTTP | `POST` | `/<path>` | `REQ-...` | `AC-...` |

## API-<CAPABILITY>-001: <Operation Title>

- **Purpose**: <one sentence describing the transport operation>
- **Requirements**: `REQ-...`
- **Acceptance criteria**: `AC-...`
- **Stability**: New / Existing / Compatibility-sensitive
- **Authorization**: <required permission, scope, and server-side enforcement>

### Request

| Location | Name | Type | Required | Validation and meaning |
|---|---|---|---:|---|
| Path/query/header/body | `<name>` | `<type>` | Yes/No | <constraint and meaning> |

```json
<minimal request example, or delete this block when tables are sufficient>
```

### Success

| Condition | HTTP/gRPC status | Result code | Data |
|---|---:|---|---|
| <successful condition> | `<2xx/code>` | `<stable code>` | `<schema or DTO>` |

Define new or changed response fields once here. Link an existing schema instead of copying it.

| Field | Type | Required | Meaning |
|---|---|---:|---|
| `<field or nested.path>` | `<type>` | Yes/No | <client-visible meaning> |

```json
<minimal success example, or delete this block when tables are sufficient>
```

### Failures

| Condition | HTTP/gRPC status | Result code | Retryable | Client-visible behavior |
|---|---:|---|---:|---|
| <failure condition> | `<status/code>` | `<stable code>` | Yes/No | <safe result> |

### Side Effects and Guarantees

- **State transition**: <state change or `No state change`>
- **Transaction boundary**: <atomic effects or `Not applicable`>
- **Idempotency**: <retry behavior or `Not applicable`>
- **Concurrency**: <locking/version check/last-write rule or `Not applicable`>
- **Data exposure**: <fields omitted, redacted, or resource scope>

### Compatibility

- **Breaking change**: Yes/No — <reason>
- **Backward compatibility**: <default, alias, version, or `Not applicable`>
- **Deprecation**: <replacement and timeline or `Not applicable`>

## Contract Traceability

| Operation | Requirements | Acceptance criteria | Design component | Verification |
|---|---|---|---|---|
| `API-...` | `REQ-...` | `AC-...` | <handler/service/component> | <test/probe> |

Every operation must appear exactly once. Every requirement and acceptance criterion must remain
covered by the parent artifacts.

## Blocking Contract Questions

| ID | Question | Controls | Requested reference/example | Owner | Status | Resolution |
|---|---|---|---|---|---|---|
| `CQ-001` | <transport decision required> | <field/status/auth/compatibility> | <existing contract/example/direction> | <person/role> | Open | — |

Use `None` only when the contract is complete or marked `Not required`. The agent must ask every
Open question and wait before contract approval or design drafting.

## Approval

- [ ] Requirements are Approved and revision-pinned above.
- [ ] Applicability is classified as transport change or `Not required`.
- [ ] Every new or changed operation has a stable ID.
- [ ] Every operation maps to requirements and acceptance criteria.
- [ ] Auth, scope, request, success, failures, retries, and compatibility are explicit.
- [ ] New or changed request and response fields have exact types and meanings.
- [ ] Stable result/error codes exist or are explicitly `Not applicable`.
- [ ] No business behavior exists only in this file.
- [ ] Blocking Contract Questions is `None`.
- **Approved by**: —
- **Approved at**: —
