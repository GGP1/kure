package session

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"strings"
	"time"

	"github.com/GGP1/kure/sig"
	"github.com/GGP1/kure/terminal"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// cleanup resets signal cleanups and sets all flags as unchanged to keep using default values.
//
// It also sets the help flag internal variable to false in case it's used.
func cleanup(cmd *cobra.Command) {
	sig.Signal.ResetCleanups()
	cmd.LocalFlags().VisitAll(func(f *pflag.Flag) { f.Changed = false })
	cmd.Flags().Set("help", "false")
}

// fillScript replaces any argument placeholder in the script with the user input.
func fillScript(args []string, script string) string {
	if !strings.ContainsRune(script, '$') {
		return script
	}

	n := 1 // Start from $1 like bash
	for _, arg := range args {
		placeholder := fmt.Sprintf("$%d", n)
		script = strings.ReplaceAll(script, placeholder, arg)
		n++
	}

	return script
}

// idleTimer executes a timer after x time has passed without receiving an input from the user.
func idleTimer(done chan struct{}, timeout *timeout) {
	if timeout.duration < (5 * time.Minute) {
		// Listen for the event before exiting so it doesn't block
		<-done
		return
	}

	var d float64
	if timeout.duration == 0 {
		d = 5
	} else {
		// round(log(x^3))
		d = math.Round(math.Log10(math.Pow(float64(timeout.duration), 3)))
	}
	timer := time.NewTimer(time.Duration(d) * time.Minute)
	select {
	case <-done:
		return

	case <-timer.C:
		fmt.Print("\n")
		terminal.Ticker(done, true, func() {
			fmt.Print(timeout)
		})
	}
}

// parseCommands looks for multiple commands concatenated by the logical AND operator and
// splits them into different slices.
//
//	Given "ls && copy github && 2fa"
//
//	Return [["ls"], ["copy", "github"], ["2fa"]].
func parseCommands(args []string) [][]string {
	// The underlying array will grow only if the script has multiple "&&" in a row
	ampersands := make([]int, 0, len(args)/2)
	for i, a := range args {
		if a == "&&" {
			// Store the indices of the ampersands
			ampersands = append(ampersands, i)
		}
	}

	// Pass on the args received if no ampersand was found
	if len(ampersands) == 0 {
		return [][]string{args}
	}

	group := make([][]string, 0, len(ampersands)+1)
	lastIdx := 0
	// Append len(ampersands) commands to the group
	for _, idx := range ampersands {
		group = append(group, args[lastIdx:idx])
		lastIdx = idx + 1 // Add one to skip the ampersand
	}

	// Append the last command
	group = append(group, args[lastIdx:])

	return group
}

// parseDoubleQuotes joins two arguments enclosed by doublequotes.
//
//	Given ["file", "touch", "\"file", "with", "spaces\""]
//
//	Return ["file", "touch", "file with spaces"]
func parseDoubleQuotes(args []string) []string {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, quote) {
			if strings.HasSuffix(arg, quote) {
				args[i] = strings.Trim(arg, quote)
				continue
			}

			for j := i + 1; j < len(args); j++ {
				if strings.HasSuffix(args[j], quote) {
					// Join enclosed words, store the sequence where the first one was
					// and remove the others from the slice
					words := strings.Join(args[i:j+1], " ")
					args[i] = strings.Trim(words, quote)
					args = append(args[:i+1], args[j+1:]...)
					i = j - 1
					break
				}
			}
		}
	}
	return args
}

// scanInput takes the user input and parses double quotes and scripts
// to return a slice with the command arguments.
func scanInput(reader *bufio.Reader, timeout *timeout, scripts map[string]string) ([][]string, error) {
	done := make(chan struct{})
	go idleTimer(done, timeout)

	text, _, err := reader.ReadLine()
	if err != nil {
		if err == io.EOF {
			sig.Signal.Kill()
		}
		return nil, err
	}
	done <- struct{}{}
	textStr := string(text)

	args := strings.Split(textStr, " ")
	if strings.Contains(textStr, quote) {
		args = parseDoubleQuotes(args)
	}

	script, ok := scripts[args[0]]
	if ok {
		script = fillScript(args[1:], script)
		args = strings.Split(script, " ")
	}

	return parseCommands(args), nil
}
