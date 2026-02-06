package color

import (
	"math"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// AgeBucket defines a time threshold and its associated color.
type AgeBucket struct {
	MaxAge time.Duration
	Color  rl.Color
	Label  string
}

// DefaultAgeBuckets defines the age-to-color mapping.
// Files are colored based on their modification time relative to now.
var DefaultAgeBuckets = []AgeBucket{
	{MaxAge: 24 * time.Hour, Color: rl.NewColor(100, 210, 100, 255), Label: "< 1 day"},          // bright green
	{MaxAge: 7 * 24 * time.Hour, Color: rl.NewColor(140, 200, 80, 255), Label: "< 1 week"},      // lime
	{MaxAge: 30 * 24 * time.Hour, Color: rl.NewColor(200, 195, 60, 255), Label: "< 1 month"},    // yellow
	{MaxAge: 180 * 24 * time.Hour, Color: rl.NewColor(210, 170, 50, 255), Label: "< 6 months"},  // amber
	{MaxAge: 365 * 24 * time.Hour, Color: rl.NewColor(200, 120, 55, 255), Label: "< 1 year"},    // orange
	{MaxAge: 3 * 365 * 24 * time.Hour, Color: rl.NewColor(170, 80, 60, 255), Label: "< 3 years"}, // rust
	// anything older falls through to the ancient color
}

// AncientColor is used for files older than all defined buckets.
var AncientColor = rl.NewColor(70, 100, 160, 255) // steel blue

// ColorFromAge returns a color based on the file's modification time.
// Uses the DefaultAgeBuckets for discrete bucket mapping.
func ColorFromAge(modTime time.Time) rl.Color {
	if modTime.IsZero() {
		return OtherColor
	}

	age := time.Since(modTime)
	if age < 0 {
		// Future modification time (clock skew) - treat as newest
		return DefaultAgeBuckets[0].Color
	}

	for _, bucket := range DefaultAgeBuckets {
		if age <= bucket.MaxAge {
			return bucket.Color
		}
	}

	return AncientColor
}

// ColorFromAgeSmooth returns a smoothly interpolated color based on file age.
// Uses HSV interpolation from green (newest) through yellow/orange to blue (oldest).
func ColorFromAgeSmooth(modTime time.Time) rl.Color {
	if modTime.IsZero() {
		return OtherColor
	}

	age := time.Since(modTime)
	if age < 0 {
		age = 0
	}

	// Map age to a 0.0-1.0 range using logarithmic scaling
	// 0.0 = just now, 1.0 = 5+ years old
	maxAgeDays := 5.0 * 365.0 // 5 years
	ageDays := age.Hours() / 24.0
	t := math.Log1p(ageDays) / math.Log1p(maxAgeDays)
	if t > 1.0 {
		t = 1.0
	}

	// HSV interpolation
	// Hue: 140 (green) -> 60 (yellow) -> 30 (orange) -> 0 (red) -> 230 (blue)
	var hue float64
	if t < 0.6 {
		// Green to orange (hue 140 -> 30)
		hue = 140.0 - (t/0.6)*110.0
	} else {
		// Orange to blue (hue 30 -> 230, going through red and wrapping)
		localT := (t - 0.6) / 0.4
		hue = 30.0 - localT*160.0
		if hue < 0 {
			hue += 360.0
		}
	}

	saturation := 0.65
	value := 0.82

	return hsvToColor(hue, saturation, value)
}

// QuantizedBucket returns a bucket index (0-31) for instanced rendering batching.
func QuantizedBucket(modTime time.Time) int {
	if modTime.IsZero() {
		return 31
	}

	age := time.Since(modTime)
	if age < 0 {
		return 0
	}

	// Logarithmic mapping to 32 buckets
	maxAgeDays := 5.0 * 365.0
	ageDays := age.Hours() / 24.0
	t := math.Log1p(ageDays) / math.Log1p(maxAgeDays)
	if t > 1.0 {
		t = 1.0
	}

	bucket := int(t * 31.0)
	if bucket > 31 {
		bucket = 31
	}
	return bucket
}

// BucketColor returns the pre-computed color for a given bucket index.
func BucketColor(bucket int) rl.Color {
	t := float64(bucket) / 31.0
	var hue float64
	if t < 0.6 {
		hue = 140.0 - (t/0.6)*110.0
	} else {
		localT := (t - 0.6) / 0.4
		hue = 30.0 - localT*160.0
		if hue < 0 {
			hue += 360.0
		}
	}
	return hsvToColor(hue, 0.65, 0.82)
}

// hsvToColor converts HSV (hue 0-360, saturation 0-1, value 0-1) to rl.Color.
func hsvToColor(h, s, v float64) rl.Color {
	h = math.Mod(h, 360.0)
	if h < 0 {
		h += 360.0
	}

	c := v * s
	x := c * (1.0 - math.Abs(math.Mod(h/60.0, 2.0)-1.0))
	m := v - c

	var r, g, b float64
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}

	return rl.NewColor(
		uint8((r+m)*255.0),
		uint8((g+m)*255.0),
		uint8((b+m)*255.0),
		255,
	)
}

// ColorFromSize returns a color based on file size relative to a maximum.
// Small files are cool (teal), large files are warm (amber/red).
func ColorFromSize(size int64, maxSize int64) rl.Color {
	if maxSize <= 0 {
		return FileColor
	}

	// Logarithmic scaling so small differences at the low end are visible
	t := math.Log1p(float64(size)) / math.Log1p(float64(maxSize))
	if t > 1.0 {
		t = 1.0
	}
	if t < 0.0 {
		t = 0.0
	}

	// HSV: hue 180 (teal/cyan, small) -> 40 (amber, medium) -> 0 (red, large)
	var hue float64
	if t < 0.5 {
		// Teal to amber
		hue = 180.0 - (t/0.5)*140.0
	} else {
		// Amber to red
		hue = 40.0 - ((t-0.5)/0.5)*40.0
	}

	saturation := 0.55 + t*0.2 // slightly more saturated for larger files
	value := 0.75 + (1.0-t)*0.1

	return hsvToColor(hue, saturation, value)
}

// LerpColor linearly interpolates between two colors.
func LerpColor(a, b rl.Color, t float32) rl.Color {
	if t <= 0 {
		return a
	}
	if t >= 1 {
		return b
	}
	return rl.NewColor(
		uint8(float32(a.R)+float32(int(b.R)-int(a.R))*t),
		uint8(float32(a.G)+float32(int(b.G)-int(a.G))*t),
		uint8(float32(a.B)+float32(int(b.B)-int(a.B))*t),
		uint8(float32(a.A)+float32(int(b.A)-int(a.A))*t),
	)
}
