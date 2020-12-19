package session

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/cmd/root"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	bolt "go.etcd.io/bbolt"
)

var (
	prefix  string
	timeout time.Duration
)

var example = `
* Run a session without timeout and using "$" as the prefix
kure session -p $

* Run a session for 1 hour
kure session -t 1h

* Show the session time left (once into one)
timeout

* Exit the session
exit`

// NewCmd returns a new command.
func NewCmd(db *bolt.DB, r io.Reader, interrupt chan os.Signal) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session",
		Short: "Run a session",
		Long: `Sessions are used for doing multiple operations by providing the master password once, it's encrypted
and stored inside a locked buffer, decrypted when needed and destroyed right after it.

The user can set a timeout to automatically close the session after X amount of time. By default it never ends.

Once into the session:
• it's optional to use the word "kure" to run a command.
• type "timeout" to see the time left.
• type "exit" or press Ctrl+C to quit.`,
		Example: example,
		PreRunE: cmdutil.RequirePassword(db),
		RunE:    runSession(db, r, interrupt),
	}

	f := cmd.Flags()
	f.StringVarP(&prefix, "prefix", "p", "kure:~#", "customize the text that precedes your commands")
	f.DurationVarP(&timeout, "timeout", "t", 0, "session timeout")

	return cmd
}

func runSession(db *bolt.DB, r io.Reader, interrupt chan os.Signal) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		zero := time.Duration(0)

		if p := viper.GetString("session.prefix"); p != "" {
			prefix = p
		}
		// Set the default only if it's not zero and the user haven't specified one
		if t := viper.GetDuration("session.timeout"); t != zero && timeout == zero {
			timeout = t
		}

		scanner := bufio.NewScanner(r)
		start := time.Now()

		go startSession(scanner, start, zero, interrupt)

		if timeout == zero {
			// Block forever
			block := make(chan struct{})
			<-block
		}

		<-time.After(timeout)

		return nil
	}
}

// startSession initializes the session.
func startSession(scanner *bufio.Scanner, start time.Time, zero time.Duration, interrupt chan os.Signal) {
	for {
		fmt.Printf("\n%s ", prefix)

		scanner.Scan()
		text := strings.TrimSpace(scanner.Text())
		args := strings.Split(text, " ")

		// Session commands
		switch args[0] {
		case "exit":
			interrupt <- os.Interrupt
			return

		case "kure":
			// Make using "kure" optional
			args[0] = ""

		case "timeout":
			if timeout == zero {
				fmt.Println("The session has no timeout.")
				continue
			}
			fmt.Printf("Time left: %.2f minutes", timeout.Minutes()-time.Since(start).Minutes())
			continue
		}

		r := root.Cmd()
		r.SetArgs(args[:])
		if err := r.Execute(); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
		}
	}
}
