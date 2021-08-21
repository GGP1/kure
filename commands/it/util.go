package it

import (
	"github.com/GGP1/kure/sig"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
)

var template = `{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
{{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}`

func format() survey.AskOpt {
	return survey.WithIcons(func(is *survey.IconSet) {
		is.Question.Text = ""
		is.SelectFocus.Format = "red"
		is.Help.Format = "cyan"
		is.HelpInput.Format = "red"
		is.MarkedOption.Format = "red"
	})
}

// ask is a wrapper of survey.Ask that sets the format and interrupts the process
// on an interrupt error.
func ask(qs []*survey.Question, response interface{}) error {
	if err := survey.Ask(qs, response, format()); err != nil {
		if err == terminal.InterruptErr {
			sig.Signal.Kill()
		}
		return err
	}

	return nil
}

// askOne is a wrapper of survey.AskOne that sets the format and interrupts the process
// on an interrupt error.
func askOne(p survey.Prompt, response interface{}) error {
	if err := survey.AskOne(p, response, format()); err != nil {
		if err == terminal.InterruptErr {
			sig.Signal.Kill()
		}
		return err
	}

	return nil
}

func selectQs(message, help string, options []string) []*survey.Question {
	pageSize := len(options)
	if pageSize > 20 {
		pageSize = 20
	}
	return []*survey.Question{
		{
			Name: "name",
			Prompt: &survey.Select{
				Message:  message,
				Help:     help,
				Options:  options,
				PageSize: pageSize,
				VimMode:  true,
			},
		},
	}
}
