package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Unset master password",
	Run: func(cmd *cobra.Command, args []string) {
		viper.Set("user.password", "")

		path := getConfigPath()

		if err := viper.WriteConfigAs(path); err != nil {
			fatalf("failed creating config file: %v", err)
		}

		fmt.Println("You logged out")
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
