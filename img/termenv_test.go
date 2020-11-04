package img

import "testing"

func TestAscii(t *testing.T) {
	c := ascii.color("#abcdef")
	if c.sequence(false) != "" {
		t.Errorf("Expected empty sequence, got %s", c.sequence(false))
	}
}

func TestANSIProfile(t *testing.T) {
	p := ansi

	c := p.color("0")
	exp := "30"
	if c.sequence(false) != exp {
		t.Errorf("Expected %s, got %s", exp, c.sequence(false))
	}
	if _, ok := c.(ANSIColor); !ok {
		t.Errorf("Expected type termenv.ANSIColor, got %T", c)
	}

	c = p.color("15")
	exp = "97"
	if c.sequence(false) != exp {
		t.Errorf("Expected %s, got %s", exp, c.sequence(false))
	}
	if _, ok := c.(ANSIColor); !ok {
		t.Errorf("Expected type termenv.ANSIColor, got %T", c)
	}
}

func TestANSI256Profile(t *testing.T) {
	p := ansi256

	c := p.color("#abcdef")
	exp := "38;5;153"
	if c.sequence(false) != exp {
		t.Errorf("Expected %s, got %s", exp, c.sequence(false))
	}
	if _, ok := c.(ANSI256Color); !ok {
		t.Errorf("Expected type termenv.ANSI256Color, got %T", c)
	}

	c = p.color("139")
	exp = "38;5;139"
	if c.sequence(false) != exp {
		t.Errorf("Expected %s, got %s", exp, c.sequence(false))
	}
	if _, ok := c.(ANSI256Color); !ok {
		t.Errorf("Expected type termenv.ANSI256Color, got %T", c)
	}

	c = p.color("2")
	exp = "32"
	if c.sequence(false) != exp {
		t.Errorf("Expected %s, got %s", exp, c.sequence(false))
	}
	if _, ok := c.(ANSIColor); !ok {
		t.Errorf("Expected type termenv.ANSIColor, got %T", c)
	}
}

func TestTrueColorProfile(t *testing.T) {
	p := trueColor

	c := p.color("#abcdef")
	exp := "38;2;171;205;239"
	if c.sequence(false) != exp {
		t.Errorf("Expected %s, got %s", exp, c.sequence(false))
	}
	if _, ok := c.(RGBColor); !ok {
		t.Errorf("Expected type termenv.HexColor, got %T", c)
	}

	c = p.color("139")
	exp = "38;5;139"
	if c.sequence(false) != exp {
		t.Errorf("Expected %s, got %s", exp, c.sequence(false))
	}
	if _, ok := c.(ANSI256Color); !ok {
		t.Errorf("Expected type termenv.ANSI256Color, got %T", c)
	}

	c = p.color("2")
	exp = "32"
	if c.sequence(false) != exp {
		t.Errorf("Expected %s, got %s", exp, c.sequence(false))
	}
	if _, ok := c.(ANSIColor); !ok {
		t.Errorf("Expected type termenv.ANSIColor, got %T", c)
	}
}
