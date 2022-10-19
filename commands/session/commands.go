package session

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/GGP1/kure/sig"
	"github.com/GGP1/kure/terminal"
)

var (
	commands = map[string]command{
		"":        func(_ params) {},
		"block":   blockFn,
		"exit":    exitFn,
		"quit":    exitFn,
		"pwd":     pwdFn,
		"sleep":   sleepFn,
		"timeout": timeoutFn,
		"timer":   timerFn,
		"ttadd":   ttaddFn,
		"ttset":   ttsetFn,
	}

	cmdParams = params{
		in:     os.Stdin,
		out:    os.Stdout,
		outErr: os.Stderr,
	}
)

type command func(params)

// params contains all commands' required parameters.
type params struct {
	in, out, outErr io.ReadWriter

	timeout *timeout
	args    []string
}

// runSessionCommand checks for any session command and returns whether
// one of them has been executed or not, if none was found it returns false.
func runSessionCommand(args []string, timeout *timeout) bool {
	// The arguments length will be zero only if the user input is "kure"
	if len(args) == 0 {
		return false
	}

	cmd, ok := commands[args[0]]
	if !ok {
		return false
	}

	cmdParams.args = args[1:]
	cmdParams.timeout = timeout
	cmd(cmdParams)

	return true
}

func blockFn(p params) {
	fmt.Fprint(p.out, "Press Enter to continue")
	dump := ""
	fmt.Fscanln(p.in, &dump)
}

func exitFn(_ params) {
	sig.Signal.Kill()
}

func pwdFn(p params) {
	dir, _ := os.Getwd()
	fmt.Fprintln(p.out, dir)
}

func sleepFn(p params) {
	if len(p.args) < 1 {
		fmt.Fprintln(p.outErr, "error: invalid duration, use sleep [duration]")
		return
	}
	d, err := time.ParseDuration(p.args[0])
	if err != nil {
		fmt.Fprintf(p.outErr, "error: invalid duration %q\n", p.args[0])
	}
	time.Sleep(d)
}

func timeoutFn(p params) {
	if p.timeout.duration == 0 {
		fmt.Fprintln(p.out, "The session has no timeout.")
		return
	}
	fmt.Fprintln(p.out, p.timeout)
}

func timerFn(p params) {
	if p.timeout.duration == 0 {
		fmt.Fprintln(p.out, "The session has no timeout.")
		return
	}

	fmt.Fprintln(p.out, "Press Enter to stop the timer")
	done := make(chan struct{})

	go terminal.Ticker(done, true, func() {
		fmt.Fprint(p.out, p.timeout)
	})

	dump := ""
	fmt.Fscanln(p.in, &dump)
	done <- struct{}{}
}

func ttaddFn(p params) {
	if len(p.args) < 1 {
		fmt.Fprintln(p.outErr, "error: invalid duration, use ttadd [duration]")
		return
	}

	d, err := time.ParseDuration(p.args[0])
	if err != nil {
		fmt.Fprintf(p.outErr, "error: invalid duration %q\n", p.args[0])
		return
	}

	p.timeout.add(d)
	fmt.Fprintln(p.out, p.timeout)
}

func ttsetFn(p params) {
	if len(p.args) < 1 {
		fmt.Fprintln(p.outErr, "error: invalid duration, use ttset [duration]")
		return
	}

	d, err := time.ParseDuration(p.args[0])
	if err != nil {
		fmt.Fprintf(p.outErr, "error: invalid duration %q\n", p.args[0])
		return
	}

	p.timeout.set(d)
	fmt.Fprintln(p.out, p.timeout)
}
