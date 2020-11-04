package cmd

import (
	"bytes"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/GGP1/kure/crypt"
	"github.com/GGP1/kure/db"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var backupCmd = &cobra.Command{
	Use:   "backup [http | encrypt | decrypt] [port] [path]",
	Short: "Create database backups",
	Run: func(cmd *cobra.Command, args []string) {
		var err error

		if p := viper.GetInt("http.port"); p != 0 {
			httpPort = uint16(p)
		}

		if decrypt {
			if path == "" {
				fatal(errInvalidPath)
			}

			path, err = filepath.Abs(path)
			if err != nil {
				fatal(errInvalidPath)
			}

			file, err := crypt.DecryptEncFile(path)
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

			path, err = filepath.Abs(path)
			if err != nil {
				fatal(errInvalidPath)
			}

			buf := new(bytes.Buffer)
			if err := db.WriteTo(buf); err != nil {
				fatal(err)
			}

			if err := crypt.CreateEncFile(buf.Bytes(), path); err != nil {
				fatal(err)
			}
			return
		}

		if httpB {
			if err := db.RequirePassword(); err != nil {
				fatal(err)
			}
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
	backupCmd.Flags().StringVar(&path, "path", "", "backup file path")
}
