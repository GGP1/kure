package cmd

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/GGP1/kure/crypt"
	"github.com/GGP1/kure/db"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	httpB, encrypt, decrypt bool
	httpPort                uint16
	path                    string
)

var errInvalidPath = errors.New("please specify a path")

var backupCmd = &cobra.Command{
	Use:   "backup [http | encrypt | decrypt] [port] [path]",
	Short: "Create database backups",
	Run: func(cmd *cobra.Command, args []string) {
		if p := viper.GetInt("http.port"); p != 0 {
			httpPort = uint16(p)
		}

		if decrypt {
			if path == "" {
				fatal(errInvalidPath)
			}

			file, err := crypt.DecryptFile(path)
			if err != nil {
				fatal(err)
			}

			fmt.Println(string(file))
			return
		}

		if encrypt {
			if path == "" {
				fatal(errInvalidPath)
			}

			buf := new(bytes.Buffer)
			if err := db.WriteTo(buf); err != nil {
				fatal(err)
			}

			if err := crypt.EncryptedFile(buf.Bytes(), path); err != nil {
				fatal(err)
			}
			return
		}

		if httpB {
			http.HandleFunc("/", db.HTTPBackup)

			addr := fmt.Sprintf(":%d", httpPort)

			fmt.Printf("Serving file on port %s", addr)
			if err := http.ListenAndServe(addr, nil); err != nil {
				fatal(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)
	backupCmd.Flags().BoolVar(&httpB, "http", false, "run a server and write the db file")
	backupCmd.Flags().Uint16Var(&httpPort, "port", 4000, "server port")

	backupCmd.Flags().BoolVar(&encrypt, "encrypt", false, "create encrypted backup")
	backupCmd.Flags().BoolVar(&decrypt, "decrypt", false, "decrypt encrypted backup")
	backupCmd.Flags().StringVar(&path, "path", "./backup", "backup file path")
}
