package it

import (
	"github.com/GGP1/kure/db/file"

	"github.com/AlecAivazis/survey/v2"
	bolt "go.etcd.io/bbolt"
)

func fileMultiselect(db *bolt.DB) ([]string, error) {
	files, err := file.ListNames(db)
	if err != nil {
		return nil, err
	}

	namesQs := []*survey.Question{
		{
			Name: "names",
			Prompt: &survey.MultiSelect{
				Message: "Choose files:",
				Options: files,
				VimMode: true,
			},
		},
	}

	names := []string{}
	if err := ask(namesQs, &names); err != nil {
		return nil, err
	}

	return names, nil
}

func fileMvNames(db *bolt.DB) ([]string, error) {
	files, err := file.ListNames(db)
	if err != nil {
		return nil, err
	}

	// Request src
	qs := selectQs("Source", "", files)
	src := struct{ Name string }{}
	if err := ask(qs, &src); err != nil {
		return nil, err
	}

	// Request dst
	dstQs := &survey.Input{
		Message: "Destination:",
	}
	dst := ""
	if err := askOne(dstQs, &dst); err != nil {
		return nil, err
	}

	return []string{src.Name, dst}, nil
}
