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
	"github.com/spf13/viper"
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
	s := bufio.NewScanner(r)
	cType := cmdutil.Scan(s, "Type")
	num := cmdutil.Scan(s, "Number")
	secCode := cmdutil.Scan(s, "Security code")
	expDate := cmdutil.Scan(s, "Expire date")
	notes := cmdutil.Scanlns(s, "Notes")

	if err := s.Err(); err != nil {
		return nil, errors.Wrap(err, "failed scanning")
	}

	// This is the easiest way to avoid creating a card after a signal,
	// however, it may not be the best solution
	if viper.GetBool("interrupt") {
		block := make(chan struct{})
		<-block
	}

	card := &pb.Card{
		Name:         name,
		Type:         cType,
		Number:       num,
		SecurityCode: secCode,
		ExpireDate:   expDate,
		Notes:        notes,
	}

	return card, nil
}
