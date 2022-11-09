// Package tfa handles two-factor authentication codes.
package tfa

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"github.com/GGP1/kure/auth"
	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/commands/2fa/add"
	"github.com/GGP1/kure/commands/2fa/rm"
	"github.com/GGP1/kure/db/totp"
	"github.com/GGP1/kure/orderedmap"
	"github.com/GGP1/kure/pb"
	"github.com/GGP1/kure/terminal"
	"github.com/GGP1/kure/tree"

	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

const example = `
* List one and copy to the clipboard
kure 2fa Sample -c

* List all
kure 2fa

* Display information about the setup key
kure 2fa Sample -i`

type tfaOptions struct {
	copy, info bool
	timeout    time.Duration
}

// NewCmd returns a new command.
func NewCmd(db *bolt.DB) *cobra.Command {
	opts := tfaOptions{}
	cmd := &cobra.Command{
		Use:   "2fa <name>",
		Short: "List two-factor authentication codes",
		Long: `List two-factor authentication codes.

Use the [-i info] flag to display information about the setup key, it also generates a QR code with the key in URL format that can be scanned by any authenticator.`,
		Example: example,
		Args:    cmdutil.MustExistLs(db, cmdutil.TOTP),
		PreRunE: auth.Login(db),
		RunE:    run2FA(db, &opts),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset variables (session)
			opts = tfaOptions{}
		},
	}

	cmd.AddCommand(add.NewCmd(db, os.Stdin), rm.NewCmd(db, os.Stdin))

	f := cmd.Flags()
	f.BoolVarP(&opts.copy, "copy", "c", false, "copy code to clipboard")
	f.BoolVarP(&opts.info, "info", "i", false, "display information about the setup key")
	f.DurationVarP(&opts.timeout, "timeout", "t", 0, "clipboard clearing timeout")

	return cmd
}

func run2FA(db *bolt.DB, opts *tfaOptions) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")
		name = cmdutil.NormalizeName(name)

		if name == "" {
			totps, err := totp.ListNames(db)
			if err != nil {
				return err
			}

			tree.Print(totps)
			return nil
		}

		t, err := totp.Get(db, name)
		if err != nil {
			return err
		}

		if opts.info {
			return printKeyInfo(t)
		}

		code := GenerateTOTP(t.Raw, time.Now(), int(t.Digits))
		if opts.copy {
			return cmdutil.WriteClipboard(cmd, opts.timeout, "TOTP", code)
		}

		fmt.Println(strings.Title(t.Name), code)
		return nil
	}
}

// GenerateTOTP returns a Time-based One-Time Password code.
func GenerateTOTP(key string, t time.Time, digits int) string {
	// Do not check error as the key was validated when added
	keyBytes, _ := base32.StdEncoding.DecodeString(key)
	h := hmac.New(sha1.New, keyBytes)

	// 30 is the default time-step size in seconds (recommended
	// as per https://tools.ietf.org/html/rfc6238#section-5.2)
	counter := math.Floor(float64(t.Unix()) / 30)
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(counter))
	h.Write(buf)
	sum := h.Sum(nil)

	// "Dynamic truncation" in RFC 4226
	// http://tools.ietf.org/html/rfc4226#section-5.4
	offset := sum[len(sum)-1] & 0xf
	value := int64(((int(sum[offset]) & 0x7f) << 24) |
		((int(sum[offset+1] & 0xff)) << 16) |
		((int(sum[offset+2] & 0xff)) << 8) |
		(int(sum[offset+3]) & 0xff))

	mod := int32(value % int64(math.Pow10(digits)))
	format := fmt.Sprintf("%%0%dd", digits)

	return fmt.Sprintf(format, mod)
}

func printKeyInfo(t *pb.TOTP) error {
	// https://github.com/google/google-authenticator/wiki/Key-Uri-Format
	URL := fmt.Sprintf("otpauth://totp/%s?secret=%s&digits=%d", strings.Title(t.Name), t.Raw, t.Digits)

	if err := terminal.DisplayQRCode(URL); err != nil {
		return err
	}
	mp := orderedmap.New()
	mp.Set("URL", URL)
	mp.Set("Key", t.Raw)
	mp.Set("Digits", fmt.Sprint(t.Digits))

	box := cmdutil.BuildBox(t.Name, mp)
	fmt.Println(box)
	return nil
}
