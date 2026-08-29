package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"replaynet/internal/proxy"
	"replaynet/internal/replay"
	"replaynet/internal/session"
	"replaynet/internal/visualizer"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "proxy":
		runProxy(os.Args[2:])
	case "replay":
		runReplay(os.Args[2:])
	default:
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: replaynet proxy --listen :9000 --upstream http://localhost:3000 --session out.rnet [--inspect :9001]")
	fmt.Fprintln(os.Stderr, "       replaynet replay session.rnet --listen :9002 [--fault at=N,type=latency,ms=N] [--inspect :9003]")
}

func runProxy(args []string) {
	fs := flag.NewFlagSet("proxy", flag.ExitOnError)
	listen := fs.String("listen", ":9000", "")
	upstream := fs.String("upstream", "", "")
	sessionPath := fs.String("session", "session.rnet", "")
	inspect := fs.String("inspect", "", "")
	fs.Parse(args)

	if *upstream == "" {
		fmt.Fprintln(os.Stderr, "error: --upstream is required")
		os.Exit(1)
	}

	var viz *visualizer.Server
	if *inspect != "" {
		viz = visualizer.New()
		go func() {
			fmt.Println("inspector listening on", *inspect)
			http.ListenAndServe(*inspect, viz.Handler())
		}()
	}

	var onEvent func(session.Event)
	if viz != nil {
		onEvent = viz.Broadcast
	}

	p, err := proxy.New(*upstream, *sessionPath, onEvent)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	defer p.Close()

	fmt.Println("proxying", *listen, "->", *upstream)
	fmt.Println("recording to", *sessionPath)

	if err := http.ListenAndServe(*listen, p); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func runReplay(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "error: session file required")
		os.Exit(1)
	}
	sessionPath := args[0]

	fs := flag.NewFlagSet("replay", flag.ExitOnError)
	listen := fs.String("listen", ":9002", "")
	inspect := fs.String("inspect", "", "")
	var faultArgs repeatedFlag
	fs.Var(&faultArgs, "fault", "")
	fs.Parse(args[1:])

	sess, err := session.Load(sessionPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error loading session:", err)
		os.Exit(1)
	}

	var viz *visualizer.Server
	if *inspect != "" {
		viz = visualizer.New()
		go func() {
			fmt.Println("inspector listening on", *inspect)
			http.ListenAndServe(*inspect, viz.Handler())
		}()
	}

	var onEvent func(session.Event)
	if viz != nil {
		onEvent = viz.Broadcast
	}

	faults, err := parseFaults(faultArgs)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error parsing --fault:", err)
		os.Exit(1)
	}

	srv := replay.New(sess, faults, onEvent)

	fmt.Println("replaying", sessionPath, "on", *listen)

	if err := http.ListenAndServe(*listen, srv); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func parseFaults(raw []string) ([]replay.FaultRule, error) {
	var rules []replay.FaultRule
	for _, r := range raw {
		rule := replay.FaultRule{}
		parts := strings.Split(r, ",")
		for _, part := range parts {
			kv := strings.SplitN(part, "=", 2)
			if len(kv) != 2 {
				return nil, fmt.Errorf("malformed fault clause: %q", part)
			}
			key, val := kv[0], kv[1]
			switch key {
			case "at":
				n, err := strconv.Atoi(val)
				if err != nil {
					return nil, err
				}
				rule.AtEventIndex = n
			case "type":
				switch val {
				case "latency":
					rule.Type = replay.FaultLatency
				case "drop":
					rule.Type = replay.FaultDrop
				case "status":
					rule.Type = replay.FaultStatus
				default:
					return nil, fmt.Errorf("unknown fault type: %q", val)
				}
			case "ms":
				n, err := strconv.Atoi(val)
				if err != nil {
					return nil, err
				}
				rule.LatencyMs = n
			case "code":
				n, err := strconv.Atoi(val)
				if err != nil {
					return nil, err
				}
				rule.StatusOverride = n
			}
		}
		rules = append(rules, rule)
	}
	return rules, nil
}
