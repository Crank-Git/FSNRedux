package color

import (
	"testing"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestColorFromAge_Newest(t *testing.T) {
	c := ColorFromAge(time.Now())
	// Should be in the "< 1 day" bucket (bright green)
	expected := DefaultAgeBuckets[0].Color
	if c != expected {
		t.Errorf("newest file color: got %v, want %v", c, expected)
	}
}

func TestColorFromAge_Ancient(t *testing.T) {
	old := time.Now().Add(-10 * 365 * 24 * time.Hour) // 10 years ago
	c := ColorFromAge(old)
	if c != AncientColor {
		t.Errorf("ancient file color: got %v, want %v", c, AncientColor)
	}
}

func TestColorFromAge_ZeroTime(t *testing.T) {
	c := ColorFromAge(time.Time{})
	if c != OtherColor {
		t.Errorf("zero time color: got %v, want %v", c, OtherColor)
	}
}

func TestColorFromAge_FutureTime(t *testing.T) {
	future := time.Now().Add(24 * time.Hour)
	c := ColorFromAge(future)
	expected := DefaultAgeBuckets[0].Color
	if c != expected {
		t.Errorf("future time color: got %v, want %v", c, expected)
	}
}

func TestColorFromAge_EachBucket(t *testing.T) {
	bucketAges := []time.Duration{
		12 * time.Hour,            // < 1 day
		3 * 24 * time.Hour,        // < 1 week
		15 * 24 * time.Hour,       // < 1 month
		90 * 24 * time.Hour,       // < 6 months
		200 * 24 * time.Hour,      // < 1 year
		2 * 365 * 24 * time.Hour,  // < 3 years
	}

	for i, age := range bucketAges {
		modTime := time.Now().Add(-age)
		c := ColorFromAge(modTime)
		expected := DefaultAgeBuckets[i].Color
		if c != expected {
			t.Errorf("bucket %d (%s): got %v, want %v",
				i, DefaultAgeBuckets[i].Label, c, expected)
		}
	}
}

func TestColorFromAgeSmooth_Range(t *testing.T) {
	// Just verify it doesn't panic for various ages
	ages := []time.Duration{
		0,
		time.Hour,
		24 * time.Hour,
		30 * 24 * time.Hour,
		365 * 24 * time.Hour,
		10 * 365 * 24 * time.Hour,
	}

	for _, age := range ages {
		modTime := time.Now().Add(-age)
		c := ColorFromAgeSmooth(modTime)
		if c.A != 255 {
			t.Errorf("alpha should be 255 for age %v, got %d", age, c.A)
		}
	}
}

func TestQuantizedBucket_Bounds(t *testing.T) {
	// Newest -> bucket 0
	b := QuantizedBucket(time.Now())
	if b != 0 {
		t.Errorf("newest bucket: got %d, want 0", b)
	}

	// Very old -> bucket 31
	old := time.Now().Add(-100 * 365 * 24 * time.Hour)
	b = QuantizedBucket(old)
	if b != 31 {
		t.Errorf("oldest bucket: got %d, want 31", b)
	}

	// Zero time -> bucket 31
	b = QuantizedBucket(time.Time{})
	if b != 31 {
		t.Errorf("zero time bucket: got %d, want 31", b)
	}
}

func TestBucketColor_AllBuckets(t *testing.T) {
	// Verify all 32 buckets produce valid colors
	for i := 0; i < 32; i++ {
		c := BucketColor(i)
		if c.A != 255 {
			t.Errorf("bucket %d alpha should be 255, got %d", i, c.A)
		}
		// Verify it's not black (all channels > 0 for visible colors)
		if c.R == 0 && c.G == 0 && c.B == 0 {
			t.Errorf("bucket %d is black, expected visible color", i)
		}
	}
}

func TestLerpColor(t *testing.T) {
	black := rl.NewColor(0, 0, 0, 255)
	white := rl.NewColor(255, 255, 255, 255)

	// t=0 -> black
	c := LerpColor(black, white, 0)
	if c != black {
		t.Errorf("lerp 0: got %v, want %v", c, black)
	}

	// t=1 -> white
	c = LerpColor(black, white, 1)
	if c != white {
		t.Errorf("lerp 1: got %v, want %v", c, white)
	}

	// t=0.5 -> mid gray
	c = LerpColor(black, white, 0.5)
	if c.R < 120 || c.R > 135 {
		t.Errorf("lerp 0.5: R=%d, expected ~127", c.R)
	}
}

func TestHsvToColor_PrimaryColors(t *testing.T) {
	// Red
	r := hsvToColor(0, 1, 1)
	if r.R < 250 || r.G > 5 || r.B > 5 {
		t.Errorf("red: got %v", r)
	}

	// Green
	g := hsvToColor(120, 1, 1)
	if g.R > 5 || g.G < 250 || g.B > 5 {
		t.Errorf("green: got %v", g)
	}

	// Blue
	b := hsvToColor(240, 1, 1)
	if b.R > 5 || b.G > 5 || b.B < 250 {
		t.Errorf("blue: got %v", b)
	}
}
