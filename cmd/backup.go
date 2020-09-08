package cmd

import (
	"fmt"
	"net/http"

	"github.com/GGP1/kure/db"

	"github.com/spf13/cobra"
)

var (
	httpB     bool
	port      uint16
	encrypt   bool
	decrypt   bool
	path      string
	backupCmd = &cobra.Command{
		Use:   "backup [http] [port] [encrypt] [decrypt] [path]",
		Short: "Create database backups",
		Run: func(cmd *cobra.Command, args []string) {
			if decrypt {
				pwd, err := passInput()
				if err != nil {
					fmt.Println("error:", err)
					return
				}

				file, err := db.DecryptBackup(path, pwd)
				if err != nil {
					fmt.Println("error:", err)
				}
				fmt.Println(string(file))
				return
			}

			if encrypt {
				pwd, err := passInput()
				if err != nil {
					fmt.Println("error:", err)
					return
				}

				if err := db.EncryptedBackup(path, pwd); err != nil {
					fmt.Println("error:", err)
				}
				return
			}

			if httpB {
				http.HandleFunc("/", db.HTTPBackup)

				addr := fmt.Sprintf(":%d", port)

				if err := http.ListenAndServe(addr, nil); err != nil {
					fmt.Println("error:", err)
				}
			}
		},
	}
)

func init() {
	RootCmd.AddCommand(backupCmd)
	backupCmd.Flags().BoolVar(&httpB, "http", false, "run a server and write the db file")
	backupCmd.Flags().Uint16Var(&port, "port", 4000, "server port")

	backupCmd.Flags().BoolVar(&encrypt, "encrypt", false, "create encrypted backup")
	backupCmd.Flags().BoolVar(&decrypt, "decrypt", false, "decrypt encrypted backup")
	backupCmd.Flags().StringVar(&path, "path", "./backup", "backup file path")
}
