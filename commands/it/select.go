package it

import (
	"fmt"
	"strings"

	"github.com/GGP1/kure/db/card"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/db/file"
	"github.com/GGP1/kure/db/totp"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

// selectCommands recursively looks for commands and returns a slice with all of them.
func selectCommands(parent *cobra.Command) ([]string, error) {
	if !parent.HasSubCommands() {
		return nil, nil
	}

	cmdList := parent.Commands()
	list := make([]string, len(cmdList))
	for i, c := range cmdList {
		list[i] = c.Name()
	}

	// Only when it isn't root, preprend "self",
	// which is used for executing the current command
	if parent.HasParent() {
		list = append([]string{"self"}, list...)
	}

	qs := selectQs("Choose a command:", parent.UsageString(), list)
	cmd := struct{ Name string }{}

	if err := ask(qs, &cmd); err != nil {
		return nil, err
	}

	if cmd.Name == "self" {
		return []string{}, nil
	}

	current, _, err := parent.Find([]string{cmd.Name})
	if err != nil {
		return nil, err
	}

	// Repeat the process with potential childs
	child, err := selectCommands(current)
	if err != nil {
		return nil, err
	}

	result := append([]string{cmd.Name}, child...)
	return result, nil
}

func selectFlags(root *cobra.Command, commands []string) ([]string, error) {
	cmd, _, err := root.Find(commands)
	if err != nil {
		return nil, errors.Wrap(err, "command not found")
	}

	if !cmd.HasFlags() {
		return nil, nil
	}

	flagQs := &survey.Input{
		Message: "Flags:",
		Help:    "\n" + cmd.LocalFlags().FlagUsages(),
	}

	flags, err := askOne(flagQs)
	if err != nil {
		return nil, err
	}

	return strings.Split(flags, " "), nil
}

func selectName(db *bolt.DB, commands []string) (string, error) {
	var (
		list    []string
		err     error
		message = fmt.Sprintf("Choose a %s:", commands[0])
	)

	switch commands[0] {
	case "2fa":
		list, err = totp.ListNames(db)
		message = "Choose a TOTP:"

	case "card":
		list, err = card.ListNames(db)

	case "file":
		list, err = file.ListNames(db)

	case "ls", "copy", "edit", "rm":
		list, err = entry.ListNames(db)
		message = "Choose an entry:"
	}
	if err != nil {
		return "", err
	}

	// If any of the commands is "ls" or "2fa", add the "all" option
	// to list all the elements
	for _, cmd := range commands {
		if cmd == "ls" || cmd == "2fa" {
			list = append([]string{"all"}, list...)
			break
		}
	}

	qs := selectQs(message, "", list)
	chosen := struct{ Name string }{}

	if err := ask(qs, &chosen); err != nil {
		return "", err
	}

	// The user selected all, hence return no name
	if chosen.Name == "all" {
		chosen.Name = ""
	}

	return chosen.Name, nil
}

func selectManager(db *bolt.DB) (string, error) {
	list := []string{"1Password", "Bitwarden", "Keepass", "KeepassXC", "Lastpass"}
	qs := selectQs("Choose a manager:", "", list)
	manager := struct{ Name string }{}

	if err := ask(qs, &manager); err != nil {
		return "", err
	}

	return manager.Name, nil
}

func inputName() (string, error) {
	nameQs := &survey.Input{
		Message: "Name:",
		Help:    "The name mustn't be empty nor include \"//\"",
	}

	name, err := askOne(nameQs)
	if err != nil {
		return "", err
	}

	return name, nil
}
