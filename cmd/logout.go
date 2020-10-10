package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Unset master password",
	Run: func(cmd *cobra.Command, args []string) {
		viper.Set("user.password", "")

		filename := fmt.Sprintf("%s/config.yaml", os.Getenv("KURE_CONFIG"))
		if filename != "" {
			if err := viper.WriteConfigAs(filename); err != nil {
				fatalf(errCreatingConfig, err)
			}
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				fatal(err)
			}

			path = fmt.Sprintf("%s/config.yaml", home)

			if err := viper.WriteConfigAs(path); err != nil {
				fatalf(errCreatingConfig, err)
			}
		}

		fmt.Println("You logged out")
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
