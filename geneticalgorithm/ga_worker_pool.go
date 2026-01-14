package geneticalgorithm

import (
	"6CCS3PRJ/thomasholden/testharness"
	"6CCS3PRJ/thomasholden/util"
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
	passSeq    [][]int
	fileNum    chan int
	Results    chan Result
	SourceCmd  []string
	passMan    *PassManager
	approxTime time.Duration
}

const (
	COMPILATION_WEIGHT  float32 = 0.0
	MEMORY_USAGE_WEIGHT float32 = 0.0
	BINARY_SIZE_WEIGHT  float32 = 0.0
	EXECUTION_WEIGHT    float32 = 1 - COMPILATION_WEIGHT - MEMORY_USAGE_WEIGHT - BINARY_SIZE_WEIGHT
	EXECUTION_CYCLES    int     = 1
)

func NewWorkerPool(sourceCmd []string, workers, numJobs int, passManager *PassManager, execTime time.Duration) *WorkerPool {
	pool := new(WorkerPool)

	pool.passSeq = make([][]int, 0, numJobs)
	pool.fileNum = make(chan int, workers)

	for i := 0; i < workers; i++ {
		pool.fileNum <- i
	}

	pool.SourceCmd = sourceCmd

	pool.passMan = passManager

	pool.approxTime = execTime
	fmt.Printf("\nO3 Time: %v\n", execTime)

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
		p.passMan.makeValid(seq),
		diff, false,
	)
	if err != nil {
		// fmt.Printf("\n%v\n", err)
		p.Results <- Result{
			score:   util.MAX_FLOAT32,
			passSeq: seq,
		}
		return
	}
	defer os.Remove(newFile.Name())

	var execution_min time.Duration
	var memory_min int64
	for begin := time.Now(); time.Since(begin) < 3*p.approxTime; {
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
