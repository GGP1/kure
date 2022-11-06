package create

import (
	"os"
	"os/exec"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/config"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const example = `
kure config create -p path/to/file`

type createOptions struct {
	path string
}

// NewCmd returns a new command.
func NewCmd() *cobra.Command {
	opts := createOptions{}
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create a configuration file",
		Example: example,
		RunE:    runCreate(&opts),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset variables (session)
			opts = createOptions{}
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.path, "path", "p", "", "destination file path")

	return cmd
}

func runCreate(opts *createOptions) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		if opts.path == "" {
			return cmdutil.ErrInvalidPath
		}

		if err := config.WriteStruct(opts.path); err != nil {
			return err
		}

		editor := cmdutil.SelectEditor()
		bin, err := exec.LookPath(editor)
		if err != nil {
			return errors.Errorf("%q executable not found", editor)
		}

		edit := exec.Command(bin, opts.path)
		edit.Stdin = os.Stdin
		edit.Stdout = os.Stdout

		if err := edit.Run(); err != nil {
			return errors.Wrap(err, "executing command")
		}

		return nil
	}
}
