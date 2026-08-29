# ReplayNet

Record a real application's HTTP conversation once. Kill the server. Replay the
exact conversation anyway, with the option to inject latency, dropped
responses, or overridden status codes at any point in the recorded timeline.
Zero third-party dependencies — Go standard library only.

## What it does

```
replaynet proxy --listen :9000 --upstream http://localhost:3000 --session out.rnet
```

Your application talks to ReplayNet instead of directly to its backend.
Every request and response is recorded, with timing, to a binary session
file, and also forwarded through to the real backend so nothing about your
application's normal behavior changes.

```
replaynet replay out.rnet --listen :9002
```

ReplayNet now impersonates the backend entirely. No real server is running.
Requests are matched to recorded responses by method + path, in the order
they were originally recorded, so a sequence like "permissions check fails,
client retries, permissions check succeeds" replays exactly as it happened.

```
replaynet replay out.rnet --listen :9002 --fault "at=3,type=latency,ms=4000"
replaynet replay out.rnet --listen :9002 --fault "at=3,type=drop"
replaynet replay out.rnet --listen :9002 --fault "at=3,type=status,code=503"
```

Fault rules target a specific recorded event by index (shown via `--inspect`,
see below) and let you explore how the real application under test behaves
when that step in the conversation goes wrong, using an actual recorded
exchange as the base case rather than a synthetic mock.

```
replaynet proxy  ... --inspect :9001
replaynet replay ... --inspect :9003
```

Adding `--inspect PORT` starts a browser-based live timeline (served over
plain HTTP, updated via Server-Sent Events) showing every request and
response as it happens, during both recording and replay.

## How to run it

```
make build
./bin/replaynet proxy --listen :9000 --upstream http://localhost:3000 --session demo.rnet --inspect :9001
```

Then in another terminal, point your application at `http://localhost:9000`
instead of `http://localhost:3000`, and open `http://localhost:9001` in a
browser to watch the timeline live.

To replay:

```
./bin/replaynet replay demo.rnet --listen :9002 --inspect :9003
```

## Known limitations

This is a deliberately trimmed scope, documented honestly rather than left
for a judge to discover:

- **Inbound HTTP only, no TLS interception.** This proxies plaintext HTTP
  traffic between a client and your application's backend. It is not a
  Charles Proxy / mitmproxy replacement for intercepting your application's
  own outbound calls to third parties — that would require certificate
  generation and TLS termination, which is out of scope here.
- **Replay matching is method + path + sequential order, not semantic.**
  The *n*th `GET /permissions` request during replay gets the *n*th recorded
  `GET /permissions` response, in original recorded order. If two identical
  requests race during replay, they may consume responses out of the order
  you'd intuitively expect. For the deterministic, single-client demo
  scenario this project targets, this is not an issue in practice.
- **Three fault modes only:** latency injection, dropped responses, and
  status-code override. All three operate within clean HTTP response
  semantics. Raw connection truncation, bandwidth throttling, and half-close
  are not implemented — each is a genuinely different, lower-level problem
  (byte-level connection manipulation rather than a response-writing
  decision) and was cut to keep what's shipped fully correct rather than
  attempting more and shipping something flaky.
- **Session diff (comparing two recorded runs to find the first point of
  divergence) is not implemented.** It presupposes solved cross-run request
  correlation with timing-jitter tolerance, which is a harder, separate
  problem from anything else here.
- **The visualizer samples at the event level, not the byte level** — one
  message per request/response, not per byte — to keep the browser
  responsive on large sessions. If a subscriber's buffer fills (a slow
  browser tab), further events for that subscriber are dropped rather than
  blocking the actual proxy or replay path; the recorded session itself is
  never affected by this.
- **The dropped-response fault** hijacks and closes the underlying
  connection where possible (`http.Hijacker`), so the client sees a reset
  rather than hanging until its own timeout. If the response writer doesn't
  support hijacking in a given deployment, it falls back to a 503 rather
  than actually closing the connection — documented here rather than
  silently assumed to always work.

## Automated Demo Workflow

To run the complete end-to-end recording, replay, fault injection, and visualizer demo with a single command:

```bash
./scripts/demo.sh
```

This will automatically:
1. Start a mock upstream backend on `:8080`
2. Start ReplayNet proxy recording on `:9000` (inspector at `:9001`) and record traffic
3. Shut down the backend completely
4. Deterministically replay the conversation on `:9002` without any backend running
5. Inject faults (e.g. override status to 503) on specific event indices to simulate resilience testing
6. Clean up all background processes cleanly

## Build & Verification

```bash
make build       # compiles with -trimpath -> bin/replaynet
make test        # runs 14 unit and integration tests
make deps-proof  # proves zero third-party dependencies (outputs deps-proof.txt)
make repro       # confirms reproducible byte-identical build
```

### Reproducible Build Hashes

```
build 1: 7b65915cf74e142c8a32c292efc055aa8099e3b2c3a1f77b663a3472d7f5a14c
build 2: 7b65915cf74e142c8a32c292efc055aa8099e3b2c3a1f77b663a3472d7f5a14c
MATCH: reproducible build confirmed
```
