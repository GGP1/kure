package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
)

var (
	clip     bool
	term     bool
	clearCmd = &cobra.Command{
		Use:   "clear",
		Short: "Clear clipboard/terminal.",
		Long:  "Manually clean the clipboard, the terminal or both.",
		Run: func(cmd *cobra.Command, args []string) {
			str := strings.Join(args, " ")

			if strings.Contains(str, "both") {
				clip = true
				term = true
			}

			if clip {
				err := clipboard.WriteAll("")
				if err != nil {
					fmt.Println(err)
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
)

func init() {
	RootCmd.AddCommand(clearCmd)
	clearCmd.Flags().BoolVarP(&clip, "clipboard", "c", false, "clear clipboard")
	clearCmd.Flags().BoolVarP(&term, "terminal", "t", false, "clear terminal")
}
