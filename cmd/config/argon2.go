package config

import (
	"fmt"
	"runtime"

	cmdutil "github.com/GGP1/kure/cmd"

	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var argon2Example = `
kure config argon2`

func argon2SubCmd(db *bolt.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "argon2",
		Short:   "Display the currently used argon2 parameters",
		Example: argon2Example,
		RunE:    runArgon2(db),
	}

	cmd.AddCommand(testSubCmd())

	return cmd
}

func runArgon2(db *bolt.DB) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		return db.View(func(tx *bolt.Tx) error {
			// kure_argon2 bucket will be created only if the user used "restore argon2" command at least once
			b := tx.Bucket([]byte("kure_argon2"))
			if b == nil {
				// Show defaults
				fmt.Println("Iterations: 1\nMemory: 1048576\nThreads:", runtime.NumCPU())
				return nil
			}

			iterations := b.Get([]byte("iterations"))
			memory := b.Get([]byte("memory"))
			threads := b.Get([]byte("threads"))

			fmt.Printf("Iterations: %s\nMemory: %s\nThreads: %s\n", iterations, memory, threads)
			return nil
		})
	}
}
