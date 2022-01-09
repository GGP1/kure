package export

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/GGP1/kure/auth"
	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/db/totp"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

const example = `
kure export <manager-name> -p path/to/file`

type exportOptions struct {
	path string
}

// NewCmd returns a new command.
func NewCmd(db *bolt.DB) *cobra.Command {
	opts := exportOptions{}

	cmd := &cobra.Command{
		Use:   "export <manager-name>",
		Short: "Export entries",
		Long: `Export entries to other password managers. 
		
This command creates a CSV file with all the entries unencrypted, make sure to delete it after it's used.

Supported:
	• 1Password
	• Bitwarden
   	• Keepass/X/XC
   	• Lastpass`,
		Example: example,
		Args:    managersSupported(),
		PreRunE: auth.Login(db),
		RunE:    runExport(db, &opts),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset variables (session)
			opts = exportOptions{}
		},
	}

	cmd.Flags().StringVarP(&opts.path, "path", "p", "", "destination file path")

	return cmd
}

func runExport(db *bolt.DB, opts *exportOptions) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		manager := strings.Join(args, " ")
		manager = strings.ToLower(manager)

		if opts.path == "" {
			return cmdutil.ErrInvalidPath
		}
		ext := filepath.Ext(opts.path)
		if ext == "" || ext == "." {
			opts.path += ".csv"
		}

		headers, records, err := fmtEntries(db, manager)
		if err != nil {
			return err
		}

		if err := createCSV(headers, records, opts.path); err != nil {
			return err
		}

		abs, _ := filepath.Abs(opts.path)
		fmt.Printf("Created CSV file at %s\n", abs)
		return nil
	}
}

func createCSV(headers []string, records [][]string, path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
	if err != nil {
		return errors.Wrap(err, "creating the file")
	}

	w := csv.NewWriter(f)
	if err := w.Write(headers); err != nil {
		return errors.Wrap(err, "writing headers")
	}

	if err := w.WriteAll(records); err != nil {
		return errors.Wrap(err, "writing records")
	}

	if err := f.Close(); err != nil {
		return errors.Wrap(err, "closing file")
	}

	return nil
}

// fmtEntries takes all the entries in the database and formats them
// in headers and records to meet each manager requirements.
func fmtEntries(db *bolt.DB, manager string) ([]string, [][]string, error) {
	entries, err := entry.List(db)
	if err != nil {
		return nil, nil, err
	}

	headers := make([]string, 1)
	records := make([][]string, len(entries))

	switch manager {
	case "keepass", "keepassx":
		headers = []string{"Account", "Login Name", "Password", "Web Site", "Comments"}

		for i, e := range entries {
			records[i] = []string{e.Name, e.Username, e.Password, e.URL, e.Notes}
		}

	case "keepassxc":
		headers = []string{"Group", "Title", "Username", "Password", "URL", "Notes"}

		for i, e := range entries {
			dir, name := splitName(e.Name)
			records[i] = []string{dir, name, e.Username, e.Password, e.URL, e.Notes}
		}

	case "1password":
		headers = []string{"Title", "Website", "Username", "Password", "Notes", "Member Number", "Recovery Codes"}

		for i, e := range entries {
			records[i] = []string{e.Name, e.URL, e.Username, e.Password, e.Notes, "", ""}
		}

	case "lastpass":
		headers = []string{"URL", "Username", "Password", "Extra", "Name", "Grouping", "Fav"}

		for i, e := range entries {
			dir, name := splitName(e.Name)
			records[i] = []string{e.URL, e.Username, e.Password, e.Notes, name, dir, ""}
		}

	case "bitwarden":
		headers = []string{"Folder", "Favorite", "Type", "Name", "Notes", "Fields", "Login_uri", "Login_username", "Login_password", "Login_totp"}

		for i, e := range entries {
			rawTOTP := getTOTP(db, e.Name)
			dir, name := splitName(e.Name)
			records[i] = []string{dir, "", "login", name, e.Notes, "", e.URL, e.Username, e.Password, rawTOTP}
		}
	}

	return headers, records, nil
}

// getTOTP returns the raw TOTP if it exists and an empty string otherwise.
func getTOTP(db *bolt.DB, name string) string {
	t, err := totp.Get(db, name)
	if err != nil {
		return ""
	}
	return t.Raw
}

// splitName is like filepath Dir() and Base() but if the directory is empty it returns "".
func splitName(path string) (dir, name string) {
	dir = filepath.Dir(path)
	if dir == "." {
		dir = ""
	}
	name = filepath.Base(path)

	return
}

func managersSupported() cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		manager := strings.Join(args, " ")

		switch strings.ToLower(manager) {
		case "1password", "bitwarden", "keepass", "keepassx", "keepassxc", "lastpass":

		default:
			return errors.Errorf(`%q is not supported

Managers supported: 1Password, Bitwarden, Keepass/X/XC, Lastpass`, manager)
		}
		return nil
	}
}
