package session

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"time"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/config"
	"github.com/GGP1/kure/sig"
	"github.com/pkg/errors"

	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
)

const example = `
* Run a session without timeout and using "$" as the prefix
kure session -p $

* Run a session for 1 hour
kure session -t 1h`

type sessionOptions struct {
	prefix  string
	timeout time.Duration
}

// NewCmd returns a new command.
func NewCmd(r io.Reader) *cobra.Command {
	opts := sessionOptions{}
	cmd := &cobra.Command{
		Use:   "session",
		Short: "Run a session",
		Long: `Sessions let you do multiple operations by providing the master password once.
		
They support running scripts using the logical AND (&&) operator and executing pre-defined ones from the configuration file by using their aliases.

During a session, the master password is encrypted and stored inside a protected buffer.

Session commands:
• block - block execution (to be manually unlocked).
• exit|quit|Ctrl+C - close the session.
• pwd - show current directory.
• timeout - show time left.
• ttadd [duration] - increase/decrease timeout.
• ttset [duration] - set a new timeout.
• sleep [duration] - sleep for x time.`,
		Example: example,
		RunE:    runSession(r, &opts),
	}

	f := cmd.Flags()
	f.StringVarP(&opts.prefix, "prefix", "p", "kure:~ $", "text that precedes your commands")
	f.DurationVarP(&opts.timeout, "timeout", "t", 0, "session timeout")

	return cmd
}

func runSession(r io.Reader, opts *sessionOptions) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, _ []string) error {
		// Use config values if they are set and the flag wasn't used
		if p := "session.prefix"; config.IsSet(p) && !cmd.Flags().Changed("prefix") {
			opts.prefix = config.GetString(p)
		}
		if t := "session.timeout"; config.IsSet(t) && !cmd.Flags().Changed("timeout") {
			opts.timeout = config.GetDuration(t)
		}

		timeout := &timeout{
			duration: opts.timeout,
			start:    time.Now(),
			timer:    time.NewTimer(opts.timeout),
		}

		rl, err := readline.NewEx(&readline.Config{
			Prompt: opts.prefix + " ",
			Stdin:  io.NopCloser(r),
		})
		if err != nil {
			return errors.Wrap(err, "creating terminal")
		}
		defer rl.Close()
		sig.Signal.AddCleanup(func() error { return rl.Close() })

		go startSession(cmd, rl, timeout)

		if timeout.duration == 0 {
			if !timeout.timer.Stop() {
				<-timeout.timer.C
			}
		}

		<-timeout.timer.C
		return nil
	}
}

func startSession(cmd *cobra.Command, rl *readline.Instance, timeout *timeout) {
	root := cmd.Root()
	// The configuration is populated on start and changes inside the session won't have effect until restart.
	scripts := config.GetStringMapString("session.scripts")

	for {
		// Force a garbage collection so the memory used by argon2 isn't reserved
		// for us by the system while idle
		runtime.GC()

		commands, err := scanInput(rl, timeout, scripts)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			continue
		}

		if err := execute(root, commands, timeout); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
		}
	}
}

func execute(root *cobra.Command, commands [][]string, timeout *timeout) error {
	for _, args := range commands {
		args = removeEmptyItems(args)
		if len(args) == 0 || args[0] == "" {
			continue
		}

		if args[0] == "kure" {
			args = args[1:]
		}

		if ran := runSessionCommand(args, timeout); ran {
			continue
		}

		root.SetArgs(args)
		subCmd, _, _ := root.Find(args)
		if subCmd.Name() == "session" {
			continue
		}

		if err := root.Execute(); err != nil {
			if subCmd.PostRun != nil {
				// Force PostRun to reset options variables (as it isn't executed on failure)
				subCmd.PostRun(nil, nil)
			}
			return err
		}

		cleanup(subCmd)
	}
	return nil
}
