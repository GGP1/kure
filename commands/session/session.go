package session

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/GGP1/kure/auth"
	cmdutil "github.com/GGP1/kure/commands"

	"github.com/awnumar/memguard"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
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

// NewCmd returns a new command.
func NewCmd(db *bolt.DB, r io.Reader) *cobra.Command {
	opts := sessionOptions{}

	cmd := &cobra.Command{
		Use:   "session",
		Short: "Run a session",
		Long: `Sessions are used for doing multiple operations by providing the master password once, it's encrypted
and stored inside a locked buffer, decrypted when needed and destroyed right after it.

The user can set a timeout to automatically close the session after X amount of time. By default it never ends.

Once into the session:
• it's optional to use the word "kure" to run a command.
• type "timeout" to see the time left.
• type "exit" or press Ctrl+C to quit.
• type "pwd" to get the current working directory.`,
		Example: example,
		PreRunE: auth.Login(db),
		RunE:    runSession(db, r, &opts),
	}

	f := cmd.Flags()
	f.StringVarP(&opts.prefix, "prefix", "p", "kure:~ $", "text that precedes your commands")
	f.DurationVarP(&opts.timeout, "timeout", "t", 0, "session timeout")

	return cmd
}

func runSession(db *bolt.DB, r io.Reader, opts *sessionOptions) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		// Use config values if they are set and the flag wasn't used
		if p := "session.prefix"; viper.IsSet(p) && !cmd.Flags().Changed("prefix") {
			opts.prefix = viper.GetString(p)
		}
		if t := "session.timeout"; viper.IsSet(t) && !cmd.Flags().Changed("timeout") {
			opts.timeout = viper.GetDuration(t)
		}

		scanner := bufio.NewScanner(r)
		start := time.Now()
		go startSession(cmd, db, scanner, start, opts)

		if opts.timeout == 0 {
			// Block forever
			block := make(chan struct{})
			<-block
		}

		<-time.After(opts.timeout)
		return nil
	}
}

func startSession(cmd *cobra.Command, db *bolt.DB, scanner *bufio.Scanner, start time.Time, opts *sessionOptions) {
	for {
		fmt.Printf("\n%s ", opts.prefix)

		scanner.Scan()
		text := strings.TrimSpace(scanner.Text())
		args := strings.Split(text, " ")
		// Lower case only the command as the args will be formatted by each command
		args[0] = strings.ToLower(args[0])

		// Session commands
		switch args[0] {
		case "exit", "quit", "logout":
			db.Close()
			memguard.SafeExit(0)
			return

		case "kure":
			// Make using "kure" optional
			args = args[1:]

		case "pwd":
			dir, _ := os.Getwd()
			fmt.Println(dir)
			continue

		case "timeout":
			if opts.timeout == 0 {
				fmt.Println("The session has no timeout.")
				continue
			}
			fmt.Println("Time left:", opts.timeout-time.Since(start))
			continue
		}

		root := cmd.Root()
		root.SetArgs(args[:])
		subCmd, _, _ := root.Find(args[:])

		if err := root.Execute(); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)

			if subCmd.PostRun != nil {
				// Force PostRun to reset options variables (it isn't executed on failure)
				subCmd.PostRun(nil, nil)
			}
		}

		// Set all flags as unchanged to keep using default values
		subCmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
			f.Changed = false
		})
	}
}
