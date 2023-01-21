package config

import (
	"testing"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/config"

	"github.com/stretchr/testify/assert"
)

func TestRead(t *testing.T) {
	db := cmdutil.SetContext(t)
	config.SetFilename("./testdata/mock_config.yaml")

	cmd := NewCmd(db, nil)
	err := cmd.Execute()
	assert.NoError(t, err, "Failed reading config")
}

func TestReadError(t *testing.T) {
	db := cmdutil.SetContext(t)
	config.SetFilename("")

	cmd := NewCmd(db, nil)
	err := cmd.Execute()
	assert.Error(t, err)
}
