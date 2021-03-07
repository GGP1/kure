package importt

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
	"github.com/GGP1/kure/pb"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var example = `
* Import
kure import keepass -p path/to/file

* Import and delete the file:
kure import 1password -e -p path/to/file`

type importOptions struct {
	path  string
	erase bool
}

// NewCmd returns a new command.
func NewCmd(db *bolt.DB) *cobra.Command {
	opts := importOptions{}

	cmd := &cobra.Command{
		Use:   "import <manager-name>",
		Short: "Import entries",
		Long: `Import entries from other password managers. Format: CSV.

If an entry already exists it will be overwritten.

Delete the CSV used with the erase flag, the file will be deleted only if no errors were encountered.

Supported:
	• 1Password
	• Bitwarden
   	• Keepass/X/XC
	• Lastpass`,
		Example: example,
		Args:    managersSupported(),
		PreRunE: auth.Login(db),
		RunE:    runImport(db, &opts),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset flag (session)
			opts = importOptions{}
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.path, "path", "p", "", "source file path")
	f.BoolVarP(&opts.erase, "erase", "e", false, "erase the file on exit (only if there are no errors)")

	return cmd
}

func runImport(db *bolt.DB, opts *importOptions) cmdutil.RunEFunc {
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

		records, err := readCSV(opts.path)
		if err != nil {
			return err
		}

		if err := createEntries(db, manager, records); err != nil {
			return err
		}

		if opts.erase {
			if err := cmdutil.Erase(opts.path); err != nil {
				return err
			}
			fmt.Printf("Erased file at %s\n", opts.path)
		}

		fmt.Printf("Successfully imported the entries from %s\n", manager)
		return nil
	}
}

func createEntries(db *bolt.DB, manager string, records [][]string) error {
	// [1:] used to avoid headers
	records = records[:][1:]
	entries := make([]*pb.Entry, len(records))

	switch manager {
	case "keepass", "keepassx":
		for i, record := range records {
			entries[i] = &pb.Entry{
				Name:     cmdutil.NormalizeName(record[0]),
				Username: record[1],
				Password: record[2],
				URL:      record[3],
				Notes:    record[4],
				Expires:  "Never",
			}
		}

	case "keepassxc":
		for i, record := range records {
			entries[i] = &pb.Entry{
				// Join folder and name
				Name:     cmdutil.NormalizeName(fmt.Sprintf("%s/%s", record[0], record[1])),
				Username: record[2],
				Password: record[3],
				URL:      record[4],
				Notes:    record[5],
				Expires:  "Never",
			}
		}

	case "1password":
		for i, record := range records {
			entries[i] = &pb.Entry{
				Name:     cmdutil.NormalizeName(record[0]),
				Username: record[2],
				Password: record[3],
				URL:      record[1],
				Notes:    fmt.Sprintf("%s.\nMember number: %s.\nRecovery Codes: %s", record[4], record[5], record[6]),
				Expires:  "Never",
			}
		}

	case "lastpass":
		for i, record := range records {
			entries[i] = &pb.Entry{
				// Join folder and name
				Name:     cmdutil.NormalizeName(fmt.Sprintf("%s/%s", record[5], record[4])),
				Username: record[1],
				Password: record[2],
				URL:      record[0],
				Notes:    record[3],
				Expires:  "Never",
			}
		}

	case "bitwarden":
		for i, record := range records {
			// Join folder and name
			name := cmdutil.NormalizeName(fmt.Sprintf("%s/%s", record[0], record[3]))
			entries[i] = &pb.Entry{
				Name:     name,
				Username: record[7],
				Password: record[8],
				URL:      record[6],
				Notes:    record[4],
				Expires:  "Never",
			}

			// Create TOTP if the entry has one
			createTOTP(db, name, record[9])
		}
	}

	for _, e := range entries {
		if err := entry.Create(db, e); err != nil {
			return err
		}
	}

	return nil
}

func createTOTP(db *bolt.DB, name, raw string) error {
	if raw == "" {
		return nil
	}

	t := &pb.TOTP{
		Name: name,
		Raw:  raw,
		// Bitwarden uses 6 digits by default
		Digits: 6,
	}

	return totp.Create(db, t)
}

func readCSV(path string) ([][]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "opening file")
	}
	defer f.Close()

	fInfo, err := f.Stat()
	if err != nil {
		return nil, errors.Wrap(err, "obtaining file information")
	}

	if fInfo.Size() == 0 {
		return nil, errors.New("the CSV file is empty")
	}

	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		return nil, errors.Wrap(err, "reading csv data")
	}

	return records, nil
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
