package root_test

import (
	"testing"

	"github.com/GGP1/kure/commands/root"

	"github.com/stretchr/testify/assert"
)

func TestExecute(t *testing.T) {
	err := root.NewCmd(nil).Execute()
	assert.NoError(t, err)
}

func TestHasDescription(t *testing.T) {
	cmd := root.NewCmd(nil)

	for _, c := range cmd.Commands() {
		if c.Short == "" && c.Long == "" {
			t.Errorf("%q command doesn't have a description", c.Name())
		}
	}
}

func TestHasExample(t *testing.T) {
	cmd := root.NewCmd(nil)
	exceptions := map[string]struct{}{
		"restore":    {},
		"help":       {},
		"completion": {},
	}

	for _, c := range cmd.Commands() {
		if !c.HasExample() {
			name := c.Name()
			if _, ok := exceptions[name]; !ok {
				t.Errorf("%q command doesn't have an example", name)
			}
		}
	}
}

func TestRunnable(t *testing.T) {
	cmd := root.NewCmd(nil)
	exceptions := map[string]struct{}{
		"card":       {},
		"file":       {},
		"completion": {},
	}

	for _, c := range cmd.Commands() {
		if !c.Runnable() {
			name := c.Name()
			if _, ok := exceptions[name]; !ok {
				t.Errorf("%q command isn't runnable", name)
			}
		}
	}
}

func TestStatelessCommand(t *testing.T) {
	cases := []struct {
		commandName string
		expected    bool
	}{
		{
			commandName: "",
			expected:    false,
		},
		{
			commandName: "help",
			expected:    true,
		},
		{
			commandName: "add",
			expected:    false,
		},
		{
			commandName: "gen",
			expected:    true,
		},
		{
			commandName: "clear",
			expected:    true,
		},
		{
			commandName: "it",
			expected:    false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.commandName, func(t *testing.T) {
			got := root.StatelessCommand(tc.commandName)
			assert.Equal(t, tc.expected, got)
		})
	}
}
