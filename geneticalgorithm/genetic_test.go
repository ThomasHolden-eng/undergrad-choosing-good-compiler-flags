package geneticalgorithm

import (
	"os"
	"reflect"
	"slices"
	"testing"

	"6CCS3PRJ/thomasholden/passmanager"
	"6CCS3PRJ/thomasholden/util"
	"6CCS3PRJ/thomasholden/workerpool"
)

// TestestreatePopulation tests the creation of a population.
func TestCreatePopulation(t *testing.T) {
	tempDir := t.TempDir()
	dummyFile := tempDir + "/hello_world.c"
	err := os.WriteFile(dummyFile, []byte("int main() { return 0; }"), 0644)
	if err != nil {
		t.Skip("Failed to create dummy file for testing: " + err.Error())
	}

	tests := []struct {
		name       string
		minPopSize int
		maxLength  int
		command    string
		expectErr  bool
	}{
		{
			name:       "valid command and sizes",
			minPopSize: 10,
			maxLength:  50,
			command:    "clang " + dummyFile,
			expectErr:  false,
		},
		{
			name:       "invalid shell command",
			minPopSize: 10,
			maxLength:  50,
			command:    "clang 'hello_world.c",
			expectErr:  true,
		},
		{
			name:       "invalid file path",
			minPopSize: 10,
			maxLength:  50,
			command:    "clang hello_world.c",
			expectErr:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pop, err := CreatePopulation(test.minPopSize, test.maxLength, test.command)

			if test.expectErr {
				if err == nil {
					t.Fatalf("expected error for command %q, got nil", test.command)
				}
			} else {
				if err == nil && pop != nil {
					if pop.maxLength != test.maxLength {
						t.Errorf("expected maxLength %d, got %d", test.maxLength, pop.maxLength)
					}
					if len(pop.members) == 0 {
						t.Errorf("population members should be initialized, got length 0")
					}
				} else if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
			}
		})
	}
}

// TestGenerateRandomIntSlice tests the generation of a random int slice.
func TestGenerateRandomIntSlice(t *testing.T) {
	tests := []struct {
		name      string
		passMan   *passmanager.PassManager
		maxLength int
	}{
		{
			name:      "zero length",
			passMan:   passmanager.GetPassManager(),
			maxLength: 0,
		},
		{
			name:      "small length",
			passMan:   passmanager.GetPassManager(),
			maxLength: 5,
		},
		{
			name:      "~O3 length",
			passMan:   passmanager.GetPassManager(),
			maxLength: 75,
		},
		{
			name:      "big length",
			passMan:   passmanager.GetPassManager(),
			maxLength: 150,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pop := &Population{
				passMan:   test.passMan,
				maxLength: test.maxLength,
			}

			returnSlice := pop.generateRandomIntSlice()
			if len(returnSlice) > pop.maxLength {
				t.Errorf("expected max length %d, got %d", pop.maxLength, len(returnSlice))
			}
		})
	}
}

// TestElitism tests the elitism function.
func TestElitism(t *testing.T) {
	tests := []struct {
		name        string
		results     []workerpool.Result
		marker      int
		expectPanic bool
		expectedLen int
	}{
		{
			name: "success with valid results",
			results: []workerpool.Result{
				{Score: 1.0, PassSeq: []int{1, 2}},
				{Score: 2.5, PassSeq: []int{3, 4}},
			},
			marker:      2,
			expectPanic: false,
			expectedLen: 2,
		},
		{
			name: "all results failing (panic)",
			results: []workerpool.Result{
				{Score: util.MAX_FLOAT64, PassSeq: []int{1}},
			},
			marker:      1,
			expectPanic: true,
			expectedLen: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := &Population{
				members: make([][]int, 0),
			}

			if test.expectPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("expected elitism to panic when all sequences fail, but it did not")
					}
				}()
			}

			reproMarker := p.elitism(test.results, test.marker)

			if !test.expectPanic {
				if len(p.members) != test.expectedLen {
					t.Errorf("expected members length %d, got %d", test.expectedLen, len(p.members))
				}
				if reproMarker < 0 {
					t.Errorf("expected valid reproMarker, got %d", reproMarker)
				}
			}
		})
	}
}

// TestReproduction tests the reproduction function.
func TestReproduction(t *testing.T) {
	p := &Population{
		members:   make([][]int, 2, 5),
		passMan:   &passmanager.PassManager{MaxPass: 10},
		maxLength: 10,
	}
	p.members[0] = []int{1, 2, 3}
	p.members[1] = []int{4, 5, 6}

	t.Run("fills capacity", func(t *testing.T) {
		p.reproduction(1)
		if len(p.members) != 5 {
			t.Errorf("expected members length 5, got %d", len(p.members))
		}
	})
}

