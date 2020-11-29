package card

import (
	"fmt"
	"strings"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/db/card"
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
		Use:     "ls <name> [-f filter] [-H hide]",
		Short:   "List cards",
		Example: lsExample,
		RunE:    runLs(db),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset flags defaults (session)
			filter, hide = false, false
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	f := cmd.Flags()
	f.BoolVarP(&filter, "filter", "f", false, "filter cards")
	f.BoolVarP(&hide, "hide", "H", false, "hide card CVC")

	return cmd
}

func runLs(db *bolt.DB) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")

		if name == "" {
			cards, err := card.ListNames(db)
			if err != nil {
				return errors.Wrap(err, "error")
			}

			paths := make([]string, len(cards))

			for i, card := range cards {
				paths[i] = card.Name
			}

			tree.Print(paths)
			return nil
		}

		if filter {
			cards, err := card.ListByName(db, name)
			if err != nil {
				return errors.Wrap(err, "error")
			}

			if len(cards) == 0 {
				return errors.New("error: no cards were found")
			}

			for _, card := range cards {
				printCard(card)
			}
			return nil
		}

		card, err := card.Get(db, name)
		if err != nil {
			return errors.Wrap(err, "error")
		}

		printCard(card)
		return nil
	}
}

func printCard(c *pb.Card) {
	cmdutil.PrintObjectName(c.Name)

	if hide {
		c.CVC = "••••"
	}

	fmt.Printf(`│ Type        │ %s
│ Number      │ %s
│ CVC         │ %s
│ Expire Date │ %s
`, c.Type, c.Number, c.CVC, c.ExpireDate)

	fmt.Println("└─────────────+───────────────────────────────────────────>")
}
