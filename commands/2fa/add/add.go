package add

import (
	"bufio"
	"encoding/base32"
	"fmt"
	"io"
	"strings"

	"github.com/GGP1/kure/auth"
	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/totp"
	"github.com/GGP1/kure/pb"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var example = `
kure 2fa add Sample`

type addOptions struct {
	digits int
}

// NewCmd returns a new command.
func NewCmd(db *bolt.DB, r io.Reader) *cobra.Command {
	var opts addOptions

	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Add a two-factor authentication code",
		Long: `Add a two-factor authentication code. The name must be one already used by an entry.

Services tipically show an hyperlinked "Enter manually", "Enter this text code" or similar messages, copy the hexadecimal code given and submit it when requested by Kure. After this, your entry will have a synchronized token with the service.`,
		Example: example,
		Args:    cmdutil.MustExist(db, cmdutil.Entry), // There must exist an entry with the same name
		PreRunE: auth.Login(db),
		RunE:    opts.runAdd(db, r),
		PostRun: func(cmd *cobra.Command, args []string) {
			opts = addOptions{}
		},
	}

	cmd.Flags().IntVarP(&opts.digits, "digits", "d", 6, "TOTP length {6|7|8}")

	return cmd
}

func (opts *addOptions) runAdd(db *bolt.DB, r io.Reader) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")
		name = cmdutil.NormalizeName(name)

		if opts.digits < 6 || opts.digits > 8 {
			return errors.Errorf("invalid digits number [%d], it must be either 6, 7 or 8", opts.digits)
		}

		key := cmdutil.Scanln(bufio.NewReader(r), "Key")
		// Adjust key
		key = strings.ReplaceAll(key, " ", "")
		key += strings.Repeat("=", -len(key)&7)
		key = strings.ToUpper(key)

		if _, err := base32.StdEncoding.DecodeString(key); err != nil {
			return errors.Wrap(err, "invalid key")
		}

		t := &pb.TOTP{
			Name:   name,
			Raw:    key,
			Digits: int32(opts.digits),
		}

		if err := totp.Create(db, t); err != nil {
			return err
		}

		fmt.Printf("\n%q TOTP added\n", name)
		return nil
	}
}
