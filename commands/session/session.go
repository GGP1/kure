package session

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/GGP1/kure/auth"
	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/config"
	"github.com/GGP1/kure/sig"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	bolt "go.etcd.io/bbolt"
)

var example = `
* Run a session without timeout and using "$" as the prefix
kure session -p $

* Run a session for 1 hour
kure session -t 1h`

type sessionOptions struct {
	prefix  string
	timeout time.Duration
}

type timeout struct {
	t     time.Duration
	start time.Time
	timer *time.Timer
}

// NewCmd returns a new command.
func NewCmd(db *bolt.DB, r io.Reader) *cobra.Command {
	opts := sessionOptions{}

	cmd := &cobra.Command{
		Use:   "session",
		Short: "Run a session",
		Long: `Sessions are used for doing multiple operations by providing the master password once, it's encrypted and stored inside a locked buffer, decrypted when needed and destroyed right after it.

Scripts can be created in the configuration file and executed inside sessions by using their aliases and, optionally, passing arguments.

Once into a session:
• use "&&" to execute a commands sequence.
• it's optional to use the word "kure" to run a command.

Session commands:
• block - block execution (to be manually unlocked).
• exit|quit|Ctrl+C - close the session.
• pwd - show current directory.
• timeout - show time left.
• ttadd [duration] - increase/decrease timeout.
• ttset [duration] - set a new timeout.
• sleep [duration] - sleep for x time.`,
		Example: example,
		PreRunE: auth.Login(db),
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
			t:     opts.timeout,
			start: time.Now(),
			timer: time.NewTimer(opts.timeout),
		}

		go startSession(cmd, r, opts.prefix, timeout)

		if timeout.t == 0 {
			timeout.timer.Stop()
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
		// for us by the system while sleeping
		runtime.GC()
		fmt.Printf("%s ", prefix)

		text, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				sig.Signal.Kill()
			}
			fmt.Fprintln(os.Stderr, "error:", err)
			continue
		}

		args := strings.Split(string(text), " ")

		script, ok := scripts[args[0]]
		if ok {
			script = fillScript(args[1:], script)
			args = strings.Split(script, " ")
		}

		if err := execute(root, args, timeout); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
		}
	}
}

// cleanup resets signal cleanups and sets all flags as unchanged to keep using default values.
//
// It also sets the help flag internal variable to false in case it's used.
func cleanup(cmd *cobra.Command) {
	sig.Signal.ResetCleanups()
	cmd.LocalFlags().VisitAll(func(f *pflag.Flag) { f.Changed = false })
	cmd.Flags().Set("help", "false")
}

func execute(root *cobra.Command, args []string, timeout *timeout) error {
	cmdsGroup := parseCmds(args)

	for _, args := range cmdsGroup {
		if len(args) == 0 {
			continue
		}
		if args[0] == "kure" {
			args = args[1:]
		}

		cont := sessionCommand(args, timeout)
		if cont {
			continue
		}

		root.SetArgs(args)
		subCmd, _, _ := root.Find(args)

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

// fillScript returns the script with the arguments replaced.
func fillScript(args []string, script string) string {
	if !strings.Contains(script, "$") {
		return script
	}

	arg := 1 // Start from $1 like bash
	for i := 0; i < len(args); i++ {
		name := args[i]

		if strings.HasPrefix(name, "\"") {
			// Look for the closing quote
			for j, a := range args[i:] {
				if strings.HasSuffix(a, "\"") {
					words := strings.Join(args[i+1:i+j+1], " ") // Add ones to exclude first and include second element
					name = strings.TrimPrefix(name, "\"") + " " + strings.TrimSuffix(words, "\"")
					i += j // Skip joined words
					break
				}
			}
		}

		script = strings.ReplaceAll(script, fmt.Sprintf("$%d", arg), name)
		arg++
	}

	return script
}

// Given
// 		"ls && copy github && 2fa"
// return
// 		[{"ls"}, {"copy", "github"}, {"2fa"}].
func parseCmds(args []string) [][]string {
	// The underlying array will grow only if the script has multiple "&&" in a row
	ampersands := make([]int, 0, len(args)/2)
	for i, a := range args {
		if a == "&&" {
			// Store the indices of the ampersands
			ampersands = append(ampersands, i)
		}
	}

	// Pass on the args received if no ampersand was found
	if len(ampersands) == 0 {
		return [][]string{args}
	}

	group := make([][]string, 0, len(ampersands)+1)
	lastIdx := 0
	// Append len(ampersands) commands to the group
	for _, idx := range ampersands {
		group = append(group, args[lastIdx:idx])
		lastIdx = idx + 1 // Add one to skip the ampersand
	}

	// Append the last command
	group = append(group, args[lastIdx:])

	return group
}
