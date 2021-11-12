package cmdutil

import (
	"bufio"
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/GGP1/kure/config"
	"github.com/GGP1/kure/db/card"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/db/file"
	"github.com/GGP1/kure/db/totp"
	"github.com/GGP1/kure/orderedmap"
	"github.com/GGP1/kure/pb"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	bolt "go.etcd.io/bbolt"
)

func TestBuildBox(t *testing.T) {
	expected := `╭────── Box ─────╮
│ Jedi   │ Luke  │
│ Hobbit │ Frodo │
│        │ Sam   │
│ Wizard │ Harry │
╰────────────────╯`

	mp := orderedmap.New()
	mp.Set("Jedi", "Luke")
	mp.Set("Hobbit", `Frodo
Sam`)
	mp.Set("Wizard", "Harry")

	got := BuildBox("test/box", mp)
	if got != expected {
		t.Errorf("Expected %s, got %s", expected, got)
	}
}

func TestConfirm(t *testing.T) {
	cases := []struct {
		desc     string
		input    string
		expected bool
	}{
		{desc: "Yes", input: "y", expected: true},
		{desc: "No", input: "n", expected: false},
		{desc: "Retry", input: "a\ny", expected: true},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			buf := bytes.NewBufferString(tc.input)

			got := Confirm(buf, "Are you sure you want to proceed?")
			if got != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, got)
			}
		})
	}
}

func TestDisplayQRCode(t *testing.T) {
	os.Stdout = os.NewFile(0, "") // Mute stdout
	if err := DisplayQRCode("secret"); err != nil {
		t.Errorf("Failed displaying QR code: %v", err)
	}
}

func TestDisplayQRCodeErrors(t *testing.T) {
	cases := []struct {
		desc   string
		secret string
	}{
		{desc: "Fail", secret: ""},
		{desc: "Secret too long", secret: longSecret},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			os.Stdout = os.NewFile(0, "") // Mute stdout
			if err := DisplayQRCode(tc.secret); err == nil {
				t.Error("Expected an error and got nil")
			}
		})
	}
}

func TestErase(t *testing.T) {
	f, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatalf("Failed creating temporary file: %v", err)
	}
	f.Close()

	if err := Erase(f.Name()); err != nil {
		t.Errorf("Failed erasing file: %v", err)
	}

	if err := Erase(f.Name()); err == nil {
		t.Error("Expected the file to be erased but it wasn't")
	}
}

func TestExistsTrue(t *testing.T) {
	db := SetContext(t, "../db/testdata/database")

	name := "naboo/tatooine"
	createObjects(t, db, name)

	cases := []struct {
		desc   string
		object object
	}{
		{
			desc:   "card",
			object: Card,
		},
		{
			desc:   "entry",
			object: Entry,
		},
		{
			desc:   "file",
			object: File,
		},
		{
			desc:   "totp",
			object: TOTP,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if err := Exists(db, name, tc.object); err == nil {
				t.Error("Expected exists to fail but got nil")
			}

			if err := Exists(db, "naboo/tatooine/hoth", tc.object); err == nil {
				t.Error("Expected exists to fail but got nil")
			}

			if err := Exists(db, "naboo", tc.object); err == nil {
				t.Error("Expected exists to fail but got nil")
			}
		})
	}
}

func TestExistsFalse(t *testing.T) {
	db := SetContext(t, "../db/testdata/database")

	cases := []struct {
		desc   string
		name   string
		object object
	}{
		{
			desc:   "card",
			name:   "test",
			object: Card,
		},
		{
			desc:   "entry",
			name:   "test",
			object: Entry,
		},
		{
			desc:   "file",
			name:   "testing/test",
			object: File,
		},
		{
			desc:   "totp",
			name:   "testing",
			object: TOTP,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if err := Exists(db, tc.name, tc.object); err != nil {
				t.Errorf("Exists() failed: %v", err)
			}
		})
	}
}

func TestFmtExpires(t *testing.T) {
	cases := []struct {
		desc     string
		expires  string
		expected string
	}{
		{
			desc:     "Never",
			expires:  "Never",
			expected: "Never",
		},
		{
			desc:     "dd/mm/yy",
			expires:  "26/06/2029",
			expected: "Tue, 26 Jun 2029 00:00:00 +0000",
		},
		{
			desc:     "yy/mm/dd",
			expires:  "2029/06/26",
			expected: "Tue, 26 Jun 2029 00:00:00 +0000",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := FmtExpires(tc.expires)
			if err != nil {
				t.Errorf("Failed formatting expires: %v", err)
			}

			if got != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, got)
			}
		})
	}

	t.Run("Invalid format", func(t *testing.T) {
		if _, err := FmtExpires("invalid format"); err == nil {
			t.Error("Expected an error and got nil")
		}
	})
}

