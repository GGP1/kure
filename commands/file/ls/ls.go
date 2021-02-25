package ls

import (
	"fmt"
	"strings"
	"time"

	"github.com/GGP1/kure/auth"
	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/file"
	"github.com/GGP1/kure/orderedmap"
	"github.com/GGP1/kure/pb"
	"github.com/GGP1/kure/tree"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

const (
	_ = 1 << (10 * iota)
	// KB - 1024 bytes
	KB
	// MB - 1048576 bytes
	MB
	// GB - 1073741824 bytes
	GB
	// TB - 1099511627776 bytes
	TB
)

var example = `
* List a file and copy its content to the clipboard
kure file ls Sample -c

* Filter files by name
kure file ls Sample -f

* List all files
kure file ls`

type lsOptions struct {
	filter bool
}

// NewCmd returns a new command.
func NewCmd(db *bolt.DB) *cobra.Command {
	opts := lsOptions{}

	cmd := &cobra.Command{
		Use:     "ls <name>",
		Short:   "List files",
		Example: example,
		Args:    cmdutil.MustExistLs(db, cmdutil.File),
		PreRunE: auth.Login(db),
		RunE:    runLs(db, &opts),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset variables (session)
			opts = lsOptions{}
		},
	}

	cmd.Flags().BoolVarP(&opts.filter, "filter", "f", false, "filter by name")

	return cmd
}

func runLs(db *bolt.DB, opts *lsOptions) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")
		name = cmdutil.NormalizeName(name)

		// List all
		if name == "" {
			files, err := file.ListNames(db)
			if err != nil {
				return err
			}

			tree.Print(files)
			return nil
		}

		// Filter by name
		if opts.filter {
			files, err := file.ListNames(db)
			if err != nil {
				return err
			}

			var filtered []string
			for _, file := range files {
				if strings.Contains(file, name) {
					filtered = append(filtered, file)
				}
			}

			if len(filtered) == 0 {
				return errors.New("no files were found")
			}

			tree.Print(filtered)
			return nil
		}

		// List one
		f, err := file.GetCheap(db, name)
		if err != nil {
			return err
		}

		printFile(f)
		return nil
	}
}

func printFile(f *pb.FileCheap) {
	parts := strings.Split(f.Name, "/")
	path := strings.Join(parts[:len(parts)-1], "/")
	bytes := f.Size
	size := fmt.Sprintf("%d bytes", bytes)
	createdAt := time.Unix(f.CreatedAt, 0)
	updatedAt := time.Unix(f.UpdatedAt, 0)

	switch {
	case bytes >= TB:
		size = fmt.Sprintf("%d TB", bytes/TB)
	case bytes >= GB:
		size = fmt.Sprintf("%d GB", bytes/GB)
	case bytes >= MB:
		size = fmt.Sprintf("%d MB", bytes/MB)
	case bytes >= KB:
		size = fmt.Sprintf("%d KB", bytes/KB)
	}

	mp := orderedmap.New()
	mp.Set("Path", "/"+path)
	mp.Set("Size", size)
	mp.Set("Created at", fmt.Sprintf("%v", createdAt))
	if !updatedAt.IsZero() {
		mp.Set("Updated at", fmt.Sprintf("%v", updatedAt))
	}

	box := cmdutil.BuildBox(f.Name, mp)
	fmt.Println("\n" + box)
}
