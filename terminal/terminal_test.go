package terminal_test

import (
	"bufio"
	"bytes"
	"os"
	"testing"

	"github.com/GGP1/kure/terminal"
)

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

			got := terminal.Confirm(buf, "Are you sure you want to proceed?")
			if got != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, got)
			}
		})
	}
}

func TestDisplayQRCode(t *testing.T) {
	os.Stdout = os.NewFile(0, "") // Mute stdout
	if err := terminal.DisplayQRCode("secret"); err != nil {
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
			if err := terminal.DisplayQRCode(tc.secret); err == nil {
				t.Error("Expected an error and got nil")
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

			got := terminal.Scanln(r, "test")
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

			got := terminal.Scanlns(r, "test")
			if got != tc.expected {
				t.Errorf("Expected %s, got: %s", tc.expected, got)
			}
		})
	}
}


const longSecret = `hpidf9YBs?5j(]j5vg a#b4pzVk4es\QS G:}t&w~((u[mL\>bMP3Nbhhl.
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
