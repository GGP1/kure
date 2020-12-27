package file

import (
	"bytes"
	"io"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/db/file"

	"github.com/atotto/clipboard"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var copy bool

var catExample = `
* Write one file
kure cat fileName

* Write one file and copy content to the clipboard:
kure cat fileName -c

* Write multiple files
kure cat file1 file2 file3`

func catSubCmd(db *bolt.DB, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cat <name>",
		Short:   "Read file and write to standard output",
		Example: catExample,
		PreRunE: cmdutil.RequirePassword(db),
		RunE:    runCat(db, w),
	}

	cmd.Flags().BoolVarP(&copy, "copy", "c", false, "copy file content to the clipboard")

	return cmd
}

func runCat(db *bolt.DB, w io.Writer) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		for _, name := range args {
			lockedBuf, f, err := file.Get(db, name)
			if err != nil {
				return err
			}

			if copy {
				if err := clipboard.WriteAll(string(f.Content)); err != nil {
					return errors.Wrap(err, "failed writing to clipboard")
				}
			}

			buf := bytes.NewReader(f.Content)
			lockedBuf.Destroy()

			if _, err := io.Copy(w, buf); err != nil {
				return errors.Wrap(err, "failed copying file content")
			}
		}
		return nil
	}
}
