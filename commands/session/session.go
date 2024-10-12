package session

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/config"

	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
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
func NewCmd(db *bolt.DB, r io.Reader) *cobra.Command {
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
	return func(cmd *cobra.Command, args []string) error {
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

		go startSession(cmd, r, opts.prefix, timeout)

		if timeout.duration == 0 {
			if !timeout.timer.Stop() {
				<-timeout.timer.C
			}
		}

		<-timeout.timer.C
		return nil
	}
}

func startSession(cmd *cobra.Command, r io.Reader, prefix string, timeout *timeout) {
	reader := bufio.NewReader(r)
	root := cmd.Root()
	// The configuration is populated on start and changes inside the session won't have effect until restart.
	scripts := config.GetStringMapString("session.scripts")

	for {
		// Force a garbage collection so the memory used by argon2 isn't reserved
		// for us by the system while idle
		runtime.GC()

		fmt.Printf("%s ", prefix)
		commands, err := scanInput(reader, timeout, scripts)
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