func TestMustExist(t *testing.T) {
	db := SetContext(t, "../db/testdata/database")
	cmd := &cobra.Command{}
	objects := []object{Card, Entry, File, TOTP}

	name := "test/testing"
	createObjects(t, db, name)

	t.Run("Success", func(t *testing.T) {
		for _, obj := range objects {
			cmd.Args = MustExist(db, obj)
			if err := cmd.Args(cmd, []string{name}); err != nil {
				t.Error(err)
			}
		}
	})

	t.Run("Fail", func(t *testing.T) {
		cases := []struct {
			desc string
			name string
		}{
			{
				desc: "Record does not exist",
				name: "test",
			},
			{
				desc: "Empty name",
				name: "",
			},
			{
				desc: "Invalid name",
				name: "test//testing",
			},
		}

		for _, tc := range cases {
			t.Run(tc.desc, func(t *testing.T) {
				cmd.Args = MustExist(db, Card)
				if err := cmd.Args(cmd, []string{tc.name}); err == nil {
					t.Error("Expected an error and got nil")
				}
			})
		}
	})

	t.Run("Folders", func(t *testing.T) {
		// Take folders into account
		cmd.Flags().AddFlag(&pflag.Flag{
			Name:    "dir",
			Changed: true,
		})

		cmd.Args = MustExist(db, Card)
		if err := cmd.Args(cmd, []string{"test"}); err != nil {
			t.Error(err)
		}

		t.Run("Fail", func(t *testing.T) {
			if err := cmd.Args(cmd, []string{"not-exists"}); err == nil {
				t.Error("Expected an error and got nil")
			}
		})
	})
}

func TestMustExistLs(t *testing.T) {
	db := SetContext(t, "../db/testdata/database")
	cmd := &cobra.Command{}
	cmd.Flags().Bool("filter", false, "")
	objects := []object{Card, Entry, File, TOTP}

	name := "test"
	createObjects(t, db, name)

	cases := []struct {
		desc   string
		name   string
		filter bool
	}{
		{
			desc: "Found name",
			name: name,
		},
		{
			desc: "Empty name",
			name: "",
		},
		{
			desc:   "Filtering",
			name:   "t",
			filter: true,
		},
	}

	t.Run("Success", func(t *testing.T) {
		for _, tc := range cases {
			t.Run(tc.desc, func(t *testing.T) {
				for _, obj := range objects {
					cmd.Args = MustExistLs(db, obj)
					if tc.filter {
						cmd.Flags().Set("filter", "true")
					}

					if err := cmd.Args(cmd, []string{tc.name}); err != nil {
						t.Error(err)
					}
				}
			})
		}
	})

	t.Run("Fail", func(t *testing.T) {
		cmd.Args = MustExistLs(db, Entry)
		cmd.Flag("filter").Changed = false

		if err := cmd.Args(cmd, []string{"non-existent"}); err == nil {
			t.Error("Expected an error and got nil")
		}
	})
}

func TestMustNotExist(t *testing.T) {
	db := SetContext(t, "../db/testdata/database")
	cmd := &cobra.Command{}
	objects := []object{Card, Entry, File, TOTP}

	t.Run("Success", func(t *testing.T) {
		for _, obj := range objects {
			cmd.Args = MustNotExist(db, obj)
			if err := cmd.Args(cmd, []string{"test"}); err != nil {
				t.Error(err)
			}
		}
	})

	t.Run("Fail", func(t *testing.T) {
		entry.Create(db, &pb.Entry{Name: "test"})
		cases := []struct {
			desc string
			name string
		}{
			{
				desc: "Exists",
				name: "test",
			},
			{
				desc: "Empty name",
				name: "",
			},
			{
				desc: "Invalid name",
				name: "testing//test",
			},
		}

		for _, tc := range cases {
			t.Run(tc.desc, func(t *testing.T) {
				cmd.Args = MustNotExist(db, Entry)

				if err := cmd.Args(cmd, []string{tc.name}); err == nil {
					t.Error("Expected an error and got nil")
				}
			})
		}
	})
}

func TestNormalizeName(t *testing.T) {
	cases := []struct {
		desc     string
		name     string
		expected string
	}{
		{
			desc:     "Normalize",
			name:     " / Go/Forum / ",
			expected: "go/forum",
		},
		{
			desc:     "Empty",
			name:     "",
			expected: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got := NormalizeName(tc.name)

			if got != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, got)
			}
		})
	}
}

func TestScanln(t *testing.T) {
	cases := []struct {
		desc     string
		input    string
		expected string
	}{
		{desc: "Scan", input: "test  \n", expected: "test"},
		{desc: "Empty scan", input: "\n", expected: ""},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			buf := bytes.NewBufferString(tc.input)
			r := bufio.NewReader(buf)

			got := Scanln(r, "test")
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
		{desc: "Scan lines", input: "test\nscanlns\n<\n", expected: "test\nscanlns"},
		{desc: "Break", input: "<\n", expected: ""},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			buf := bytes.NewBufferString(tc.input)
			r := bufio.NewReader(buf)

			got := Scanlns(r, "test")
			if got != tc.expected {
				t.Errorf("Expected %s, got: %s", tc.expected, got)
			}
		})
	}
}

