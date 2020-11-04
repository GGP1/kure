// Package tree is used to build up a tree from a slice of paths and print
// it to the standard output.
package tree

import (
	"fmt"
	"strings"
)

// Folder represents a system folder.
type Folder struct {
	Name     string
	Children []*Folder
}

// Print prints the paths passed as a tree on the console.
func Print(paths []string) {
	root := Root(paths)

	start := "│  "
	for i, r := range root.Children {
		if i == len(root.Children)-1 {
			fmt.Println("└──", r.Name)
			start = "   "
		} else {
			fmt.Println("├──", r.Name)
		}

		printChildren(r, "", start)
	}
}

// Root returns the tree root.
func Root(paths []string) *Folder {
	root := &Folder{}

	for _, p := range paths {
		build(root, p)
	}

	return root
}

// build constructs the path tree.
func build(root *Folder, path string) {
	child := &Folder{}

	if !strings.Contains(path, "/") {
		child.Name = path
		root.Children = append(root.Children, child)
		return
	}

	parts := strings.Split(path, "/")

	child.Name = parts[0]
	remains := strings.Join(parts[1:], "/")

	// Repeat the process with the rest of the parts
	build(child, remains)

	// If the root already exists, check heritage to the deepest level
	// and merge if two parents have the same name
	for _, r := range root.Children {
		if r.Name == child.Name {
			finished := checkHeritage(r, child.Children)
			if finished {
				r.Children = append(r.Children, child.Children...)
			}
			return
		}
	}
	root.Children = append(root.Children, child)
}

// checkHeritage checks if two nodes of the tree have the same name
// to merge them once the last level has been checked.
//
// It returns a boolean to notify when the recursion has ended, this
// helps to determine when to append the children to the parent.
func checkHeritage(parent *Folder, children []*Folder) bool {
	for _, p := range parent.Children {
		for _, s := range children {
			if p.Name == s.Name {
				finished := checkHeritage(p, s.Children)
				if finished {
					p.Children = append(p.Children, s.Children...)
				}
				return false
			}
		}
	}
	return true
}

func printChildren(root *Folder, indent, start string) {
	for i, r := range root.Children {
		fmt.Print(start)

		add := " │  "

		if i == len(root.Children)-1 {
			fmt.Println(indent, "└──", r.Name)
			add = "    "
		} else {
			fmt.Println(indent, "├──", r.Name)
		}

		printChildren(r, indent+add, start)
	}
}
