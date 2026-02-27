package geneticalgorithm

import (
	"slices"
	"testing"
)

// slice1 is the first parent slice
var slice1 []int = []int{0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20}

// slice2 is the second parent slice
var slice2 []int = []int{1, 3, 5, 7, 9, 11, 13, 15, 17, 19, 21}

// crossoverTeardown verifies that the slices remain unmodified
func crossoverTeardown(t *testing.T) {
	var slice1Original []int = []int{0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20}
	var slice2Original []int = []int{1, 3, 5, 7, 9, 11, 13, 15, 17, 19, 21}
	if !slices.Equal(slice1, slice1Original) {
		t.Logf("Crossover should not affect original slice, %v : %v", slice1, slice1Original)
		slice1 = slice1Original
		t.FailNow()
	}
	if !slices.Equal(slice2, slice2Original) {
		t.Logf("Crossover should not affect original slice, %v : %v", slice2, slice2Original)
		slice2 = slice2Original
		t.FailNow()
	}
}

// TestReproduce tests the high level reproduce function
//
// Performs many iterations to ensure the randomness of the function
// doesn't allow for lucky passes.
func TestReproduce(t *testing.T) {
	defer crossoverTeardown(t)
	for range 300 {
		slice3, slice4 := reproduce(slice1, slice2, 21, 22)

		if len(slice3) > 22 {
			t.Fatalf("Length should not exceed maxLen, %v : %d", slice3, len(slice3))
		}
		if len(slice4) > 22 {
			t.Fatalf("Length should not exceed maxLen, %v : %d", slice4, len(slice4))
		}
		if slices.Equal(slice1, slice3) || slices.Equal(slice2, slice3) {
			t.Fatalf("Child slice should never be a clone of parent: %v", slice3)
		}
		if slices.Equal(slice1, slice4) || slices.Equal(slice2, slice4) {
			t.Fatalf("Child slice should never be a clone of parent: %v", slice4)
		}
	}
}

// TestTwoPoint tests two point crossover
func TestTwoPoint(t *testing.T) {
	defer crossoverTeardown(t)
	for range 100 {
		slice3, slice4 := twoPointCrossover(slice1, slice2)

		if slices.Equal(slice1, slice3) || slices.Equal(slice2, slice3) {
			t.Fatalf("Two Point child slice should never be a clone of parent: %v", slice3)
		}
		if slices.Equal(slice1, slice4) || slices.Equal(slice2, slice4) {
			t.Fatalf("Two Point child slice should never be a clone of parent: %v", slice4)
		}
	}
}

// TestLCS tests longest common subsequence crossover
func TestLCS(t *testing.T) {
	defer crossoverTeardown(t)
	for range 100 {
		slice3, slice4 := lcsCrossover(slice1, slice2)

		if slices.Equal(slice1, slice3) || slices.Equal(slice2, slice3) {
			t.Fatalf("LCS child slice should never be a clone of parent: %v", slice3)
		}
		if slices.Equal(slice1, slice4) || slices.Equal(slice2, slice4) {
			t.Fatalf("LCS child slice should never be a clone of parent: %v", slice4)
		}
	}
}

// TestOnePoint tests one point crossover
func TestOnePoint(t *testing.T) {
	defer crossoverTeardown(t)
	for range 100 {
		slice3, slice4 := onePointCrossover(slice1, slice2)

		if slices.Equal(slice1, slice3) || slices.Equal(slice2, slice3) {
			t.Fatalf("One Point child slice should never be a clone of parent: %v", slice3)
		}
		if slices.Equal(slice1, slice4) || slices.Equal(slice2, slice4) {
			t.Fatalf("One Point child slice should never be a clone of parent: %v", slice4)
		}
	}
}

// TestEmpty tests empty slice handling for all crossovers
func TestEmpty(t *testing.T) {
	a, b := reproduce([]int{}, []int{}, 10, 10)
	if len(a) <= 0 || len(b) <= 0 {
		t.Fatalf("Reproduce should ensure offspring are not of 0 length")
	}

	c, d := onePointCrossover([]int{}, []int{})
	e, f := twoPointCrossover([]int{}, []int{})
	g, h := lcsCrossover([]int{}, []int{})
	if len(c) != 0 || len(d) != 0 ||
		len(e) != 0 || len(f) != 0 ||
		len(g) != 0 || len(h) != 0 {
		t.Fatalf("Crossover should fallback to cloning for len input == 0")
	}
}
