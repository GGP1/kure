package stats

import (
	"testing"

	cmdutil "github.com/GGP1/kure/commands"

	"github.com/stretchr/testify/assert"
)

func TestStats(t *testing.T) {
	db := cmdutil.SetContext(t)

	t.Run("Success", func(t *testing.T) {
		cmd := NewCmd(db)
		err := cmd.Execute()
		assert.NoError(t, err)
	})

	t.Run("Database connection closed", func(t *testing.T) {
		db.Close()
		err := NewCmd(db).Execute()
		assert.Error(t, err)
	})
}
