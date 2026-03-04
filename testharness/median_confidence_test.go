package testharness

import (
	"testing"
)

// TestEstimateMedianRSE tests the mathematical logic of the RSE approximation.
func TestEstimateMedianRSE(t *testing.T) {
	tests := []struct {
		name   string
		input  []float64
		isZero bool
		approx float64
	}{
		{
			name:   "Constant values",
			input:  []float64{10, 10, 10, 10, 10},
			isZero: true,
		},
		{
			name:   "All zeros edge case",
			input:  []float64{0, 0, 0, 0, 0},
			isZero: true,
		},
		{
			name:   "Zero median with non zero IQR doesn't cause 0 division",
			input:  []float64{-10, 0, 10},
			isZero: false,
			approx: 5.0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mRse := estimateMedianRSE(test.input)

			if test.isZero && mRse != 0.0 {
				t.Errorf("Expected 0, got %v", mRse)
			} else if mRse+1 < test.approx || mRse-1 > test.approx {
				t.Errorf("Unexpected relative standard error for %v : %f", test.input, mRse)
			}
		})
	}
}

// TestMedianConfidence tests the boundary conditions of the threshold comparison.
func TestMedianConfidence(t *testing.T) {
	tests := []struct {
		name      string
		input     []float64
		threshold float64
		pass      bool
	}{
		{
			name:      "RSE is 0, below threshold",
			input:     []float64{5, 5, 5, 5, 5},
			threshold: 0.1,
			pass:      true,
		},
		{
			name:      "RSE is 0, at threshold",
			input:     []float64{5, 5, 5, 5, 5},
			threshold: 0.0,
			pass:      false,
		},
		{
			name:      "No variance should pass even if median is small",
			input:     []float64{0, 0, 0, 0},
			threshold: 1.0,
			pass:      true,
		},
		{
			name:      "RSE should not explode when median is 0",
			input:     []float64{-10, 0, 10},
			threshold: 10.0,
			pass:      true,
		},
		{
			name:      "Negative median doesn't yield negative RSE",
			input:     []float64{-100, -100, -10, -1, -1},
			threshold: 0.0,
			pass:      false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pass := MedianConfidence(test.input, test.threshold)
			if pass != test.pass {
				t.Errorf("MedianConfidence() = %v, pass %v", pass, test.pass)
			}
		})
	}
}
