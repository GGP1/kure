package config

import (
	"testing"

	cmdutil "github.com/GGP1/kure/commands"

	"github.com/spf13/viper"
)

func TestRead(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	viper.SetConfigFile("./testdata/mock_config.yaml")

	cmd := NewCmd(db, nil)
	if err := cmd.Execute(); err != nil {
		t.Errorf("Failed reading config: %v", err)
	}
}

func TestReadError(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	viper.SetConfigFile("")

	cmd := NewCmd(db, nil)
	if err := cmd.Execute(); err == nil {
		t.Error("Expected an error and got nil")
	}
}
