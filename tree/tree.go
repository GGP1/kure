// Package tree is used to build up a tree from a slice of paths and print
// it to the standard output.
package tree

import (
	"fmt"
	"strings"
)

// node represents a system folder or file.
type node struct {
	name     string
	children []*node
}

// Print prints the paths passed as a tree on the console.
func Print(paths []string) {
	root := newTree(paths)
	start := "│  "
	for i, r := range root.children {
		lastChild := i == len(root.children)-1
		symbol, _ := getTokens(lastChild)
		if lastChild {
			start = "   "
		}

		fmt.Println(symbol, r.name)
		printChildren(r, "", start)
	}
}

// newTree builds the tree and returns the root node.
func newTree(paths []string) *node {
	root := &node{}
	for _, p := range paths {
		buildBranch(root, strings.Split(p, "/"))
	}
	return root
}

// buildBranch adds a branch to the root or merges two branches at the deepest matching level.
//
// path will be something like [root, folder, subfolder, file].
func buildBranch(root *node, path []string) {
	child := &node{name: path[0]}

	// len(path) will be never < 1 and if there is only
	// one element it must be unique as we already verified
	// it when the user added the record
	if len(path) == 1 {
		root.children = append(root.children, child)
		return
	}

	temp := child
	// Add each child to its corresponding parent
	for _, name := range path[1:] {
		c := &node{name: name}
		child.children = append(child.children, c)
		child = c
	}
	child = temp

	for _, r := range root.children {
		// If a node already exists, look for matches until the deepest level
		if r.name == child.name {
			if !deeperMatch(r, child) {
				// If no match was found at a deeper level,
				// perform the append at this one
				r.children = append(r.children, child.children...)
			}
			return
		}
	}

	// child is a new node
	root.children = append(root.children, child)
}

// deeperMatch checks if two nodes of the tree have the same name
// to merge them once the last level has been checked.
//
// It returns a boolean to notify whether or not a match was found.
func deeperMatch(parent, child *node) bool {
	for _, a := range parent.children {
		for _, b := range child.children {
			// If there is a match repeat the process with their children
			if a.name == b.name {
				if !deeperMatch(a, b) {
					// If no match was found at a deeper level,
					// perform the append at this one
					a.children = append(a.children, b.children...)
				}
				return true
			}
		}
	}

	return false
}

// printChildren uses recursion for printing every folder children and adds
// indentation every time it prints the last element of a branch.
func printChildren(root *node, indent, start string) {
	for i, r := range root.children {
		fmt.Print(start)
		symbol, add := getTokens(i == len(root.children)-1)
		fmt.Println(indent, symbol, r.name)

		printChildren(r, indent+add, start)
	}
}

func getTokens(lastChild bool) (string, string) {
	if lastChild {
		return "└──", "    "
	}
	return "├──", " │  "
}
