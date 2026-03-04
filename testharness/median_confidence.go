package testharness

import (
	"6CCS3PRJ/thomasholden/util"
	"math"
)

// MedianConfidence returns true if relative error of the slice is less than threshold.
func MedianConfidence[T util.Number](list []T, threshold float64) bool {
	return estimateMedianRSE(list) < threshold
}

// estimateRSE provides a value for the relative standard error of a slice w.r.t the median.
//
// We approximate sample σ, s with the formula: s ≈ IQR / 1.35 ≈ IQR * 0.74 [1]
// Then use the formula: σ^2 = 1/(4*n*f(m)^2) [2] for the population variation
// and find population σ = 1/(2*sqrt(n)*f(m)). Using that f(m) = 1/sqrt(2*π*s^2),
// σ = sqrt(2*π*s^2)/(2*sqrt(n)) = sqrt(π)*s/sqrt(2*n) = sqrt(π/2) * s/sqrt(n)
//
// The σ of the sampling distribution is the standard error [3].
//
// [1] Higgins JPT, Thomas J, Chandler J, Cumpston M, Li T, Page MJ, Welch VA (editors).
// Cochrane Handbook for Systematic Reviews of Interventions. 2nd Edition. Chichester (UK):
// John Wiley & Sons; 2019.
// [2] Williams, D. (2001). Weighing the Odds. Cambridge University Press. p. 165. ISBN 052100618X.
// [3] Everitt, B. S. (2003). The Cambridge Dictionary of Statistics. Cambridge University Press.
// ISBN 978-0-521-81099-9.
func estimateMedianRSE[T util.Number](list []T) float64 {
	const (
		ratio                float64 = 1 / 1.35
		fmMedianSqrt         float64 = 1.253314137 // math.Sqrt(math.Pi / 2.0) can't be const
		ratioDotFmMedianSqrt float64 = ratio * fmMedianSqrt
	)

	iqr := util.IQR(list) // Cannot be negative
	median := math.Abs(util.Median(list))
	sqrtN := math.Sqrt(float64(len(list)))

	if median <= 1 { // Edge case wherein we just return σ of the sampling distribution
		return (ratioDotFmMedianSqrt * iqr) / sqrtN
	}

	return (ratioDotFmMedianSqrt * iqr) / (sqrtN * median)
}
