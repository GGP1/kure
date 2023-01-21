package backup

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	cmdutil "github.com/GGP1/kure/commands"

	"github.com/stretchr/testify/assert"
)

func TestBackupFile(t *testing.T) {
	db := cmdutil.SetContext(t)
	filename := "backup-test"

	cmd := NewCmd(db)
	f := cmd.Flags()
	f.Set("path", filename)

	err := cmd.Execute()
	assert.NoError(t, err, "Failed creating the backup file")

	err = os.Remove(filename)
	assert.NoError(t, err, "Failed removing the backup file")
}

func TestBackupServer(t *testing.T) {
	db := cmdutil.SetContext(t)

	rec := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "localhost:4000/", nil)
	assert.NoError(t, err, "Failed sending the request")

	hf := httpBackup(db)
	hf.ServeHTTP(rec, req)

	res := rec.Result()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	gotCt := res.Header.Get("Content-Type")
	expectedCt := "application/octet-stream"
	assert.Equal(t, expectedCt, gotCt)
}

func TestBackupErrors(t *testing.T) {
	db := cmdutil.SetContext(t)

	cases := []struct {
		desc string
		port string
		http string
		path string
	}{
		{
			desc: "HTTP",
			http: "true",
			port: "0",
		},
		{
			desc: "Invalid path",
			path: "",
		},
		{
			desc: "Mkdir error",
			path: "backup.go/",
		},
	}

	cmd := NewCmd(db)

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			f := cmd.Flags()
			f.Set("path", tc.path)
			f.Set("http", tc.http)
			f.Set("port", tc.port)

			err := cmd.Execute()
			assert.Error(t, err)
		})
	}
}

func TestWriteTo(t *testing.T) {
	db := cmdutil.SetContext(t)

	var buf bytes.Buffer
	err := writeTo(db, &buf)
	assert.NoError(t, err, "Failed writing database")

	assert.NotEqual(t, buf.Len(), 0)
}

func TestPostRun(t *testing.T) {
	NewCmd(nil).PostRun(nil, nil)
}
