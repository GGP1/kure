package cmd

import (
	"os"
	"os/exec"
	"runtime"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
)

var both, clip, term bool

var clearCmd = &cobra.Command{
	Use:   "clear [-b both] [-c clipboard] [-t terminal]",
	Short: "Clear clipboard/terminal or both",
	Long: `Manually clean clipboard, terminal (and its history) or both of them.
Windows users must clear the history manually with ALT+F7, executing "cmd" command 
or by re-opening the cmd (as it saves session history only).`,
	Run: func(cmd *cobra.Command, args []string) {
		if clip == true || term == true {
			both = false
		}

		if both {
			clip = true
			term = true
		}

		if clip {
			if err := clipboard.WriteAll(""); err != nil {
				fatal(err)
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

			h := exec.Command("/bin/bash", "history -c", "history -cw")
			h.Stdout = os.Stdout
			h.Run()
		}
	},
}

func init() {
	rootCmd.AddCommand(clearCmd)
	clearCmd.Flags().BoolVarP(&both, "both", "b", true, "clear clipboard, terminal and history")
	clearCmd.Flags().BoolVarP(&clip, "clipboard", "c", false, "clear clipboard")
	clearCmd.Flags().BoolVarP(&term, "terminal", "t", false, "clear terminal")
}