func TestSelectEditor(t *testing.T) {
	t.Run("Default editor", func(t *testing.T) {
		expected := "nano"
		config.Set("editor", expected)
		defer config.Reset()

		got := SelectEditor()
		if got != expected {
			t.Errorf("Expected %q, got %q", expected, got)
		}
	})

	t.Run("EDITOR", func(t *testing.T) {
		expected := "editor"
		os.Setenv("EDITOR", expected)
		defer os.Unsetenv("EDITOR")

		got := SelectEditor()
		if got != expected {
			t.Errorf("Expected %q, got %q", expected, got)
		}
	})

	t.Run("VISUAL", func(t *testing.T) {
		expected := "visual"
		os.Setenv("VISUAL", expected)
		defer os.Unsetenv("VISUAL")

		got := SelectEditor()
		if got != expected {
			t.Errorf("Expected %q, got %q", expected, got)
		}
	})

	t.Run("Default", func(t *testing.T) {
		got := SelectEditor()
		if got != "vim" {
			t.Errorf("Expected vim, got %q", got)
		}
	})
}

func TestSetContext(t *testing.T) {
	path := "../db/testdata/database"
	db := SetContext(t, path)

	gotPath := db.Path()
	if gotPath != path {
		t.Errorf("Expected path to be %q, got %q", path, gotPath)
	}

	gotOpenTx := db.Stats().OpenTxN
	if gotOpenTx != 0 {
		t.Errorf("Expected to have 0 opened transactions and got %d", gotOpenTx)
	}
}

func TestWatchFile(t *testing.T) {
	done := make(chan struct{}, 1)
	errCh := make(chan error, 1)

	filename := "test_watch_file"
	if err := os.WriteFile(filename, []byte("test"), 0600); err != nil {
		t.Fatal(err)
	}

	go WatchFile(filename, done, errCh)

	// Sleep to write after the file is being watched
	time.Sleep(50 * time.Millisecond)
	if err := os.WriteFile(filename, []byte("test watch file"), 0600); err != nil {
		t.Fatal(err)
	}

	select {
	case <-done:

	case <-errCh:
		t.Error("Watching file failed")
	}

	// Remove test file
	if err := os.Remove(filename); err != nil {
		t.Fatal(err)
	}
}

func TestWatchFileErrors(t *testing.T) {
	cases := []struct {
		desc     string
		filename string
		initial  bool
	}{
		{
			desc:     "Initial stat error",
			filename: "test_error.json",
			initial:  true,
		},
		{
			desc:     "For loop stat error",
			filename: "test_error.json",
			initial:  false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			done := make(chan struct{}, 1)
			errCh := make(chan error, 1)

			if !tc.initial {
				if err := os.WriteFile(tc.filename, []byte("test error"), 0644); err != nil {
					t.Fatalf("Failed creating file: %v", err)
				}
			}

			go WatchFile(tc.filename, done, errCh)

			// Sleep to wait until the file is created and fail once inside the for loop
			if !tc.initial {
				time.Sleep(10 * time.Millisecond)
				if err := os.Remove(tc.filename); err != nil {
					t.Fatalf("Failed removing the file: %v", err)
				}
			}

			select {
			case <-done:
				t.Error("Expected an error and it succeeded")

			case <-errCh:
			}
		})
	}
}

func TestWriteClipboard(t *testing.T) {
	if clipboard.Unsupported {
		t.Skip("No clipboard utilities available")
	}

	cmd := &cobra.Command{}

	t.Run("Default timeout", func(t *testing.T) {
		config.Set("clipboard.timeout", 10*time.Millisecond)
		defer config.Reset()

		if err := WriteClipboard(cmd, 0, "", "test"); err != nil {
			t.Fatal(err)
		}

		got, err := clipboard.ReadAll()
		if err != nil {
			t.Error(err)
		}

		if got != "" {
			t.Errorf("Expected the clipboard to be empty and got %q", got)
		}
	})

	t.Run("t > 0", func(t *testing.T) {
		if err := WriteClipboard(cmd, 10*time.Millisecond, "", "test"); err != nil {
			t.Fatal(err)
		}

		got, err := clipboard.ReadAll()
		if err != nil {
			t.Error(err)
		}

		if got != "" {
			t.Errorf("Expected the clipboard to be empty and got %q", got)
		}
	})

	t.Run("t = 0", func(t *testing.T) {
		clip := "test"
		if err := WriteClipboard(cmd, 0, "", clip); err != nil {
			t.Fatal(err)
		}

		got, err := clipboard.ReadAll()
		if err != nil {
			t.Error(err)
		}

		if got != clip {
			t.Errorf("Expected %q, got %q", clip, got)
		}
	})
}

func createObjects(t *testing.T, db *bolt.DB, name string) {
	t.Helper()
	if err := entry.Create(db, &pb.Entry{Name: name}); err != nil {
		t.Fatal(err)
	}
	if err := card.Create(db, &pb.Card{Name: name}); err != nil {
		t.Fatal(err)
	}
	if err := file.Create(db, &pb.File{Name: name}); err != nil {
		t.Fatal(err)
	}
	if err := totp.Create(db, &pb.TOTP{Name: name}); err != nil {
		t.Fatal(err)
	}
}

var longSecret = `hpidf9YBs?5j(]j5vg a#b4pzVk4es\QS G:}t&w~((u[mL\>bMP3Nbhhl.
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
