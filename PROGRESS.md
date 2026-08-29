# PROGRESS

Status as of initial scaffold. Full technical contract is in
`docs/TECHNICAL_SPEC.md` — read that first if you're an AI agent picking
this up, this file is just the status summary and next-step pointer.

No comments are used anywhere in the Go source, by project convention.
Keep it that way in anything you add.

## Done and verified

All of this has been built, compiled, and tested — not just scaffolded.

- **`internal/session`** (§2 of the spec) — Event struct, binary
  encode/decode (`format.go`, `encoding.go`), `Writer`/`Load` file API
  (`session.go`). `tests/session_roundtrip_test.go` passes, covering empty
  bodies, >1MB bodies, unicode paths, multi-value headers, and a full
  writer→file→load round trip.
- **`internal/proxy`** (§3) — `RecordingProxy` records request/response
  pairs with timing while forwarding traffic through to a real upstream.
  Handles upstream connection failures by recording a synthetic response
  event rather than silently dropping it. 50MB body cap enforced.
- **`internal/replay`** (§4) — `ReplayServer` builds a method+path index
  from adjacent request/response pairs in the recorded session, serves
  responses in original recorded order, returns 404 past the end of
  recorded responses for a given key. Fault injection (`fault.go`)
  implements all three planned modes: latency, drop (via connection
  hijack, falls back to 503 if hijack unsupported), status override.
- **`internal/visualizer`** — SSE server broadcasting compact per-event
  JSON messages (not per-byte), non-blocking broadcast with per-subscriber
  buffered channels that drop rather than block the proxy/replay path if a
  browser tab is slow. Static UI (`static/index.html`, `app.js`,
  `style.css`) is plain HTML/CSS/vanilla JS, embedded via `//go:embed`, no
  framework, no build step.
- **`cmd/replaynet`** — CLI with `proxy` and `replay` subcommands, `--fault`
  supports repeated flags, `--inspect PORT` wires the visualizer in on
  both paths.
- **`tests/replay_match_test.go`** — end-to-end test using a real
  `httptest.NewServer` fake upstream: records a login/profile/permissions
  (fails once, succeeds on retry) sequence, then replays with the fake
  upstream shut down and asserts every response matches exactly, in order.
  This is the core "recreate the bug without the original server" claim,
  verified.
- **Live CLI smoke test performed manually** (not just `go test`): built
  the actual binary, ran a real fake upstream + proxy over real TCP ports,
  recorded a 500-then-200 retry sequence, killed both processes entirely,
  ran `replaynet replay` and confirmed the exact same sequence reproduced
  with zero backend running. Also confirmed `--fault at=N,type=status,code=503`
  correctly overrides a specific recorded event by index.
- **Reproducible build** — `scripts/reproducible-build.sh` builds twice
  with `-trimpath`, hashes both, confirmed byte-identical output. Bonus is
  in working order, not just planned.
- **`go build ./...` and `go test ./...` both pass clean.**
- README, STDLIB.md, LICENSE (MIT), .zero-dep.toml, deps-proof.txt,
  Makefile all present per the submission checklist in the main hackathon
  spec.

## Explicitly not built (by design — see README "Known limitations")

These are deliberate scope cuts, not oversights. Don't build them without
re-reading why they were cut first (README + spec §1 scope decision):

- TLS interception / outbound MITM proxying
- Semantic request matching (currently method+path+sequential-order only)
- Truncate / bandwidth-throttle / half-close / disconnect-after fault modes
- Session diff (`replaynet diff`)

## What's actually left to do

Roughly in priority order:

1. **`--inspect` has not been exercised live end-to-end** (SSE endpoint +
   static file serving compiles and the logic is unit-reasoned, but nobody
   has opened a browser at `/` and watched a real recording/replay update
   the timeline live). Do this next — start `replaynet proxy --inspect :9001`,
   open `http://localhost:9001`, generate some traffic, confirm the
   timeline renders and updates. This is your best demo asset, don't
   discover a bug in it the night before recording the video.
2. **No test yet for the visualizer's drop-on-full-buffer behavior**
   (spec §5.2's non-blocking constraint). A quick test: subscribe a
   channel, don't drain it, broadcast >64 events, confirm the proxy/replay
   path never blocks and the subscriber just misses some messages.
3. **No test yet for the `FaultDrop` hijack path specifically** — the
   status-override fault was verified live, latency and drop have not been
   individually confirmed with a real client observing a connection reset
   vs. a hang. Worth a manual curl test with `-v` to watch the actual TCP
   behavior.
4. **Upstream body re-reading on retries**: `outReq` in `proxy.go` is built
   from the already-read `reqBody` bytes via `bytes.NewReader`, which is
   correct, but there's no test yet confirming large-body proxying (e.g. a
   multi-MB PUT) works end-to-end through the live proxy, only through the
   session format's own round-trip test. Worth one live test with a large
   file.
5. **README's stated 50MB body cap has no test** confirming a request over
   the limit actually gets a 413 rather than some other failure mode.
6. **Package Killer target (Toxiproxy) download/usage numbers are not yet
   verified** — STDLIB.md says to check before citing a number; that check
   hasn't been done yet. Do this close to submission time so the number is
   current.
7. **Demo video not started.** Suggested structure per the original pitch:
   record a real multi-step sequence with a deliberate failure, kill the
   backend, replay, show it recreate the bug, then use `--fault` to "alter
   history" and show the app behave differently on a re-run. The
   visualizer running live in a browser during this is the strongest
   single shot in the video — plan the recording around having that
   working first (see item 1).
8. **No `.rnet` corruption/partial-write handling.** If the process is
   killed mid-`Append`, `Load` will hit `io.ErrUnexpectedEOF` reading a
   truncated final record and return an error rather than gracefully
   truncating and loading everything before it. This isn't in the locked
   spec as a requirement, but it's a plausible judge question ("what
   happens if recording crashes mid-write?") worth having an honest answer
   to — either fix it to truncate-and-load-partial, or add one line to
   README's Known Limitations explicitly stating recording is not itself
   crash-safe (only replay determinism is guaranteed once a session is
   fully recorded).

## How to pick this up

```
git init
git add -A
git commit -m "initial scaffold: session format, proxy, replay, visualizer, CLI"
make build
make test
```

Read `docs/TECHNICAL_SPEC.md` §8 ("Notes for AI-assisted implementation")
before generating new code in any of these packages — it flags the exact
spots (fault-drop behavior, the session round-trip test's role as a gate)
where an assistant is likely to quietly change something that matters.
