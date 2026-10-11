package dsp_test

import (
	"math"
	"testing"

	"pw-comp/dsp"
)

func TestFastPower2EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		x    float64
		want float64
	}{
		{name: "overflow", x: 1e6, want: math.Inf(1)},
		{name: "positive infinity", x: math.Inf(1), want: math.Inf(1)},
		{name: "underflow", x: -1e6, want: 0},
		{name: "negative infinity", x: math.Inf(-1), want: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := dsp.FastPower2(tc.x); got != tc.want {
				t.Errorf("FastPower2(%g) = %g, want %g", tc.x, got, tc.want)
			}
		})
	}

	if got := dsp.FastPower2(math.NaN()); !math.IsNaN(got) {
		t.Errorf("FastPower2(NaN) = %g, want NaN", got)
	}
}

func TestFastPower2Accuracy(t *testing.T) {
	t.Parallel()

	for x := -60.0; x <= 60.0; x += 0.37 {
		want := math.Exp2(x)
		if relErr := math.Abs(dsp.FastPower2(x)-want) / want; relErr > 1e-3 {
			t.Errorf("FastPower2(%g): relative error %g exceeds 1e-3", x, relErr)
		}
	}
}

func TestFastSqrtNonPositive(t *testing.T) {
	t.Parallel()

	for _, x := range []float64{0, -1e-18, -1} {
		if got := dsp.FastSqrt(x); got != 0 {
			t.Errorf("FastSqrt(%g) = %g, want 0", x, got)
		}
	}
}
