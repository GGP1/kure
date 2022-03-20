package tree_test

import (
	"strings"
	"testing"

	"github.com/GGP1/kure/tree"
)

func TestTreePrint(t *testing.T) {
	paths := []string{
		"kure/atoll/password",
		"kure/atoll/passphrase",
		"sync/atomic",
		"unsafe/pointer",
	}

	tree.Print(paths)
	// Output:
	// ├── kure
	// │   └── atoll
	// │       ├── password
	// │       └── passphrase
	// ├── sync
	// │   └── atomic
	// └── unsafe
	//     └── pointer
}

func TestTreeStructure(t *testing.T) {
	paths := []string{
		"The Hobbit",
		"The Lord of the Rings/The fellowship of the ring",
		"The Lord of the Rings/The two towers",
		"The Lord of the Rings/The return of the king",
	}

	root := tree.Build(paths)
	folders := make(map[string]struct{}, len(paths))

	for _, p := range paths {
		if _, ok := folders[p]; !ok {
			s, _, _ := strings.Cut(p, "/")
			folders[s] = struct{}{}
		}
	}

	expected := len(folders)
	if len(root.Children) != expected {
		t.Errorf("Expected %d branches, got %d", expected, len(root.Children))
	}

	for i, r := range root.Children {
		name, _, _ := strings.Cut(paths[i], "/")

		if r.Name != name {
			t.Errorf("Expected branch name to be %q, got %q", name, r.Name)
		}

		if i == len(root.Children)-1 {
			if len(r.Children) == 0 {
				t.Errorf("Expected %q branch to have a child named %q", r.Name, r.Children[0])
			}
		}
	}
}

func BenchmarkTree(b *testing.B) {
	paths := []string{
		"bench/mark/tree",
		"root",
		"multi/planetary/life",
		"go/src/github.com/GGP1/kure",
		"super/long/path/contaning/folders/subfolders/and/files",
		"go/src/github.com/<username>/<project>",
	}

	for i := 0; i < b.N; i++ {
		tree.Build(paths)
	}
}
