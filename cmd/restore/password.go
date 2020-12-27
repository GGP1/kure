package restore

import (
	"fmt"
	"sync"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/crypt"
	"github.com/GGP1/kure/db/card"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/db/file"
	"github.com/GGP1/kure/db/note"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	bolt "go.etcd.io/bbolt"
)

func passwordSubCmd(db *bolt.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "password",
		Short: "Re-encrypt all the records with a new password",
		Long: `Re-encrypt all the records with a new password.
		
Interrupting this process may cause irreversible damage to your information, please do not exit after typing the new password.`,
		PreRunE: cmdutil.RequirePassword(db),
		RunE:    runPassword(db),
	}

	return cmd
}

// There should be no errors, that's why they are omitted when creating or removing elements.
func runPassword(db *bolt.DB) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		cardsBuf, cards, err := card.List(db)
		if err != nil {
			return err
		}
		entriesBuf, entries, err := entry.List(db)
		if err != nil {
			return err
		}
		filesBuf, files, err := file.List(db)
		if err != nil {
			return err
		}
		notesBuf, notes, err := note.List(db)
		if err != nil {
			return err
		}

		fmt.Println("+──────── NEW ────────+")
		newPassword, err := crypt.AskPassword(true)
		if err != nil {
			return err
		}

		// Create them with the new password
		viper.Set("user.password", newPassword)

		var wg sync.WaitGroup
		wg.Add(len(cards) + len(entries) + len(files) + len(notes))

		// Overwrite objects encrypted with the new password
		// Errors are omitted as there shouldn't be any
		createCards(db, cardsBuf, cards, &wg)
		createEntries(db, entriesBuf, entries, &wg)
		createFiles(db, filesBuf, files, &wg)
		createNotes(db, notesBuf, notes, &wg)

		wg.Wait()

		fmt.Print("\nPassword updated successfully")
		return nil
	}
}
