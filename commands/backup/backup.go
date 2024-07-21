package backup

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/config"
	"github.com/GGP1/kure/sig"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

const example = `
* Create a file backup
kure backup --path path/to/file

* Serve the database on a local server, port 7777
kure backup --http --port 7777

* Download database
curl localhost:7777 > database_name`

type backupOptions struct {
	path  string
	port  uint16
	httpB bool
}

// NewCmd returns a new command.
func NewCmd(db *bolt.DB) *cobra.Command {
	opts := backupOptions{}
	cmd := &cobra.Command{
		Use:     "backup",
		Short:   "Create database backup",
		Example: example,
		RunE:    opts.runBackup(db),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset variables (session)
			opts = backupOptions{
				port: 8080,
			}
		},
	}

	f := cmd.Flags()
	f.BoolVar(&opts.httpB, "http", false, "serve database file on a local server")
	f.StringVar(&opts.path, "path", "", "destination file path")
	f.Uint16Var(&opts.port, "port", 8080, "server port")

	return cmd
}

func (opts *backupOptions) runBackup(db *bolt.DB) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		if opts.httpB {
			return serveFile(db, opts.port)
		}

		return fileBackup(db, opts.path)
	}
}

// serveFile serves the file on localhost.
func serveFile(db *bolt.DB, port uint16) error {
	if port == 0 {
		return errors.New("invalid port")
	}

	server := &http.Server{
		Addr: fmt.Sprintf(":%d", port),
	}
	sig.Signal.AddCleanup(func() error {
		// Do not exit after a signal as we are handling the shutdown
		sig.Signal.KeepAlive()
		fmt.Println("Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			return errors.Wrap(err, "graceful shutdown")
		}

		if err := server.Close(); err != nil {
			return errors.Wrap(err, "closing server")
		}
		return nil
	})

	// Register route only once, otherwise it will panic if
	// called multiple times inside a session
	var once sync.Once
	once.Do(func() {
		http.HandleFunc("/", httpBackup(db))
	})
	fmt.Printf("Serving database on http://localhost:%d (Press Ctrl+C to quit)\n", port)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return errors.Wrap(err, "starting server")
	}

	return nil
}

// fileBackup writes the database to a new file.
func fileBackup(db *bolt.DB, path string) error {
	if path == "" {
		return cmdutil.ErrInvalidPath
	}

	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return errors.Wrap(err, "making directory")
	}

	if err := os.Chdir(dir); err != nil {
		return errors.Wrap(err, "changing working directory")
	}

	newDB, err := bolt.Open(filepath.Base(path), 0o600, nil)
	if err != nil {
		return errors.Wrap(err, "opening database backup")
	}

	if err := bolt.Compact(newDB, db, 0); err != nil {
		return errors.Wrap(err, "copying database backup")
	}

	if err := newDB.Close(); err != nil {
		return errors.Wrap(err, "closing database backup")
	}

	abs, _ := filepath.Abs(path)
	fmt.Println("Backup created at", abs)
	return nil
}

// httpBackup writes a consistent view of the database to a http endpoint.
func httpBackup(db *bolt.DB) http.HandlerFunc {
	name := filepath.Base(config.GetString("database.path"))
	disposition := fmt.Sprintf(`attachment; filename=%q`, name)

	return func(w http.ResponseWriter, r *http.Request) {
		err := db.View(func(tx *bolt.Tx) error {
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("Content-Disposition", disposition)
			w.Header().Set("Content-Length", strconv.Itoa(int(tx.Size())))
			if _, err := tx.WriteTo(w); err != nil {
				return errors.Wrap(err, "writing the database")
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
			return errors.Wrap(err, "writing the database")
		}
		return nil
	})
}
