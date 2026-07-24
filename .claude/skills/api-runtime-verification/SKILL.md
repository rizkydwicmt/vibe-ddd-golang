---
name: api-runtime-verification
description: Protocol for safely booting and probing the local API end-to-end. Use for smoke tests, response-envelope verification, route probes, and runtime validation after API changes.
---

# API Runtime Verification Protocol

## Purpose

Prove the built service boots and serves the expected response contract using local disposable
infrastructure, while leaving pre-existing processes and data untouched.

## Preconditions

- The configured database target is confirmed local/disposable.
- The API port and database port are not owned by unrelated processes.
- Build and targeted tests have already passed, unless runtime reproduction is the task.
- Mutating probes have isolated disposable data and are required by the requested behavior.

If any target may be shared, stop and ask.

## Flow

1. Record pre-existing API/database processes and containers. Reuse safe local infrastructure;
   create a disposable Postgres container only when permitted by the parent task.
2. Build and start the API while capturing its exact PID and logs.
3. Poll `/healthz` for at most 30 seconds. On timeout, preserve the relevant log excerpt and stop
   the process created by the task.
4. Require `/readyz` to return HTTP 200.
5. Probe `/api/v1/users` and require the standard envelope fields `requestId`, `code`, and
   non-null `debug`.
6. For an optional focus, prefer read-only routes. A mutating probe must use disposable data,
   avoid deletes of pre-existing records, and assert both HTTP status and stable result code.
7. Check logs for panic or stack traces.
8. Stop the exact API PID created by the task. Remove only containers created by the task.
9. Report each probe with route, HTTP status, application code, and pass/fail result.

## Baseline commands

```bash
make build
./bin/vibe-ddd-golang &
curl -sf http://localhost:8080/healthz
curl -sf http://localhost:8080/readyz
curl -sf http://localhost:8080/api/v1/users \
  | jq -e '.requestId and .code and (.debug != null)'
```

Adapt configuration through the existing Viper configuration mechanism; do not introduce
ad-hoc environment reads in application code.

## Acceptance

- Health and readiness return HTTP 200.
- Every probed API response uses the standard envelope.
- No unexpected 5xx, panic, or stack trace occurs.
- No shared/pre-existing data or process is mutated.
- Every process/container created by the task is cleaned up.

## Hard rules

- Never infer that a database is disposable from its hostname alone; confirm configuration.
- Never use `pkill`, broad process matching, or remove pre-existing containers.
- Runtime success does not replace tests; it complements them.
- Report actual observed statuses and codes, never expected values as results.
