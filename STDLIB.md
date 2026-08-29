# STDLIB.md

Every package this project would normally reach for, and the standard
library feature used instead.

| You'd install | Use instead | Where |
|---|---|---|
| `http-proxy` / `http-proxy-middleware` (Node) or a Go reverse-proxy helper | Hand-rolled forwarding handler over `net/http` | `internal/proxy/proxy.go` — reads the incoming request, builds and issues an outbound request via `http.Client`, and writes the response through, rather than using `httputil.ReverseProxy` or a package |
| A binary serialization library (protobuf, msgpack) | `encoding/binary` with a hand-designed record format | `internal/session/format.go`, `internal/session/encoding.go` — fixed-width and length-prefixed fields, big-endian throughout |
| A hashing package | `crypto/sha256` | Body checksums recorded per event in `internal/session/format.go` |
| `ws` / `socket.io` (Node) or a WebSocket library | Server-Sent Events over plain `net/http` | `internal/visualizer/server.go` — avoids implementing a WebSocket handshake for a use case (server-to-browser push) that SSE covers with a plain HTTP response stream |
| A frontend framework (React, Vue) | `embed` + vanilla HTML/CSS/JS | `internal/visualizer/static/` — no build step, no framework, bundled directly into the binary |
| A CLI framework (cobra, urfave/cli) | `flag` | `cmd/replaynet/main.go`, `cmd/replaynet/flags.go` |
| A test framework / assertion library | `testing`, `net/http/httptest` | `tests/` — including a real fake upstream via `httptest.NewServer` for end-to-end record/replay verification |
| A mocking library for the fake upstream in tests | `net/http/httptest.NewServer` with a real handler | `tests/replay_match_test.go` |
| A UUID/session-ID library | Session identification handled via file path rather than an embedded ID; if a generated ID is added later, `crypto/rand` covers it without a package |
| Toxiproxy (as an installed dependency) | This project's own fault-injection replay logic | `internal/replay/fault.go` — same conceptual space (network-condition fault injection for testing), reimplemented rather than depended on |

## Package Killer

Positioned against **Toxiproxy**, a widely-used chaos-engineering fault
injection tool for network testing. ReplayNet covers the same conceptual
need — deliberately injecting latency, drops, and failures into a network
conversation for testing purposes — through a from-scratch implementation
with zero dependencies, plus the additional record/replay capability
Toxiproxy does not provide. Verify current adoption/download figures before
citing a specific number in any final write-up rather than reusing an
unverified figure.
