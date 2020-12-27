package card

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/db/card"
	"github.com/GGP1/kure/pb"
	"github.com/awnumar/memguard"
	"github.com/pkg/errors"

	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var addExample = `
* Add a new card
kure card add cardName`

func addSubCmd(db *bolt.DB, r io.Reader) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add <name>",
		Short:   "Add a card",
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

		if err := cmdutil.Exists(db, name, "card"); err != nil {
			return err
		}

		lockedBuf, c, err := input(db, name, r)
		if err != nil {
			return err
		}

		if err := card.Create(db, lockedBuf, c); err != nil {
			return err
		}

		fmt.Printf("\n%q added\n", name)
		return nil
	}
}

func input(db *bolt.DB, name string, r io.Reader) (*memguard.LockedBuffer, *pb.Card, error) {
	s := bufio.NewScanner(r)

	lockedBuf, card := pb.SecureCard()
	card.Name = name
	card.Type = cmdutil.Scan(s, "Type")
	card.Number = cmdutil.Scan(s, "Number")
	card.SecurityCode = cmdutil.Scan(s, "Security code")
	card.ExpireDate = cmdutil.Scan(s, "Expire date")
	card.Notes = cmdutil.Scanlns(s, "Notes")

	if err := s.Err(); err != nil {
		return nil, nil, errors.Wrap(err, "failed scanning")
	}

	return lockedBuf, card, nil
}
