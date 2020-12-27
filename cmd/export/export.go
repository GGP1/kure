package export

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/db/entry"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var path string

var example = `
kure export <manager-name> -p path/to/file`

// NewCmd returns a new command.
func NewCmd(db *bolt.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export <manager-name>",
		Short: "Export entries",
		Long: `Export entries to other password managers. Format: CSV.

Supported:
   • Bitwarden
   • Keepass
   • Lastpass
   • 1Password`,
		Example: example,
		PreRunE: cmdutil.RequirePassword(db),
		RunE:    runExport(db),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset flags (session)
			path = ""
		},
	}

	cmd.Flags().StringVarP(&path, "path", "p", "", "destination file path")

	return cmd
}

func runExport(db *bolt.DB) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		manager := strings.Join(args, " ")
		if manager == "" {
			return errors.New("please specify the password manager name")
		}

		if path == "" {
			return errors.New("invalid path")
		}
		if filepath.Ext(path) == "" {
			path += ".csv"
		}

		headers, records, err := formatEntries(db, manager)
		if err != nil {
			return err
		}

		f, err := os.Create(path)
		if err != nil {
			return errors.Wrap(err, "failed creating the file")
		}

		w := csv.NewWriter(f)
		if err := w.Write(headers); err != nil {
			return errors.Wrap(err, "failed writing headers")
		}

		if err := w.WriteAll(records); err != nil {
			return errors.Wrap(err, "failed writing records")
		}

		if err := f.Close(); err != nil {
			return errors.Wrap(err, "failed closing file")
		}

		fmt.Printf("Created CSV file at %s\n", path)
		return nil
	}
}

// formatEntries takes all the entries in the database and formats them
// in headers and records to meet each manager requirements.
func formatEntries(db *bolt.DB, manager string) ([]string, [][]string, error) {
	lockedBuf, entries, err := entry.List(db)
	if err != nil {
		return nil, nil, err
	}
	defer lockedBuf.Destroy()

	headers := make([]string, 1)
	records := make([][]string, len(entries))

	switch strings.ToLower(manager) {
	case "keepass", "kp":
		headers = []string{"Account", "Login Name", "Password", "Web Site", "Comments"}

		for i, entry := range entries {
			records[i] = []string{entry.Name, entry.Username, entry.Password, entry.URL, entry.Notes}
		}

	case "1password", "onepassword", "1p":
		headers = []string{"Title", "Website", "Username", "Password", "Notes", "Member Number", "Recovery Codes"}

		for i, entry := range entries {
			records[i] = []string{entry.Name, entry.URL, entry.Username, entry.Password, entry.Notes, "", ""}
		}

	case "lastpass", "lp":
		headers = []string{"URL", "Username", "Password", "Extra", "Name", "Grouping", "Fav"}

		for i, entry := range entries {
			name, folders := splitName(entry.Name)
			records[i] = []string{entry.URL, entry.Username, entry.Password, entry.Notes, name, folders, ""}
		}

	case "bitwarden", "bw":
		headers = []string{"Folder", "Favorite", "Type", "Name", "Notes", "Fields", "Login_uri", "Login_username", "Login_password", "Login_totp"}

		for i, entry := range entries {
			name, folders := splitName(entry.Name)
			records[i] = []string{folders, "", "login", name, entry.Notes, "", entry.URL, entry.Username, entry.Password, ""}
		}

	default:
		return nil, nil, errors.Errorf("%q is not supported", manager)
	}

	return headers, records, nil
}

// splitName takes a path and returns the basename and the folders separated.
func splitName(path string) (basename, folders string) {
	split := strings.Split(path, "/")
	if len(split) == 1 {
		return path, ""
	}

	return filepath.Base(path), filepath.Dir(path)
}
