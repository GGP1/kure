package session

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/GGP1/kure/sig"
)

// params contains all commands' required parameters.
type params struct {
	in, out, outErr io.ReadWriter

	args    []string
	timeout *timeout
}

var commands = map[string]func(params){
	"": func(_ params) {},
	"block": func(p params) {
		fmt.Fprint(p.out, "Press Enter to continue")
		dump := ""
		fmt.Fscanln(p.in, &dump)
	},
	"exit": func(_ params) {
		sig.Signal.Kill()
	},
	"quit": func(_ params) {
		sig.Signal.Kill()
	},
	"pwd": func(p params) {
		dir, _ := os.Getwd()
		fmt.Fprintln(p.out, dir)
	},
	"sleep": func(p params) {
		if len(p.args) < 1 {
			fmt.Fprintln(p.outErr, "error: invalid duration, use sleep [duration]")
			return
		}
		d, err := time.ParseDuration(p.args[0])
		if err != nil {
			fmt.Fprintf(p.outErr, "error: invalid duration %q\n", p.args[0])
		}
		time.Sleep(d)
	},
	"timeout": func(p params) {
		if p.timeout.t == 0 {
			fmt.Fprintln(p.out, "The session has no timeout.")
			return
		}
		fmt.Fprintln(p.out, "Time left:", p.timeout.t-time.Since(p.timeout.start))
	},
	"ttadd": func(p params) {
		if len(p.args) < 1 {
			fmt.Fprintln(p.outErr, "error: invalid duration, use ttadd [duration]")
			return
		}

		d, err := time.ParseDuration(p.args[0])
		if err != nil {
			fmt.Fprintf(p.outErr, "error: invalid duration %q\n", p.args[0])
			return
		}

		if d == 0 {
			return
		}

		if p.timeout.t == 0 {
			p.timeout.start = time.Now()
		}
		p.timeout.timer.Reset(p.timeout.t - time.Since(p.timeout.start) + d)
		p.timeout.t += d
		fmt.Fprintln(p.out, "Time left:", p.timeout.t-time.Since(p.timeout.start))
	},
	"ttset": func(p params) {
		if len(p.args) < 1 {
			fmt.Fprintln(p.outErr, "error: invalid duration, use ttset [duration]")
			return
		}

		d, err := time.ParseDuration(p.args[0])
		if err != nil {
			fmt.Fprintf(p.outErr, "error: invalid duration %q\n", p.args[0])
			return
		}

		if d == 0 {
			p.timeout.timer.Stop()
			p.timeout.t = 0
			return
		}

		p.timeout.start = time.Now()
		p.timeout.t = d
		p.timeout.timer.Reset(d)
		fmt.Fprintln(p.out, "Time left:", p.timeout.t-time.Since(p.timeout.start))
	},
}

// sessionCommand checks for any session command and returns a boolean representing
// a "continue" in the loop where it was called.
func sessionCommand(args []string, timeout *timeout) bool {
	// The arguments length will be zero only if the user input is "kure"
	if len(args) == 0 {
		return false
	}

	cmd, ok := commands[args[0]]
	if !ok {
		return false
	}

	cmd(params{
		in:      os.Stdin,
		out:     os.Stdout,
		outErr:  os.Stderr,
		args:    args[1:],
		timeout: timeout,
	})

	return true
}
