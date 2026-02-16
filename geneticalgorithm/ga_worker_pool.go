package geneticalgorithm

import (
	"6CCS3PRJ/thomasholden/testharness"
	"6CCS3PRJ/thomasholden/util"
	"fmt"

	"sync"
	"time"
)

// Result holds the results for a singular pass sequence's execution
type Result struct {
	passSeq []int
	score   float64
}

// WorkerPool provides functionality for pooling the execution of LLVM optimisation executions.
type WorkerPool struct {
	passSeq    [][]int
	SourceCmd  []string // the command from which the project should be built
	cache      util.Cache[string, float64]
	fileNum    chan int
	Results    chan Result // a channel containing all results
	passMan    *PassManager
	approxTime time.Duration
	approxMem  int64
}

const (
	COMPILATION_WEIGHT  float64 = 0.0 // The proportion of the score to be made up of compile time.
	MEMORY_USAGE_WEIGHT float64 = 0.0 // The proportion of the score to be made up of execution's max RSS.
	BINARY_SIZE_WEIGHT  float64 = 0.0 // The proportion of the score to be made up of compiled binary size.

	// The proportion of the score to be made up of execution time.
	EXECUTION_WEIGHT float64 = 1 - COMPILATION_WEIGHT - MEMORY_USAGE_WEIGHT - BINARY_SIZE_WEIGHT
)

// NewWorkerPool returns a pointer to a worker pool.
func NewWorkerPool(sourceCmd []string, workers, numJobs int, baseline []int) (*WorkerPool, error) {
	pool := new(WorkerPool)

	pool.passSeq = make([][]int, 0, numJobs)
	pool.fileNum = make(chan int, workers)

	// Preps the channel that controls how many LLVM sessions can be active at once
	for i := 0; i < workers; i++ {
		pool.fileNum <- i
	}

	pool.SourceCmd = sourceCmd

	pool.passMan = GetPassManager()

	pool.cache.Init()
	pool.approxTime = time.Hour * 24

	validO3 := pool.passMan.makeValid(baseline)
	o3Time, o3MaxRSS, err := testharness.GetSetBaselineStats(sourceCmd, validO3, pool.manageCPULoad)
	if err != nil {
		return nil, err
	}

	pool.approxTime = o3Time
	pool.approxMem = o3MaxRSS
	fmt.Printf("\nO3 Time: %v\n", o3Time)
	pool.cache.Set(validO3, valueFromJobStats(0, o3Time, 0, 0))

	return pool, nil
}

// AddPassSequence appends a sequence to be queued for execution.
func (p *WorkerPool) AddPassSequence(seq []int) {
	p.passSeq = append(p.passSeq, seq)
}

// StartPool begins worker pool execution, returning the results.
//
// Handles pooled parallel execution.
func (p *WorkerPool) StartPool() []Result {
	var wg sync.WaitGroup
	p.Results = make(chan Result, len(p.passSeq))

	for _, seq := range p.passSeq {
		wg.Go(func() { p.runJob(seq, <-p.fileNum, false) })
	}

	p.manageCPULoad(&wg)

	wg.Wait()

	p.passSeq = make([][]int, 0, len(p.passSeq))
	close(p.Results)
	return util.ArrayFromChannel(p.Results)
}

// manageCPULoad handles dynamically adding workers to maintain CPU load towards the
// end of the pool's duration.
func (p *WorkerPool) manageCPULoad(wg *sync.WaitGroup) {
	for range cap(p.fileNum) - 1 {
		wg.Go(func() { p.runJob([]int{}, <-p.fileNum, true) })
	}
}

// runJob executes a single pass sequences optimisation pipeline to obtain its resulting score.
func (p *WorkerPool) runJob(seq []int, diff int, dummy bool) {
	if !dummy {
		defer PBar.Add(1)
	}
	defer func() { p.fileNum <- diff }()

	validSeq := p.passMan.makeValid(seq)

	if score, ok := p.cache.Get(validSeq); ok && !dummy {
		p.Results <- Result{
			score:   score,
			passSeq: seq,
		}
		return
	}

	compilation_t, execution_t, binary_size, maxrss, err := testharness.CompileAndExecute(validSeq, diff, false, p.SourceCmd, p.approxTime)
	finalScore := valueFromJobStats(compilation_t, time.Duration(execution_t), binary_size, maxrss)
	defer p.cache.Set(validSeq, finalScore)
	if err != nil {
		p.queueErrorResult(seq)
	} else if !dummy {
		p.Results <- Result{
			score:   finalScore,
			passSeq: seq,
		}
	}
}

// valueFromJobStats returns our heuristic value based upon the optimisation pipelines statistics.
func valueFromJobStats(compilation, execution time.Duration, size int, memory int64) float64 {
	return (COMPILATION_WEIGHT*float64(compilation.Microseconds()) +
		EXECUTION_WEIGHT*float64(execution.Microseconds()) +
		BINARY_SIZE_WEIGHT*float64(size) +
		MEMORY_USAGE_WEIGHT*float64(memory))
}

// queueErrorResult adds a result to the Result channel indicating sequence failure.
func (p *WorkerPool) queueErrorResult(seq []int) {
	p.Results <- Result{
		score:   util.MAX_FLOAT64,
		passSeq: seq,
	}
}
