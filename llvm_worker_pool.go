package main

import (
	"6CCS3PRJ/thomasholden/testharness"

	"os"
	"sync"
	"time"
)

type Result struct {
	score   float32
	passSeq []int
}

type WorkerPool struct {
	passSeq   [][]int
	fileNum   chan int
	Results   chan Result
	SourceCmd []string
}

const (
	COMPILATION_WEIGHT  float32       = 0.0001
	MEMORY_USAGE_WEIGHT float32       = 0.005
	BINARY_SIZE_WEIGHT  float32       = 0.0001
	EXECUTION_WEIGHT    float32       = 1 - COMPILATION_WEIGHT - MEMORY_USAGE_WEIGHT - BINARY_SIZE_WEIGHT
	EXECUTION_CYCLES    int           = 25
	MAX_DURATION        time.Duration = time.Hour * 292
	MAX_INT64           int64         = 9223372036854775807
)

func NewWorkerPool(sourceCmd []string, workers, numJobs int) *WorkerPool {
	pool := new(WorkerPool)

	pool.passSeq = make([][]int, 0, numJobs)
	pool.fileNum = make(chan int, workers)
	pool.Results = make(chan Result, numJobs)

	for i := 0; i < workers; i++ {
		pool.fileNum <- i
	}

	pool.SourceCmd = sourceCmd

	return pool
}

func (p *WorkerPool) AddPassSequence(seq []int) {
	p.passSeq = append(p.passSeq, seq)
}

func (p *WorkerPool) StartPool() []Result {
	var wg sync.WaitGroup

	for _, seq := range p.passSeq {
		wg.Go(func() { p.runJob(seq, <-p.fileNum) })
	}

	wg.Wait()
	close(p.Results)
	return testharness.ArrayFromChannel(p.Results)
}

func (p *WorkerPool) runJob(seq []int, diff int) {
	defer func() { p.fileNum <- diff }()
	var scores []float32

	compilation_t, binary_size, newFile, err := testharness.CompileFileWithPasses(p.SourceCmd, seq, diff)
	if err != nil {
		panic(err)
	}
	defer os.Remove(newFile.Name())

	var execution_min time.Duration
	var memory_min int64
	for begin := time.Now(); time.Since(begin) < 15*time.Second; {
		execution_min = MAX_DURATION
		memory_min = MAX_INT64
		for range EXECUTION_CYCLES {
			execution_t, memory_usage, err := testharness.RunFile(newFile)
			if err != nil {
				panic(err)
			}
			if execution_t < execution_min && execution_t != 0 {
				execution_min = execution_t
			}
			if memory_usage < memory_min && memory_usage != 0 {
				memory_min = memory_usage
			}
		}

		heuristic := valueFromJobStats(compilation_t, execution_min, binary_size, memory_min)

		scores = append(scores, heuristic)
	}
	newFile.Close()

	p.Results <- Result{
		score:   *testharness.Median(scores),
		passSeq: seq,
	}
}

func valueFromJobStats(compilation, execution time.Duration, size int, memory int64) float32 {
	return (COMPILATION_WEIGHT*float32(compilation.Microseconds()) +
		EXECUTION_WEIGHT*float32(execution.Microseconds()) +
		BINARY_SIZE_WEIGHT*float32(size) +
		MEMORY_USAGE_WEIGHT*float32(memory))
}
