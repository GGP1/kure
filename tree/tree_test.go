package tree

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrint(t *testing.T) {
	paths := []string{
		"kure/atoll/password",
		"kure/atoll/passphrase",
		"sync/atomic",
		"unsafe/pointer",
	}

	Print(paths)
	// Output:
	// ├── kure/
	// │   └── atoll/
	// │       ├── password
	// │       └── passphrase
	// ├── sync/
	// │   └── atomic
	// └── unsafe/
	//     └── pointer
}

func TestTreeStructure(t *testing.T) {
	paths := []string{
		"The Hobbit",
		"The Lord of the Rings/The fellowship of the ring",
		"The Lord of the Rings/The two towers",
		"The Lord of the Rings/The return of the king",
	}

	root := newTree(paths)
	folders := make(map[string]struct{}, len(paths))

	for _, p := range paths {
		if _, ok := folders[p]; !ok {
			s, _, _ := strings.Cut(p, "/")
			folders[s] = struct{}{}
		}
	}

	expected := len(folders)
	assert.Equal(t, expected, len(root.children))

	for i, r := range root.children {
		name, _, _ := strings.Cut(paths[i], "/")

		assert.Equal(t, name, r.name)

		if i == len(root.children)-1 {
			assert.NotEmpty(t, r.children)
		}
	}
}

func TestPrintTree(t *testing.T) {
	root := &node{
		children: []*node{
			{
				name: "kure",
				top:  true,
				children: []*node{
					{
						name: "atoll",
						children: []*node{
							{name: "password"},
							{name: "passphrase"},
						},
					},
				},
			},
			{name: "sync", top: true, children: []*node{{name: "atomic"}}},
			{name: "unsafe", top: true, children: []*node{{name: "pointer"}}},
		},
	}

	printTree(root, "")
	// Output:
	// ├── kure/
	// │   └── atoll/
	// │       ├── password
	// │       └── passphrase
	// ├── sync/
	// │   └── atomic
	// └── unsafe/
	//     └── pointer
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
		newTree(paths)
	}
}
