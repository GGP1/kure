package file

import (
	"fmt"
	"strings"
	"time"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/db/file"
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

var filter bool

var lsExample = `
* List a file
kure file ls fileName

* Filter files by name
kure file ls fileName -f

* List all files
kure file ls`

func lsSubCmd(db *bolt.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ls <name> [-f filter]",
		Short:   "List files",
		Example: lsExample,
		RunE:    runLs(db),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset flags defaults (session)
			filter = false
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.Flags().BoolVarP(&filter, "filter", "f", false, "filter files by name")

	return cmd
}

func runLs(db *bolt.DB) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")
		if name == "" {
			files, err := file.ListNames(db)
			if err != nil {
				return errors.Wrap(err, "error")
			}

			paths := make([]string, len(files))

			for i, file := range files {
				paths[i] = file.Name
			}

			tree.Print(paths)
			return nil
		}

		if filter {
			files, err := file.ListByName(db, name)
			if err != nil {
				return errors.Wrap(err, "error")
			}

			if len(files) == 0 {
				return errors.New("error: no wallets were found")
			}

			for _, file := range files {
				printFile(file)
			}
			return nil
		}

		file, err := file.Get(db, name)
		if err != nil {
			return errors.Wrap(err, "error")
		}

		printFile(file)
		return nil
	}
}

func printFile(f *pb.File) {
	cmdutil.PrintObjectName(f.Name)

	t := time.Unix(f.CreatedAt, 0)
	bytes := len(f.Content)
	size := fmt.Sprintf("%d bytes", bytes)

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

	fmt.Printf(`│ Path       │ %s
│ Filename   │ %s
│ Size       │ %s
│ Created at │ %s
`, f.Name, f.Filename, size, t)

	fmt.Println("└────────────+────────────────────────────────────────────>")
}
