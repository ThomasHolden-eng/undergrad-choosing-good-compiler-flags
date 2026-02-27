package geneticalgorithm

import (
	"6CCS3PRJ/thomasholden/util"
	"slices"
	"testing"
	"time"

	"github.com/schollz/progressbar/v3"
)

// pbarInit ensure PBar is initialized for tests.
func pbarInit() {
	if PBar == nil {
		PBar = progressbar.DefaultSilent(-1)
	}
}

// getTestPassManager creates a pass manager for testing purposes.
func getTestPassManager() *PassManager {
	return GetPassManager()
}

// TestValueFromJobStats tests the ValueFromJobStats function.
func TestValueFromJobStats(t *testing.T) {
	pbarInit()

	tests := []struct {
		name        string
		compilation time.Duration
		execution   time.Duration
		size        int
		memory      int64
		want        float64
	}{
		{
			name:        "Zero values",
			compilation: 0,
			execution:   0,
			size:        0,
			memory:      0,
			want:        0.0,
		},
		{
			name:        "Random values",
			compilation: time.Second,
			execution:   time.Second,
			size:        1024,
			memory:      2048,
			want: (COMPILATION_WEIGHT*float64(time.Second.Microseconds()) +
				EXECUTION_WEIGHT*float64(time.Second.Microseconds()) +
				BINARY_SIZE_WEIGHT*float64(1024) +
				MEMORY_USAGE_WEIGHT*float64(2048)),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := valueFromJobStats(test.compilation, test.execution, test.size, test.memory)
			if got != test.want {
				t.Errorf("valueFromJobStats() = %v, want %v", got, test.want)
			}
		})
	}
}

// TestWorkerPoolAddPassSequence tests the AddPassSequence function.
func TestWorkerPoolAddPassSequence(t *testing.T) {
	pbarInit()

	pool := &WorkerPool{
		passSeq: make([][]int, 0),
	}

	seq := []int{1, 2, 3}
	pool.AddPassSequence(seq)

	if len(pool.passSeq) != 1 {
		t.Fatalf("Expected passSeq length 1, got %d", len(pool.passSeq))
	}

	if !slices.Equal(pool.passSeq[0], seq) {
		t.Errorf("Expected sequence %v, got %v", seq, pool.passSeq[0])
	}
}

// TestWorkerPoolStartPoolCached verifies the worker pool logic by mocking the cache.
//
// This allows us to test the pool management, threading, and result aggregation
// without triggering the compiler.
func TestWorkerPoolStartPoolCached(t *testing.T) {
	pbarInit()

	workers := 2
	numJobs := 3

	// Bypass NewWorkerPool's external calls
	pool := &WorkerPool{
		passSeq:    make([][]int, 0, numJobs),
		fileNum:    make(chan int, workers),
		SourceCmd:  []string{"clang", "main.c"},
		passMan:    getTestPassManager(),
		approxTime: 0 * time.Second,
		approxMem:  0,
	}

	for i := 0; i < workers; i++ {
		pool.fileNum <- i
	}

	pool.cache.Init(256)

	inputSequences := [][]int{
		{10, 20},
		{30, 40},
		{50, 60},
	}

	expectedScores := map[string]float64{
		pool.passMan.makeValid(inputSequences[0]): 100.0,
		pool.passMan.makeValid(inputSequences[1]): 200.0,
		pool.passMan.makeValid(inputSequences[2]): 300.0,
	}

	for _, seq := range inputSequences {
		pool.AddPassSequence(seq)
		validSeq := pool.passMan.makeValid(seq)
		pool.cache.Set(validSeq, expectedScores[validSeq])
	}

	t.Log("Starting Pool")
	results := pool.StartPool()
	t.Log("Finishing Pool")

	if len(results) != numJobs {
		t.Fatalf("Expected %d results, got %d", numJobs, len(results))
	}

	resultMap := make(map[string]float64)
	for _, r := range results {
		key := pool.passMan.makeValid(r.passSeq)
		resultMap[key] = r.score
	}

	for key, wantScore := range expectedScores {
		gotScore, ok := resultMap[key]
		if !ok {
			t.Errorf("Result missing for sequence %s", key)
			continue
		}
		if gotScore != wantScore {
			t.Errorf("Score mismatch for %s: got %v, want %v", key, gotScore, wantScore)
		}
	}

	t.Log("Waiting on channel")
	select {
	case _, ok := <-pool.Results:
		if ok {
			t.Error("Results channel should be closed")
		}
	default:
	}
}

// TestWorkerPoolQueueErrorResult tests the queueErrorResult function.
func TestWorkerPoolQueueErrorResult(t *testing.T) {
	pbarInit()

	pool := &WorkerPool{
		Results: make(chan Result, 1),
	}

	seq := []int{999}
	pool.queueErrorResult(seq)

	close(pool.Results)

	res := <-pool.Results
	if res.score != util.MAX_FLOAT64 {
		t.Errorf("Expected error score %v, got %v", util.MAX_FLOAT64, res.score)
	}
}
