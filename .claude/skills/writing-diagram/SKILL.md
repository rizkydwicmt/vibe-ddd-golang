---
name: writing-diagram
description: Protocol for writing Mermaid architecture diagrams in docs/diagram/. Use when creating or updating a diagram, sequence diagram, flowchart, ERD, or system-context view, or when an ADR needs a visual companion.
---

# Diagram Writing Protocol

## Purpose

Diagrams in `docs/diagram/` are the visual companion to the ADRs: boundaries, request
paths, dependency direction, runtime topology. This skill codifies
`docs/diagram/README.md` and the template `docs/diagram/0000-diagram-template.md`.

This skill governs durable diagrams under `docs/diagram/`. Feature specs use their Visual
Architecture Check to decide whether a feature-local Mermaid view belongs in `design.md`.
Promote that view here only when it is reusable or durable. ADR and diagram routing are
independent: a diagram may exist without an ADR.

## Flow

1. Read the index table in `docs/diagram/README.md`; next free 4-digit `NNNN` is your number.
2. Copy `docs/diagram/0000-diagram-template.md` to `docs/diagram/NNNN-kebab-case-title.md`.
3. Fill the header: ADR reference (or `—` when none exists yet), related diagram, one-line
   purpose, one-two sentences of scope (including what the diagram is *not* for). Replace the
   template's `NNNN` heading with the exact 4-digit filename prefix.
4. **The template's sections are a menu, not a mandate.** Keep only the views that serve
   the one idea this diagram documents; delete the rest:
   - System Context — `flowchart LR` + subgraph boundaries
   - Primary Flow — `sequenceDiagram` + `autonumber`
   - KISS Default Path — `flowchart TD` with escape hatches dashed
   - Detailed Sequence — `sequenceDiagram` + `autonumber`
   - Verification/Decision — `flowchart TD`, all rejections funnel to one failure node
   - Data Model — `erDiagram`, only entities this decision adds/changes
   - Failure Handling — `flowchart TD`
   - Implementation Checklist — numbered steps
5. Add a row to the index table in `docs/diagram/README.md` **in the same change**.
6. When the documented structure or flow changes later, update the diagram (and its ADR)
   in the same change.

## Picking the Mermaid type

| Topic | Type |
|---|---|
| Boundaries, ownership, topology | `flowchart LR` with `subgraph` |
| Request/response, lifecycle, protocol | `sequenceDiagram` with `autonumber` |
| Branching validation / decision logic | `flowchart TD`, single failure node |
| Schema / entities | `erDiagram` |

## Mermaid gotchas

- Quote any label containing `(){}[]/:,` or other special chars: `A["parse (JSON)"]` —
  unquoted labels are the most common render failure.
- Keep template examples renderable. Never put `<placeholder>` tokens inside a Mermaid fence;
  use plain sample labels and replace them with concrete names.
- Keep node labels short; put detail in prose under the diagram.
- Keep participant names short and consistent across every sequence in one file.
- Databases render as `B[("db")]`, escape hatches as dashed edges `-.->`.

## Hard Rules

- Mermaid source only — no binary images unless text source is genuinely impractical.
- Filename `NNNN-kebab-case-title.md`, zero-padded 4 digits. Never renumber.
- The first heading number exactly matches the filename prefix.
- One primary idea per diagram. Unrelated flows split into separate files.
- Generic service-template terms — no product or customer names (documenting an example
  domain like `user`/`payment` is the only exception).
- Link the owning ADR when one exists; `—` otherwise. Every added or changed durable wire format,
  flow, state model, boundary, topology, or data model gets a diagram under `docs/diagram/`.
- Reject decorative diagrams. A diagram must contain at least two meaningful nodes, states, or
  entities and one relationship, transition, or message.
- Update the `docs/diagram/README.md` index table in the same change. No orphan diagrams.
- Small enough to review in a pull request.

## Common mistakes

| ❌ Wrong | ✅ Right |
|---|---|
| One file with auth flow + schema + deploy topology | three diagrams, cross-linked |
| `A[handle request (v2)]` | `A["handle request (v2)"]` |
| `PARENT ||--o{ CHILD : <relationship>` | `PARENT ||--o{ CHILD : owns` |
| `0001-request-flow.md` with `# NN.` | `0001-request-flow.md` with `# 0001.` |
| Diagram added, index untouched | index row added in the same commit |
| Keeping all 8 template sections half-filled | only the sections that serve the idea |
| Each rejection branch its own dead-end node | all rejections funnel to one failure node |
| PNG exported from a whiteboard tool | Mermaid source in the .md |
