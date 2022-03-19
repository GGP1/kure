package root

import (
	"fmt"
	"os"
	"runtime/debug"

	tfa "github.com/GGP1/kure/commands/2fa"
	"github.com/GGP1/kure/commands/add"
	"github.com/GGP1/kure/commands/backup"
	"github.com/GGP1/kure/commands/card"
	"github.com/GGP1/kure/commands/clear"
	"github.com/GGP1/kure/commands/config"
	"github.com/GGP1/kure/commands/copy"
	"github.com/GGP1/kure/commands/edit"
	"github.com/GGP1/kure/commands/export"
	"github.com/GGP1/kure/commands/file"
	"github.com/GGP1/kure/commands/gen"
	importt "github.com/GGP1/kure/commands/import"
	"github.com/GGP1/kure/commands/it"
	"github.com/GGP1/kure/commands/ls"
	"github.com/GGP1/kure/commands/restore"
	"github.com/GGP1/kure/commands/rm"
	"github.com/GGP1/kure/commands/session"
	"github.com/GGP1/kure/commands/stats"

	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var (
	version bool
	cmd     = &cobra.Command{
		Use:           "kure",
		Short:         "Kure ~ CLI password manager",
		SilenceErrors: true,
		SilenceUsage:  true,
		CompletionOptions: cobra.CompletionOptions{
			HiddenDefaultCmd: true,
		},
		Run: func(cmd *cobra.Command, args []string) {
			if version {
				printVersion()
				return
			}
			_ = cmd.Usage()
		},
	}
)

// DevCmd returns the root command with all its sub commands and without a database object.
//
// It should be used for documentation or testing purposes only.
func DevCmd() *cobra.Command {
	registerCmds(nil)
	return cmd
}

// Execute adds all the subcommands to the root and executes it.
func Execute(db *bolt.DB) error {
	cmd.Flags().BoolVarP(&version, "version", "v", false, "version for kure")
	registerCmds(db)

	return cmd.Execute()
}

// registerCmds adds all the commands to the root.
func registerCmds(db *bolt.DB) {
	cmd.AddCommand(tfa.NewCmd(db))
	cmd.AddCommand(add.NewCmd(db, os.Stdin))
	cmd.AddCommand(backup.NewCmd(db))
	cmd.AddCommand(card.NewCmd(db))
	cmd.AddCommand(clear.NewCmd())
	cmd.AddCommand(config.NewCmd(db, os.Stdin))
	cmd.AddCommand(copy.NewCmd(db))
	cmd.AddCommand(edit.NewCmd(db))
	cmd.AddCommand(export.NewCmd(db))
	cmd.AddCommand(file.NewCmd(db))
	cmd.AddCommand(gen.NewCmd())
	cmd.AddCommand(importt.NewCmd(db))
	cmd.AddCommand(it.NewCmd(db))
	cmd.AddCommand(ls.NewCmd(db))
	cmd.AddCommand(restore.NewCmd(db))
	cmd.AddCommand(rm.NewCmd(db, os.Stdin))
	cmd.AddCommand(session.NewCmd(db, os.Stdin))
	cmd.AddCommand(stats.NewCmd(db))
}

func printVersion() {
	bi, _ := debug.ReadBuildInfo()

	var lastCommitHash string
	for _, setting := range bi.Settings {
		if setting.Key == "vcs.revision" {
			lastCommitHash = setting.Value
			break
		}
	}

	fmt.Printf("%s (%s) - [%s]\n", bi.Main.Version, lastCommitHash, bi.GoVersion)
}
