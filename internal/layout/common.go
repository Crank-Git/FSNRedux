package layout

import "math"

// scaleHeight converts a file size to a visual height using logarithmic scaling.
// This prevents massive files from dominating the view and tiny files from being invisible.
func scaleHeight(size int64, opts Options) float32 {
	if size <= 0 {
		return opts.MinHeight
	}

	// Logarithmic scaling: log2(size_in_kb + 1) * scale
	sizeKB := float64(size) / 1024.0
	h := float32(math.Log2(sizeKB+1)) * opts.HeightScale

	if h < opts.MinHeight {
		h = opts.MinHeight
	}
	if h > opts.MaxHeight {
		h = opts.MaxHeight
	}
	return h
}
