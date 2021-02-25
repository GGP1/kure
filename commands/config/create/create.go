package create

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	cmdutil "github.com/GGP1/kure/commands"

	"github.com/pelletier/go-toml"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var example = `
kure config create -p path/to/file`

type config struct {
	Clipboard struct {
		Timeout string
	}
	Database struct {
		Path string
	}
	Editor  string
	Keyfile struct {
		Path string
	}
	Session struct {
		Prefix  string
		Timeout string
	}
}

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

		cfgContent, err := marshaler(config{}, opts.path)
		if err != nil {
			return err
		}

		// Lower data instead of using tags for each data format.
		if err := os.WriteFile(opts.path, bytes.ToLower(cfgContent), 0600); err != nil {
			return errors.Wrap(err, "writing configuration skeleton")
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

func marshaler(v interface{}, path string) ([]byte, error) {
	ext := filepath.Ext(path)
	if ext == "" || ext == "." {
		return nil, errors.New("invalid file extension")
	}
	format := strings.ToLower(ext[1:])

	switch format {
	case "json":
		return json.MarshalIndent(v, "", "  ")

	case "yaml", "yml":
		return yaml.Marshal(v)

	case "toml":
		return toml.Marshal(v)

	default:
		return nil, errors.Errorf("%q is not supported. Formats supported: json, yaml and toml.", format)
	}
}
