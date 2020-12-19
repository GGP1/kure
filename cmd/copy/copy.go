package copy

import (
	"fmt"
	"strings"
	"time"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/db/entry"

	"github.com/atotto/clipboard"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var (
	timeout  time.Duration
	username bool
)

var example = `
* Copy password and clean after 15m
kure copy entryName -t 15m

* Copy username
kure copy entryName -u`

// NewCmd returns a new command.
func NewCmd(db *bolt.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "copy <name>",
		Short:   "Copy entry credentials to the clipboard",
		Aliases: []string{"cp"},
		Example: example,
		PreRunE: cmdutil.RequirePassword(db),
		RunE:    runCopy(db),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset flags (session)
			timeout = 0
			username = false
		},
	}

	f := cmd.Flags()
	f.DurationVarP(&timeout, "timeout", "t", 0, "clipboard cleaning timeout")
	f.BoolVarP(&username, "username", "u", false, "copy entry username")

	return cmd
}

func runCopy(db *bolt.DB) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")
		if name == "" {
			return errors.New("invalid name")
		}

		entry, err := entry.Get(db, name)
		if err != nil {
			return err
		}

		field := "Password"
		copy := entry.Password
		if username {
			field = "Username"
			copy = entry.Username
		}

		if err := clipboard.WriteAll(copy); err != nil {
			return errors.Wrap(err, "failed writing to the clipboard")
		}

		fmt.Printf("%s copied to clipboard\n", field)

		if timeout > 0 {
			<-time.After(timeout)
			clipboard.WriteAll("")
		}

		return nil
	}
}
