package geneticalgorithm

import (
	"6CCS3PRJ/thomasholden/testharness"
	"6CCS3PRJ/thomasholden/util"
	"fmt"

	"os"
	"sync"
	"time"
)

// Result holds the results for a singular pass sequence's execution
type Result struct {
	score   float64
	passSeq []int
}

// WorkerPool provides functionality for pooling the execution of LLVM optimisation executions.
type WorkerPool struct {
	passSeq    [][]int
	fileNum    chan int
	Results    chan Result // a channel containing all results
	SourceCmd  []string    // the command from which the project should be built
	passMan    *PassManager
	approxTime time.Duration
	cache      map[string]float64
}

const (
	COMPILATION_WEIGHT  float64 = 0.0 // The proportion of the score to be made up of compile time.
	MEMORY_USAGE_WEIGHT float64 = 0.0 // The proportion of the score to be made up of execution's max RSS.
	BINARY_SIZE_WEIGHT  float64 = 0.0 // The proportion of the score to be made up of compiled binary size.

	// The proportion of the score to be made up of execution time.
	EXECUTION_WEIGHT float64 = 1 - COMPILATION_WEIGHT - MEMORY_USAGE_WEIGHT - BINARY_SIZE_WEIGHT
)

// NewWorkerPool returns a pointer to a worker pool.
func NewWorkerPool(sourceCmd []string, workers, numJobs int, passManager *PassManager, execTime time.Duration) *WorkerPool {
	pool := new(WorkerPool)

	pool.passSeq = make([][]int, 0, numJobs)
	pool.fileNum = make(chan int, workers)

	// Preps the channel that controls how many LLVM sessions can be active at once
	for i := 0; i < workers; i++ {
		pool.fileNum <- i
	}

	pool.SourceCmd = sourceCmd

	pool.passMan = passManager

	pool.approxTime = execTime
	fmt.Printf("\nO3 Time: %v\n", execTime)

	pool.cache = make(map[string]float64)

	return pool
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
		wg.Go(func() { p.runJob(seq, <-p.fileNum) })
	}

	wg.Wait()

	p.passSeq = make([][]int, 0, len(p.passSeq))
	close(p.Results)
	return util.ArrayFromChannel(p.Results)
}

// runJob executes a single pass sequences optimisation pipeline to obtain its resulting score.
func (p *WorkerPool) runJob(seq []int, diff int) {
	defer func() { p.fileNum <- diff }()
	defer PBar.Add(1)
	var scores []float64

	key := ""
	for pass := range seq {
		key += string(pass) + ","
	}
	if score, ok := p.cache[key]; ok {
		p.Results <- Result{
			score:   score,
			passSeq: seq,
		}
		return
	}

	compilation_t, binary_size, newFile, err := testharness.CompileFileWithPasses(
		p.SourceCmd,
		p.passMan.makeValid(seq),
		diff, false,
	)
	if err != nil {
		// fmt.Printf("\n%v\n", err)
		p.Results <- Result{
			score:   util.MAX_FLOAT64,
			passSeq: seq,
		}
		return
	}
	defer os.Remove(newFile.Name())

	for begin := time.Now(); (time.Since(begin) < 5*p.approxTime ||
		!MedianConfidence(scores, 0.05)) &&
		time.Since(begin) < 12*p.approxTime; {

		execution_t, memory_usage, err := testharness.RunFile(newFile)
		if err != nil {
			p.Results <- Result{
				score:   util.MAX_FLOAT64,
				passSeq: seq,
			}
			return
		}

		heuristic := valueFromJobStats(compilation_t, execution_t, binary_size, memory_usage)

		scores = append(scores, heuristic)
	}
	newFile.Close()

	p.cache[key] = util.Median(scores)
	p.Results <- Result{
		score:   util.Median(scores),
		passSeq: seq,
	}
}

// valueFromJobStats returns our heuristic value based upon the optimisation pipelines statistics.
func valueFromJobStats(compilation, execution time.Duration, size int, memory int64) float64 {
	return (COMPILATION_WEIGHT*float64(compilation.Microseconds()) +
		EXECUTION_WEIGHT*float64(execution.Microseconds()) +
		BINARY_SIZE_WEIGHT*float64(size) +
		MEMORY_USAGE_WEIGHT*float64(memory))
}
