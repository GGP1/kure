package gen

import (
	"fmt"
	"math"

	cmdutil "github.com/GGP1/kure/cmd"

	"github.com/GGP1/atoll"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	qr, repeat       bool
	length           uint64
	format           []int
	include, exclude string
)

var errInvalidLength = errors.New("error: invalid length")

var example = `
* Generate a random password
kure gen -l 18 -f 1,2,3 -i %&/ -e ? -r

* Generate and show the QR code image
kure gen -l 20 -q`

// NewCmd returns a new command.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gen [-l length] [-f format] [-i include] [-e exclude] [-r repeat] [-q qr]",
		Short: "Generate a random password",
		Long: `Generate a random password.
		
When using [-q qr] flag, make sure the terminal is bigger than the image or it will spoil.`,
		Example: example,
		RunE:    runGen(),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset flags defaults (session)
			length = 0
			format = nil
			include, exclude = "", ""
			qr, repeat = false, true
		},
	}

	f := cmd.Flags()
	f.Uint64VarP(&length, "length", "l", 0, "password length")
	f.IntSliceVarP(&format, "format", "f", nil, "password format")
	f.StringVarP(&include, "include", "i", "", "characters to include in the password")
	f.StringVarP(&exclude, "exclude", "e", "", "characters to exclude from the password")
	f.BoolVarP(&repeat, "repeat", "r", true, "allow duplicated characters or not (default false)")
	f.BoolVarP(&qr, "qr", "q", false, "show the QR code image on the terminal")

	cmd.AddCommand(phraseSubCmd())

	return cmd
}

func runGen() cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		if length < 1 || length > math.MaxUint64 {
			return errInvalidLength
		}

		if format == nil {
			passFormat := viper.GetIntSlice("entry.format")
			if len(passFormat) == 0 {
				return errors.New("error: please specify a format")
			}
			format = passFormat
		}

		if entryRepeat := viper.GetBool("entry.repeat"); entryRepeat != false {
			repeat = entryRepeat
		}

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

		password, err := p.Generate()
		if err != nil {
			return err
		}

		if qr {
			if err := cmdutil.DisplayQRCode(password); err != nil {
				return err
			}
		}

		entropy := calculateEntropy(length, uFormat)

		fmt.Printf("Password: %s\nBits of entropy: %.2f\n", password, entropy)
		return nil
	}
}

func calculateEntropy(length uint64, format []uint8) float64 {
	var poolLength uint16

	for _, level := range format {
		// https://github.com/GGP1/atoll#entropy
		switch level {
		case 1:
			poolLength += 26
		case 2:
			poolLength += 26
		case 3:
			poolLength += 10
		case 4:
			poolLength++
		case 5:
			poolLength += 32
		case 6:
			poolLength += 95
		}
	}

	entropy := math.Log2(math.Pow(float64(poolLength), float64(length)))

	return entropy
}
