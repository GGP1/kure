package clear

import (
	"os"
	"os/exec"
	"runtime"

	cmdutil "github.com/GGP1/kure/cmd"

	"github.com/atotto/clipboard"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var both, clip, term bool

var example = `
* Clear both terminal and clipboard
kure clear -b

* Clear terminal
kure clear -t

* Clear clipboard
kure clear -c`

// NewCmd returns a new command.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear clipboard/terminal (and history) or both",
		Long: `Manually clear clipboard, terminal (and its history) or both of them. Kure clears all by default.

Windows users must clear the history manually with ALT+F7, executing "cmd" command 
or by re-opening the cmd (as it saves session history only).`,
		Example: example,
		RunE:    runClear(),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset flags (session)
			both, clip, term = true, false, false
		},
	}

	f := cmd.Flags()
	f.BoolVarP(&both, "both", "b", true, "clear clipboard, terminal and history")
	f.BoolVarP(&clip, "clipboard", "c", false, "clear clipboard")
	f.BoolVarP(&term, "terminal", "t", false, "clear terminal")

	return cmd
}

func runClear() cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		// In case the user wants only one option, set both to false (that is true by default)
		if clip == true || term == true {
			both = false
		}

		if both {
			clip = true
			term = true
		}

		if clip {
			if err := clipboard.WriteAll(""); err != nil {
				return errors.Wrap(err, "failed clearing clipboard")
			}
		}

		if term {
			if runtime.GOOS == "windows" {
				c := exec.Command("cmd", "/c", "cls")
				c.Stdout = os.Stdout
				c.Run()
				return nil
			}

			c := exec.Command("clear")
			c.Stdout = os.Stdout
			c.Run()

			h := exec.Command("/bin/bash", "history -c", "history -cw")
			h.Stdout = os.Stdout
			h.Run()
		}

		return nil
	}
}