// TestMutation tests the mutation function.
func TestMutation(t *testing.T) {
	p := &Population{
		members:   make([][]int, 5),
		passMan:   &passmanager.PassManager{MaxPass: 20},
		maxLength: 10,
	}
	p.members[0] = []int{1, 2, 3}
	p.members[1] = []int{4, 5, 6}
	p.members[2] = []int{1, 2, 3}
	p.members[3] = []int{4, 5, 6}
	p.members[4] = []int{1, 2, 3}

	p.mutation(1)

	t.Run("modifies correct slice range", func(t *testing.T) {
		if len(p.members) != 5 {
			t.Errorf("expected members length 5, got %d", len(p.members))
		}
		count := 0
		for i, seq := range p.members[2:] {
			if slices.Equal(seq, p.members[i%2]) {
				count++
			}
		}
		if count == 3 {
			t.Errorf("expected at least 1 mutation, got None")
		}
	})

	t.Run("doesn't modify incorrect slice range", func(t *testing.T) {
		if !slices.Equal(p.members[0], []int{1, 2, 3}) {
			t.Errorf("should not modify elites, but modified at 0")
		}
		if !slices.Equal(p.members[1], []int{4, 5, 6}) {
			t.Errorf("should not modify elites, but modified at 1")
		}
	})
}

// TestSortPopulation tests the sortPopulation function.
func TestSortPopulation(t *testing.T) {
	tests := []struct {
		name     string
		input    []workerpool.Result
		expected []workerpool.Result
	}{
		{
			name:     "already sorted",
			input:    []workerpool.Result{{Score: 1.0}, {Score: 2.0}, {Score: 3.0}},
			expected: []workerpool.Result{{Score: 1.0}, {Score: 2.0}, {Score: 3.0}},
		},
		{
			name:     "reverse sorted",
			input:    []workerpool.Result{{Score: 3.0}, {Score: 2.0}, {Score: 1.0}},
			expected: []workerpool.Result{{Score: 1.0}, {Score: 2.0}, {Score: 3.0}},
		},
		{
			name:     "with equals",
			input:    []workerpool.Result{{Score: 2.0}, {Score: 1.0}, {Score: 2.0}},
			expected: []workerpool.Result{{Score: 1.0}, {Score: 2.0}, {Score: 2.0}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := &Population{}
			p.sortPop(test.input)

			if !reflect.DeepEqual(test.input, test.expected) {
				t.Errorf("expected %v, got %v", test.expected, test.input)
			}
		})
	}
}

// TestConvergenceTest tests convergence is calculated correctly.
func TestConvergenceTest(t *testing.T) {
	tests := []struct {
		name           string
		members        [][]int
		lastSafePop    []int
		initialConv    int
		expectedConv   int
		expectedResult bool
	}{
		{
			name:           "diverging population",
			members:        [][]int{{1, 2}, {3, 4}, {5, 6}, {7, 8}},
			lastSafePop:    []int{9, 9},
			initialConv:    3,
			expectedConv:   0,
			expectedResult: false,
		},
		{
			name:           "converging population",
			members:        [][]int{{1, 2}, {3, 4}, {5, 6}, {7, 8}},
			lastSafePop:    []int{1, 2},
			initialConv:    2,
			expectedConv:   3,
			expectedResult: false,
		},
		{
			name:           "reached convergence threshold",
			members:        [][]int{{1, 2}, {3, 4}, {5, 6}, {7, 8}},
			lastSafePop:    []int{1, 2},
			initialConv:    4,
			expectedConv:   5,
			expectedResult: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := &Population{
				members:     test.members,
				lastSafePop: test.lastSafePop,
				convCount:   test.initialConv,
			}

			result := p.testConvergence()

			if p.convCount != test.expectedConv {
				t.Errorf("expected convCount %d, got %d", test.expectedConv, p.convCount)
			}
			if result != test.expectedResult {
				t.Errorf("expected result %v, got %v", test.expectedResult, result)
			}

			safeMarker := int(SAFE_POP * float64(len(p.members)))
			safeMarker = max(safeMarker-1, 0)
			if !slices.Equal(p.lastSafePop, p.members[safeMarker]) {
				t.Errorf("expected lastSafePop to be updated to %v, got %v", p.members[safeMarker], p.lastSafePop)
			}
		})
	}
}
