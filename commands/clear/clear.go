package clear

import (
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"runtime"
	"strings"

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
	f.BoolVarP(&opts.hist, "history", "H", false, "remove kure commands from terminal history")

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
			return clearTerminalWindows(opts)
		}

		return clearTerminalUnix(opts)
	}
}

func clearTerminalWindows(opts *clearOptions) error {
	if opts.term {
		c := exec.Command("cmd", "/c", "cls")
		c.Stdout = os.Stdout
		if err := c.Run(); err != nil {
			return errors.Wrap(err, "clearing terminal")
		}
	}

	if opts.hist {
		output, err := exec.Command("powershell", "(Get-PSReadLineOption).HistorySavePath").Output()
		if err != nil {
			return errors.Wrap(err, "getting powershell history file path")
		}
		path := strings.TrimRight(string(output), "\r\n")

		if err := clearHistoryFile(path); err != nil {
			return errors.Wrap(err, "clearing terminal history")
		}
	}

	return nil
}

func clearTerminalUnix(opts *clearOptions) error {
	if opts.term {
		c := exec.Command("clear")
		c.Stdout = os.Stdout
		if err := c.Run(); err != nil {
			return errors.Wrap(err, "clearing terminal")
		}
	}

	if opts.hist {
		if history, err := exec.LookPath("history"); err == nil {
			if err := exec.Command(history, "-a").Run(); err != nil {
				return errors.Wrap(err, "flushing session commands to terminal history")
			}
		}

		if histFile, ok := os.LookupEnv("HISTFILE"); ok {
			if err := clearHistoryFile(histFile); err != nil {
				return errors.Wrap(err, "clearing terminal history")
			}
		}
	}

	return nil
}

func clearHistoryFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return errors.Wrap(err, "opening file")
	}

	stat, err := f.Stat()
	if err != nil {
		return errors.Wrap(err, "getting file stats")
	}

	b := make([]byte, stat.Size())
	buf := bytes.NewBuffer(b)
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Bytes()
		if bytes.HasPrefix(bytes.TrimSpace(line), []byte("kure ")) {
			continue
		}
		buf.Write(line)
		buf.WriteByte('\n')
	}

	if err := f.Close(); err != nil {
		return errors.Wrap(err, "closing file")
	}

	if err := scanner.Err(); err != nil {
		return errors.Wrap(err, "scanning file")
	}

	if err := os.WriteFile(path, buf.Bytes(), 0600); err != nil {
		return errors.Wrap(err, "writing file")
	}

	return nil
}
