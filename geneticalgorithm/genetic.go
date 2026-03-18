package geneticalgorithm

import (
	"6CCS3PRJ/thomasholden/passmanager"
	"6CCS3PRJ/thomasholden/testharness"
	"6CCS3PRJ/thomasholden/util"
	"6CCS3PRJ/thomasholden/workerpool"

	"errors"
	"fmt"
	"math/rand"
	"runtime"
	"slices"
	"time"

	"github.com/google/shlex" // Parsing logic should be exported as not part of the genetic algorithm
	// Progress bar should be moved else where as also independent of GA
)

// Population manages the genomes of the genetic algorithm.
type Population struct {
	members     [][]int                  // slices of int slices, where each int slice is a pass sequence.
	passMan     *passmanager.PassManager // pointer to a PassManager.
	maxLength   int                      // the maximum length of any genome at any given time.
	workers     *workerpool.WorkerPool   // pointer to a pool of parallelised workers.
	lastSafePop []int                    // copy of the last safe population member for comparison.
	convCount   int                      // counter to track steps without improvement.
}

const (
	SAFE_POP       float64 = 0.25 // Proportion of population to apply reproduction & selection.
	MUTATION_STEPS float64 = 1.2  // Number of times to run mutation (* population size).
)

// CreatePopulation initialized a Population struct and its variables.
func CreatePopulation(minPopSize, maxLength int, command string) (*Population, error) {
	workers := runtime.NumCPU() / 2
	runtime.GOMAXPROCS(workers)
	size := ((minPopSize / runtime.NumCPU()) + 1) * runtime.NumCPU()

	pop := &Population{}
	pop.members = make([][]int, size)
	pop.passMan = passmanager.GetPassManager()
	pop.maxLength = maxLength
	pop.lastSafePop = []int{}

	pop.members[0] = pop.passMan.GetO3Int()
	pop.members[0] = pop.members[0][:min(len(pop.members[0]), maxLength)]
	for i := range len(pop.members) - 1 {
		pop.members[i+1] = pop.generateRandomIntSlice()
	}

	commandSplit, err := shlex.Split(command)
	if err != nil {
		return nil, errors.New("command lexing failed: " + err.Error())
	}

	pop.workers, err = workerpool.NewWorkerPool(commandSplit, workers, size, pop.members[0])
	if err != nil {
		return nil, err
	}

	return pop, nil
}

// generateRandomIntSlice returns a slice of ints within the pass range of length [1-p.maxLength].
func (p *Population) generateRandomIntSlice() []int {
	returnSlice := p.passMan.GetO3Int()
	returnSlice = returnSlice[:min(len(returnSlice), p.maxLength)]
	if p.maxLength <= 0 {
		return returnSlice
	}

	for range rand.Intn(len(returnSlice)) + 1 {
		returnSlice = mutate(returnSlice, p.passMan.MaxPass, p.maxLength)
	}

	return slices.Clip(returnSlice)
}

// nextGeneration executes a single generation of the genetic algorithm.
//
// This is composed of execution -> elitism -> reproduction -> mutation.
func (p *Population) nextGeneration() bool {
	for _, member := range p.members {
		p.workers.AddPassSequence(member)
	}
	results := p.workers.StartPool()
	p.sortPop(results)
	for i, result := range results {
		if i < 10 {
			fmt.Printf("\t%v\n", result)
		}
	}
	p.members = make([][]int, 0, len(p.members))

	safeMarker := int(SAFE_POP * float64(len(results)))
	reproMarker := p.elitism(results, safeMarker)

	p.reproduction(reproMarker)

	p.mutation(safeMarker)

	return p.testConvergence()
}

// elitism selects the top k population members depending on results.
func (p *Population) elitism(results []workerpool.Result, marker int) int {
	reproMarker := -1
	for i := range marker {
		if results[i].Score < util.MAX_FLOAT64 {
			p.members = append(p.members, slices.Clone(results[i].PassSeq))
			reproMarker = i
		}
	}
	if reproMarker < 0 {
		panic("error in GA: all sequences failing")
	}
	return reproMarker
}

// reproduction fills the remaining population via reproduction among elites.
func (p *Population) reproduction(marker int) {
	for len(p.members) < cap(p.members) {
		passPtr1 := rand.Int() % (marker + 1)
		passPtr2 := rand.Int() % (marker + 1)

		newSeq1, newSeq2 := reproduce(
			p.members[passPtr1],
			p.members[passPtr2],
			p.passMan.MaxPass,
			p.maxLength,
		)
		p.members = append(p.members, newSeq1)
		if len(p.members) < cap(p.members) {
			p.members = append(p.members, newSeq2)
		}
	}
}

// mutation applies a random mutation to the reproduced population.
func (p *Population) mutation(marker int) {
	steps := int(float64(len(p.members)) * MUTATION_STEPS)
	for range steps {
		popPtr := rand.Intn(len(p.members)-marker-1) + marker + 1
		p.members[popPtr] = mutate(p.members[popPtr], p.passMan.MaxPass, p.maxLength)
	}
}

// RunNGenerations executes nextGeneration n times.
func (p *Population) RunNGenerations(n int) int {
	for i := range n {
		converged := p.nextGeneration()

		if converged && i > 10 {
			return i + 1
		}
	}
	return n
}

// RunGenerationsForDuration executes nextGeneration until the duration is exceeded.
func (p *Population) RunGenerationsForDuration(d time.Duration) int {
	startTime := time.Now()
	i := 0
	for ; time.Since(startTime) < d; i++ {
		converged := p.nextGeneration()

		if converged && i > 10 {
			return i + 1
		}
	}
	return i
}

// sortPop sorts the population depending on their achieved results.
func (p *Population) sortPop(results []workerpool.Result) {
	slices.SortFunc(results, func(a, b workerpool.Result) int {
		if a.Score < b.Score {
			return -1
		} else if a.Score > b.Score {
			return 1
		}
		return 0
	})
}

// testConvergence evaluates whether the population has converged yet.
//
// This is decided based on whether the top population member has improved over 10 generations.
func (p *Population) testConvergence() bool {
	safeMarker := int(SAFE_POP * float64(len(p.members)))
	safeMarker = max(safeMarker-1, 0)
	currentSafePop := p.members[safeMarker]

	if slices.Equal(p.lastSafePop, currentSafePop) {
		p.convCount += 1
	} else {
		p.convCount = 0
	}
	p.lastSafePop = append([]int{}, currentSafePop...)

	return p.convCount >= 5
}

// Best returns the best performing population member's passes and execution time.
func (p *Population) Best() (string, error) {
	validOptimal := p.passMan.MakeValid(p.members[0])
	_, _, _, _, err := testharness.CompileAndExecute(validOptimal, 0, true, p.workers.SourceCmd, p.workers.ApproxTime)
	return validOptimal, err
}
