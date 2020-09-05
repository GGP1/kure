package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/GGP1/kure/db"

	"github.com/spf13/cobra"
)

var (
	titles    []string
	deleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete an entry.",
		Run: func(cmd *cobra.Command, args []string) {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Are you sure? [y/n]: ")

			text, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println(err)
				return
			}
			res := strings.ToLower(text)

			if strings.Contains(res, "y") || strings.Contains(res, "yes") {
				err := db.DeleteEntry(titles)
				if err != nil {
					fmt.Println(err)
				}
			}
		},
	}
)

func init() {
	RootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().StringSliceVarP(&titles, "titles", "t", []string{}, "entry title")
	deleteCmd.MarkFlagRequired("title")
}
