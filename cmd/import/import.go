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
		Short: "Import entries from other password managers",
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

		var entries []*pb.Entry
		// [1:] used to avoid headers
		switch strings.ToLower(manager) {
		case "keepass", "kp":
			for _, record := range records[:][1:] {
				e := &pb.Entry{
					Name:     strings.ToLower(record[0]),
					Username: record[1],
					Password: record[2],
					URL:      record[3],
					Notes:    record[4],
					Expires:  "Never",
				}

				entries = append(entries, e)
			}

		case "1password", "onepassword", "1p":
			for _, record := range records[:][1:] {
				notes := fmt.Sprintf("%s. Member number: %s. Recovery Codes: %s", record[4], record[5], record[6])

				e := &pb.Entry{
					Name:     strings.ToLower(record[0]),
					Username: record[2],
					Password: record[3],
					URL:      record[1],
					Notes:    notes,
					Expires:  "Never",
				}

				entries = append(entries, e)
			}

		case "lastpass", "lp":
			for _, record := range records[:][1:] {
				e := &pb.Entry{
					// Join folder with name
					Name:     strings.ToLower(fmt.Sprintf("%s/%s", record[5], record[4])),
					Username: record[1],
					Password: record[2],
					URL:      record[0],
					Notes:    record[3],
					Expires:  "Never",
				}

				entries = append(entries, e)
			}

		case "bitwarden", "bw":
			for _, record := range records[:][1:] {
				e := &pb.Entry{
					// Join folder with name
					Name:     strings.ToLower(fmt.Sprintf("%s/%s", record[0], record[3])),
					Username: record[7],
					Password: record[8],
					URL:      record[6],
					Notes:    record[4],
					Expires:  "Never",
				}

				entries = append(entries, e)
			}

		default:
			return errors.Errorf("%q is not supported", manager)
		}

		for _, e := range entries {
			if err := entry.Create(db, e); err != nil {
				return err
			}
		}

		return nil
	}
}
