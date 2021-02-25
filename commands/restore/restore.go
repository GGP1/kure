package restore

import (
	"fmt"
	"os"
	"sync"

	"github.com/GGP1/kure/auth"
	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/card"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/db/file"
	"github.com/GGP1/kure/db/totp"
	"github.com/GGP1/kure/pb"

	"github.com/pkg/errors"
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
		// List all records with the old credentials
		cards, err := card.List(db)
		if err != nil {
			return err
		}
		entries, err := entry.List(db)
		if err != nil {
			return err
		}
		files, err := file.List(db)
		if err != nil {
			return err
		}
		totps, err := totp.List(db)
		if err != nil {
			return err
		}

		fmt.Println("\n────────── NEW CREDENTIALS ──────────")

		// Initialize registration and re-encrypt the records with the new credentials
		if err := auth.Register(db, os.Stdin); err != nil {
			return err
		}

		fmt.Println("Restoring database, this may take a few minutes...")

		var (
			wg   sync.WaitGroup
			errs []error
		)
		wg.Add(len(cards) + len(entries) + len(files) + len(totps))

		for _, c := range cards {
			go func(c *pb.Card) {
				if err := card.Create(db, c); err != nil {
					errs = append(errs, errors.Wrap(err, c.Name))
				}
				wg.Done()
			}(c)
		}
		for _, e := range entries {
			go func(e *pb.Entry) {
				if err := entry.Create(db, e); err != nil {
					errs = append(errs, errors.Wrap(err, e.Name))
				}
				wg.Done()
			}(e)
		}
		for _, f := range files {
			go func(f *pb.File) {
				if err := file.Create(db, f); err != nil {
					errs = append(errs, errors.Wrap(err, f.Name))
				}
				wg.Done()
			}(f)
		}
		for _, t := range totps {
			go func(t *pb.TOTP) {
				if err := totp.Create(db, t); err != nil {
					errs = append(errs, errors.Wrap(err, t.Name))
				}
				wg.Done()
			}(t)
		}

		wg.Wait()

		for _, err := range errs {
			fmt.Fprintln(os.Stderr, "error:", err)
		}

		fmt.Println("\nDatabase restored successfully")
		return nil
	}
}
