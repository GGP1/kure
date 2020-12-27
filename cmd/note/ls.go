package note

import (
	"fmt"
	"strings"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/db/note"
	"github.com/GGP1/kure/orderedmap"
	"github.com/GGP1/kure/pb"
	"github.com/GGP1/kure/tree"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var filter bool

var lsExample = `
* List one and hide sensible information (optional)
kure note ls noteName -H

* Filter by name
kure note ls noteName -f

* List all
kure note ls`

// lsSubCmd returns the copy subcommand
func lsSubCmd(db *bolt.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ls <name>",
		Short:   "List notes",
		Example: lsExample,
		PreRunE: cmdutil.RequirePassword(db),
		RunE:    runLs(db),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset flags (session)
			filter = false
		},
	}

	f := cmd.Flags()
	f.BoolVarP(&filter, "filter", "f", false, "filter notes")

	return cmd
}

func runLs(db *bolt.DB) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")

		switch name {
		case "":
			notes, err := note.ListNames(db)
			if err != nil {
				return err
			}
			tree.Print(notes)

		default:
			if filter {
				notes, err := note.ListNames(db)
				if err != nil {
					return err
				}

				var list []string
				for _, note := range notes {
					if strings.Contains(note, name) {
						list = append(list, note)
					}
				}

				if len(list) == 0 {
					return errors.New("no notes were found")
				}

				tree.Print(list)
				break
			}

			noteBuf, note, err := note.Get(db, name)
			if err != nil {
				return err
			}

			printNote(name, note)
			noteBuf.Destroy()
		}

		return nil
	}
}

func printNote(name string, n *pb.Note) {
	// Map's key/value pairs are stored inside locked buffers
	oMap := orderedmap.New(1)
	oMap.Set("Text", n.Text)

	lockedBuf, box := cmdutil.BuildBox(name, oMap)
	fmt.Println("\n" + box)
	lockedBuf.Destroy()
}
