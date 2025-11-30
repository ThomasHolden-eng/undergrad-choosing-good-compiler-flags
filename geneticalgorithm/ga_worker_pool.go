package geneticalgorithm

import (
	"6CCS3PRJ/thomasholden/testharness"
	"6CCS3PRJ/thomasholden/util"
	"bytes"
	"fmt"

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
	passMan   *PassManager
}

const (
	COMPILATION_WEIGHT  float32 = 0.0001
	MEMORY_USAGE_WEIGHT float32 = 0.005
	BINARY_SIZE_WEIGHT  float32 = 0.0001
	EXECUTION_WEIGHT    float32 = 1 - COMPILATION_WEIGHT - MEMORY_USAGE_WEIGHT - BINARY_SIZE_WEIGHT
	EXECUTION_CYCLES    int     = 5
)

func NewWorkerPool(sourceCmd []string, workers, numJobs int, passManager *PassManager) *WorkerPool {
	pool := new(WorkerPool)

	pool.passSeq = make([][]int, 0, numJobs)
	pool.fileNum = make(chan int, workers)

	for i := 0; i < workers; i++ {
		pool.fileNum <- i
	}

	pool.SourceCmd = sourceCmd

	pool.passMan = passManager

	return pool
}

func (p *WorkerPool) AddPassSequence(seq []int) {
	p.passSeq = append(p.passSeq, seq)
}

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

func (p *WorkerPool) runJob(seq []int, diff int) {
	defer func() { p.fileNum <- diff }()
	defer PBar.Add(1)
	var scores []float32

	compilation_t, binary_size, newFile, err := testharness.CompileFileWithPasses(
		p.SourceCmd,
		p.makeValid(seq),
		diff, false,
	)
	if err != nil {
		fmt.Printf("\n%v\n", err)
		p.Results <- Result{
			score:   util.MAX_FLOAT32,
			passSeq: seq,
		}
		return
	}
	defer os.Remove(newFile.Name())

	var execution_min time.Duration
	var memory_min int64
	for begin := time.Now(); time.Since(begin) < 1*time.Second; {
		execution_min = util.MAX_DURATION
		memory_min = util.MAX_INT64
		for range EXECUTION_CYCLES {
			execution_t, memory_usage, err := testharness.RunFile(newFile)
			if err != nil {
				break
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
		score:   float32(*util.Median(scores)),
		passSeq: seq,
	}
}

func valueFromJobStats(compilation, execution time.Duration, size int, memory int64) float32 {
	return (COMPILATION_WEIGHT*float32(compilation.Microseconds()) +
		EXECUTION_WEIGHT*float32(execution.Microseconds()) +
		BINARY_SIZE_WEIGHT*float32(size) +
		MEMORY_USAGE_WEIGHT*float32(memory))
}

func (p *WorkerPool) makeValid(seq []int) string {
	const (
		INMODULE byte = iota
		INFUNC   byte = iota
		INCGSCC  byte = iota
		INLOOP   byte = iota
	)

	getClosing := func(scope byte) string {
		if scope != INLOOP {
			return ")"
		}
		return "))"
	}

	var buffer bytes.Buffer
	currentScope := INMODULE

	firstVal := seq[0]
	if p.passMan.cgsccRange.Within(firstVal) {
		buffer.WriteString("cgscc(")
		buffer.WriteString(p.passMan.IntToPass(firstVal))
	} else if p.passMan.functionRange.Within(firstVal) {
		buffer.WriteString("function(")
		buffer.WriteString(p.passMan.IntToPass(firstVal))
	} else if p.passMan.loopRange.Within(firstVal) {
		buffer.WriteString("function(loop(")
		buffer.WriteString(p.passMan.IntToPass(firstVal))
	} else {
		buffer.WriteString(p.passMan.IntToPass(firstVal))
	}

	var next string
	for _, val := range seq[1:] {
		if p.passMan.cgsccRange.Within(val) {
			next, currentScope = p.nextAppend(val, currentScope, INMODULE, INCGSCC, "cgscc(", getClosing)
			buffer.WriteString(next)
		} else if p.passMan.functionRange.Within(val) {
			next, currentScope = p.nextAppend(val, currentScope, INMODULE, INFUNC, "function(", getClosing)
			buffer.WriteString(next)
		} else if p.passMan.loopRange.Within(val) {
			next, currentScope = p.nextAppend(val, currentScope, INMODULE, INLOOP, "function(loop(", getClosing)
			buffer.WriteString(next)
		} else {
			next, currentScope = p.nextAppend(val, currentScope, INMODULE, INMODULE, "", getClosing)
			buffer.WriteString(next)
		}
	}

	if currentScope != INMODULE {
		buffer.WriteString(")")
	}

	return buffer.String()
}

func (p *WorkerPool) nextAppend(pass int, currentScope, defaultScope, desiredScope byte,
	prefix string, getClosing func(byte) string) (string, byte) {

	retString := ""

	switch currentScope {
	case desiredScope:
		retString += ","
		retString += p.passMan.IntToPass(pass)
	case defaultScope:
		retString += "," + prefix
		retString += p.passMan.IntToPass(pass)
		currentScope = desiredScope
	default:
		retString += getClosing(currentScope) + "," + prefix
		retString += p.passMan.IntToPass(pass)
		currentScope = desiredScope
	}

	return retString, currentScope
}
