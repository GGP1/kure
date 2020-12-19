package cmd

import (
	"bufio"
	"bytes"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/GGP1/kure/db/card"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/db/file"
	"github.com/GGP1/kure/db/note"
	"github.com/GGP1/kure/pb"

	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

func TestBuildBox(t *testing.T) {
	expected := `┌───── Test ─────┐
│ Jedi   │ Luke  │
│ Hobbit │ Frodo │
│        │ Sam   │
│ Wizard │ Harry │
└────────────────┘`

	// The iteration order it's not guaranteed and the test
	// may fail until the use of an ordered map
	mp := map[string]string{
		"Jedi": "Luke",
		"Hobbit": `Frodo
Sam`,
		"Wizard": "Harry",
	}

	got := BuildBox("test/test", mp)
	if got != expected {
		t.Errorf("Expected %s, got %s", expected, got)
	}
}

func TestDisplayQRCode(t *testing.T) {
	cases := []struct {
		desc   string
		secret string
		pass   bool
	}{
		{desc: "Low", secret: "secret", pass: true},
		{desc: "High", secret: "secret", pass: true},
		{desc: "Highest", secret: "secret", pass: true},
		{desc: "Fail", secret: "", pass: false},
		{desc: "Secret too long", secret: longSecret, pass: false},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			err := DisplayQRCode(tc.secret)
			assertError(t, "DisplayQRCode()", err, tc.pass)
		})
	}
}

func TestExistsTrue(t *testing.T) {
	db := SetContext(t, "../db/testdata/database")
	defer db.Close()

	cases := []struct {
		name   string
		object string
		create func() error
	}{
		{
			name:   "test",
			object: "card",
			create: func() error { return card.Create(db, &pb.Card{Name: "test"}) },
		},
		{
			name:   "test",
			object: "entry",
			create: func() error { return entry.Create(db, &pb.Entry{Name: "test", Expires: "Never"}) },
		},
		{
			name:   "test",
			object: "note",
			create: func() error { return note.Create(db, &pb.Note{Name: "test"}) },
		},
	}

	for _, tc := range cases {
		t.Run(tc.object, func(t *testing.T) {
			if err := tc.create(); err != nil {
				t.Fatal(err)
			}

			if err := Exists(db, tc.name, tc.object); err == nil {
				t.Error("Expected exists to fail but got nil")
			}
		})
	}
}

func TestExistsFalse(t *testing.T) {
	db := SetContext(t, "../db/testdata/database")
	defer db.Close()

	cases := []struct {
		name   string
		object string
	}{
		{name: "test", object: "card"},
		{name: "test", object: "entry"},
		{name: "test", object: "note"},
	}

	for _, tc := range cases {
		t.Run(tc.object, func(t *testing.T) {
			if err := Exists(db, tc.name, tc.object); err != nil {
				t.Errorf("Exists() failed: %v", err)
			}
		})
	}
}

func TestExistsInvalidObject(t *testing.T) {
	db := SetContext(t, "../db/testdata/database")
	defer db.Close()

	if err := Exists(db, "test", "invalid"); err == nil {
		t.Error("Expected exists to fail but got nil")
	}
}

func TestGetConfigPath(t *testing.T) {
	cases := []struct {
		desc string
		path string
	}{
		{desc: "Env var with extension", path: "/home/kure/.kure.yaml"},
		{desc: "Env var without extension", path: "/home/kure"},
		{desc: "Home directory", path: ""},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if err := os.Setenv("KURE_CONFIG", tc.path); err != nil {
				t.Errorf("Failed to set the environment variable: %v", err)
			}

			path, err := GetConfigPath()
			if err != nil {
				t.Errorf("Failed getting the configuration file path: %v", err)
			}

			if path == "" {
				t.Errorf("Expected a path to the file and got %q", path)
			}

			filename := ".kure.yaml"
			if !strings.Contains(path, filename) {
				t.Errorf("Invalid path, expected it to contain %q but got: %q", filename, path)
			}
		})
	}
}

func TestGetConfigPathError(t *testing.T) {
	env := "HOME"
	switch runtime.GOOS {
	case "windows":
		env = "USERPROFILE"
	case "plan9":
		env = "home"
	}

	os.Setenv(env, "")
	if err := os.Setenv("KURE_CONFIG", ""); err != nil {
		t.Errorf("Failed to set the environment variable: %v", err)
	}

	path, err := GetConfigPath()
	if err == nil {
		t.Fatalf("Expected GetConfigPath() to fail but got: %s", path)
	}
}

func TestProceed(t *testing.T) {
	cases := []struct {
		desc    string
		input   string
		proceed bool
	}{
		{desc: "Continue", input: "y", proceed: true},
		{desc: "Stop", input: "n", proceed: false},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			buf := bytes.NewBufferString(tc.input)

			proceed := Proceed(buf)
			if proceed != tc.proceed {
				t.Errorf("Expected %v, got %v", tc.proceed, proceed)
			}
		})
	}
}

