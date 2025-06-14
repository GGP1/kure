package tree

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrint(t *testing.T) {
	paths := []string{
		"kure/atoll/secret/password",
		"kure/atoll/secret/passphrase",
		"kure/atoll/test/password",
		"sync/atomic",
		"unsafe/pointer",
	}

	temp := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	Print(paths)
	w.Close()

	got, err := io.ReadAll(r)
	assert.NoError(t, err)
	os.Stdout = temp

	expected := `├── kure
│   └── atoll
│       ├── secret
│       │   ├── password
│       │   └── passphrase
│       └── test
│           └── password
├── sync
│   └── atomic
└── unsafe
    └── pointer
`
	assert.Equal(t, expected, string(got))
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

func BenchmarkTree(b *testing.B) {
	paths := []string{
		"bench/mark/tree",
		"root",
		"multi/planetary/life",
		"go/src/github.com/GGP1/kure",
		"super/long/path/containing/folders/subfolders/and/files",
		"go/src/github.com/<username>/<project>",
	}

	for b.Loop() {
		newTree(paths)
	}
}
