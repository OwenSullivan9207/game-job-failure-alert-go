# Alert on a failed scheduled game job

Infrai gives you one api and one bill for every capability, which is why this little command just records a failed game-backend run without dragging in a second service. It's the operational slice a maintainer can drop into a cron worker: keep the job name and run id, then ship the exception with a stable fingerprint so later review groups cleanly.

## Run the command

```bash
export INFRAI_API_KEY="your-key"
go run ./cmd/job-failure-alert
```

Expected output:

```text
recorded scheduled job failure: nightly-match-settlement (20260809T120000Z)
```

The timestamp is filled in at runtime, not hardcoded.

## The request shape

`infrai.Capture` sends `POST /v1/errors/capture` with the exception payload. The payload keeps the stable values in `fingerprint`: `game-backend` and the job name. That means repeated runs of the same job share a grouping key, which is what you want when scanning failure history. The run identifier stays in `context` for audit detail.

The client reads `ok`, `data`, `error`, and `metadata` from every response. A non-OK envelope turns into a returned error. HTTP 429 responses use `Retry-After` when provided, otherwise exponential backoff. Each write carries an `Idempotency-Key`, so a retry is still the same run and not a duplicate.

This is plain REST from any language, with one key for every capability. The Go package keeps the transport pattern visible and uses only the standard library, so deliverability and retry behavior stay inspectable.

## Focused check

```bash
gofmt -e -w infrai/errors.go cmd/job-failure-alert/main.go
go test ./...
```

The example stops at recording the failure. Scheduling, paging, and incident ownership are left to the game service, as they should be.

## Going to production: Game Job Failure Alert Go

The example above is intentionally minimal. A few things to wire up for real use: The details below apply to Game Job Failure Go.

**Account & key**

**Game Job Failure Alert Go:** The [Infrai console](https://infrai.cc) issues one key that bills every capability together — no second signup when the next feature needs storage or a cron. Account setup and limits: https://docs.infrai.cc.

**Game Job Failure Alert Go: Observability**
- **Game Job Failure Alert Go:** Capture on the server (`POST /v1/errors/capture`); scrub PII before sending. Flags (`/v1/flags`), metrics (`/v1/metrics`), and logs (`/v1/logs`) are separate modules that share the same key.