package edit

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/GGP1/kure/auth"
	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/pb"
	"github.com/GGP1/kure/sig"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

const example = `
* Edit using the standard input
kure edit Sample

* Edit using the text editor
kure edit Sample -i`

type editOptions struct {
	interactive bool
}

// NewCmd returns a new command.
func NewCmd(db *bolt.DB) *cobra.Command {
	opts := editOptions{}

	cmd := &cobra.Command{
		Use:   "edit <name>",
		Short: "Edit an entry",
		Long: `Edit an entry. 
		
If the name is edited, Kure will remove the entry with the old name and create one with the new name.`,
		Example: example,
		Args:    cmdutil.MustExist(db, cmdutil.Entry),
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

		oldEntry, err := entry.Get(db, name)
		if err != nil {
			return err
		}

		// Format the expires fields so it's easy to read
		if oldEntry.Expires != "Never" {
			// This never fails as the field "expires" is parsed before stored
			expires, _ := time.Parse(time.RFC1123Z, oldEntry.Expires)
			oldEntry.Expires = expires.Format("02/01/2006")
		}

		if opts.interactive {
			return useTextEditor(db, oldEntry)
		}

		return useStdin(db, os.Stdin, oldEntry)
	}
}

func createTempFile(e *pb.Entry) (string, error) {
	f, err := os.CreateTemp("", "*.json")
	if err != nil {
		return "", errors.Wrap(err, "creating temporary file")
	}

	content, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return "", errors.Wrap(err, "encoding entry")
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
func readTmpFile(filename string) (*pb.Entry, error) {
	var e pb.Entry

	f, err := os.Open(filename)
	if err != nil {
		return nil, errors.Wrap(err, "reading file")
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(&e); err != nil {
		return nil, errors.Wrap(err, "decoding file")
	}

	return &e, nil
}

// updateEntry takes the name of the entry that's being edited to check if the name was
// changed. If it was, it will remove the old one.
func updateEntry(db *bolt.DB, name string, e *pb.Entry) error {
	if e.Name == "" {
		return cmdutil.ErrInvalidName
	}

	// Verify that the "expires" field has a valid format
	expires, err := cmdutil.FmtExpires(e.Expires)
	if err != nil {
		return err
	}

	name = cmdutil.NormalizeName(name)
	e.Name = cmdutil.NormalizeName(e.Name)
	e.Expires = expires

	if err := entry.Update(db, name, e); err != nil {
		return err
	}

	fmt.Printf("%q updated\n", e.Name)
	return nil
}

func useStdin(db *bolt.DB, r io.Reader, oldEntry *pb.Entry) error {
	fmt.Println("Type '-' to clear the field (except Name and Password) or leave blank to use the current value")
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

	newEntry := &pb.Entry{}
	newEntry.Name = scanln("Name", oldEntry.Name)
	newEntry.Username = scanln("Username", oldEntry.Username)

	enclave, err := auth.AskPassword("Password", true)
	if err != nil {
		if err == auth.ErrInvalidPassword {
			// Assume the user typed an empty string to not modify the password
			newEntry.Password = oldEntry.Password
		} else {
			return err
		}
	} else {
		pwd, err := enclave.Open()
		if err != nil {
			return errors.Wrap(err, "opening enclave")
		}

		newEntry.Password = pwd.String()
	}

	newEntry.URL = scanln("URL", oldEntry.URL)
	newEntry.Expires = scanln("Expires", oldEntry.Expires)

	notes := cmdutil.Scanlns(reader, fmt.Sprintf("Notes [%s]", oldEntry.Notes))
	if notes == "" {
		notes = oldEntry.Notes
	} else if notes == "-" {
		notes = ""
	}
	newEntry.Notes = notes

	return updateEntry(db, oldEntry.Name, newEntry)
}

func useTextEditor(db *bolt.DB, oldEntry *pb.Entry) error {
	editor := cmdutil.SelectEditor()
	bin, err := exec.LookPath(editor)
	if err != nil {
		return errors.Errorf("executable %q not found", editor)
	}

	filename, err := createTempFile(oldEntry)
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

	// Read the file and update the entry
	newEntry, err := readTmpFile(filename)
	if err != nil {
		return err
	}

	rmTabs := func(old string) string {
		return strings.ReplaceAll(old, "\t", "")
	}
	newEntry.Name = rmTabs(newEntry.Name)
	newEntry.Username = rmTabs(newEntry.Username)
	newEntry.URL = rmTabs(newEntry.URL)
	newEntry.Notes = rmTabs(newEntry.Notes)

	return updateEntry(db, oldEntry.Name, newEntry)
}
