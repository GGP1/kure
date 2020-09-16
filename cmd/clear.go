package cmd

import (
	"log"
	"os"
	"os/exec"
	"runtime"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
)

var (
	both bool
	clip bool
	term bool
)

var clearCmd = &cobra.Command{
	Use:   "clear [-b both] [-c clipboard] [-t terminal]",
	Short: "Clear clipboard/terminal",
	Long:  "Manually clean the clipboard, the terminal or both.",
	Run: func(cmd *cobra.Command, args []string) {
		if both {
			clip = true
			term = true
		}

		if clip {
			if err := clipboard.WriteAll(""); err != nil {
				log.Fatal("error: ", err)
			}
		}

		if term {
			if runtime.GOOS == "windows" {
				c := exec.Command("cmd", "/c", "cls")
				c.Stdout = os.Stdout
				c.Run()
				return
			}

			c := exec.Command("clear")
			c.Stdout = os.Stdout
			c.Run()
		}
	},
}

func init() {
	RootCmd.AddCommand(clearCmd)
	clearCmd.Flags().BoolVarP(&both, "both", "b", false, "clear both clipboard and terminal")
	clearCmd.Flags().BoolVarP(&clip, "clipboard", "c", false, "clear clipboard")
	clearCmd.Flags().BoolVarP(&term, "terminal", "t", false, "clear terminal")
}
