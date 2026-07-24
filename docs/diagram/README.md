# Diagram Index

This directory holds architecture and workflow diagrams for the service template. Each
diagram should explain one durable structure or flow that is easier to understand visually
than in prose.

## Purpose

Diagrams are the **visual companion** to the architecture docs — they explain *how* major
parts relate, move, or depend on each other. Keep them general, implementation-focused, and
free of product-specific naming unless the diagram is documenting a concrete example domain.

If you are a new engineer, scan this index after reading the top-level README and ADRs. The
diagrams should help you understand boundaries, request paths, dependency direction, and
runtime topology quickly.

## Index

| #  | Title | ADR |
|----|-------|-----|

## How to add a diagram

1. Create `NNNN-kebab-case-title.md`, where `NNNN` is the next free 4-digit number.
2. Use Mermaid for source-controlled diagrams whenever possible.
3. Give the diagram a short title, purpose, and legend when symbols are not obvious.
4. Link the ADR that explains the decision, or use `—` when no ADR exists yet.
5. Add a row to the index table above.
6. Update the diagram when the documented structure or flow changes.

## Conventions

- Filenames: `NNNN-kebab-case-title.md`, zero-padded to 4 digits.
- One primary idea per diagram. Split unrelated flows into separate diagrams.
- Prefer generic service-template terms over product or customer names.
- Keep diagrams small enough to review in a pull request.
- Cross-link related ADRs, README sections, and code entry points.
- Store generated images only when Mermaid or text source is not practical.
