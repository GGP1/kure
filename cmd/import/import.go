package importt

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/pb"

	"github.com/awnumar/memguard"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var path string

var example = `
kure import <manager-name> -p path/to/file`

// NewCmd returns a new command.
func NewCmd(db *bolt.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import <manager-name>",
		Short: "Import entries",
		Long: `Import entries from other password managers. Format: CSV.

Supported:
   • Bitwarden
   • Keepass
   • Lastpass
   • 1Password`,
		Example: example,
		RunE:    runImport(db),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset flag (session)
			path = ""
		},
	}

	cmd.Flags().StringVarP(&path, "path", "p", "", "path to csv file")

	return cmd
}

func runImport(db *bolt.DB) cmdutil.RunEFunc {
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

		f, err := os.Open(path)
		if err != nil {
			return errors.Wrap(err, "failed opening file")
		}
		defer f.Close()

		r := csv.NewReader(f)
		records, err := r.ReadAll()
		if err != nil {
			return errors.Wrap(err, "failed reading csv data")
		}

		// [1:] used to avoid headers
		switch strings.ToLower(manager) {
		case "keepass", "kp":
			for _, record := range records[:][1:] {
				lockedBuf, e := pb.SecureEntry()
				e.Name = strings.ToLower(record[0])
				e.Username = record[1]
				e.Password = record[2]
				e.URL = record[3]
				e.Notes = record[4]
				e.Expires = "Never"

				if err := entry.Create(db, lockedBuf, e); err != nil {
					fmt.Fprintln(os.Stderr, "error:", err)
				}
			}

		case "1password", "onepassword", "1p":
			for _, record := range records[:][1:] {
				notes := fmt.Sprintf("%s. Member number: %s. Recovery Codes: %s", record[4], record[5], record[6])

				lockedBuf, e := pb.SecureEntry()
				e.Name = strings.ToLower(record[0])
				e.Username = record[2]
				e.Password = record[3]
				e.URL = record[1]
				e.Notes = notes
				e.Expires = "Never"
				memguard.WipeBytes([]byte(notes))

				if err := entry.Create(db, lockedBuf, e); err != nil {
					fmt.Fprintln(os.Stderr, "error:", err)
				}
			}

		case "lastpass", "lp":
			for _, record := range records[:][1:] {
				lockedBuf, e := pb.SecureEntry()
				// Join folder and name
				e.Name = strings.ToLower(fmt.Sprintf("%s/%s", record[5], record[4]))
				e.Username = record[1]
				e.Password = record[2]
				e.URL = record[0]
				e.Notes = record[3]
				e.Expires = "Never"

				if err := entry.Create(db, lockedBuf, e); err != nil {
					fmt.Fprintln(os.Stderr, "error:", err)
				}
			}

		case "bitwarden", "bw":
			for _, record := range records[:][1:] {
				lockedBuf, e := pb.SecureEntry()
				// Join folder and name
				e.Name = strings.ToLower(fmt.Sprintf("%s/%s", record[0], record[3]))
				e.Username = record[7]
				e.Password = record[8]
				e.URL = record[6]
				e.Notes = record[4]
				e.Expires = "Never"

				if err := entry.Create(db, lockedBuf, e); err != nil {
					fmt.Fprintln(os.Stderr, "error:", err)
				}
			}

		default:
			return errors.Errorf("%q is not supported", manager)
		}

		fmt.Printf("Sucessfully imported the entries from %s\n", manager)
		return nil
	}
}
