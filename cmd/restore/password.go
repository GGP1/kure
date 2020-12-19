package restore

import (
	"bytes"
	"crypto/subtle"
	"fmt"
	"sync"
	"syscall"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/crypt"
	"github.com/GGP1/kure/db/card"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/db/file"
	"github.com/GGP1/kure/db/note"

	"github.com/awnumar/memguard"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	bolt "go.etcd.io/bbolt"
	"golang.org/x/crypto/ssh/terminal"
)

func passwordSubCmd(db *bolt.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "password",
		Short: "Re-encrypt all the records with a new password",
		Long: `Re-encrypt all the records with a new password.
		
Interrupting this process may cause irreversible damage to your information, please do not exit after typing the new password.`,
		RunE: runPassword(db),
	}

	return cmd
}

// There should be no errors, that's why they are omitted when creating or removing elements.
func runPassword(db *bolt.DB) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		if err := oldPassword(); err != nil {
			return err
		}

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
		notes, err := note.List(db)
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
		createCards(db, cards, &wg)
		createEntries(db, entries, &wg)
		createFiles(db, files, &wg)
		createNotes(db, notes, &wg)

		wg.Wait()

		fmt.Print("\nPassword updated successfully")
		return nil
	}
}

func oldPassword() error {
	fmt.Print("Enter old password: ")
	password, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return errors.Wrap(err, "reading password")
	}
	fmt.Print("\n")

	if subtle.ConstantTimeCompare(password, nil) == 1 {
		return errors.New("invalid password")
	}

	pwd := memguard.NewBufferFromBytes(bytes.TrimSpace(password))
	viper.Set("user.password", pwd.Seal())

	return nil
}
