---
name: agentic-delivery
description: Delivery protocol for autonomous repository changes. Use when implementing a complete change that must be planned, validated, reviewed, repaired, and reported without unsafe or open-ended agent loops.
---

# Agentic Delivery Protocol

## Purpose

Deliver a repository change end-to-end while keeping decisions reviewable. Domain-specific
skills define **what correct code is**; this skill defines **how work progresses**.

## Delivery loop

1. **Resolve scope** — restate the requested outcome, affected area, and acceptance checks.
   Read `CLAUDE.md`, then the relevant skill. Do not explore unrelated packages.
2. **Clarify** — compare the request with repository contracts and identify missing, ambiguous,
   or contradictory decisions. Ask the user concise blocking questions and explicitly invite
   Jira/PRD links, API examples, screenshots, existing implementations, or preferred direction.
   Wait for the answer before design or implementation. Do not convert silence into an assumption.
3. **Inspect** — read the smallest reference implementation and current registrations.
   Check `git status` before editing; preserve unrelated user changes.
4. **Implement** — make the smallest complete diff. Follow the domain skill's build order.
   Do not add dependencies, abstractions, configuration, transports, or compatibility layers
   unless the request requires them.
5. **Format and test narrowly** — format touched Go files, then run the closest package tests.
   Fix failures caused by the change before broad validation.
6. **Validate broadly** — run build, race tests, new-code lint, and generated-file drift checks
   that apply to the diff. Generated output must be reviewed before it is kept.
7. **Independent review** — delegate a read-only review to the narrowest matching reviewer.
   Pass file paths and the authoritative skill; include `writing-spec` when an approved spec or
   API contract governs the change. Never copy the rubric into the prompt.
8. **Repair loop** — fix actionable `critical` and `warn` findings, then repeat the affected
   checks. Maximum two repair rounds; after that, stop and report the unresolved blocker.
9. **Runtime proof** — run safe smoke probes when local dependencies already exist or can be
   started without touching shared state. Never invent destructive probes.
10. **Report** — summarize changed files, checks with actual results, skipped checks and reasons,
   remaining risks, and any user decision still required.

## Clarification gate

Clarification is mandatory before design or implementation when missing information could alter:

- observable behavior, scope, acceptance criteria, or user workflow;
- API fields, validation, defaults, result codes, compatibility, or transport behavior;
- state transitions, authorization, audit behavior, destructive effects, or data ownership;
- persistence, migration, concurrency, retry, ordering, or idempotency semantics;
- dependency choice, architecture boundary, or an Accepted ADR;
- which existing implementation, screenshot, contract, or product example should be followed.

Ask the smallest set of blocking questions. For each question, state what decision it controls
and request any useful reference or example. Do not edit code, generate migrations/contracts, or
mark a design/task Approved until the user answers or explicitly authorizes a stated assumption.

Questions are unnecessary only when the answer is explicit in an Approved spec, Accepted ADR,
current public contract, or authoritative repository convention. Cite that authority in the plan.

## Autonomy boundaries

Proceed without asking when the action is local, reversible, and implied by the request:

- edit source, tests, generated docs, and repository documentation;
- run formatters, targeted tests, build, lint, and read-only checks;
- start and stop processes or disposable containers created for this task;
- fix review findings that preserve the requested contract.

Stop and ask before:

- choosing unspecified API fields, business rules, status transitions, or authorization policy;
- proceeding with a material ambiguity instead of requesting references, examples, or direction;
- changing a public API contract, wire format, cross-domain dependency, or accepted ADR;
- adding a dependency or runtime service;
- executing destructive migrations or mutating shared/non-local data;
- using credentials, deploying, pushing, committing, or modifying remote resources;
- resolving unrelated pre-existing failures.

## Failure policy

- Distinguish **change failure** from **environment failure** and **pre-existing failure**.
- Never weaken tests, lint, validation, or error handling to make a check pass.
- Do not retry an unchanged failing command more than once.
- Capture the failing command and concise error. Continue with independent checks when safe.
- Clean up background processes created by the task even when validation fails.

## Acceptance record

Every completed delivery reports these fields:

```text
Scope: <implemented outcome>
Changed: <paths or concise groups>
Passed: <commands/checks>
Skipped: <check — reason>
Review: <no findings or resolved/unresolved findings>
Risk: <remaining risk or none>
```

## Hard rules

- One orchestrator owns the delivery. Subagents are bounded workers or read-only reviewers,
  never recursive orchestrators.
- Skills are policy; commands sequence skills; agents execute narrow roles; CI is final authority.
- Never duplicate a domain skill's rubric in a command or agent definition.
- Never claim a check passed unless its command completed successfully in this task.
