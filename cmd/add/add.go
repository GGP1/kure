package add

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"strings"
	"time"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/pb"

	"github.com/GGP1/atoll"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	bolt "go.etcd.io/bbolt"
)

var (
	custom, repeat   bool
	length           uint64
	format           []int
	include, exclude string
)

var (
	errInvalidPath   = errors.New("invalid path")
	errInvalidName   = errors.New("invalid name")
	errInvalidLength = errors.New("error: invalid length")
)

var example = `
* Add an entry using a custom password
kure add entryName -c

* Add an entry generating a customized random password
kure add -l 27 -f 1,2,3,4,5 -i & -e / -r`

// NewCmd returns a new command.
func NewCmd(db *bolt.DB, r io.Reader) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add <name> [-c custom] [-l length] [-f format] [-i include] [-e exclude] [-r repeat]",
		Short:   "Add an entry",
		Aliases: []string{"a", "new", "create"},
		Example: example,
		RunE:    runAdd(db, r),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset flags defaults (session)
			custom, repeat = false, true
			length = 0
			format = nil
			include, exclude = "", ""
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	f := cmd.Flags()
	f.BoolVarP(&custom, "custom", "c", false, "use a custom password")
	f.Uint64VarP(&length, "length", "l", 0, "password length")
	f.IntSliceVarP(&format, "format", "f", nil, "password format (1,2,3,4,5)")
	f.StringVarP(&include, "include", "i", "", "string of characters to include in the password")
	f.StringVarP(&exclude, "exclude", "e", "", "string of characters to exclude from the password")
	f.BoolVarP(&repeat, "repeat", "r", true, "allow duplicated characters or not")

	cmd.AddCommand(phraseSubCmd(db, r))

	return cmd
}

func runAdd(db *bolt.DB, r io.Reader) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")
		if name == "" {
			return errInvalidName
		}
		if !strings.Contains(name, "/") && len(name) > 57 {
			return errors.New("entry name must contain 57 letters or less")
		}

		if err := cmdutil.RequirePassword(db); err != nil {
			return err
		}

		e, err := entryInput(db, name, custom, r)
		if err != nil {
			return errors.Wrap(err, "error")
		}

		if !custom {
			if length < 1 || length > math.MaxUint64 {
				return errInvalidLength
			}

			// If the user didn't specify format levels, use default
			if format == nil {
				if defFormat := viper.GetIntSlice("entry.format"); len(defFormat) > 0 {
					format = defFormat
				}
			}
			if eRepeat := viper.GetBool("entry.repeat"); eRepeat != false {
				repeat = eRepeat
			}

			uFormat := make([]uint8, len(format))

			for i := range format {
				uFormat[i] = uint8(format[i])
			}

			password := &atoll.Password{
				Length:  length,
				Format:  uFormat,
				Include: include,
				Exclude: exclude,
				Repeat:  repeat,
			}

			p, err := atoll.NewSecret(password)
			if err != nil {
				return errors.Wrap(err, "error")
			}

			e.Password = p
		}

		if err := entry.Create(db, e); err != nil {
			return errors.Wrap(err, "error")
		}

		fmt.Printf("\nSuccessfully created %q entry.\n", name)
		return nil
	}
}

func entryInput(db *bolt.DB, name string, custom bool, r io.Reader) (*pb.Entry, error) {
	name = strings.ToLower(name)

	if err := exists(db, name); err != nil {
		return nil, err
	}

	var (
		password string
		err      error
	)

	scanner := bufio.NewScanner(r)
	username := cmdutil.Scan(scanner, "Username")
	if custom {
		password = cmdutil.Scan(scanner, "Password")
	}
	url := cmdutil.Scan(scanner, "URL")
	notes := cmdutil.Scanlns(scanner, "Notes")
	expires := cmdutil.Scan(scanner, "Expires")

	expires, err = formatExpiration(expires)
	if err != nil {
		return nil, err
	}

	e := &pb.Entry{
		Name:     name,
		Username: username,
		URL:      url,
		Notes:    notes,
		Expires:  expires,
	}
	if password != "" {
		e.Password = password
	}

	return e, nil
}

// exists compares "path" with the existing entry names that contain "path",
// split both name and each of the entry names and look for matches on the same level.
//
// Given a path "Naboo/Padmé" and comparing it with "Naboo/Padmé Amidala":
//
// "Padmé" != "Padmé Amidala", skip.
//
// Given a path "jedi/Yoda" and comparing it with "jedi/Obi-Wan Kenobi":
//
// "jedi/Obi-Wan Kenobi" does not contain "jedi/Yoda", skip.
func exists(db *bolt.DB, path string) error {
	entries, err := entry.ListNames(db)
	if err != nil {
		return err
	}

	parts := strings.Split(path, "/")
	n := len(parts) - 1 // entry name index
	name := parts[n]    // name without folders

	for _, e := range entries {
		if strings.Contains(e.Name, path) {
			entryName := strings.Split(e.Name, "/")[n]

			if entryName == name {
				return errors.Errorf("already exists an entry or folder named %q, use <kure edit> to edit", path)
			}
		}
	}

	return nil
}

// formatExpiration returns name and expires fields formatted.
func formatExpiration(expires string) (string, error) {
	expires = strings.ToLower(expires)

	switch expires {
	case "never", "", " ", "0", "0s":
		expires = "Never"

	default:
		expires = strings.ReplaceAll(expires, "-", "/")

		// If the first format fails, try the second
		exp, err := time.Parse("2006/01/02", expires)
		if err != nil {
			exp, err = time.Parse("02/01/2006", expires)
			if err != nil {
				return "", errors.New("invalid time format. Valid formats: d/m/y or y/m/d")
			}
		}

		expires = exp.Format(time.RFC1123Z)
	}

	return expires, nil
}
