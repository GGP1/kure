package note

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/db/note"
	"github.com/GGP1/kure/pb"
	"github.com/awnumar/memguard"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var addExample = `
* Add a new note
kure note add noteName`

func addSubCmd(db *bolt.DB, r io.Reader) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add <name>",
		Short:   "Add a note",
		Aliases: []string{"create", "new"},
		Example: addExample,
		PreRunE: cmdutil.RequirePassword(db),
		RunE:    runAdd(db, r),
	}

	return cmd
}

func runAdd(db *bolt.DB, r io.Reader) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")
		if name == "" {
			return errInvalidName
		}

		name = strings.TrimSpace(strings.ToLower(name))

		if err := cmdutil.Exists(db, name, "note"); err != nil {
			return err
		}

		lockedBuf, n, err := input(db, name, r)
		if err != nil {
			return err
		}

		if err := note.Create(db, lockedBuf, n); err != nil {
			return err
		}

		fmt.Printf("\n%q added\n", name)
		return nil
	}
}

func input(db *bolt.DB, name string, r io.Reader) (*memguard.LockedBuffer, *pb.Note, error) {
	s := bufio.NewScanner(r)

	lockedBuf, note := pb.SecureNote()
	note.Name = name
	note.Text = cmdutil.Scanlns(s, "Text")

	if err := s.Err(); err != nil {
		return nil, nil, errors.Wrap(err, "failed scanning")
	}

	return lockedBuf, note, nil
}
