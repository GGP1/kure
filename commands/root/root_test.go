package root_test

import (
	"testing"

	"github.com/GGP1/kure/commands/root"

	"github.com/stretchr/testify/assert"
)

func TestExecute(t *testing.T) {
	err := root.Execute(nil)
	assert.NoError(t, err)
}

func TestHasDescription(t *testing.T) {
	cmd := root.DevCmd()

	for _, c := range cmd.Commands() {
		if c.Short == "" && c.Long == "" {
			t.Errorf("%q command doesn't have a description", c.Name())
		}
	}
}

func TestHasExample(t *testing.T) {
	cmd := root.DevCmd()
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
	cmd := root.DevCmd()
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
