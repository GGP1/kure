package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var cfgFile, folder string

// RootCmd is the root command
var RootCmd = &cobra.Command{
	Use:   "kure",
	Short: "CLI password manager.",
}

func init() {
	cobra.OnInitialize()

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path")
	RootCmd.PersistentFlags().StringVar(&folder, "folder", "", "select folder")
}

// must returns the error that occurred and exists.
func must(err error) {
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}

// scan takes the input of the field and returns the value.
func scan(scanner *bufio.Scanner, field string, value string) string {
	fmt.Printf("%s: ", field)

	scanner.Scan()
	value = scanner.Text()

	return value
}
