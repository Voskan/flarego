// pkg/flamegraph/palette.go
// Colour utilities for flame graphs.  The UI expects each frame to be
// decorated with a HEX colour string so that nodes can be tinted by:
//   - Package/module  (stable hash → palette)
//   - Delta sign       (green for shrink, red for growth)
//   - State bands      (GC, Heap, Blocked)
//
// The package exports a Colourer function type that consumers (encoders, UI
// helpers) can use to look up colours without coupling to specific palettes.
//
// Implementation details:
//   - A tiny deterministic hash (FNV-1a) maps strings → hue in HSL space.
//   - Conversion HSL → RGB → HEX avoids third‑party deps.
//   - Predefined hues for special pseudo‑stacks ensure visual consistency.
package flamegraph

import (
	"fmt"
	"hash/fnv"
	"math"
)

// Colourer maps frame names to CSS hex colours (e.g., "#ff8800").
type Colourer func(name string, value int64) string

// DefaultColourer is used by the UI when no theme override is provided.
// Pseudo-stacks prefixed with "(" (e.g., "(GC)") receive fixed colours; other
// names hash to a pastel palette.
var DefaultColourer Colourer = func(name string, value int64) string {
    // fixed colours for pseudo stacks
    switch name {
    case "(GC)":
        return "#b39ddb" // light purple
    case "(Heap)":
        return "#80cbc4" // teal
    case "(Blocked)":
        return "#ef9a9a" // salmon
    }

    // delta-aware: negative values (shrinking) in green, positive in red
    if value < 0 {
        return desaturate("#43a047", 0.3) // green
    }

    // hash to pastel colour for stable node tinting
    h := hashHue(name)
    r, g, b := hslToRGB(h, 0.6, 0.70) // pastel
    return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// hashHue returns a hue in [0,360) based on FNV-1a hash of the string.
func hashHue(s string) float64 {
    h := fnv.New32a()
    _, _ = h.Write([]byte(s))
    return float64(h.Sum32()%360)
}

// hslToRGB converts H,S,L in [0..1] to r,g,b 0..255.
func hslToRGB(h, s, l float64) (uint8, uint8, uint8) {
    h = math.Mod(h, 360) / 360
    var r, g, b float64
    if s == 0 {
        r = l; g = l; b = l
    } else {
        var p, q float64
        if l < 0.5 {
            q = l * (1 + s)
        } else {
            q = l + s - l*s
        }
        p = 2*l - q
        r = hueToRGB(p, q, h+1.0/3)
        g = hueToRGB(p, q, h)
        b = hueToRGB(p, q, h-1.0/3)
    }
    return uint8(r * 255), uint8(g * 255), uint8(b * 255)
}

func hueToRGB(p, q, t float64) float64 {
    if t < 0 {
        t += 1
    }
    if t > 1 {
        t -= 1
    }
    switch {
    case t < 1.0/6:
        return p + (q-p)*6*t
    case t < 1.0/2:
        return q
    case t < 2.0/3:
        return p + (q-p)*(2.0/3-t)*6
    default:
        return p
    }
}

// desaturate blends colour with light grey for muted states.
func desaturate(hex string, factor float64) string {
    var r, g, b uint8
    fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b)
    blend := func(c uint8) uint8 {
        grey := uint8(200)
        return uint8(float64(c)*(1-factor) + float64(grey)*factor)
    }
    r2, g2, b2 := blend(r), blend(g), blend(b)
    return fmt.Sprintf("#%02x%02x%02x", r2, g2, b2)
}
