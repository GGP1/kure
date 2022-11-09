package edit

import (
	"os"
	"os/exec"

	"github.com/GGP1/kure/auth"
	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/config"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

const example = `
kure config edit`

// NewCmd returns a new command.
func NewCmd(db *bolt.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "edit",
		Short:   "Edit the current configuration file",
		Example: example,
		PreRunE: auth.Login(db),
		RunE:    runEdit(),
	}

	return cmd
}

func runEdit() cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		path := config.Filename()

		f, err := os.OpenFile(path, os.O_RDWR, 0600)
		if err != nil {
			return errors.Wrap(err, "opening configuration file")
		}
		defer f.Close()

		editor := cmdutil.SelectEditor()
		bin, err := exec.LookPath(editor)
		if err != nil {
			return errors.Errorf("%q executable not found", editor)
		}

		edit := exec.Command(bin, path)
		edit.Stdin = os.Stdin
		edit.Stdout = os.Stdout

		if err := edit.Run(); err != nil {
			return errors.Wrap(err, "running edit command")
		}

		return nil
	}
}
