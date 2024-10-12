package root

import (
	"fmt"
	"os"
	"runtime/debug"

	cmdutil "github.com/GGP1/kure/commands"
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
	"github.com/GGP1/kure/commands/rotate"
	"github.com/GGP1/kure/commands/session"
	"github.com/GGP1/kure/commands/stats"

	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var statelessCommands = map[string]struct{}{
	"clear":     {},
	"gen":       {},
	"help":      {},
	"-v":        {},
	"--version": {},
}

type rootOptions struct {
	version bool
}

// NewCmd returns a new command.
func NewCmd(db *bolt.DB) *cobra.Command {
	opts := rootOptions{}
	cmd := &cobra.Command{
		Use:           "kure",
		Short:         "kure ~ CLI password manager with sessions",
		SilenceErrors: true,
		SilenceUsage:  true,
		CompletionOptions: cobra.CompletionOptions{
			HiddenDefaultCmd: true,
		},
		RunE: runRoot(&opts),
	}

	cmd.Flags().BoolVarP(&opts.version, "version", "v", false, "display kure version")
	cmd.AddCommand(
		tfa.NewCmd(db),
		add.NewCmd(db, os.Stdin),
		backup.NewCmd(db),
		card.NewCmd(db),
		clear.NewCmd(),
		config.NewCmd(db),
		copy.NewCmd(db),
		edit.NewCmd(db),
		export.NewCmd(db),
		file.NewCmd(db),
		gen.NewCmd(),
		importt.NewCmd(db),
		it.NewCmd(db),
		ls.NewCmd(db),
		restore.NewCmd(db),
		rotate.NewCmd(db),
		rm.NewCmd(db, os.Stdin),
		session.NewCmd(db, os.Stdin),
		stats.NewCmd(db),
	)

	return cmd
}

func runRoot(opts *rootOptions) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		if opts.version {
			printVersion()
			return nil
		}

		_ = cmd.Usage()
		return nil
	}
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

	fmt.Printf("[%s] %s %s\n", bi.GoVersion, bi.Main.Version, lastCommitHash)
}

// IsStatelessCommand returns true if the specified command does not require opening the database.
func IsStatelessCommand(command string) bool {
	_, ok := statelessCommands[command]
	return ok
}
