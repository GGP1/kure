package cmd

import "github.com/spf13/cobra"

// RootCmd is the root command
var RootCmd = &cobra.Command{
	Use:   "kure",
	Short: "CLI password manager.",
}
