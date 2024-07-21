package rotate

import (
	"testing"

	"github.com/GGP1/atoll"
	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/pb"

	"github.com/stretchr/testify/assert"
)

func TestRotate(t *testing.T) {
	db := cmdutil.SetContext(t)

	name := "test"
	password := "testing"
	err := entry.Create(db, &pb.Entry{Name: name, Password: password})
	assert.NoError(t, err)

	cmd := NewCmd(db)
	cmd.SetArgs([]string{name})

	err = cmd.Execute()
	assert.NoError(t, err)

	updatedEntry, err := entry.Get(db, name)
	assert.NoError(t, err)

	assert.NotEqual(t, password, updatedEntry.Password)
	assert.Equal(t, len(password), len(updatedEntry.Password))

	oldSecret := atoll.SecretFromString(password)
	newSecret := atoll.SecretFromString(updatedEntry.Password)

	assert.Equal(t, oldSecret.Entropy(), newSecret.Entropy())
}

func TestRotateErrors(t *testing.T) {
	db := cmdutil.SetContext(t)

	cases := []struct {
		desc string
		name string
	}{
		{
			desc: "Invalid name",
			name: "",
		},
		{
			desc: "Does not exist",
			name: "non-existent",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			cmd := NewCmd(db)
			cmd.SetArgs([]string{tc.name})

			err := cmd.Execute()
			assert.Error(t, err)
		})
	}
}
