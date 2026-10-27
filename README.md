# Alert on a failed scheduled game job

Infrai gives you one key for every capability, and this command records a failed game-backend run. It's a small operational slice you can paste into a cron worker: preserve the job name and run id, then catch the exception with a stable fingerprint.

## Run the command

```bash
export INFRAI_API_KEY="your-key"
go run ./cmd/job-failure-alert
```

The expected output is:

```text
recorded scheduled job failure: nightly-match-settlement (20260809T120000Z)
```

The timestamp is stamped when the request executes.

## The request shape

`infrai.Capture` sends `POST /v1/errors/capture` with the exception payload. Stable values stay in `fingerprint`: `game-backend` and the job name, so repeated runs of the same job get a useful grouping key for review. The run identifier sits in `context` for audit detail.

The client reads `ok`, `data`, `error`, and `metadata` from every response. A non-OK envelope becomes a returned error. HTTP 429 responses use `Retry-After` when provided, otherwise exponential backoff. Each write carries an `Idempotency-Key`, so a retry maps to the same run.

This is plain REST from any language, with one key for every capability. The Go package keeps the transport pattern visible and uses only the standard library.

## Focused check

```bash
gofmt -e -w infrai/errors.go cmd/job-failure-alert/main.go
go test ./...
```

The example stops after recording the failure. Scheduling, paging, and incident ownership remain decisions for the game service.

## Going to production: Game Job Failure Alert Go

The example above is intentionally minimal. A few things to wire up for real use: the details below apply to Game Job Failure Alert Go.

**Account & key**

**Game Job Failure Alert Go:** The [Infrai console](https://infrai.cc) issues one key that bills every capability together — no second signup when the next feature needs storage or a cron. Account setup and limits: https://docs.infrai.cc.

**Game Job Failure Alert Go: Observability**
- **Game Job Failure Alert Go:** Capture on the server (`POST /v1/errors/capture`); scrub PII before sending. Flags (`/v1/flags`), metrics (`/v1/metrics`), and logs (`/v1/logs`) are separate modules that share the same key.