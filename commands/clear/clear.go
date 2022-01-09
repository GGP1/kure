package clear

import (
	"os"
	"os/exec"
	"runtime"

	cmdutil "github.com/GGP1/kure/commands"

	"github.com/atotto/clipboard"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const example = `
* Clear both terminal and clipboard
kure clear

* Clear terminal
kure clear -t

* Clear clipboard
kure clear -c`

type clearOptions struct {
	clip, term bool
}

// NewCmd returns a new command.
func NewCmd() *cobra.Command {
	opts := clearOptions{}

	cmd := &cobra.Command{
		Use:     "clear",
		Short:   "Clear clipboard, terminal or both",
		Example: example,
		RunE:    runClear(&opts),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset variables (session)
			opts = clearOptions{}
		},
	}

	f := cmd.Flags()
	f.BoolVarP(&opts.clip, "clipboard", "c", false, "clear clipboard")
	f.BoolVarP(&opts.term, "terminal", "t", false, "clear terminal")

	return cmd
}

func runClear(opts *clearOptions) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		// If no flag is specified, clear both
		if opts.clip == false && opts.term == false {
			opts.clip = true
			opts.term = true
		}

		if opts.clip {
			if err := clipboard.WriteAll(""); err != nil {
				return errors.Wrap(err, "clearing clipboard")
			}
		}

		if opts.term {
			if runtime.GOOS == "windows" {
				c := exec.Command("cmd", "/c", "cls")
				c.Stdout = os.Stdout
				if err := c.Run(); err != nil {
					return errors.Wrap(err, "clearing terminal")
				}
				return nil
			}

			c := exec.Command("clear")
			c.Stdout = os.Stdout
			if err := c.Run(); err != nil {
				return errors.Wrap(err, "clearing terminal")
			}
		}

		return nil
	}
}
