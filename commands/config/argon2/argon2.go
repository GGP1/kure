package argon2

import (
	"fmt"

	"github.com/GGP1/kure/auth"
	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/commands/config/argon2/test"
	authDB "github.com/GGP1/kure/db/auth"

	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var argon2Example = `
kure config argon2`

// NewCmd returns a new command.
func NewCmd(db *bolt.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "argon2",
		Short:   "Display currently used argon2 parameters",
		Aliases: []string{"argon"},
		Example: argon2Example,
		PreRunE: auth.Login(db),
		RunE:    runArgon2(db),
	}

	cmd.AddCommand(test.NewCmd())

	return cmd
}

func runArgon2(db *bolt.DB) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		params, err := authDB.GetParameters(db)
		if err != nil {
			return err
		}

		fmt.Printf("Iterations: %d\nMemory: %d\nThreads: %d\n",
			params.Iterations, params.Memory, params.Threads)
		return nil
	}
}
