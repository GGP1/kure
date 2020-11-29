package stats

import (
	"testing"

	cmdutil "github.com/GGP1/kure/cmd"
)

func TestStats(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	defer db.Close()

	cmd := NewCmd(db)
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Errorf("Stats() failed %v", err)
	}
}
