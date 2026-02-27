package geneticalgorithm

import (
	"slices"
	"testing"
)

// TestMutate tests the mutation function.
func TestMutate(t *testing.T) {
	tests := []struct {
		name   string
		seq    []int
		maxVal int
		maxLen int
	}{
		{
			name:   "empty input and length",
			seq:    []int{},
			maxVal: 10,
			maxLen: 0},
		{
			name:   "empty input but not length",
			seq:    []int{},
			maxVal: 10,
			maxLen: 5},
		{
			name:   "standard input",
			seq:    []int{1, 2, 3, 4, 5},
			maxVal: 10,
			maxLen: 5},
	}

	for range 10 { // Repeat to avoid lucky runs due to rand
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				seq := mutate(slices.Clone(test.seq), test.maxVal, test.maxLen)
				if len(seq) > test.maxLen {
					t.Errorf("Length should not exceed maxLen, %v : %d", seq, len(seq))
				}
				if slices.Equal(seq, test.seq) && (len(test.seq) > 0 || len(seq) > 0) {
					t.Errorf("Child slice should never be a clone of parent: %v", seq)
				}
				if len(seq) > 0 && slices.Max(seq) > test.maxVal {
					t.Errorf("Max value in sequence, %v, should not exceed maxVal, %d", seq, test.maxVal)
				}
			})
		}
	}
}
