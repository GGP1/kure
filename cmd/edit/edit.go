package edit

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"time"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/pb"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var name bool

var example = `
* Edit all fields 
kure edit entryName

* Edit all fields but the name
kure edit entryName -n`

// NewCmd returns a new command.
func NewCmd(db *bolt.DB, r io.Reader) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit <name> [-n name]",
		Short: "Edit an entry",
		Long: `Edit entry fields.
		
"-" = Clear field.
"" (nothing) = Do not modify field.`,
		Example: example,
		RunE:    runEdit(db, r),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset flags defaults (session)
			name = false
		},
	}

	cmd.Flags().BoolVarP(&name, "name", "n", false, "edit entry name as well")

	return cmd
}

func runEdit(db *bolt.DB, r io.Reader) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")

		oldEntry, err := entry.Get(db, name)
		if err != nil {
			return err
		}

		newEntry, err := editEntryInput(oldEntry, r)
		if err != nil {
			return err
		}

		if err := entry.Edit(db, name, newEntry); err != nil {
			return err
		}

		fmt.Printf("\nSuccessfully edited %q entry.\n", name)

		return nil
	}
}

func editEntryInput(entry *pb.Entry, r io.Reader) (*pb.Entry, error) {
	fmt.Println("Use '-' to clear a field and '' (nothing) to keep it unchanged")
	fmt.Print("\n")

	scanner := bufio.NewScanner(r)
	if name {
		n := cmdutil.Scan(scanner, "New name")
		entry.Name = n
	}
	username := cmdutil.Scan(scanner, "Username")
	password := cmdutil.Scan(scanner, "Password")
	url := cmdutil.Scan(scanner, "URL")
	notes := cmdutil.Scanlns(scanner, "Notes")
	expires := cmdutil.Scan(scanner, "Expires")

	expires = strings.ToLower(expires)

	switch expires {
	case "never", "", " ", "0":
		expires = "Never"

	default:
		expires = strings.ReplaceAll(expires, "-", "/")

		// If the first format fails, try the second
		exp, err := time.Parse("2006/01/02", expires)
		if err != nil {
			exp, err = time.Parse("02/01/2006", expires)
			if err != nil {
				return nil, errors.New("invalid time format. Valid formats: d/m/y or y/m/d")
			}
		}

		expires = exp.Format(time.RFC1123Z)
	}

	// Fields must be passed by reference, values not necessary
	updater := func(field *string, value string) {
		switch value {
		case "":
		case "-":
			*field = ""
		default:
			*field = value
		}
	}

	updater(&entry.Username, username)
	updater(&entry.Password, password)
	updater(&entry.URL, url)
	updater(&entry.Notes, notes)
	updater(&entry.Expires, expires)

	return entry, nil
}
