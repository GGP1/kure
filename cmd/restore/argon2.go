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

		if err := updateConfig(db); err != nil {
			return err
		}

		var wg sync.WaitGroup
		wg.Add(len(cards) + len(entries) + len(files) + len(notes))

		// Overwrite objects encrypted with the new configuration
		// Errors are omitted as there shouldn't be any
		createCards(db, cardsBuf, cards, &wg)
		createEntries(db, entriesBuf, entries, &wg)
		createFiles(db, filesBuf, files, &wg)
		createNotes(db, notesBuf, notes, &wg)

		wg.Wait()

		fmt.Printf(`
Argon2 parameters updated successfully
Iterations: %d
Memory: %d
Threads: %d
`, iterations, memory, threads)
		return nil
	}
}

func updateConfig(db *bolt.DB) error {
	// Update the argon2 parameters being used
	err := db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("kure_argon2"))
		if err != nil {
			return errors.Wrap(err, "failed creating bucket")
		}

		if err := b.Put([]byte("iterations"), []byte(fmt.Sprintf("%d", iterations))); err != nil {
			return err
		}
		if err := b.Put([]byte("memory"), []byte(fmt.Sprintf("%d", memory))); err != nil {
			return err
		}
		if err := b.Put([]byte("threads"), []byte(fmt.Sprintf("%d", threads))); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	path, err := cmdutil.GetConfigPath()
	if err != nil {
		return err
	}

	viper.Set("argon2.iterations", iterations)
	viper.Set("argon2.memory", memory)
	viper.Set("argon2.threads", threads)

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
