package ls

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/GGP1/kure/auth"
	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/card"
	"github.com/GGP1/kure/orderedmap"
	"github.com/GGP1/kure/pb"
	"github.com/GGP1/kure/tree"

	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var example = `
* List one, show sensitive information and QR code
kure card ls Sample -s -q

* Filter by name
kure card ls Sample -f

* List all
kure card ls`

type lsOptions struct {
	filter, qr, show bool
}

// NewCmd returns a new command.
func NewCmd(db *bolt.DB) *cobra.Command {
	opts := lsOptions{}

	cmd := &cobra.Command{
		Use:     "ls <name>",
		Short:   "List cards",
		Example: example,
		Args:    cmdutil.MustExistLs(db, cmdutil.Card),
		PreRunE: auth.Login(db),
		RunE:    runLs(db, &opts),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset variables (session)
			opts = lsOptions{}
		},
	}

	f := cmd.Flags()
	f.BoolVarP(&opts.filter, "filter", "f", false, "filter by name")
	f.BoolVarP(&opts.qr, "qr", "q", false, "show the number QR code on the terminal")
	f.BoolVarP(&opts.show, "show", "s", false, "show card number and security code")

	return cmd
}

func runLs(db *bolt.DB, opts *lsOptions) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")
		name = cmdutil.NormalizeName(name)

		// List all
		if name == "" {
			cards, err := card.ListNames(db)
			if err != nil {
				return err
			}

			tree.Print(cards)
			return nil
		}

		// Filter by name
		if opts.filter {
			cards, err := card.ListNames(db)
			if err != nil {
				return err
			}

			var filtered []string
			for _, card := range cards {
				matched, err := filepath.Match(name, card)
				if err != nil {
					return err
				}

				if matched {
					filtered = append(filtered, card)
				}
			}

			if len(filtered) == 0 {
				return errors.New("no cards were found")
			}

			tree.Print(filtered)
			return nil
		}

		// List one
		c, err := card.Get(db, name)
		if err != nil {
			return err
		}

		if opts.qr {
			if err := cmdutil.DisplayQRCode(c.Number); err != nil {
				return err
			}
		}

		printCard(name, c, opts.show)
		return nil
	}
}

func printCard(name string, c *pb.Card, show bool) {
	if !show {
		c.Number = "••••••••••••••••"
		c.SecurityCode = "•••"
	}

	mp := orderedmap.New()
	mp.Set("Type", c.Type)
	mp.Set("Number", c.Number)
	mp.Set("Security code", c.SecurityCode)
	mp.Set("Expire date", c.ExpireDate)
	mp.Set("Notes", c.Notes)

	box := cmdutil.BuildBox(name, mp)
	fmt.Println("\n" + box)
}
