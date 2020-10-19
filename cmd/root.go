package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var cfgFile string

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
		fmt.Println(err)
		os.Exit(1)
	}
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
