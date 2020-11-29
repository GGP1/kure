package card

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/db/card"
	"github.com/GGP1/kure/pb"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var addExample = `
* Add a new card
kure card add cardName`

// addSubCmd returns the add subcommand.
func addSubCmd(db *bolt.DB, r io.Reader) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "add <name>",
		Short:         "Add a card to the database",
		Aliases:       []string{"a", "new", "create"},
		Example:       addExample,
		RunE:          runAdd(db, r),
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	return cmd
}

func runAdd(db *bolt.DB, r io.Reader) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")
		if name == "" {
			return errInvalidName
		}
		if !strings.Contains(name, "/") && len(name) > 57 {
			return errors.New("card name must contain 57 letters or less")
		}

		if err := cmdutil.RequirePassword(db); err != nil {
			return err
		}

		c, err := cardInput(db, name, r)
		if err != nil {
			return errors.Wrap(err, "error")
		}

		if err := card.Create(db, c); err != nil {
			return errors.Wrap(err, "error")
		}

		fmt.Printf("\nSuccessfully created %q card.\n", name)
		return nil
	}
}

func cardInput(db *bolt.DB, name string, r io.Reader) (*pb.Card, error) {
	name = strings.ToLower(name)

	if err := exists(db, name); err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(r)
	cType := cmdutil.Scan(scanner, "Type")
	num := cmdutil.Scan(scanner, "Number")
	CVC := cmdutil.Scan(scanner, "CVC")
	expDate := cmdutil.Scan(scanner, "Expire date")

	card := &pb.Card{
		Name:       name,
		Type:       cType,
		Number:     num,
		CVC:        CVC,
		ExpireDate: expDate,
	}

	return card, nil
}

// exists compares "path" with the existing card names that contain "path",
// split both name and each of the card names and look for matches on the same level.
//
// Given a path "Naboo/Padmé" and comparing it with "Naboo/Padmé Amidala":
//
// "Padmé" != "Padmé Amidala", skip.
//
// Given a path "jedi/Yoda" and comparing it with "jedi/Obi-Wan Kenobi":
//
// "jedi/Obi-Wan Kenobi" does not contain "jedi/Yoda", skip.
func exists(db *bolt.DB, path string) error {
	cards, err := card.ListNames(db)
	if err != nil {
		return err
	}

	parts := strings.Split(path, "/")
	n := len(parts) - 1 // card name index
	name := parts[n]    // name without folders

	for _, c := range cards {
		if strings.Contains(c.Name, path) {
			cardName := strings.Split(c.Name, "/")[n]

			if cardName == name {
				return errors.Errorf("already exists a card or folder named %q", path)
			}
		}
	}

	return nil
}
