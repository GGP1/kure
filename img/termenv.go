package img

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
)

// Adapted from: github.com/muesli/termenv.

const (
	csi           = "\x1b["
	ascii profile = iota
	ansi
	ansi256
	trueColor

	foreground = "38"
	background = "48"
	resetSeq   = "0"
)

var ansiHex = []string{
	"#000000",
	"#ffffff",
}

// Color
type clr interface {
	sequence(bg bool) string
}

// style is a string that various rendering styles can be applied to.
type style struct {
	string
	styles []string
}

func newStyle(s ...string) style {
	return style{
		string: strings.Join(s, " "),
	}
}

// Styled renders s with all applied styles.
func (t style) Styled(s string) string {
	if len(t.styles) == 0 {
		return s
	}

	seq := strings.Join(t.styles, ";")
	if seq == "" {
		return s
	}

	return fmt.Sprintf("%s%sm%s%sm", csi, seq, s, csi+resetSeq)
}

func (t style) String() string {
	return t.Styled(t.string)
}

// foreground sets a foreground color.
func (t style) foreground(c clr) style {
	if c != nil {
		t.styles = append(t.styles, c.sequence(false))
	}
	return t
}

// background sets a background color.
func (t style) background(c clr) style {
	if c != nil {
		t.styles = append(t.styles, c.sequence(true))
	}
	return t
}

// NoColor ..
type NoColor struct{}

// ANSIColor is a color (0-15) as defined by the ANSI Standard.
type ANSIColor int

// ANSI256Color is a color (16-255) as defined by the ANSI Standard.
type ANSI256Color int

// RGBColor is a hex-encoded color, e.g. "#abcdef".
type RGBColor string

func (c NoColor) sequence(bg bool) string {
	return ""
}

func (c ANSIColor) sequence(bg bool) string {
	col := int(c)
	bgMod := func(c int) int {
		if bg {
			return c + 10
		}
		return c
	}

	if col < 8 {
		return fmt.Sprintf("%d", bgMod(col)+30)
	}
	return fmt.Sprintf("%d", bgMod(col-8)+90)
}

func (c ANSI256Color) sequence(bg bool) string {
	prefix := foreground
	if bg {
		prefix = background
	}
	return fmt.Sprintf("%s;5;%d", prefix, c)
}

func (c RGBColor) sequence(bg bool) string {
	f, err := hex(string(c))
	if err != nil {
		return ""
	}

	prefix := foreground
	if bg {
		prefix = background
	}
	return fmt.Sprintf("%s;2;%d;%d;%d", prefix, uint8(f.R*255), uint8(f.G*255), uint8(f.B*255))
}

type profile int

func (p profile) convert(c clr) clr {
	if p == ascii {
		return NoColor{}
	}

	switch v := c.(type) {
	case ANSIColor:
		return v

	case ANSI256Color:
		if p == ansi {
			return ansi256ToANSIColor(v)
		}
		return v

	case RGBColor:
		h, err := hex(string(v))
		if err != nil {
			return nil
		}
		if p < trueColor {
			ac := hexToANSI256Color(h)
			if p == ansi {
				return ansi256ToANSIColor(ac)
			}
			return ac
		}
		return v
	}

	return c
}

func (p profile) color(s string) clr {
	if len(s) == 0 {
		return nil
	}

	var c clr
	if strings.HasPrefix(s, "#") {
		c = RGBColor(s)
	} else {
		i, err := strconv.Atoi(s)
		if err != nil {
			return nil
		}

		if i < 16 {
			c = ANSIColor(i)
		} else {
			c = ANSI256Color(i)
		}
	}

	return p.convert(c)
}

func colorProfile() profile {
	if runtime.GOOS == "windows" {
		return colorProfileWindows()
	}

	return colorProfileUnix()
}

// colorProfile returns trueColor.
func colorProfileWindows() profile {
	return ansi256
}

func colorProfileUnix() profile {
	term := os.Getenv("TERM")
	colorTerm := os.Getenv("COLORTERM")

	switch strings.ToLower(colorTerm) {
	case "24bit":
		fallthrough
	case "truecolor":
		if term == "screen" || !strings.HasPrefix(term, "screen") {
			// enable TrueColor in tmux, but not for old-school screen
			return trueColor
		}
	case "yes":
		fallthrough
	case "true":
		return ansi
	}

	if strings.Contains(term, "256color") {
		return ansi256
	}
	if strings.Contains(term, "color") {
		return ansi
	}

	return ascii
}

func ansi256ToANSIColor(c ANSI256Color) ANSIColor {
	var r int
	md := math.MaxFloat64

	h, _ := hex(ansiHex[c])
	for i := 0; i <= 15; i++ {
		hb, _ := hex(ansiHex[i])
		d := h.distanceLab(hb)

		if d < md {
			md = d
			r = i
		}
	}

	return ANSIColor(r)
}

func hexToANSI256Color(c Color) ANSI256Color {
	v2ci := func(v float64) int {
		if v < 48 {
			return 0
		}
		if v < 115 {
			return 1
		}
		return int((v - 35) / 40)
	}

	// Calculate the nearest 0-based color index at 16..231
	r := v2ci(c.R * 255.0) // 0..5 each
	g := v2ci(c.G * 255.0)
	b := v2ci(c.B * 255.0)
	ci := 36*r + 6*g + b /* 0..215 */

	// Calculate the represented colors back from the index
	i2cv := [6]int{0, 0x5f, 0x87, 0xaf, 0xd7, 0xff}
	cr := i2cv[r] // r/g/b, 0..255 each
	cg := i2cv[g]
	cb := i2cv[b]

	// Calculate the nearest 0-based gray index at 232..255
	var grayIdx int
	average := (r + g + b) / 3
	if average > 238 {
		grayIdx = 23
	} else {
		grayIdx = (average - 3) / 10 // 0..23
	}
	gv := 8 + 10*grayIdx // same value for r/g/b, 0..255

	// Return the one which is nearer to the original input rgb value
	c2 := Color{R: float64(cr) / 255.0, G: float64(cg) / 255.0, B: float64(cb) / 255.0}
	g2 := Color{R: float64(gv) / 255.0, G: float64(gv) / 255.0, B: float64(gv) / 255.0}
	colorDist := c.distanceLab(c2)
	grayDist := c.distanceLab(g2)

	if colorDist <= grayDist {
		return ANSI256Color(16 + ci)
	}
	return ANSI256Color(232 + grayIdx)
}
