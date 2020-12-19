package restore

import (
	"fmt"
	"runtime"
	"sync"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/db/card"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/db/file"
	"github.com/GGP1/kure/db/note"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	bolt "go.etcd.io/bbolt"
)

var (
	iterations, memory uint32
	threads            uint8
)

var example = `
kure restore argon2 -m 50000 -i 2 -t 4`

func argon2SubCmd(db *bolt.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "argon2",
		Short: "Re-encrypt all the information with new argon2 parameters",
		Long: `Re-encrypt all the records with a new password.
		
Interrupting this process may cause irreversible damage to your information, please do not exit after typing the new password.`,
		Example: example,
		PreRunE: cmdutil.RequirePassword(db),
		RunE:    runArgon2(db),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset flags (session)
			iterations, memory = 1, 1048576
			threads = uint8(runtime.NumCPU())
		},
	}

	f := cmd.Flags()
	f.Uint32VarP(&iterations, "iterations", "i", 1, "number of passes over the memory")
	f.Uint32VarP(&memory, "memory", "m", 1048576, "amount of memory allowed for argon2 to use")
	f.Uint8VarP(&threads, "threads", "t", uint8(runtime.NumCPU()), "number of threads running in parallel")

	return cmd
}

// There should be no errors, that's why they are omitted when creating or removing elements.
func runArgon2(db *bolt.DB) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		if iterations < 1 || memory < 1 {
			return errors.New("iterations and memory should be higher than 0")
		}
		if threads < 1 {
			return errors.New("the number of threads must be higher than 0")
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

		if err := updateConfig(); err != nil {
			return err
		}

		var wg sync.WaitGroup
		wg.Add(len(cards) + len(entries) + len(files) + len(notes))

		// Overwrite objects encrypted with the new configuration
		// Errors are omitted as there shouldn't be any
		createCards(db, cards, &wg)
		createEntries(db, entries, &wg)
		createFiles(db, files, &wg)
		createNotes(db, notes, &wg)

		wg.Wait()

		fmt.Print("\nArgon2 parameters updated successfully")
		return nil
	}
}

func updateConfig() error {
	path, err := cmdutil.GetConfigPath()
	if err != nil {
		return err
	}

	if iterations != 0 {
		viper.Set("argon2.iterations", iterations)
	}
	if memory != 0 {
		viper.Set("argon2.memory", memory)
	}
	if threads != 0 {
		viper.Set("argon2.threads", threads)
	}

	// Unset password so the field it's not stored in the config file
	password := viper.Get("user.password")
	viper.Set("user.password", nil)

	if err := viper.WriteConfigAs(path); err != nil {
		return errors.Wrap(err, "failed writing config")
	}

	// Reset password to its previous value
	viper.Set("user.password", password)
	viper.SetConfigFile(path)

	return nil
}
