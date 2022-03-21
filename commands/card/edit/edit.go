package edit

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/GGP1/kure/auth"
	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/card"
	"github.com/GGP1/kure/pb"
	"github.com/GGP1/kure/sig"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

const example = `
* Edit using the standard input
kure card edit Sample

* Edit using the text editor
kure card edit Sample -i`

type editOptions struct {
	interactive bool
}

// NewCmd returns a new command.
func NewCmd(db *bolt.DB) *cobra.Command {
	opts := editOptions{}

	cmd := &cobra.Command{
		Use:   "edit <name>",
		Short: "Edit a card",
		Long: `Edit a card.

If the name is edited, Kure will remove the old card and create one with the new name.`,
		Example: example,
		Args:    cmdutil.MustExist(db, cmdutil.Card),
		PreRunE: auth.Login(db),
		RunE:    runEdit(db, &opts),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset variables (session)
			opts = editOptions{}
		},
	}

	cmd.Flags().BoolVarP(&opts.interactive, "it", "i", false, "use the text editor")

	return cmd
}

func runEdit(db *bolt.DB, opts *editOptions) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")
		name = cmdutil.NormalizeName(name)

		oldCard, err := card.Get(db, name)
		if err != nil {
			return err
		}

		if opts.interactive {
			return useTextEditor(db, oldCard)
		}

		return useStdin(db, os.Stdin, oldCard)
	}
}

func createTempFile(c *pb.Card) (string, error) {
	f, err := os.CreateTemp("", "*.json")
	if err != nil {
		return "", errors.Wrap(err, "creating temporary file")
	}

	content, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return "", errors.Wrap(err, "encoding card")
	}

	if _, err := f.Write(content); err != nil {
		return "", errors.Wrap(err, "writing temporary file")
	}

	if err := f.Close(); err != nil {
		return "", errors.Wrap(err, "closing temporary file")
	}

	return f.Name(), nil
}

// readTmpFile reads the modified file and formats the card.
func readTmpFile(filename string) (*pb.Card, error) {
	var c pb.Card

	f, err := os.Open(filename)
	if err != nil {
		return nil, errors.Wrap(err, "reading file")
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(&c); err != nil {
		return nil, errors.Wrap(err, "decoding file")
	}

	return &c, nil
}

// updateCard takes the name of the card that's being edited to check if the name was
// changed. If it was, it will remove the old one.
func updateCard(db *bolt.DB, name string, c *pb.Card) error {
	if c.Name == "" {
		return cmdutil.ErrInvalidName
	}

	name = cmdutil.NormalizeName(name)
	c.Name = cmdutil.NormalizeName(c.Name)

	if err := card.Update(db, name, c); err != nil {
		return err
	}

	fmt.Println(c.Name, "updated")
	return nil
}

func useStdin(db *bolt.DB, r io.Reader, oldCard *pb.Card) error {
	fmt.Println("Type '-' to clear the field or leave blank to use the current value")
	reader := bufio.NewReader(r)

	scanln := func(field, value string) string {
		input := cmdutil.Scanln(reader, fmt.Sprintf("%s [%s]", field, value))
		if input == "-" {
			return ""
		} else if input != "" {
			return input
		}
		return value
	}

	newCard := &pb.Card{
		Name:         scanln("Name", oldCard.Name),
		Type:         scanln("Type", oldCard.Type),
		Number:       scanln("Number", oldCard.Number),
		SecurityCode: scanln("Security code", oldCard.SecurityCode),
		ExpireDate:   scanln("Expire date", oldCard.ExpireDate),
	}

	notes := cmdutil.Scanlns(reader, fmt.Sprintf("Notes [%s]", oldCard.Notes))
	if notes == "" {
		notes = oldCard.Notes
	} else if notes == "-" {
		notes = ""
	}
	newCard.Notes = notes

	return updateCard(db, oldCard.Name, newCard)
}

func useTextEditor(db *bolt.DB, oldCard *pb.Card) error {
	editor := cmdutil.SelectEditor()
	bin, err := exec.LookPath(editor)
	if err != nil {
		return errors.Errorf("executable %q not found", editor)
	}

	filename, err := createTempFile(oldCard)
	if err != nil {
		return err
	}

	sig.Signal.AddCleanup(func() error { return cmdutil.Erase(filename) })
	defer cmdutil.Erase(filename)

	// Open the temporary file with the selected text editor
	edit := exec.Command(bin, filename)
	edit.Stdin = os.Stdin
	edit.Stdout = os.Stdout

	if err := edit.Start(); err != nil {
		return errors.Wrapf(err, "running %s", editor)
	}

	done := make(chan struct{}, 1)
	errCh := make(chan error, 1)
	go cmdutil.WatchFile(filename, done, errCh)

	// Block until an event is received or an error occurs
	select {
	case <-done:
	case err := <-errCh:
		return err
	}

	if err := edit.Wait(); err != nil {
		return err
	}

	newCard, err := readTmpFile(filename)
	if err != nil {
		return err
	}

	rmTabs := func(old string) string {
		return strings.ReplaceAll(old, "\t", "")
	}
	newCard.Name = rmTabs(newCard.Name)
	newCard.Type = rmTabs(newCard.Type)
	newCard.Number = rmTabs(newCard.Number)
	newCard.SecurityCode = rmTabs(newCard.SecurityCode)
	newCard.ExpireDate = rmTabs(newCard.ExpireDate)
	newCard.Notes = rmTabs(newCard.Notes)

	return updateCard(db, oldCard.Name, newCard)
}
