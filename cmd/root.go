package cmd

import (
	"bufio"
	"fmt"

	"github.com/spf13/cobra"
)

var cfgFile string

// RootCmd is the root command
var RootCmd = &cobra.Command{
	Use:   "kure",
	Short: "CLI password manager.",
}

func init() {
	cobra.OnInitialize()

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path")
}

// scan takes the input of the field and retuns the value
func scan(scanner *bufio.Scanner, field string, value string) string {
	fmt.Printf("%s: ", field)

	scanner.Scan()
	value = scanner.Text()

	return value
}
