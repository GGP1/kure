package rm

import (
	"bytes"
	"testing"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/totp"
	"github.com/GGP1/kure/pb"

	"github.com/stretchr/testify/assert"
)

func TestRm(t *testing.T) {
	db := cmdutil.SetContext(t, "../../../db/testdata/database")

	err := totp.Create(db, &pb.TOTP{Name: "test"})
	assert.NoError(t, err)

	cases := []struct {
		desc         string
		name         string
		confirmation string
	}{
		{
			desc:         "Abort",
			name:         "test",
			confirmation: "n",
		},
		{
			desc:         "Proceed",
			name:         "test",
			confirmation: "y",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			buf := bytes.NewBufferString(tc.confirmation)
			cmd := NewCmd(db, buf)
			cmd.SetArgs([]string{tc.name})

			err := cmd.Execute()
			assert.NoError(t, err, "Failed removing the TOTP")
		})
	}
}

func TestRmErrors(t *testing.T) {
	db := cmdutil.SetContext(t, "../../../db/testdata/database")

	cases := []struct {
		desc         string
		name         string
		confirmation string
	}{
		{
			desc: "Invalid name",
			name: "",
		},
		{
			desc:         "Does not exists",
			name:         "non-existent",
			confirmation: "y",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			buf := bytes.NewBufferString(tc.confirmation)
			cmd := NewCmd(db, buf)
			cmd.SetArgs([]string{tc.name})

			err := cmd.Execute()
			assert.Error(t, err)
		})
	}
}
