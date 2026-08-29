# PROGRESS

Status of ReplayNet. Full technical contract is in `docs/TECHNICAL_SPEC.md`.

No comments are used anywhere in the Go source, by project convention.
Keep it that way in anything you add.

## Done and verified

All of this has been built, compiled, and tested — not just scaffolded.

- **`internal/session`** (§2 of the spec) — Event struct, binary
  encode/decode (`format.go`, `encoding.go`), `Writer`/`Load` file API
  (`session.go`). `tests/session_roundtrip_test.go` passes, covering empty
  bodies, >1MB bodies, unicode paths, multi-value headers, writer→file→load
  round trip, and graceful partial-write truncation recovery.
- **`internal/proxy`** (§3) — `RecordingProxy` records request/response
  pairs with timing while forwarding traffic through to a real upstream.
  Handles upstream connection failures by recording a synthetic response
  event rather than silently dropping it. 50MB body cap enforced with 413
  responses (`tests/proxy_record_test.go`).
- **`internal/replay`** (§4) — `ReplayServer` builds a method+path index
  from adjacent request/response pairs in the recorded session, serves
  responses in original recorded order, returns 404 past the end of
  recorded responses for a given key. Fault injection (`fault.go`)
  implements all three planned modes: latency, drop (via connection
  hijack, falls back to 503 if hijack unsupported), status override.
  Fully verified in `tests/fault_injection_test.go`.
- **`internal/visualizer`** (§5) — SSE server broadcasting compact per-event
  JSON messages, non-blocking broadcast with per-subscriber buffered channels
  that drop rather than block proxy/replay throughput (`tests/visualizer_test.go`).
  Static UI (`static/index.html`, `app.js`, `style.css`) includes a live
  swimlane topology diagram with packet flow animations, metrics summary,
  real-time timeline table with `#Index` badges for fault targeting,
  search/filters, and auto-scroll.
- **`cmd/replaynet`** (§6) — CLI with `proxy` and `replay` subcommands,
  `--fault` supports repeated flags, `--inspect PORT` wires the visualizer in on
  both paths.
- **`tests/replay_match_test.go`** — end-to-end test using a real
  `httptest.NewServer` fake upstream: records a login/profile/permissions
  (fails once, succeeds on retry) sequence, then replays with the fake
  upstream shut down and asserts every response matches exactly, in order.
- **Live CLI & Browser Verification**: Verified live in Chrome with
  real HTTP traffic recorded over proxy, rendered in real time on the
  visualizer timeline at `:9001`, and replayed deterministically.
- **Reproducible build** — `scripts/reproducible-build.sh` builds twice
  with `-trimpath`, hashes both, confirmed byte-identical output.
- **`go build ./...` and `go test -v ./...`** pass clean (14 unit/integration tests).
- README, STDLIB.md, LICENSE (MIT), .zero-dep.toml, deps-proof.txt,
  Makefile all verified.

## Explicitly not built (by design — see README "Known limitations")

These are deliberate scope cuts, not oversights. Don't build them without
re-reading why they were cut first (README + spec §1 scope decision):

- TLS interception / outbound MITM proxying
- Semantic request matching (currently method+path+sequential-order only)
- Truncate / bandwidth-throttle / half-close / disconnect-after fault modes
- Session diff (`replaynet diff`)

- Package Killer target (Toxiproxy) stats verified and documented in STDLIB.md (12.3k+ GitHub stars).

## Remaining items

1. **Demo video recording** — record a full demonstration showing
   recording a failing step, replaying without the backend, altering history
   with `--fault`, and observing the live visualizer inspector.
