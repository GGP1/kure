package session

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/GGP1/kure/sig"
)

// command represents a session command.
type command func(params) bool

// params contains all commands' required parameters.
type params struct {
	in, out, outErr io.ReadWriter

	args    []string
	start   time.Time
	timeout time.Duration
}

var commands = map[string]command{
	"": func(_ params) bool {
		return true
	},
	"block": func(p params) bool {
		fmt.Fprint(p.out, "Press Enter to continue")
		dump := ""
		fmt.Fscanln(p.in, &dump)
		return true
	},
	"exit": func(_ params) bool {
		sig.Signal.Kill()
		return false
	},
	"quit": func(_ params) bool {
		sig.Signal.Kill()
		return false
	},
	"pwd": func(p params) bool {
		dir, _ := os.Getwd()
		fmt.Fprintln(p.out, dir)
		return true
	},
	"sleep": func(p params) bool {
		d, err := time.ParseDuration(p.args[1])
		if err != nil {
			fmt.Fprintf(p.outErr, "error: invalid duration %q\n", p.args[1])
		}
		time.Sleep(d)
		return true
	},
	"timeout": func(p params) bool {
		if p.timeout == 0 {
			fmt.Fprintln(p.out, "The session has no timeout.")
			return true
		}
		fmt.Fprintln(p.out, "Time left:", p.timeout-time.Since(p.start))
		return true
	},
}

// sessionCommand checks for any session command and returns a boolean representing
// a "continue" in the loop where it was called.
func sessionCommand(args []string, start time.Time, opts *sessionOptions) bool {
	cmd, ok := commands[args[0]]
	if !ok {
		return false
	}

	return cmd(params{
		in:      os.Stdin,
		out:     os.Stdout,
		outErr:  os.Stderr,
		args:    args,
		start:   start,
		timeout: opts.timeout,
	})
}
