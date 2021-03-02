package add

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/GGP1/kure/auth"
	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/card"
	"github.com/GGP1/kure/pb"

	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var example = `
* Add a new card
kure card add Sample`

// NewCmd returns a new command.
func NewCmd(db *bolt.DB, r io.Reader) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add <name>",
		Short:   "Add a card",
		Aliases: []string{"create", "new"},
		Example: example,
		Args:    cmdutil.MustNotExist(db, cmdutil.Card),
		PreRunE: auth.Login(db),
		RunE:    runAdd(db, r),
	}

	return cmd
}

func runAdd(db *bolt.DB, r io.Reader) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")
		name = cmdutil.NormalizeName(name)

		c, err := input(db, name, r)
		if err != nil {
			return err
		}

		if err := card.Create(db, c); err != nil {
			return err
		}

		fmt.Printf("\n%q added\n", name)
		return nil
	}
}

func input(db *bolt.DB, name string, r io.Reader) (*pb.Card, error) {
	reader := bufio.NewReader(r)
	c := &pb.Card{
		Name:         name,
		Type:         cmdutil.Scanln(reader, "Type"),
		Number:       cmdutil.Scanln(reader, "Number"),
		SecurityCode: cmdutil.Scanln(reader, "Security code"),
		ExpireDate:   cmdutil.Scanln(reader, "Expire date"),
		Notes:        cmdutil.Scanlns(reader, "Notes"),
	}

	return c, nil
}
