package main

import "time"

type Result struct {
	score   float32
	passSeq []byte
}

type WorkerPool struct {
	passSeq    chan []byte
	fileNum    chan int
	Results    chan Result
	SourceFile string
}

const COMPILATION_WEIGHT float32 = 0.1
const EXECUTION_WEIGHT float32 = 0.9

func NewWorkerPool(file string, workers, numJobs int) *WorkerPool {
	pool := new(WorkerPool)

	pool.passSeq = make(chan []byte, numJobs)
	pool.fileNum = make(chan int, workers)
	pool.Results = make(chan Result, numJobs)

	for i := 0; i < workers; i++ {
		pool.fileNum <- i
	}

	pool.SourceFile = file

	return pool
}

func (p WorkerPool) AddPassSequence(seq []byte) {
	p.passSeq <- seq
}

func (p WorkerPool) StartPool() {
	for {
		go p.runJob(<-p.passSeq, <-p.fileNum)
	}
}

func (p WorkerPool) runJob(seq []byte, diff int) {
	compilation_t := compileSource(seq, p.SourceFile, diff)
	execution_t := executeSource(p.SourceFile, diff)

	heuristic := valueFromJobTime(compilation_t, execution_t)

	p.Results <- Result{
		score:   heuristic,
		passSeq: seq,
	}
	p.fileNum <- diff
}

func valueFromJobTime(compilation, execution time.Duration) float32 {
	return COMPILATION_WEIGHT*float32(compilation) + EXECUTION_WEIGHT*float32(execution)
}

func compileSource(seq []byte, filename string, diff int) time.Duration {
	return time.Second
}

func executeSource(filename string, diff int) time.Duration {
	return time.Second
}
