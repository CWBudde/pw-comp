package dsp

import "math"

// Polynomial coefficients for continuous error function approximation.
// These coefficients provide a fast log2 approximation using a 5th-order polynomial.
//
//nolint:gochecknoglobals // Mathematical constants used across all FastLog2 calls
var cl2Continuous5 = []float64{
	-0.0821343513178931783,
	0.649732456739820052,
	-2.13417801862571777,
	4.08642207062728868,
	-1.51984215742349793,
}

// FastLog2 provides a fast approximation of log2(x) using polynomial evaluation.
// This is significantly faster than math.Log2() with acceptable accuracy for audio DSP.
//
// The approximation works by:
// 1. Extracting the exponent and mantissa using math.Frexp
// 2. Approximating log2(mantissa) using a polynomial
// 3. Combining: log2(x) = exponent + log2(mantissa).
func FastLog2(x float64) float64 {
	// Handle edge cases to avoid log(0)
	if x <= 0 {
		return -math.Inf(1)
	}

	// Extract exponent and mantissa
	// Frexp breaks x into mantissa (0.5 <= frac < 1.0) and exponent
	frac, exp := math.Frexp(x)

	// Approximate log2(frac) using Horner's method for polynomial evaluation
	logMantissa := cl2Continuous5[0]*frac + cl2Continuous5[1]
	logMantissa = logMantissa*frac + cl2Continuous5[2]
	logMantissa = logMantissa*frac + cl2Continuous5[3]
	logMantissa = logMantissa*frac + cl2Continuous5[4]

	// Compute log2(x) = exponent + log2(mantissa)
	// Note: exp-1 because Frexp returns mantissa in [0.5, 1.0) range
	return float64(exp-1) + logMantissa
}

// Exponent range of float64: 2^1024 overflows to +Inf, 2^-1075 underflows to 0.
const (
	fastPower2MaxExp = 1024
	fastPower2MinExp = -1075
)

// Polynomial coefficients for 2^x approximation (3rd order minimum error).
// These coefficients provide a fast power-of-2 approximation for the fractional part.
//
//nolint:gochecknoglobals // Mathematical constants used across all FastPower2 calls
var cp2MinError3 = []float64{
	0.693292707161004662,
	0.242162975514835621,
	0.0548668824216034384,
}

// FastPower2 computes 2^x using 3rd-order polynomial approximation.
// This is significantly faster than math.Pow(2.0, x) with acceptable accuracy for audio DSP.
//
// The approximation works by:
// 1. Splitting x into integer and fractional parts
// 2. Approximating 2^fracPart using a polynomial
// 3. Scaling by 2^intPart through the exponent bits (math.Ldexp): 2^x = 2^fracPart * 2^intPart.
//
// Inputs above the float64 exponent range return +Inf and inputs below it return 0,
// matching math.Pow(2, x). NaN is propagated. The compressor only calls it with
// gain values in log2 space (x <= 0), far inside the valid range.
func FastPower2(x float64) float64 {
	// Guard the float64 exponent range before converting to int, which is
	// implementation-defined for out-of-range values
	switch {
	case math.IsNaN(x):
		return x
	case x >= fastPower2MaxExp:
		return math.Inf(1)
	case x <= fastPower2MinExp:
		return 0
	}

	// Split into integer and fractional parts
	intPart := int(math.Round(x))
	fracPart := x - float64(intPart)

	// Evaluate polynomial for fractional part using Horner's method
	// poly(f) = 1 + f * (c0 + f * (c1 + f * c2))
	polyResult := 1.0 + fracPart*(cp2MinError3[0]+
		fracPart*(cp2MinError3[1]+fracPart*cp2MinError3[2]))

	// Scale by 2^intPart via the exponent bits (very fast)
	return math.Ldexp(polyResult, intPart)
}

// FastSqrt computes sqrt(x) using Babylonian/Newton-Raphson method.
// with bit manipulation for a fast initial guess.
//
// The approximation works by:
// 1. Using bit manipulation to get an initial guess
// 2. Performing two Newton-Raphson iterations for refinement
// Formula: x_new = 0.5 * (x_old + value/x_old).
//
// FastSqrt returns 0 for x <= 0. This is exact for x == 0 and intentionally
// clamps negative inputs (which can only arise from rounding in callers such as
// the compressor's knee term) instead of returning NaN, so a gain calculation
// never propagates NaN into the audio path.
func FastSqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}

	// Get initial approximation via bit manipulation
	// This exploits the IEEE 754 representation to estimate sqrt
	bits := math.Float64bits(x)
	bits = ((bits - (1 << 52)) >> 1) + (1 << 61)
	result := math.Float64frombits(bits)

	// Two Newton-Raphson iterations for refinement
	// First iteration: standard formula
	result = 0.5 * (result + x/result)

	// Second iteration: apply the same formula again
	result = 0.5 * (result + x/result)

	return result
}
