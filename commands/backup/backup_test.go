package backup

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	cmdutil "github.com/GGP1/kure/commands"
)

func TestBackupFile(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	filename := "backup-test"

	cmd := NewCmd(db)
	f := cmd.Flags()
	f.Set("path", filename)

	if err := cmd.Execute(); err != nil {
		t.Errorf("Failed creating the backup file: %v", err)
	}

	if err := os.Remove(filename); err != nil {
		t.Errorf("Failed removing the backup file: %v", err)
	}
}

func TestBackupServer(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")

	rec := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "localhost:4000/", nil)
	if err != nil {
		t.Fatalf("Failed sending the request: %v", err)
	}

	hf := httpBackup(db)
	hf.ServeHTTP(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %s", res.Status)
	}

	gotCt := res.Header.Get("Content-Type")
	expectedCt := "application/octet-stream"
	if gotCt != expectedCt {
		t.Errorf("Expected %q, got %q", expectedCt, gotCt)
	}
}

func TestBackupErrors(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")

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

			if err := cmd.Execute(); err == nil {
				t.Error("Expected and error but got nil")
			}
		})
	}
}

func TestWriteTo(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")

	var buf bytes.Buffer
	if err := writeTo(db, &buf); err != nil {
		t.Fatalf("Failed writing database: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("Expected buffer length not to be zero")
	}
}

func TestPostRun(t *testing.T) {
	NewCmd(nil).PostRun(nil, nil)
}
