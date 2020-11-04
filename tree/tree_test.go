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
		"The fellowship of the ring",
		"The two towers",
		"The return of the king",
		"Preceded/The Hobbit",
	}

	root := tree.Root(paths)
	expected := len(paths)

	if len(root.Children) != expected {
		t.Errorf("Expected %d branches, got: %d", expected, len(root.Children))
	}

	for i, r := range root.Children {
		name := strings.Split(paths[i], "/")

		if r.Name != name[0] {
			t.Errorf("Expected branch name to be: %s, got: %s", name[0], r.Name)
		}

		if i == len(root.Children)-1 {
			if len(r.Children) == 0 {
				t.Errorf("Expected \"%s\" to have a branch named \"%s\"", r.Name, r.Children[0])
			}
		}
	}
}
