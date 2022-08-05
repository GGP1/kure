package config

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/GGP1/kure/auth"
	cmdutil "github.com/GGP1/kure/commands"
	argon2cmd "github.com/GGP1/kure/commands/config/argon2"
	"github.com/GGP1/kure/commands/config/create"
	"github.com/GGP1/kure/commands/config/edit"
	"github.com/GGP1/kure/config"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

const example = `
* Read configuration file
kure config`

// NewCmd returns a new command.
func NewCmd(db *bolt.DB, r io.Reader) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "config",
		Short:   "Read the configuration file",
		Aliases: []string{"cfg"},
		Example: example,
		PreRunE: auth.Login(db),
		RunE:    runConfig(r),
	}

	cmd.AddCommand(argon2cmd.NewCmd(db), create.NewCmd(), edit.NewCmd(db))

	return cmd
}

func runConfig(r io.Reader) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		path := config.Filename()
		data, err := os.ReadFile(path)
		if err != nil {
			return errors.Wrap(err, "reading configuration file")
		}

		content := strings.TrimSpace(string(data))
		fmt.Printf(`
File location: %s
		
%s
`, path, content)

		return nil
	}
}
