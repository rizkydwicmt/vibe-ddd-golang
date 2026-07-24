# NNNN. <Diagram Title In Title Case>

- **ADR reference**: [NNNN. <ADR title>](../adr/NNNN-<slug>.md)
- **Related diagram**: [NNNN. <related diagram title>](NNNN-<slug>.md)
- **Purpose**: one line listing the views this document provides (system context, sequences, flowcharts, data model, failure handling, …).

One or two sentences of scope. State who/what this diagram is for and, if relevant, what it is explicitly *not* for (point readers to the correct diagram instead). Keep it factual.

> Keep diagrams in sync with the owning ADR. If the ADR and this diagram disagree, update both in the same change (see [0000. ADR Authoring Standard](../adr/0000-adr-authoring-standard-kiss.md)).

## System Context

High-level boxes and the edges between them. Show ownership boundaries (services, databases, external providers), not internal logic.

```mermaid
flowchart LR
  subgraph CallerBoundary["Caller Boundary"]
    Caller["Caller"]
  end

  subgraph ServiceBoundary["Service Boundary"]
    Service["API Service"]
    Database[("Primary Database")]
  end

  Caller --> Service
  Service <--> Database
```

## <Primary Flow / Provisioning / Setup>

The main happy-path sequence. Use `autonumber`. Keep participant names short and consistent across every sequence in the file.

```mermaid
sequenceDiagram
  autonumber
  participant Caller
  participant Service
  participant Store as Data Store

  Caller->>Service: Submit request
  Service->>Store: Persist state
  Service-->>Caller: Return result
```

Rules:

- List the invariants this flow must preserve as short bullets.
- State anything that happens exactly once, or only under a grace window.

## KISS Default Path

The simplest path a normal caller follows. Mark exceptions as explicit escape hatches.

```mermaid
flowchart TD
  A["Receive request"] --> B["Validate input"]
  B --> C["Execute operation"]
  C --> D["Return result"]

  X["Escape hatch: approved exception"] -.-> B
```

KISS defaults:

- State the boring default (one version, one middleware, one SDK surface, etc.).
- List what is an exception rather than the normal path.

## <Request / Response / Detailed> Sequence

```mermaid
sequenceDiagram
  autonumber
  participant Client
  participant Service
  participant Database
  participant Provider as External Provider

  Client->>Service: Submit request
  Service->>Database: Load and validate state
  Service->>Provider: Call when required
  Provider-->>Service: Return provider result
  Service->>Database: Persist final state
  Service-->>Client: Return response
```

## Verification / Decision Flowchart

Use a top-down flowchart for branching validation logic. Funnel every rejection branch to a single, clearly named failure node.

```mermaid
flowchart TD
  A["Receive request"] --> B{"Input valid?"}
  B -- "No" --> X["Return failure response"]
  B -- "Yes" --> C{"Operation allowed?"}
  C -- "No" --> X
  C -- "Yes" --> OK["Execute success path"]
```

## Data Model Extension

Only the entities this decision adds or changes. No cross-database foreign keys if services own separate databases — say so explicitly.

```mermaid
erDiagram
  PARENT ||--o{ CHILD : owns

  PARENT {
    TEXT id PK
    TEXT status
    TIMESTAMPTZ created_at
  }

  CHILD {
    TEXT id PK
    TEXT parent_id
    TIMESTAMPTZ created_at
  }
```

Notes:

- Call out which fields live in Redis vs Postgres vs logs.
- Name the source of truth and any convenience projections.

## Failure Handling

```mermaid
flowchart TD
  A["Failure occurs"] --> B{"Before durable write?"}
  B -- "Yes" --> C["Return minimal safe response"]
  B -- "No" --> D{"Failure class"}
  D -- "Validation" --> E["Return 4xx response envelope"]
  D -- "Provider rejected" --> F["Return provider result envelope"]
  D -- "Transient" --> G["Return retryable response envelope"]
  D -- "Internal" --> H["Return safe 5xx response envelope"]
  C --> K["Log redacted reason only"]
  E --> K
  F --> K
  G --> K
  H --> K
```

## Implementation Checklist

1. <First safe, incremental step.>
2. <Schema / migration change, if any.>
3. <Service wiring (middleware, handlers, events).>
4. <SDK / client surface changes.>
5. <Tests: unit, integration, contract, migration.>
6. <Observability: logs, metrics, redaction rules.>
7. <Operations: rotation, rollback, replay, backfill.>
