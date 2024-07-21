package rotate

import (
	"fmt"
	"strings"
	"time"

	"github.com/GGP1/atoll"
	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/terminal"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

const example = `
* Rotate a password by generating a random one that uses the same parameters
kure rotate Sample

* Rotate a password using a new custom one
kure rotate Sample -c`

type rotateOptions struct {
	copy, custom bool
	timeout      time.Duration
}

// NewCmd returns a new command.
func NewCmd(db *bolt.DB) *cobra.Command {
	opts := rotateOptions{}
	cmd := &cobra.Command{
		Use:     "rotate <name>",
		Short:   "Rotate an entry's password",
		Example: example,
		Args:    cmdutil.MustExist(db, cmdutil.Entry),
		RunE:    runRotate(db, &opts),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset variables (session)
			opts = rotateOptions{}
		},
	}

	f := cmd.Flags()
	f.BoolVarP(&opts.copy, "copy", "c", false, "copy new password to clipboard")
	f.BoolVar(&opts.custom, "custom", false, "use a custom password")
	f.DurationVarP(&opts.timeout, "timeout", "t", 0, "clipboard clearing timeout")

	return cmd
}

func runRotate(db *bolt.DB, opts *rotateOptions) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")
		name = cmdutil.NormalizeName(name)

		e, err := entry.Get(db, name)
		if err != nil {
			return err
		}

		fmt.Printf("Old password: %s\n", e.Password)

		if opts.custom {
			e.Password, err = readPassword()
			if err != nil {
				return err
			}
		} else {
			// Generate a password with the same parameters as the old one
			secret := atoll.SecretFromString(e.Password)
			password, err := secret.Generate()
			if err != nil {
				return err
			}

			e.Password = string(password)
		}

		if err := entry.Update(db, name, e); err != nil {
			return err
		}

		if opts.copy {
			return cmdutil.WriteClipboard(cmd, opts.timeout, "Password", e.Password)
		}

		fmt.Printf("\n%q password rotated\n", name)
		return nil
	}
}

func readPassword() (string, error) {
	enclave, err := terminal.ScanPassword("New password", true)
	if err != nil {
		return "", err
	}

	pwd, err := enclave.Open()
	if err != nil {
		return "", errors.Wrap(err, "opening enclave")
	}

	return pwd.String(), nil
}
