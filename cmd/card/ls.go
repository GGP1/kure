package card

import (
	"fmt"
	"strings"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/db/card"
	"github.com/GGP1/kure/orderedmap"
	"github.com/GGP1/kure/pb"
	"github.com/GGP1/kure/tree"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var filter, hide bool

var lsExample = `
* List one and hide sensible information (optional)
kure card ls cardName -H

* Filter by name
kure card ls cardName -f

* List all
kure card ls`

// lsSubCmd returns the copy subcommand
func lsSubCmd(db *bolt.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ls <name>",
		Short:   "List cards",
		Example: lsExample,
		PreRunE: cmdutil.RequirePassword(db),
		RunE:    runLs(db),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset flags (session)
			filter, hide = false, false
		},
	}

	f := cmd.Flags()
	f.BoolVarP(&filter, "filter", "f", false, "filter cards")
	f.BoolVarP(&hide, "hide", "H", false, "hide card security code")

	return cmd
}

func runLs(db *bolt.DB) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")

		switch name {
		case "":
			cards, err := card.ListNames(db)
			if err != nil {
				return err
			}
			tree.Print(cards)

		default:
			if filter {
				cards, err := card.ListNames(db)
				if err != nil {
					return err
				}

				var list []string
				for _, card := range cards {
					if strings.Contains(card, name) {
						list = append(list, card)
					}
				}

				if len(list) == 0 {
					return errors.New("no cards were found")
				}

				tree.Print(list)
				break
			}

			lockedBuf, card, err := card.Get(db, name)
			if err != nil {
				return err
			}

			printCard(name, card)
			lockedBuf.Destroy()
		}

		return nil
	}
}

func printCard(name string, c *pb.Card) {
	if hide {
		c.SecurityCode = "••••"
	}

	// Map's key/value  pairs are stored inside locked buffers
	oMap := orderedmap.New(5)
	oMap.Set("Type", c.Type)
	oMap.Set("Number", c.Number)
	oMap.Set("Security code", c.SecurityCode)
	oMap.Set("Expire date", c.ExpireDate)
	oMap.Set("Notes", c.Notes)

	lockedBuf, box := cmdutil.BuildBox(name, oMap)
	fmt.Println("\n" + box)
	lockedBuf.Destroy()
}
