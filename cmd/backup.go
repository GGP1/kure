package cmd

import (
	"fmt"
	"log"
	"net/http"

	"github.com/GGP1/kure/db"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	httpB    bool
	httpPort uint16
	encrypt  bool
	decrypt  bool
	path     string
)

var backupCmd = &cobra.Command{
	Use:   "backup [http] [port] [encrypt] [decrypt] [path]",
	Short: "Create database backups",
	Run: func(cmd *cobra.Command, args []string) {
		if p := viper.GetInt("http.port"); p != 0 {
			httpPort = uint16(p)
		}

		if decrypt {
			file, err := db.DecryptFile(path)
			if err != nil {
				log.Fatal("error: ", err)
			}
			fmt.Println(string(file))
			return
		}

		if encrypt {
			if err := db.EncryptedFile(path); err != nil {
				log.Fatal("error: ", err)
			}
			return
		}

		if httpB {
			http.HandleFunc("/", db.HTTPBackup)

			addr := fmt.Sprintf(":%d", httpPort)

			fmt.Printf("Serving file on port %s", addr)
			if err := http.ListenAndServe(addr, nil); err != nil {
				log.Fatal("error: ", err)
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(backupCmd)
	backupCmd.Flags().BoolVar(&httpB, "http", false, "run a server and write the db file")
	backupCmd.Flags().Uint16Var(&httpPort, "port", 2727, "server port")

	backupCmd.Flags().BoolVar(&encrypt, "encrypt", false, "create encrypted backup")
	backupCmd.Flags().BoolVar(&decrypt, "decrypt", false, "decrypt encrypted backup")
	backupCmd.Flags().StringVar(&path, "path", "./backup", "backup file path")
}
