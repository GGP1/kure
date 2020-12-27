package backup

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	cmdutil "github.com/GGP1/kure/cmd"
)

func TestBackup(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	defer db.Close()

	cases := []struct {
		desc string
		port string
		http string
		path string
		pass bool
	}{
		{desc: "File", path: "backup.test", pass: true},
		{desc: "HTTP", http: "true", port: "0", pass: false},
		{desc: "Invalid path", path: "", pass: false},
	}

	cmd := NewCmd(db)

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			f := cmd.Flags()
			f.Set("path", tc.path)
			f.Set("http", tc.http)
			f.Set("port", tc.port)

			err := cmd.RunE(cmd, nil)
			if err != nil && tc.pass {
				t.Errorf("Failed creating the backup file: %v", err)
			}
			if err == nil && !tc.pass {
				t.Error("Expected and error but got nil")
			}
		})
	}

	if err := os.Remove("backup.test"); err != nil {
		t.Fatalf("Failed revoving the backup: %v", err)
	}
}

func TestHTTPBackup(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	defer db.Close()

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

func TestWriteTo(t *testing.T) {
	var buf bytes.Buffer

	db := cmdutil.SetContext(t, "../../db/testdata/database")
	defer db.Close()

	if err := writeTo(db, &buf); err != nil {
		t.Fatalf("Failed writing database: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("Expected buffer length not to be zero")
	}
}

func TestPostRun(t *testing.T) {
	cmd := NewCmd(nil)
	cmd.PostRun(cmd, nil)
}
