package phrase

import (
	"fmt"
	"math"
	"strings"

	cmdutil "github.com/GGP1/kure/commands"

	"github.com/GGP1/atoll"

	"github.com/awnumar/memguard"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	thousand    = math.Pow(10, 3)
	million     = math.Pow(10, 6)
	billion     = math.Pow(10, 9)
	trillion    = math.Pow(10, 12)
	quadrillion = math.Pow(10, 15)
	quintillion = math.Pow(10, 18)
	sextillion  = math.Pow(10, 21)
)

const (
	minute     = 60
	hour       = minute * 60
	day        = hour * 24
	month      = day * 30
	year       = month * 12
	decade     = year * 10
	century    = decade * 10
	millennium = century * 10
)

const example = `
* Generate a random passphrase
kure gen phrase -l 8 -L WordList -s &

* Generate and show QR code
kure gen phrase -l 5 -q

* Generate, copy and mute standard output
kure gen -l 7 -cm`

type phraseOptions struct {
	list, separator string
	incl, excl      []string
	copy, mute, qr  bool
	length          uint64
}

// NewCmd returns a new command.
func NewCmd() *cobra.Command {
	opts := phraseOptions{}

	cmd := &cobra.Command{
		Use:   "phrase",
		Short: "Generate a random passphrase",
		Long: `Generate a random passphrase.
		
Keyspace is the number of possible combinations of the passphrase.
Average time taken to crack is based on a brute force attack scenario where the guesses per second is 1 trillion.`,
		Aliases: []string{"passphrase", "p"},
		Example: example,
		RunE:    runPhrase(&opts),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset variables (session)
			opts = phraseOptions{
				separator: " ",
			}
		},
	}

	f := cmd.Flags()
	f.BoolVarP(&opts.copy, "copy", "c", false, "copy the passphrase to the clipboard")
	f.Uint64VarP(&opts.length, "length", "l", 0, "number of words")
	f.StringVarP(&opts.separator, "separator", "s", " ", "character that separates each word")
	f.StringVarP(&opts.list, "list", "L", "WordList", "passphrase list used {NoList|WordList|SyllableList}")
	f.StringSliceVarP(&opts.incl, "include", "i", nil, "words to include in the passphrase")
	f.StringSliceVarP(&opts.excl, "exclude", "e", nil, "words to exclude from the passphrase")
	f.BoolVarP(&opts.qr, "qr", "q", false, "show QR code on terminal")
	f.BoolVarP(&opts.mute, "mute", "m", false, "mute standard output when the passphrase is copied")

	return cmd
}

func runPhrase(opts *phraseOptions) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		if opts.length < 1 || opts.length > math.MaxUint64 {
			return cmdutil.ErrInvalidLength
		}

		// Use Wordlist list as default
		l := atoll.WordList
		if opts.list != "" {
			opts.list = strings.ReplaceAll(opts.list, " ", "")

			switch strings.ToLower(opts.list) {
			case "nolist", "no":
				l = atoll.NoList

			case "wordlist", "word":
				// Do nothing as it's the default

			case "syllablelist", "syllable":
				l = atoll.SyllableList

			default:
				return errors.Errorf("invalid list: %q", opts.list)
			}
		}

		p := &atoll.Passphrase{
			Length:    opts.length,
			Separator: opts.separator,
			Include:   opts.incl,
			Exclude:   opts.excl,
			List:      l,
		}

		passphrase, err := atoll.NewSecret(p)
		if err != nil {
			return err
		}

		phraseBuf := memguard.NewBufferFromBytes([]byte(passphrase))
		memguard.WipeBytes([]byte(passphrase))
		defer phraseBuf.Destroy()

		if opts.qr {
			if err := cmdutil.DisplayQRCode(phraseBuf.String()); err != nil {
				return err
			}
		}

		entropy := p.Entropy()
		keyspace, timeToCrack := FormatSecretSecurity(atoll.Keyspace(p), atoll.SecondsToCrack(p)/2)

		if !opts.mute || !opts.copy {
			fmt.Printf(`Passphrase: %s
	
Entropy: %.2f bits
Keyspace: %s
Average time taken to crack: %s
`, phraseBuf.String(), entropy, keyspace, timeToCrack)
		}

		if opts.copy {
			if err := cmdutil.WriteClipboard(cmd, 0, "Passphrase", phraseBuf.String()); err != nil {
				return err
			}
		}
		return nil
	}
}

// FormatSecretSecurity makes the data passed easier to read.
func FormatSecretSecurity(keyspace, avgTimeToCrack float64) (space, time string) {
	if math.IsInf(keyspace, 1) {
		return "+Inf", "+Inf"
	}

	space = fmt.Sprintf("%.2f", keyspace)
	time = fmt.Sprintf("%.2f seconds", avgTimeToCrack)

	switch {
	case keyspace >= sextillion:
		space = fmt.Sprintf("%.2f sextillion", keyspace/sextillion)

	case keyspace >= quintillion:
		space = fmt.Sprintf("%.2f quintillion", keyspace/quintillion)

	case keyspace >= quadrillion:
		space = fmt.Sprintf("%.2f quadrillion", keyspace/quadrillion)

	case keyspace >= trillion:
		space = fmt.Sprintf("%.2f trillion", keyspace/trillion)

	case keyspace >= billion:
		space = fmt.Sprintf("%.2f billion", keyspace/billion)

	case keyspace >= million:
		space = fmt.Sprintf("%.2f million", keyspace/million)

	case keyspace >= thousand:
		space = fmt.Sprintf("%.2f thousand", keyspace/thousand)
	}

	switch {
	case avgTimeToCrack >= millennium:
		time = fmt.Sprintf("%.2f millenniums", avgTimeToCrack/millennium)

	case avgTimeToCrack >= century:
		time = fmt.Sprintf("%.2f centuries", avgTimeToCrack/century)

	case avgTimeToCrack >= decade:
		time = fmt.Sprintf("%.2f decades", avgTimeToCrack/decade)

	case avgTimeToCrack >= year:
		time = fmt.Sprintf("%.2f years", avgTimeToCrack/year)

	case avgTimeToCrack >= month:
		time = fmt.Sprintf("%.2f months", avgTimeToCrack/month)

	case avgTimeToCrack >= day:
		time = fmt.Sprintf("%.2f days", avgTimeToCrack/day)

	case avgTimeToCrack >= hour:
		time = fmt.Sprintf("%.2f hours", avgTimeToCrack/hour)

	case avgTimeToCrack >= minute:
		time = fmt.Sprintf("%.2f minutes", avgTimeToCrack/minute)
	}

	return prettify(space), prettify(time)
}

// prettify adds commas to the number inside the string to make it easier for humans to read.
//
// Example: 51334.21 -> 51,334.21
func prettify(str string) string {
	idx := strings.IndexByte(str, '.')
	num := str[:idx]

	sb := strings.Builder{}
	j := 0
	for i := len(num) - 1; i >= 0; i-- {
		sb.WriteByte(num[j])
		if i%3 == 0 && i != 0 {
			sb.WriteByte(',')
		}
		j++
	}

	// Append remaining characters
	sb.WriteString(str[idx:])
	return sb.String()
}
