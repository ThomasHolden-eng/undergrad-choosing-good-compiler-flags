package util

import (
	"slices"
	"strconv"
	"testing"
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := tt.setup()
			got := ArrayFromChannel(ch)
			if !slices.Equal(got, tt.expected) {
				t.Errorf("generated array %v, wanted %v", got, tt.expected)
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := linearInterpolate(tt.list, tt.tVal)
			if got != tt.expected {
				t.Errorf("got interpolated value %v, wanted %v", got, tt.expected)
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listCopy := slices.Clone(tt.list)
			got := Median(listCopy)
			if got != tt.expected {
				t.Errorf("got median of %v, want %v", got, tt.expected)
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listCopy := slices.Clone(tt.list)
			got := IQR(listCopy)
			if got != tt.expected {
				t.Errorf("got iqr of %v, wanted %v", got, tt.expected)
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.rng.Within(tt.val)
			if got != tt.expected {
				t.Errorf("Range(%v, %v).Within(%v) = %v, want %v",
					tt.rng.start, tt.rng.end, tt.val, got, tt.expected)
			}
		})
	}
}
