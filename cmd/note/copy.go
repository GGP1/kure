package note

import (
	"fmt"
	"strings"
	"time"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/db/note"

	"github.com/atotto/clipboard"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var timeout time.Duration

var copyExample = `
* Copy and clean after 30s
kure note copy -t 30s`

// copySubCmd returns the copy subcommand
func copySubCmd(db *bolt.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "copy <name>",
		Short:   "Copy note text",
		Aliases: []string{"cp"},
		Example: copyExample,
		PreRunE: cmdutil.RequirePassword(db),
		RunE:    runCopy(db),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset flags (session)
			timeout = 0
		},
	}

	f := cmd.Flags()
	f.DurationVarP(&timeout, "timeout", "t", 0, "clipboard cleaning timeout")

	return cmd
}

func runCopy(db *bolt.DB) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")
		if name == "" {
			return errInvalidName
		}

		lockedBuf, note, err := note.Get(db, name)
		if err != nil {
			return err
		}

		if err := clipboard.WriteAll(note.Text); err != nil {
			return errors.Wrap(err, "failed writing to the clipboard")
		}
		lockedBuf.Destroy()

		fmt.Println("Note copied to clipboard")

		if timeout > 0 {
			<-time.After(timeout)
			clipboard.WriteAll("")
		}

		return nil
	}
}