func TestRequirePassword(t *testing.T) {
	db := SetContext(t, "../db/testdata/database")
	defer db.Close()

	name := "require_password_test"

	cases := []struct {
		desc   string
		create func() error
		remove func() error
	}{
		{
			desc:   "Card",
			create: func() error { return card.Create(db, &pb.Card{Name: name}) },
			remove: func() error { return card.Remove(db, name) },
		},
		{
			desc:   "Entry",
			create: func() error { return entry.Create(db, &pb.Entry{Name: name, Expires: "Never"}) },
			remove: func() error { return entry.Remove(db, name) },
		},
		{
			desc:   "File",
			create: func() error { return file.Create(db, &pb.File{Name: name}) },
			remove: func() error { return file.Remove(db, name) },
		},
		{
			desc:   "Note",
			create: func() error { return note.Create(db, &pb.Note{Name: name}) },
			remove: func() error { return note.Remove(db, name) },
		},
	}

	// This mock is used to execute RequirePassword as PreRunE
	mock := func(db *bolt.DB) *cobra.Command {
		return &cobra.Command{
			Use:     "mock",
			PreRunE: RequirePassword(db),
		}
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if err := tc.create(); err != nil {
				t.Fatal(err)
			}

			cmd := mock(db)
			if err := cmd.PreRunE(cmd, nil); err != nil {
				t.Errorf("RequirePassword() failed: %v", err)
			}

			if err := tc.remove(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestScan(t *testing.T) {
	cases := []struct {
		desc     string
		input    string
		expected string
	}{
		{desc: "Scan", input: "test\n", expected: "test"},
		{desc: "Empty scan", input: "\n", expected: ""},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			buf := bytes.NewBufferString(tc.input)
			scanner := bufio.NewScanner(buf)

			got := Scan(scanner, "test")
			if got != tc.expected {
				t.Errorf("Expected %s, got: %s", tc.expected, got)
			}
		})
	}
}

func TestScanlns(t *testing.T) {
	cases := []struct {
		desc     string
		input    string
		expected string
	}{
		{desc: "Scan lines", input: "test\nscanlns\n!end\n", expected: "test\nscanlns"},
		{desc: "Break", input: "!end\n", expected: ""},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			buf := bytes.NewBufferString(tc.input)
			scanner := bufio.NewScanner(buf)

			got := Scanlns(scanner, "test")
			if got != tc.expected {
				t.Errorf("Expected %s, got: %s", tc.expected, got)
			}
		})
	}
}

func TestSetContext(t *testing.T) {
	db := SetContext(t, "../db/testdata/database")
	defer db.Close()
}

func assertError(t *testing.T, funcName string, err error, pass bool) {
	if err != nil && pass {
		t.Errorf("Failed running %s: %v", funcName, err)
	}
	if err == nil && !pass {
		t.Error("Expected an error and got nil")
	}
}

var longSecret string = `hpidf9YBs?5j(]j5vg a#b4pzVk4es\QS G:}t&w~((u[mL\>bMP3Nbhhl.
	WBnSq4?C/C%'gC%hNlK'1^uMp\%u${W'~0M6_iW$iDn8Tk%|V;bk} *Q+0|,r Ul"7:INCaeyJkpff~e+%nH.
	<!>jxAKO|XYaL]=4/r|/JVi3[pldNZ<p%DM:=6q8=F~-F&*nWLF|R2b=afySN qNxLpk1BUv@J."vC}zzpw U(w4m
	m}=%y7?Swm3hA=vSTh[_8]Y$B".!:jXT)V!UJl:1\ S6-,n(.~a^V,X%MP1>)58ek-]Si}iu!P,A7;v97icAcy}F1z
	z-}N4#}xkS!A6cAi!*uK{ZHFT|Xo+otmbWjxpX;SM$3;#vJq~^2G=gMUL"|B8(F]2yJ%nq)cq[Lq){u7Wu96p@JD5EQy)]
	"d>c<z>7Tptj?z%P?VEGk]{j3z+,aXdn>ENv $.zZ&NZRyp=<DWS/fLJF!w8oZ=XB6D*40MBLMyH)1<o~1Az=L^N|YzGZ}
	=8)gYO,YhF9&D@SZaoO])USY%a9W~m~/d_{Kul4n<g^lr+5h8HA7TWgoKsSkGT#u-5i#V*n!:A_D/^z%9j#{%D@j&Wue6~
	7L.c8#s-+='^d1&<XIJwLxPcTdk<I9z9@szKWG -Pe^.P!"_54m8l%^fzObj~L!Y??+<%]Wo _Y9L>%VY"~67')<zX1us]
	->&T8YiA"7D@FS].z,mb8;Ae/OLAgZ!0t*PQG40>FR^F>\g&qtlJrTX?cn/F=.dN&7=b1Ws{IcE4ft6,2?Z{yGK"=(Xa||
	/wRtqyRXE<|/[ -+}eet>"[NXOD8Fy~mU>rEM=iP >{pnspSTLCooJOd+-PMo/_A?h3*w62SD01o5v(z?uZQYP_O@L# HI
	Oyy:jK2U]y9[Ea}Gl*e[l++DoDE7 $y7(A.-'y/efa$WB6Y$-Gq<JPHJBzUfQvMFO1hx5.Ve~N,rfv57G;u;oDb?cx_p6Zn
	HXM&$P7;WOm%DH}U%ye[#-S>)?P[6Bw<6/j|2*|v6";et)A#4?|1_wrYTWVY)P?z1Q!8)~O2y5tXj1n#RwZLr@L':zY1C|m
	".G:EsvzRvNCBc0c}QhWN\LAn@Q-Y#]RP$H*>lx['ds.j7SX66AM1^>&9)qv;XRkQ*zj|YB)*"P2Fxt:U+9#z5__\OYc+_M
	q3Q-39eD/6RdP'wjh5"v]Z(ffW3g ^U>$9pm@:wk|0#2EzokB0%HD/>A=w'Drp4W!H;:4?X,Tqtl(P }}u<10)|'d3cI$6`
