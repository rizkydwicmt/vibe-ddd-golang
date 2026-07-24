---
name: writing-adr
description: Protocol for writing Architecture Decision Records in docs/adr/. Use when documenting an architectural decision, writing or updating an ADR, superseding a decision, or when asked "should this be an ADR".
---

# ADR Writing Protocol

## Purpose

ADRs in `docs/adr/` are the immutable historical record of *why* the service is shaped the
way it is. This skill codifies `docs/adr/README.md` and the template
`docs/adr/0000-adr-template.md` into an executable protocol.

## When an ADR is warranted

Write one when the decision is **cross-cutting, hard to reverse, or a contract**: wire
format, response envelope, auth model, persistence strategy, migration tooling, event
transport, dependency direction, versioning policy.

Do NOT write one for implementation detail: a function rename, a library bump with no
contract change, a bug fix, code layout inside one domain.

## Flow

1. Read the index table in `docs/adr/README.md`; next free 4-digit `NNNN` is your number.
2. Copy `docs/adr/0000-adr-template.md` to `docs/adr/NNNN-kebab-case-title.md`.
   Title in imperative present tense ("Use Atlas for schema migrations", not "Atlas migrations").
3. Fill header: `Status: Proposed`, today's date, deciders, related ADRs (or remove the line).
4. Fill the four sections (guidance below). Aim for 50–150 lines total.
5. Add a row to the index table in `docs/adr/README.md` **in the same change**.
6. If this supersedes an older ADR: flip the old one's status to `Superseded by NNNN` and
   cross-link both ways (metadata edits are the only allowed edits to Accepted ADRs).
7. If the decision changes wire format, request flow, or data model: pair it with a diagram
   (see the `writing-diagram` skill) and link it.

## Section guidance

- **Context** — the problem and its forces (technical, business, operational) and the
  constraints bounding the solution space. Factual. State the problem, **not the answer** —
  if the chosen solution appears here, rewrite.
- **Decision** — imperative present-tense fact, one paragraph if possible:
  *"We use X for Y. Z happens directly from W."*
- **Consequences** — three lists: Positive, Negative/tradeoffs, Neutral.
  **Negative must not be empty.** Every real decision costs something; name it.
- **Alternatives considered** — each option gets a concrete rejection reason
  ("adds a second broker to operate", not "not a good fit").

## Status lifecycle

```
Proposed ──► Accepted ──► Deprecated
              └─────────► Superseded by NNNN
```

`Proposed` while under review → `Accepted` when merged → immutable except metadata flips.

## Hard Rules

- Filename `NNNN-kebab-case-title.md`, zero-padded 4 digits. Never renumber, never reuse a number.
- One decision per ADR. Splits naturally in two → write two ADRs.
- No edits to `Accepted` ADRs except status/link metadata. Substantive change = new ADR.
- 50–150 lines.
- Keep generic to the service template — no product or customer names.
- Update the `docs/adr/README.md` index table in the same change. No orphan ADRs.
- Cross-link related ADRs in the header.

## Common mistakes

| ❌ Wrong | ✅ Right |
|---|---|
| Context says "we need Atlas because..." | Context states the migration problem; Atlas appears only in Decision |
| Negative consequences empty or "none" | at least one real cost named |
| Alternative "rejected: not suitable" | "rejected: requires a second broker to operate and monitor" |
| Editing an Accepted ADR's Decision | new ADR + `Superseded by NNNN` on the old one |
| ADR file added, index untouched | index row added in the same commit |
| Title "Atlas migrations" | "Use Atlas for schema migrations" (imperative present) |
