package backup

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	cmdutil "github.com/GGP1/kure/cmd"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	bolt "go.etcd.io/bbolt"
)

var (
	httpB bool
	port  uint16
	path  string
)

var (
	errInvalidPath = errors.New("invalid path")
	errInvalidPort = errors.New("invalid port")
)

var example = `
* Create a file backup
kure backup --path path/to/file

* Serve the database on a local server, port 2000
kure backup --http --port 2000

* Download database
curl localhost:4000 > database.name`

// NewCmd returns a new command.
func NewCmd(db *bolt.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "backup",
		Short:   "Create database backup",
		Example: example,
		PreRunE: cmdutil.RequirePassword(db),
		RunE:    runBackup(db),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset flags (session)
			httpB = false
			port = 4000
			path = ""
		},
	}

	f := cmd.Flags()
	f.BoolVar(&httpB, "http", false, "run a server and write the db file")
	f.StringVar(&path, "path", "", "destination file path")
	f.Uint16Var(&port, "port", 4000, "server port")

	return cmd
}

func runBackup(db *bolt.DB) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		// Serve on localhost
		if httpB {
			if p := viper.GetInt("http.port"); p > 0 {
				port = uint16(p)
			}
			if port == 0 {
				return errInvalidPort
			}

			http.HandleFunc("/", httpBackup(db))

			addr := fmt.Sprintf(":%d", port)
			fmt.Printf("Serving database on http://localhost%s (Press Ctrl+C to quit)\n", addr)
			return http.ListenAndServe(addr, nil)
		}

		// Create a file backup
		if path == "" {
			return errInvalidPath
		}

		dir := filepath.Dir(path)

		if err := os.MkdirAll(dir, os.ModeDir); err != nil {
			return errors.Wrap(err, "failed making directory")
		}

		if err := os.Chdir(dir); err != nil {
			return errors.Wrap(err, "failed changing directory")
		}

		f, err := os.Create(filepath.Base(path))
		if err != nil {
			return errors.Wrap(err, "failed opening file")
		}

		if err := writeTo(db, f); err != nil {
			return err
		}

		if err := f.Close(); err != nil {
			return errors.Wrap(err, "failed closing file")
		}

		fmt.Println("Backup created at", path)
		return nil
	}
}

// httpBackup writes a consistent view of the database to a http endpoint.
func httpBackup(db *bolt.DB) http.HandlerFunc {
	name := viper.GetString("database.name")
	disposition := fmt.Sprintf(`attachment; filename="%s"`, name)

	return func(w http.ResponseWriter, r *http.Request) {
		err := db.View(func(tx *bolt.Tx) error {
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("Content-Disposition", disposition)
			w.Header().Set("Content-Length", strconv.Itoa(int(tx.Size())))
			if _, err := tx.WriteTo(w); err != nil {
				return errors.Wrap(err, "write database")
			}

			return nil
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// writeTo writes the entire database to a writer.
func writeTo(db *bolt.DB, w io.Writer) error {
	return db.View(func(tx *bolt.Tx) error {
		if _, err := tx.WriteTo(w); err != nil {
			return errors.Wrap(err, "write database")
		}
		return nil
	})
}
