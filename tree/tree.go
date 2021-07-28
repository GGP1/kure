// Package tree is used to build up a tree from a slice of paths and print
// it to the standard output.
package tree

import (
	"fmt"
	"strings"
)

// Node represents a system folder or file.
type Node struct {
	Name     string
	Children []*Node
}

// Print prints the paths passed as a tree on the console.
func Print(paths []string) {
	root := Build(paths)

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

// Build constructs the tree and returns the root node.
func Build(paths []string) *Node {
	root := &Node{}

	for _, p := range paths {
		build(root, strings.Split(p, "/"))
	}

	return root
}

// build constructs the tree.
//
// path will be something like [root, folder, subfolder, file]
func build(root *Node, path []string) {
	child := &Node{Name: path[0]}

	// len(path) will be never < 1 and if there is only
	// one element it must be unique as we already verified
	// it when the user added the record
	if len(path) == 1 {
		root.Children = append(root.Children, child)
		return
	}

	temp := child
	// Add each child to its corresponding parent
	for _, name := range path[1:] {
		c := &Node{Name: name}
		child.Children = append(child.Children, c)
		child = c
	}
	child = temp

	for _, r := range root.Children {
		// If a node already exists, look for matches until the deepest level
		if r.Name == child.Name {
			if !foundMatch(r, child.Children) {
				// If no match was found in a deeper level,
				// perform the append in this one
				r.Children = append(r.Children, child.Children...)
			}
			return
		}
	}

	// child is a new node
	root.Children = append(root.Children, child)
}

// foundMatch checks if two nodes of the tree have the same name
// to merge them once the last level has been checked.
//
// It returns a boolean to notify whether or not a match was found.
func foundMatch(parent *Node, children []*Node) bool {
	for _, p := range parent.Children {
		for _, c := range children {
			// If there is a match repeat the process with their children
			if p.Name == c.Name {
				if !foundMatch(p, c.Children) {
					// If no match was found in a deeper level,
					// perform the append in this one
					p.Children = append(p.Children, c.Children...)
				}
				return true
			}
		}
	}

	return false
}

// printChildren uses recursion for printing every folder children and adds
// indentation every time it prints the last element of a branch.
func printChildren(root *Node, indent, start string) {
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
