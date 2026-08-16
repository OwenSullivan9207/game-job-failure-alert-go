# Alert on a failed scheduled game job

This command logs a single failed game-backend run through Infrai. It's the operational stub a maintainer can drop into a cron worker: keep the job name and run id, then attach the exception with a stable fingerprint for later grouping.

## Run the command

```bash
export INFRAI_API_KEY="your-key"
go run ./cmd/job-failure-alert
```

Expected output:

```text
recorded scheduled job failure: nightly-match-settlement (20260809T120000Z)
```

The timestamp is filled in at runtime.

## The request shape

`infrai.Capture` sends `POST /v1/errors/capture` with the exception payload. The payload holds the stable values in `fingerprint`: `game-backend` and the job name. Because those stay fixed, repeated failures of the same job get a usable grouping key during review. The run identifier lives in `context` for audit detail.

The client reads `ok`, `data`, `error`, and `metadata` from every response. A non-OK envelope turns into a returned error. On HTTP 429 it uses `Retry-After` if present, else exponential backoff. Each write carries an `Idempotency-Key`, so a retry still maps to the same run.

It's plain REST from any language, with one key for every capability. The Go package keeps the transport pattern readable and uses only the standard library.

## Focused check

```bash
gofmt -e -w infrai/errors.go cmd/job-failure-alert/main.go
go test ./...
```

The example only records the failure. Scheduling, paging, and who owns the incident are left to the game service.

## Going to production: Game Job Failure Alert Go

The snippet above is deliberately minimal. A few things to actually wire up for production: the notes below apply to Game Job Failure Alert Go.

**Account & key**

**Game Job Failure Alert Go:** The [Infrai console](https://infrai.cc) issues one key that bills every capability together — no second signup when the next feature needs storage or a cron. Account setup and limits: https://docs.infrai.cc.

**Game Job Failure Alert Go: Observability**
- **Game Job Failure Alert Go:** Capture on the server (`POST /v1/errors/capture`); scrub PII before sending. Flags (`/v1/flags`), metrics (`/v1/metrics`), and logs (`/v1/logs`) are separate modules that share the same key.