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
* Clear terminal and clipboard
kure clear

* Clear clipboard
kure clear -c

* Clear terminal screen
kure clear -t

* Clear kure commands from terminal history
kure clear -h`

type clearOptions struct {
	clip, term, hist bool
}

// NewCmd returns a new command.
func NewCmd() *cobra.Command {
	opts := clearOptions{}

	cmd := &cobra.Command{
		Use:     "clear",
		Short:   "Clear clipboard, terminal screen or history",
		Example: example,
		RunE:    runClear(&opts),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset variables (session)
			opts = clearOptions{}
		},
	}

	f := cmd.Flags()
	f.BoolVarP(&opts.clip, "clipboard", "c", false, "clear clipboard")
	f.BoolVarP(&opts.term, "terminal", "t", false, "clear terminal screen")
	f.BoolVarP(&opts.hist, "history", "H", false, "clear kure commands from terminal history")

	return cmd
}

func runClear(opts *clearOptions) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		// If no flags were specified, clear clipboard and terminal
		if !opts.clip && !opts.term && !opts.hist {
			opts.clip = true
			opts.term = true
		}

		if opts.clip {
			if err := clipboard.WriteAll(""); err != nil {
				return errors.Wrap(err, "clearing clipboard")
			}
		}

		if runtime.GOOS == "windows" {
			return clearWindowsTerminal(opts)
		}

		return clearUnixTerminal(opts)
	}
}

func clearWindowsTerminal(opts *clearOptions) error {
	if opts.term {
		c := exec.Command("cmd", "/c", "cls")
		c.Stdout = os.Stdout
		if err := c.Run(); err != nil {
			return errors.Wrap(err, "clearing terminal")
		}
	}

	if opts.hist {
		clearPSHist := "Set-Content -Path (Get-PSReadLineOption).HistorySavePath -Value (Get-Content -Path (Get-PSReadLineOption).HistorySavePath | Select-String -Pattern '^kure' -NotMatch)"
		if err := exec.Command("powershell", clearPSHist).Run(); err != nil {
			return errors.Wrap(err, "clearing kure commands from history file")
		}
	}

	return nil
}

func clearUnixTerminal(opts *clearOptions) error {
	if opts.term {
		c := exec.Command("clear")
		c.Stdout = os.Stdout
		if err := c.Run(); err != nil {
			return errors.Wrap(err, "clearing terminal")
		}
	}

	if opts.hist {
		if err := exec.Command("history", "-a").Run(); err != nil {
			return errors.Wrap(err, "flushing session commands to terminal history")
		}

		histFile, ok := os.LookupEnv("HISTFILE")
		if !ok {
			histFile = "~./bash_history"
		}
		if err := exec.Command("sed", "-i", "/^kure/d", histFile).Run(); err != nil {
			return errors.Wrap(err, "clearing kure commands from history file")
		}
	}

	return nil
}
