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
	"github.com/awnumar/memguard"

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
	errInvalidName   = errors.New("invalid name")
	errInvalidLength = errors.New("invalid length")
)

var example = `
* Add an entry using a custom password
kure add entryName -c

* Add an entry generating a random password
kure add entryName -l 27 -f 1,2,3,4,5 -i & -e / -r`

// NewCmd returns a new command.
func NewCmd(db *bolt.DB, r io.Reader) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add <name>",
		Short:   "Add an entry",
		Aliases: []string{"create", "new"},
		Example: example,
		PreRunE: cmdutil.RequirePassword(db),
		RunE:    runAdd(db, r),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset flags (session)
			custom, repeat = false, true
			length = 0
			format = nil
			include, exclude = "", ""
		},
	}

	cmd.AddCommand(phraseSubCmd(db, r))

	f := cmd.Flags()
	f.BoolVarP(&custom, "custom", "c", false, "use a custom password")
	f.Uint64VarP(&length, "length", "l", 0, "password length")
	f.IntSliceVarP(&format, "format", "f", nil, "password format (1,2,3,4,5)")
	f.StringVarP(&include, "include", "i", "", "string of characters to include in the password")
	f.StringVarP(&exclude, "exclude", "e", "", "string of characters to exclude from the password")
	f.BoolVarP(&repeat, "repeat", "r", true, "character repetition")

	return cmd
}

func runAdd(db *bolt.DB, r io.Reader) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")
		if name == "" {
			return errInvalidName
		}

		if !custom {
			if length < 1 || length > math.MaxUint64 {
				return errInvalidLength
			}
		}

		name = strings.TrimSpace(strings.ToLower(name))

		if err := cmdutil.Exists(db, name, "entry"); err != nil {
			return err
		}

		lockedBuf, e, err := input(db, r, name, custom)
		if err != nil {
			return err
		}

		if !custom {
			// Generate random password
			e.Password, err = genPassword(length, format, include, exclude, repeat)
			if err != nil {
				return err
			}
		}

		if err := entry.Create(db, lockedBuf, e); err != nil {
			return err
		}

		fmt.Printf("\n%q added\n", name)
		return nil
	}
}

func input(db *bolt.DB, r io.Reader, name string, custom bool) (*memguard.LockedBuffer, *pb.Entry, error) {
	s := bufio.NewScanner(r)

	lockedBuf, entry := pb.SecureEntry()
	entry.Name = name
	entry.Username = cmdutil.Scan(s, "Username")
	if custom {
		entry.Password = cmdutil.Scan(s, "Password")
	}
	entry.URL = cmdutil.Scan(s, "URL")
	entry.Notes = cmdutil.Scanlns(s, "Notes")

	expires := cmdutil.Scan(s, "Expires")

	if err := s.Err(); err != nil {
		return nil, nil, errors.Wrap(err, "scanning failed")
	}

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
				return nil, nil, errors.New("invalid time format. Valid formats: d/m/y or y/m/d")
			}
		}

		expires = exp.Format(time.RFC1123Z)
	}

	entry.Expires = expires
	memguard.WipeBytes([]byte(expires))

	return lockedBuf, entry, nil
}

// genPassword returns a customized random password or an error.
func genPassword(length uint64, format []int, include, exclude string, repeat bool) (string, error) {
	// If the user didn't specify format levels, use default
	if len(format) == 0 {
		f := viper.GetIntSlice("entry.format")
		if len(f) == 0 {
			return "", errors.New("please specify a format")
		}
		format = f
	}

	// Convert int slice to uint8 slice
	uFormat := make([]uint8, len(format))
	for i := range format {
		uFormat[i] = uint8(format[i])
	}

	p := &atoll.Password{
		Length:  length,
		Format:  uFormat,
		Include: include,
		Exclude: exclude,
		Repeat:  repeat,
	}

	password, err := atoll.NewSecret(p)
	if err != nil {
		return "", err
	}

	return password, nil
}
