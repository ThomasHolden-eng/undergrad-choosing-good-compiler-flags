package util

import (
	"slices"
	"strconv"
	"testing"
	"time"
)

func TestArrayFromChannel(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() chan int
		expected []int
	}{
		{
			name: "buffered channel with elements",
			setup: func() chan int {
				ch := make(chan int, 3)
				ch <- 1
				ch <- 2
				ch <- 3
				close(ch)
				return ch
			},
			expected: []int{1, 2, 3},
		},
		{
			name: "empty channel",
			setup: func() chan int {
				ch := make(chan int)
				close(ch)
				return ch
			},
			expected: []int{},
		},
		{
			name: "unbuffered channel populated via goroutine",
			setup: func() chan int {
				ch := make(chan int)
				go func() {
					for i := 10; i <= 12; i++ {
						ch <- i
					}
					close(ch)
				}()
				return ch
			},
			expected: []int{10, 11, 12},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ch := test.setup()
			got := ArrayFromChannel(ch)
			if !slices.Equal(got, test.expected) {
				t.Errorf("generated array %v, wanted %v", got, test.expected)
			}
		})
	}
}

func TestLinearInterpolate(t *testing.T) {
	tests := []struct {
		name     string
		list     []float64
		tVal     float64
		expected float64
	}{
		{
			name:     "empty list",
			list:     []float64{},
			tVal:     0.5,
			expected: 0.0,
		},
		{
			name:     "single element",
			list:     []float64{42.0},
			tVal:     0.5,
			expected: 42.0,
		},
		{
			name:     "two elements median",
			list:     []float64{10.0, 20.0},
			tVal:     0.5,
			expected: 15.0,
		},
		{
			name:     "three elements median",
			list:     []float64{10.0, 20.0, 30.0},
			tVal:     0.5,
			expected: 20.0,
		},
		{
			name:     "four elements lower quartile",
			list:     []float64{10.0, 20.0, 30.0, 40.0},
			tVal:     0.25,
			expected: 17.5,
		},
		{
			name:     "four elements upper quartile",
			list:     []float64{10.0, 20.0, 30.0, 40.0},
			tVal:     0.75,
			expected: 32.5,
		},
		{
			name:     "lower bound",
			list:     []float64{10.0, 20.0},
			tVal:     0.0,
			expected: 10.0,
		},
		{
			name:     "upper bound",
			list:     []float64{10.0, 20.0},
			tVal:     1.0,
			expected: 20.0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := linearInterpolate(test.list, test.tVal)
			if got != test.expected {
				t.Errorf("got interpolated value %v, wanted %v", got, test.expected)
			}
		})
	}
}

func TestMedian(t *testing.T) {
	tests := []struct {
		name     string
		list     []int
		expected float64
	}{
		{
			name:     "empty list returns MAX_FLOAT64",
			list:     []int{},
			expected: MAX_FLOAT64,
		},
		{
			name:     "odd number of elements",
			list:     []int{3, 1, 2},
			expected: 2.0,
		},
		{
			name:     "even number of elements",
			list:     []int{1, 4, 2, 3},
			expected: 2.5,
		},
		{
			name:     "negative values",
			list:     []int{-10, -50, -30},
			expected: -30.0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			listCopy := slices.Clone(test.list)
			got := Median(listCopy)
			if got != test.expected {
				t.Errorf("got median of %v, want %v", got, test.expected)
			}
		})
	}
}

func TestIQR(t *testing.T) {
	tests := []struct {
		name     string
		list     []float64
		expected float64
	}{
		{
			name:     "empty list returns MAX_FLOAT64",
			list:     []float64{},
			expected: MAX_FLOAT64,
		},
		{
			name:     "standard sequence",
			list:     []float64{10.0, 20.0, 30.0, 40.0},
			expected: 15.0,
		},
		{
			name:     "single element sequence",
			list:     []float64{10.0},
			expected: 0.0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			listCopy := slices.Clone(test.list)
			got := IQR(listCopy)
			if got != test.expected {
				t.Errorf("got iqr of %v, wanted %v", got, test.expected)
			}
		})
	}
}

