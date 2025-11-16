package main

import (
	"6CCS3PRJ/thomasholden/testharness"
	"time"
)

type Result struct {
	score   float32
	passSeq []int
}

type WorkerPool struct {
	passSeq   chan []int
	fileNum   chan int
	Results   chan Result
	SourceCmd []string
}

const (
	COMPILATION_WEIGHT  float32 = 0.1
	MEMORY_USAGE_WEIGHT float32 = 0.1
	BINARY_SIZE_WEIGHT  float32 = 0.1
	EXECUTION_WEIGHT    float32 = 0.7
)

func NewWorkerPool(sourceCmd []string, workers, numJobs int) *WorkerPool {
	pool := new(WorkerPool)

	pool.passSeq = make(chan []int, numJobs)
	pool.fileNum = make(chan int, workers)
	pool.Results = make(chan Result, numJobs)

	for i := 0; i < workers; i++ {
		pool.fileNum <- i
	}

	pool.SourceCmd = sourceCmd

	return pool
}

func (p WorkerPool) AddPassSequence(seq []int) {
	p.passSeq <- seq
}

func (p WorkerPool) StartPool() {
	for {
		go p.runJob(<-p.passSeq, <-p.fileNum)
	}
}

func (p WorkerPool) runJob(seq []int, diff int) {
	defer func() { p.fileNum <- diff }()

	compilation_t, binary_size, newFile, _ := testharness.CompileFileWithPasses(p.SourceCmd, seq, diff)
	execution_t, memory_usage, _ := testharness.RunFile(newFile, diff)

	heuristic := valueFromJobStats(compilation_t, execution_t, binary_size, memory_usage)

	p.Results <- Result{
		score:   heuristic,
		passSeq: seq,
	}
}

func valueFromJobStats(compilation, execution time.Duration, size, memory int) float32 {
	return (COMPILATION_WEIGHT*float32(compilation) +
		EXECUTION_WEIGHT*float32(execution) +
		BINARY_SIZE_WEIGHT*float32(size) +
		MEMORY_USAGE_WEIGHT*float32(memory))
}
