package add

import (
	"bufio"
	"encoding/base32"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/GGP1/kure/auth"
	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/totp"
	"github.com/GGP1/kure/pb"
	"github.com/GGP1/kure/terminal"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

const example = `
* Add with setup key
kure 2fa add Sample

* Add with URL
kure 2fa add -u`

type addOptions struct {
	digits int32
	url    bool
}

// NewCmd returns a new command.
func NewCmd(db *bolt.DB, r io.Reader) *cobra.Command {
	opts := addOptions{}
	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Add a two-factor authentication code",
		Long: `Add a two-factor authentication code.

• Using a setup key: services tipically show hyperlinked text like "Enter manually" or "Enter this text code", copy the hexadecimal code given and submit it when requested.

• Using a URL: extract the URL encoded in the QR code given and submit it when requested. Format: otpauth://totp/{service}:{account}?secret={secret}.`,
		Example: example,
		Args: func(cmd *cobra.Command, args []string) error {
			// When adding with URL the name won't be specified
			if opts.url {
				return nil
			}

			return cmdutil.MustNotExist(db, cmdutil.TOTP)(cmd, args)
		},
		PreRunE: auth.Login(db),
		RunE:    runAdd(db, r, &opts),
		PostRun: func(cmd *cobra.Command, args []string) {
			opts = addOptions{
				digits: 6,
			}
		},
	}

	f := cmd.Flags()
	f.Int32VarP(&opts.digits, "digits", "d", 6, "TOTP length {6|7|8}")
	f.BoolVarP(&opts.url, "url", "u", false, "add using a URL")

	return cmd
}

func runAdd(db *bolt.DB, r io.Reader, opts *addOptions) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")
		name = cmdutil.NormalizeName(name)

		if opts.url {
			return addWithURL(db, r)
		}

		return addWithKey(db, r, name, opts.digits)
	}
}

func addWithKey(db *bolt.DB, r io.Reader, name string, digits int32) error {
	if digits < 6 || digits > 8 {
		return errors.Errorf("invalid digits number [%d], it must be either 6, 7 or 8", digits)
	}

	key := terminal.Scanln(bufio.NewReader(r), "Key")
	// Adjust key
	key = strings.ReplaceAll(key, " ", "")
	key += strings.Repeat("=", -len(key)&7)
	key = strings.ToUpper(key)

	if _, err := base32.StdEncoding.DecodeString(key); err != nil {
		return errors.Wrap(err, "invalid key")
	}

	return createTOTP(db, name, key, digits)
}

// addWithURL creates a new TOTP using the values passed in the url.
func addWithURL(db *bolt.DB, r io.Reader) error {
	uri := terminal.Scanln(bufio.NewReader(r), "URL")
	URL, err := url.Parse(uri)
	if err != nil {
		return errors.Wrap(err, "parsing url")
	}

	query := URL.Query()
	if err := validateURL(URL, query); err != nil {
		return err
	}

	name := getName(URL.Path)
	if err := cmdutil.Exists(db, name, cmdutil.TOTP); err != nil {
		return err
	}

	digits := stringDigits(query.Get("digits"))
	secret := query.Get("secret")
	if _, err := base32.StdEncoding.DecodeString(secret); err != nil {
		return errors.Wrap(err, "invalid secret")
	}

	return createTOTP(db, name, secret, digits)
}

func createTOTP(db *bolt.DB, name, key string, digits int32) error {
	t := &pb.TOTP{
		Name:   name,
		Raw:    key,
		Digits: digits,
	}

	if err := totp.Create(db, t); err != nil {
		return err
	}

	fmt.Printf("\n%q TOTP added\n", name)
	return nil
}

// getName extracts the service name from the URL.
func getName(path string) string {
	// Given "/Example:account@mail.com", return "Example"
	path = strings.TrimPrefix(path, "/")
	name, _, _ := strings.Cut(path, ":")
	return cmdutil.NormalizeName(name)
}

// stringDigits returns the digits to use depending on the string passed.
func stringDigits(digits string) int32 {
	switch digits {
	case "8":
		return 8
	case "7":
		return 7
	default:
		return 6
	}
}

func validateURL(URL *url.URL, query url.Values) error {
	if URL.Scheme != "otpauth" {
		return errors.New("invalid scheme, must be otpauth")
	}

	if URL.Host != "totp" {
		return errors.New("invalid host, must be totp")
	}

	algorithm := query.Get("algorithm")
	if algorithm != "" && algorithm != "SHA1" {
		return errors.New("invalid algorithm, must be SHA1")
	}

	period := query.Get("period")
	if period != "" && period != "30" {
		return errors.New("invalid period, must be 30 seconds")
	}

	return nil
}
