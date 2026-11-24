package theme

import (
	"fmt"
	"strconv"
	"strings"
)

// rgb is a parsed "#RRGGBB" color.
type rgb struct{ r, g, b float64 }

// parse reads "#RGB", "#RRGGBB" or "#RRGGBBAA" (alpha dropped).
func parse(hex string) (rgb, bool) {
	h := strings.TrimPrefix(hex, "#")
	if len(h) == 3 {
		h = string([]byte{h[0], h[0], h[1], h[1], h[2], h[2]})
	}
	if len(h) != 6 && len(h) != 8 {
		return rgb{}, false
	}
	v, err := strconv.ParseUint(h[:6], 16, 32)
	if err != nil {
		return rgb{}, false
	}
	return rgb{float64(v >> 16 & 0xff), float64(v >> 8 & 0xff), float64(v & 0xff)}, true
}

func (c rgb) hex() string {
	clamp := func(f float64) int { return max(0, min(255, int(f+0.5))) }
	return fmt.Sprintf("#%02X%02X%02X", clamp(c.r), clamp(c.g), clamp(c.b))
}

// mix blends a into b by t (0 is b, 1 is a), in sRGB like design tools.
func mix(a, b string, t float64) string {
	x, ok1 := parse(a)
	y, ok2 := parse(b)
	if !ok1 || !ok2 {
		return b
	}
	return rgb{y.r + (x.r-y.r)*t, y.g + (x.g-y.g)*t, y.b + (x.b-y.b)*t}.hex()
}

// alpha is c with an alpha byte: alpha("#7C6CF6", 0x66).
func alpha(c string, a uint8) string {
	x, ok := parse(c)
	if !ok {
		return c
	}
	return fmt.Sprintf("%s%02X", x.hex(), a)
}

// ValidColor reports whether s is a color the theme takes ("#RRGGBB").
func ValidColor(s string) bool {
	_, ok := parse(s)
	return ok
}

// Hex normalises a color to "#RRGGBB" (alpha dropped).
func Hex(s string) string {
	x, ok := parse(s)
	if !ok {
		return s
	}
	return x.hex()
}
