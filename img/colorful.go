package img

import (
	"fmt"
	"image/color"
	"math"
)

// Adapted from: github.com/lucasb-eyer/go-colorful.

// Color is stored internally using sRGB (standard RGB) values in the range 0-1
type Color struct {
	R, G, B float64
}

// makeColor constructs a color something implementing color.Color
func makeColor(col color.Color) (Color, bool) {
	r, g, b, a := col.RGBA()
	if a == 0 {
		return Color{0, 0, 0}, false
	}

	// Since color.Color is alpha pre-multiplied, we need to divide the
	// RGB values by alpha again in order to get back the original RGB.
	r *= 0xffff
	r /= a
	g *= 0xffff
	g /= a
	b *= 0xffff
	b /= a

	return Color{float64(r) / 65535.0, float64(g) / 65535.0, float64(b) / 65535.0}, true
}

// hex returns the hex "html" representation of the color, as in #ff0080.
func (col Color) hex() string {
	// Add 0.5 for rounding
	return fmt.Sprintf("#%02x%02x%02x", uint8(col.R*255.0+0.5), uint8(col.G*255.0+0.5), uint8(col.B*255.0+0.5))
}

// distanceLab is a good measure of visual similarity between two colors.
//
// A result of 0 would mean identical colors, while a result of 1 or higher
// means the colors differ a lot.
func (col Color) distanceLab(c2 Color) float64 {
	l1, a1, b1 := col.lab()
	l2, a2, b2 := c2.lab()
	return math.Sqrt(sqrt(l1-l2) + sqrt(a1-a2) + sqrt(b1-b2))
}

// lab converts the given color to CIE L*a*b* space using D65 as reference white.
func (col Color) lab() (l, a, b float64) {
	return xyzToLab(col.xyz())
}

// XYZ
// http://www.sjbrown.co.uk/2004/05/14/gamma-correct-rendering/
func (col Color) xyz() (x, y, z float64) {
	return linearRGBToXyz(col.linearRGB())
}

// linearRGB converts the color into the linear RGB space
// (see http://www.sjbrown.co.uk/2004/05/14/gamma-correct-rendering/).
func (col Color) linearRGB() (r, g, b float64) {
	r = linearize(col.R)
	g = linearize(col.G)
	b = linearize(col.B)
	return r, g, b
}

func xyzToLab(x, y, z float64) (l, a, b float64) {
	// Use D65 white as reference point by default.
	// http://www.fredmiranda.com/forum/topic/1035332
	// http://en.wikipedia.org/wiki/Standard_illuminant
	// This is the default reference white point.
	d65 := [3]float64{0.95047, 1.00000, 1.08883}
	return xyzToLabWhiteRef(x, y, z, d65)
}

func xyzToLabWhiteRef(x, y, z float64, wref [3]float64) (l, a, b float64) {
	fy := labF(y / wref[1])
	l = 1.16*fy - 0.16
	a = 5.0 * (labF(x/wref[0]) - fy)
	b = 2.0 * (fy - labF(z/wref[2]))
	return l, a, b
}

// sqrt returns the square root of v.
func sqrt(v float64) float64 {
	return v * v
}

func linearRGBToXyz(r, g, b float64) (x, y, z float64) {
	x = 0.4124564*r + 0.3575761*g + 0.1804375*b
	y = 0.2126729*r + 0.7151522*g + 0.0721750*b
	z = 0.0193339*r + 0.1191920*g + 0.9503041*b
	return x, y, z
}

func linearize(v float64) float64 {
	if v <= 0.04045 {
		return v / 12.92
	}
	return math.Pow((v+0.055)/1.055, 2.4)
}

func labF(t float64) float64 {
	if t > 6.0/29.0*6.0/29.0*6.0/29.0 {
		return math.Cbrt(t)
	}
	return t/3.0*29.0/6.0*29.0/6.0 + 4.0/29.0
}

// hex parses a "html" hex color-string, either in the 3 "#f0c" or 6 "#ff1034" digits form.
func hex(scol string) (Color, error) {
	format := "#%02x%02x%02x"
	factor := 1.0 / 255.0
	if len(scol) == 4 {
		format = "#%1x%1x%1x"
		factor = 1.0 / 15.0
	}

	var r, g, b uint8
	n, err := fmt.Sscanf(scol, format, &r, &g, &b)
	if err != nil {
		return Color{}, err
	}
	if n != 3 {
		return Color{}, fmt.Errorf("color: %v is not a hex-color", scol)
	}

	return Color{float64(r) * factor, float64(g) * factor, float64(b) * factor}, nil
}
