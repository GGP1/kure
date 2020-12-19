package root

import (
	"fmt"
	"os"

	"github.com/awnumar/memguard"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var rootCmd = &cobra.Command{
	Use:           "kure",
	Short:         "CLI password manager.",
	SilenceErrors: true,
	SilenceUsage:  true,
}

// Cmd returns the root command.
func Cmd() *cobra.Command {
	return rootCmd
}

// Execute sets each sub command flag and adds it to the root.
func Execute(db *bolt.DB) {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		db.Close()
		memguard.SafeExit(1)
	}
}

// Register adds a new command to the root.
func Register(cmd *cobra.Command) {
	rootCmd.AddCommand(cmd)
}
