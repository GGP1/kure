package rm

import (
	"fmt"
	"io"
	"strings"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/card"
	"github.com/GGP1/kure/terminal"

	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

const example = `
* Remove a card
kure card rm Sample

* Remove a directory
kure card rm SampleDir/

* Remove multiple cards
kure card rm Sample Sample2 Sample3`

// NewCmd returns the a new command.
func NewCmd(db *bolt.DB, r io.Reader) *cobra.Command {
	return &cobra.Command{
		Use:     "rm <names>",
		Short:   "Remove cards or directories",
		Example: example,
		Args:    cmdutil.MustExist(db, cmdutil.Card, true),
		RunE:    runRm(db, r),
	}
}

func runRm(db *bolt.DB, r io.Reader) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		if !terminal.Confirm(r, "Are you sure you want to proceed?") {
			return nil
		}

		names := make([]string, 0, len(args))
		for _, name := range args {
			name = cmdutil.NormalizeName(name, true)

			if !strings.HasSuffix(name, "/") {
				names = append(names, name)
				fmt.Println("Remove:", name)
				continue
			}

			cards, err := card.ListNames(db)
			if err != nil {
				return err
			}

			for _, c := range cards {
				if strings.HasPrefix(c, name) {
					names = append(names, c)
					fmt.Println("Remove:", c)
				}
			}
		}

		return card.Remove(db, names...)
	}
}
