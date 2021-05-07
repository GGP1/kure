package restore

import (
	"fmt"
	"os"

	"github.com/GGP1/kure/auth"
	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/card"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/db/file"
	"github.com/GGP1/kure/db/totp"

	"github.com/pkg/errors"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

// NewCmd returns a new command.
func NewCmd(db *bolt.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore the database using new credentials",
		Long: `Restore the database using new credentials.

Overwrite the registered credentials and re-encrypt every record with the new ones.`,
		PreRunE: auth.Login(db),
		RunE:    runRestore(db),
	}

	return cmd
}

func runRestore(db *bolt.DB) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		// Operations are run synchronously to avoid exhausting the user's pc
		listBar := progressbar.New64(4)
		listBar.Describe("Loading records")
		// List all records with the old credentials
		cards, err := card.List(db)
		if err != nil {
			return err
		}
		listBar.Add64(1)
		entries, err := entry.List(db)
		if err != nil {
			return err
		}
		listBar.Add64(1)
		files, err := file.List(db)
		if err != nil {
			return err
		}
		listBar.Add64(1)
		totps, err := totp.List(db)
		if err != nil {
			return err
		}
		listBar.Add64(1)
		fmt.Print("\n")

		// Initialize registration and re-encrypt the records with the new credentials
		if err := auth.Register(db, os.Stdin); err != nil {
			return err
		}

		var errs []error
		createBar := progressbar.New(len(cards) + len(entries) + len(files) + len(totps))
		createBar.Describe("Re-encrypting records")

		for _, c := range cards {
			if err := card.Create(db, c); err != nil {
				errs = append(errs, errors.Wrap(err, c.Name))
			}
			createBar.Add64(1)
		}
		for _, e := range entries {
			if err := entry.Create(db, e); err != nil {
				errs = append(errs, errors.Wrap(err, e.Name))
			}
			createBar.Add64(1)
		}
		for _, f := range files {
			if err := file.Create(db, f); err != nil {
				errs = append(errs, errors.Wrap(err, f.Name))
			}
			createBar.Add64(1)
		}
		for _, t := range totps {
			if err := totp.Create(db, t); err != nil {
				errs = append(errs, errors.Wrap(err, t.Name))
			}
			createBar.Add64(1)
		}

		for _, err := range errs {
			fmt.Fprintln(os.Stderr, "error:", err)
		}

		return nil
	}
}
