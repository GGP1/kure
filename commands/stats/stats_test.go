package stats

import (
	"testing"

	cmdutil "github.com/GGP1/kure/commands"
)

func TestStats(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")

	t.Run("Success", func(t *testing.T) {
		cmd := NewCmd(db)
		if err := cmd.Execute(); err != nil {
			t.Errorf("Stats() failed %v", err)
		}
	})

	t.Run("Database connection closed", func(t *testing.T) {
		db.Close()

		if err := NewCmd(db).Execute(); err == nil {
			t.Error("Expected an error and got nil")
		}
	})
}
