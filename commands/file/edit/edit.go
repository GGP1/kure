package edit

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/GGP1/kure/auth"
	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/file"
	"github.com/GGP1/kure/pb"
	"github.com/GGP1/kure/sig"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var example = `
kure file edit Sample -e nvim`

type editOptions struct {
	editor string
}

// NewCmd returns a new command.
func NewCmd(db *bolt.DB) *cobra.Command {
	opts := editOptions{}

	cmd := &cobra.Command{
		Use:   "edit <name>",
		Short: "Edit a file",
		Long: `Edit a file.

Caution: a temporary file is created with a random name, it will be erased right after the first save but this is still insecure.

Command procedure:
1. Create a temporary file with the content of the stored file.
2. Execute the text editor to edit it.
3. Wait for it to be saved.
4. Read its content and update kure's file.
5. Overwrite the initially created file with random bytes and delete it.

Tips:
• Some editors will require to exit to modify the file.
• Use an image editor command to edit images.`,
		Example: example,
		Args:    cmdutil.MustExist(db, cmdutil.File),
		PreRunE: auth.Login(db),
		RunE:    runEdit(db, &opts),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset variables (session)
			opts = editOptions{}
		},
	}

	cmd.Flags().StringVarP(&opts.editor, "editor", "e", "", "file editor command")

	return cmd
}

func runEdit(db *bolt.DB, opts *editOptions) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")
		name = cmdutil.NormalizeName(name)

		oldFile, err := file.Get(db, name)
		if err != nil {
			return err
		}

		if opts.editor == "" {
			opts.editor = cmdutil.SelectEditor()
		}
		bin, err := exec.LookPath(opts.editor)
		if err != nil {
			return errors.Errorf("executable %q not found", opts.editor)
		}

		filename, err := createTempFile(filepath.Ext(oldFile.Name), oldFile.Content)
		if err != nil {
			return err
		}

		sig.Signal.AddCleanup(func() error { return cmdutil.Erase(filename) })
		defer func() error { return cmdutil.Erase(filename) }()

		// Open the temporary file with the selected text editor
		edit := exec.Command(bin, filename)
		edit.Stdin = os.Stdin
		edit.Stdout = os.Stdout

		if err := edit.Start(); err != nil {
			return errors.Wrapf(err, "running %s", opts.editor)
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

		// Read the file created and update the "kure" file
		if err := readAndUpdate(db, name, filename, oldFile); err != nil {
			return err
		}

		fmt.Printf("\n%q updated\n", name)
		return nil
	}
}

func createTempFile(ext string, content []byte) (string, error) {
	f, err := os.CreateTemp("", "*."+ext)
	if err != nil {
		return "", errors.Wrap(err, "creating temporary file")
	}

	if _, err := f.Write(content); err != nil {
		return "", errors.Wrap(err, "writing temporary file")
	}

	if err := f.Close(); err != nil {
		return "", errors.Wrap(err, "closing temporary file")
	}

	return f.Name(), nil
}

// readAndUpdate reads the modified file and updates the kure file content.
func readAndUpdate(db *bolt.DB, name, filename string, old *pb.File) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return errors.Wrap(err, "reading file")
	}
	content = bytes.TrimSpace(content)

	new := &pb.File{
		Name:      old.Name,
		Content:   content,
		Size:      int64(len(content)),
		CreatedAt: old.CreatedAt,
		UpdatedAt: time.Now().Unix(),
	}

	if err := file.Create(db, new); err != nil {
		return errors.Wrap(err, "updating file")
	}

	return nil
}