func TestForEach(t *testing.T) {
	t.Run("int to int (multiply by 2)", func(t *testing.T) {
		input := []int{1, 2, 3}
		expected := []int{2, 4, 6}
		got := ForEach(input, func(i int) int {
			return i * 2
		})
		if !slices.Equal(got, expected) {
			t.Errorf("applied foreach got %v, wanted %v", got, expected)
		}
	})

	t.Run("int to string format", func(t *testing.T) {
		input := []int{10, 20}
		expected := []string{"10", "20"}
		got := ForEach(input, strconv.Itoa)
		if !slices.Equal(got, expected) {
			t.Errorf("applied foreach got %v, wanted %v", got, expected)
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		input := []int{}
		expected := []int{}
		got := ForEach(input, func(i int) int { return i + 1 })
		if !slices.Equal(got, expected) {
			t.Errorf("foreach on empty slice got %v, want %v", got, expected)
		}
	})
}

func TestRange_Within(t *testing.T) {
	tests := []struct {
		name     string
		rng      *Range[int]
		val      int
		expected bool
	}{
		{
			name:     "value strictly inside",
			rng:      NewRange(10, 20),
			val:      15,
			expected: true,
		},
		{
			name:     "value on lower boundary",
			rng:      NewRange(10, 20),
			val:      10,
			expected: true,
		},
		{
			name:     "value on upper boundary",
			rng:      NewRange(10, 20),
			val:      20,
			expected: true,
		},
		{
			name:     "value below lower boundary",
			rng:      NewRange(10, 20),
			val:      9,
			expected: false,
		},
		{
			name:     "value above upper boundary",
			rng:      NewRange(10, 20),
			val:      21,
			expected: false,
		},
		{
			name:     "inverted range bounds (start > end)",
			rng:      NewRange(20, 10),
			val:      15,
			expected: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.rng.Within(test.val)
			if got != test.expected {
				t.Errorf("Range(%v, %v).Within(%v) = %v, want %v",
					test.rng.start, test.rng.end, test.val, got, test.expected)
			}
		})
	}
}

func TestGetMinMaxTimes(t *testing.T) {
	withinTolerance := func(a, b time.Duration) bool {
		const tolerance = 50 * time.Millisecond

		diff := a - b
		if diff < 0 {
			diff = -diff
		}
		return diff <= tolerance
	}

	tests := []struct {
		name        string
		elapsedTime time.Duration
		maxTime     time.Duration
		genCount    int
		convergence int
		wantMinT    time.Duration
		wantMaxT    time.Duration
	}{
		{
			name:        "Zero generations completed",
			elapsedTime: 10 * time.Second,
			maxTime:     100 * time.Second,
			genCount:    0,
			convergence: 0,
			wantMinT:    0,
			wantMaxT:    90 * time.Second,
		},
		{
			name:        "Early generations, bottlenecked by convergence",
			elapsedTime: 10 * time.Second,
			maxTime:     100 * time.Second,
			genCount:    5,
			convergence: 2,
			wantMinT:    16 * time.Second,
			wantMaxT:    90 * time.Second,
		},
		{
			name:        "High convergencce, bottlenecked by early generation",
			elapsedTime: 8 * time.Second,
			maxTime:     100 * time.Second,
			genCount:    8,
			convergence: 9,
			wantMinT:    3 * time.Second,
			wantMaxT:    92 * time.Second,
		},
		{
			name:        "Zero minGens expected",
			elapsedTime: 40 * time.Second,
			maxTime:     100 * time.Second,
			genCount:    20,
			convergence: 10,
			wantMinT:    0,
			wantMaxT:    60 * time.Second,
		},
		{
			name:        "minT caps to maxT",
			elapsedTime: 95 * time.Second,
			maxTime:     100 * time.Second,
			genCount:    10,
			convergence: 0,
			wantMinT:    5 * time.Second,
			wantMaxT:    5 * time.Second,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			startTime := time.Now().Add(-test.elapsedTime)
			minT, maxT := GetMinMaxTimes(startTime, test.maxTime, test.genCount, test.convergence)

			if !withinTolerance(minT, test.wantMinT) {
				t.Errorf("Got min of %v, expected %v", minT, test.wantMinT)
			}
			if !withinTolerance(maxT, test.wantMaxT) {
				t.Errorf("Got max of %v, expected %v", maxT, test.wantMaxT)
			}
		})
	}
}
