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
	"github.com/GGP1/kure/db/wallet"
	"github.com/GGP1/kure/pb"
)

func TestDisplayQRCode(t *testing.T) {
	cases := []struct {
		action string
		secret string
		pass   bool
	}{
		{action: "Low", secret: "secret", pass: true},
		{action: "High", secret: "secret", pass: true},
		{action: "Highest", secret: "secret", pass: true},
		{action: "Fail", secret: "", pass: false},
		{action: "Secret too long", secret: longSecret, pass: false},
	}

	for _, tc := range cases {
		err := DisplayQRCode(tc.secret)
		assertError(t, tc.action, "DisplayQRCode()", err, tc.pass)
	}
}

func TestGetConfigPath(t *testing.T) {
	cases := []struct {
		action string
		path   string
	}{
		{action: "Env var with extension", path: "/home/kure/.kure.yaml"},
		{action: "Env var without extension", path: "/home/kure"},
		{action: "Home directory", path: ""},
	}

	for _, tc := range cases {
		if err := os.Setenv("KURE_CONFIG", tc.path); err != nil {
			t.Errorf("%s: failed to set the environment variable: %v", tc.action, err)
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

func TestPrintObjectName(t *testing.T) {
	name := "test"

	PrintObjectName(name)
	// Output:
	//
	// +───────────────────────── Test ─────────────────────────>
}

func TestProceed(t *testing.T) {
	cases := []struct {
		action  string
		input   string
		proceed bool
	}{
		{action: "Continue", input: "y", proceed: true},
		{action: "Stop", input: "n", proceed: false},
	}

	for _, tc := range cases {
		buf := bytes.NewBufferString(tc.input)

		proceed := Proceed(buf)
		if proceed != tc.proceed {
			t.Errorf("%s: expected %v, got %v", tc.action, tc.proceed, proceed)
		}
	}
}

func TestRequirePassword(t *testing.T) {
	db := SetContext(t, "../db/testdata/database")
	defer db.Close()

	name := "require_password_test"

	cases := []struct {
		action string
		create func()
		delete func()
	}{
		{
			action: "Card",
			create: func() {
				if err := card.Create(db, &pb.Card{Name: name}); err != nil {
					t.Fatal(err)
				}
			},
			delete: func() {
				if err := card.Remove(db, name); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			action: "Entry",
			create: func() {
				if err := entry.Create(db, &pb.Entry{Name: name, Expires: "Never"}); err != nil {
					t.Fatal(err)
				}
			},
			delete: func() {
				if err := entry.Remove(db, name); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			action: "File",
			create: func() {
				if err := file.Create(db, &pb.File{Name: name}); err != nil {
					t.Fatal(err)
				}
			},
			delete: func() {
				if err := file.Remove(db, name); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			action: "Wallet",
			create: func() {
				if err := wallet.Create(db, &pb.Wallet{Name: name}); err != nil {
					t.Fatal(err)
				}
			},
			delete: func() {
				if err := wallet.Remove(db, name); err != nil {
					t.Fatal(err)
				}
			},
		},
	}

	for _, tc := range cases {
		tc.create()
		if err := RequirePassword(db); err != nil {
			t.Errorf("%s: RequirePassword() failed: %v", tc.action, err)
		}
		tc.delete()
	}

}

func TestScan(t *testing.T) {
	cases := []struct {
		action   string
		input    string
		expected string
	}{
		{action: "Scan", input: "test\n", expected: "test"},
		{action: "Empty scan", input: "\n", expected: ""},
	}

	for _, tc := range cases {
		buf := bytes.NewBufferString(tc.input)
		scanner := bufio.NewScanner(buf)

		got := Scan(scanner, "test")
		if got != tc.expected {
			t.Errorf("expected %s, got: %s", tc.expected, got)
		}
	}
}

func TestScanlns(t *testing.T) {
	cases := []struct {
		action   string
		input    string
		expected string
	}{
		{action: "Scan lines", input: "test\nscanlns\n!end\n", expected: "test\nscanlns"},
		{action: "Break", input: "!end\n", expected: ""},
	}

	for _, tc := range cases {
		buf := bytes.NewBufferString(tc.input)
		scanner := bufio.NewScanner(buf)

		got := Scanlns(scanner, "test")
		if got != tc.expected {
			t.Errorf("%s: expected %s, got: %s", tc.action, tc.expected, got)
		}
	}
}

func TestSetContext(t *testing.T) {
	db := SetContext(t, "../db/testdata/database")
	defer db.Close()
}

func assertError(t *testing.T, name, funcName string, err error, pass bool) {
	if err != nil && pass {
		t.Errorf("%s: failed running %s: %v", name, funcName, err)
	}
	if err == nil && !pass {
		t.Errorf("%s: expected an error and got nil", name)
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
