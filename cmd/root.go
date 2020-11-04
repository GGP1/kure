package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var both, clip, create, custom, decrypt, directory, encrypt, filter, hide, httpB, ignore, overwrite, password, qr, repeat, term, username bool

var (
	cfgFile, exclude, field, include, list, path, separator string
	semaphore                                               uint32
	httpPort                                                uint16
	length                                                  uint64
	excl, incl, many                                        []string
	format                                                  []int
	timeout                                                 time.Duration
)

var (
	errInvalidLength    = errors.New("invalid length")
	errInvalidName      = errors.New("invalid name")
	errInvalidPath      = errors.New("invalid path")
	errWritingClipboard = errors.New("failed writing to the clipboard")
)

var rootCmd = &cobra.Command{
	Use:   "kure",
	Short: "CLI password manager.",
}

func init() {
	cobra.OnInitialize()
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path")
}

// Execute sets each sub command flag and adds it to the root.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fatal(err)
	}
}

// proceed asks the user if he wants to continue or not.
func proceed() bool {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Are you sure you want to proceed? [y/N]: ")

	scanner.Scan()
	text := scanner.Text()
	text = strings.ToLower(text)

	if strings.Contains(text, "y") {
		return true
	}

	return false
}

// fatal prints the error and exits.
func fatal(err error) {
	fmt.Printf("error: %v\n", err)
	os.Exit(1)
}

// fatalf takes an error and its context, prints it and exits.
func fatalf(format string, v ...interface{}) {
	fmt.Println("error:", fmt.Sprintf(format, v...))
	os.Exit(1)
}

// scan takes the input of the field and updates the pointer.
func scan(scanner *bufio.Scanner, field string, value *string) {
	fmt.Printf("%s: ", field)

	scanner.Scan()
	text := scanner.Text()
	*value = strings.TrimSpace(text)
}

func scanlns(scanner *bufio.Scanner, field string, value *string) {
	fmt.Printf("%s (type !end to finish): ", field)

	var text []string
	for scanner.Scan() {
		t := scanner.Text()

		// switch not working
		if t == "!end" {
			break
		}

		if t == "" || t == "\n" {
			continue
		}

		text = append(text, t)
	}
	*value = strings.Join(text, "\n")
}

// printObjectName prints the name of the object adapting
// the top line of the chart so it stays always in the center.
func printObjectName(name string) {
	if strings.Contains(name, "/") {
		split := strings.Split(name, "/")
		name = split[len(split)-1]
	}
	name = strings.Title(name)

	dashes := 43
	halfBar := ((dashes - len(name)) / 2) - 1

	fmt.Print("+")
	for i := 0; i < halfBar; i++ {
		fmt.Print("─")
	}
	fmt.Printf(" %s ", name)

	if (len(name) % 2) == 0 {
		halfBar++
	}

	for i := 0; i < halfBar; i++ {
		fmt.Print("─")
	}
	fmt.Print(">\n")
}
