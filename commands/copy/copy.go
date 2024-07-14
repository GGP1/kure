package copy

import (
	"strings"
	"time"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/entry"

	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

const example = `
* Copy password and clean after 15m
kure copy Sample -t 15m

* Copy username
kure copy Sample -u

* Copy both username and password consecutively
kure copy Sample -a`

type copyOptions struct {
	timeout  time.Duration
	username bool
	all      bool
}

// NewCmd returns a new command.
func NewCmd(db *bolt.DB) *cobra.Command {
	opts := copyOptions{}
	cmd := &cobra.Command{
		Use:     "copy <name>",
		Short:   "Copy entry credentials to the clipboard",
		Aliases: []string{"cp"},
		Example: example,
		Args:    cmdutil.MustExist(db, cmdutil.Entry),
		RunE:    runCopy(db, &opts),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset variables (session)
			opts = copyOptions{}
		},
	}

	f := cmd.Flags()
	f.DurationVarP(&opts.timeout, "timeout", "t", 0, "clipboard clearing timeout")
	f.BoolVarP(&opts.username, "username", "u", false, "copy entry username")
	f.BoolVarP(&opts.all, "all", "a", false, "copy entry username and password consecutively")

	return cmd
}

func runCopy(db *bolt.DB, opts *copyOptions) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")
		name = cmdutil.NormalizeName(name)

		e, err := entry.Get(db, name)
		if err != nil {
			return err
		}

		if opts.all {
			if err := cmdutil.WriteClipboard(cmd, opts.timeout, "Username", e.Username); err != nil {
				return err
			}

			return cmdutil.WriteClipboard(cmd, opts.timeout, "Password", e.Password)
		}

		field := "Password"
		copy := e.Password
		if opts.username {
			field = "Username"
			copy = e.Username
		}

		return cmdutil.WriteClipboard(cmd, opts.timeout, field, copy)
	}
}
