package geneticalgorithm

import (
	"6CCS3PRJ/thomasholden/testharness"
	"6CCS3PRJ/thomasholden/util"
	"errors"
	"math/rand"
	"runtime"
	"slices"
	"time"

	"github.com/google/shlex"           // Parsing logic should be exported as not part of the genetic algorithm
	"github.com/schollz/progressbar/v3" // Progress bar should be moved else where as also independent of GA
)

type Population struct {
	members   [][]int
	passMan   *PassManager
	maxLength int
	workers   *WorkerPool
	bestScore float64
}

const (
	SAFE_POP   float64 = 0.2  // Proportion of population to apply reproduction & selection
	REPRO_POP  float64 = 0.1  // Proportion of population to apply ONLY reproduction
	RANDOM_POP float64 = 0.05 // Proportion of population to be randomly generated every generation
) // The rest of the population will be generated via reproduction from among SAFE_POP & REPRO_POP

var PBar *progressbar.ProgressBar

func CreatePopulation(minPopSize, maxLength int, command string) (*Population, error) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	size := ((minPopSize / runtime.NumCPU()) + 1) * runtime.NumCPU()

	pop := &Population{}
	pop.members = make([][]int, size)
	pop.passMan = NewPassManager()
	pop.maxLength = maxLength
	pop.bestScore = float64(util.MAX_FLOAT32)

	pop.members[0] = pop.passMan.GetO3Int()
	for i := range len(pop.members) - 1 {
		pop.members[i+1] = pop.generateRandomIntSlice()
	}

	commandSplit, err := shlex.Split(command)
	if err != nil {
		return nil, errors.New("command lexing failed")
	}

	o3Time, err := testharness.GetSequenceExecutionTime(commandSplit, pop.passMan.makeValid(pop.members[0]), false)
	if err != nil {
		return nil, errors.New("failed basic execution, is the command formatted correctly?:\n" + err.Error())
	}

	pop.workers = NewWorkerPool(commandSplit, runtime.NumCPU(), size, pop.passMan, o3Time)

	return pop, nil
}

func (p *Population) generateRandomIntSlice() []int {
	returnSlice := make([]int, 0, p.maxLength)

	for i := 1.0; len(returnSlice) < p.maxLength && i > 1.0/float64(p.maxLength); i = rand.Float64() {
		returnSlice = append(returnSlice, rand.Int()%(p.passMan.MaxPass))
	}

	return slices.Clip(returnSlice)
}

func (p *Population) nextGeneration() {
	for _, member := range p.members {
		p.workers.AddPassSequence(member)
	}
	results := p.workers.StartPool()
	p.sortPop(results)

	p.bestScore = results[0].score

	p.members = make([][]int, 0, len(p.members))

	safeMarker := int(SAFE_POP * float64(len(results)))
	reproMarker := int((SAFE_POP + REPRO_POP) * float64(len(results)))
	for i := range safeMarker {
		p.members = append(p.members, slices.Clone(results[i].passSeq))
	}

	for len(p.members) < cap(p.members) {
		if rand.Float64() < RANDOM_POP {
			p.members = append(p.members, p.generateRandomIntSlice())
			continue
		}
		passPtr1 := rand.Int() % reproMarker
		passPtr2 := rand.Int() % reproMarker

		newSeq1, newSeq2 := p.reproduce(results[passPtr1].passSeq, results[passPtr2].passSeq)
		p.members = append(p.members, newSeq1)
		if len(p.members) < cap(p.members) {
			p.members = append(p.members, newSeq2)
		}
	}
}

func (p *Population) RunNGenerations(n int) {
	PBar = progressbar.Default(int64(n) * int64(cap(p.members)))
	for range n {
		p.nextGeneration()
	}
}

func (p *Population) reproduce(a, b []int) ([]int, []int) {
	newA := slices.Clone(a)
	newB := slices.Clone(b)

	randF := rand.Float64()
	if randF <= 0.6 {
		newA, newB = p.randomCrossover(newA, newB)
	} else {
		p.mutate(newA)
		p.mutate(newB)
	}

	if len(newA) > p.maxLength {
		newA = slices.Delete(newA, p.maxLength, len(newA))
	}
	if len(newB) > p.maxLength {
		newB = slices.Delete(newB, p.maxLength, len(newB))
	}

	return newA, newB
}

func (p *Population) mutate(a []int) {
	if len(a) == 0 {
		newPass := rand.Int() % p.passMan.MaxPass
		a[0] = newPass
	}
	point := rand.Int() % len(a)

	randf := rand.Float64()
	if randf < 0.33 {
		newPass := rand.Int() % p.passMan.MaxPass
		a[point] = newPass
	} else if randf > 0.66 {
		if point == 0 {
			point = 1
		}
		a = append(a[:point-1], a[point:]...)
	} else if len(a) < p.maxLength {
		newPass := rand.Int() % p.passMan.MaxPass
		a = slices.Insert(a, point, newPass)
	} else {
		if point == 0 {
			point = 1
		}
		a = append(a[:point-1], a[point:]...)
	}

}

func (p *Population) randomCrossover(a, b []int) ([]int, []int) {
	var aPtr int
	var bPtr int
	if len(a) == 0 || len(b) == 0 { // avoid splitting at 0
		p.mutate(a)
		p.mutate(b)
	}

	aPtr = rand.Int() % (len(a))
	bPtr = rand.Int() % (len(b))

	aHead, aTail := slices.Clone(a[:aPtr]), slices.Clone(a[aPtr:])
	bHead, bTail := slices.Clone(b[:bPtr]), slices.Clone(b[bPtr:])

	newA := append(aHead, bTail...)
	newB := append(bHead, aTail...)

	return newA, newB
}

// https://doi.org/10.1109/tevc.2006.878096 TODO
func (p *Population) svlc(a, b []int) ([]int, []int) {
	aPtr := rand.Int() % len(a)
	bPtr := rand.Int() % len(b)

	aHead, aTail := slices.Clone(a[:aPtr]), slices.Clone(a[aPtr:])
	bHead, bTail := slices.Clone(b[:bPtr]), slices.Clone(b[bPtr:])

	newA := append(aHead, bTail...)
	newB := append(bHead, aTail...)

	return newA, newB
}

func (p *Population) sortPop(results []Result) {
	slices.SortFunc(results, func(a, b Result) int {
		if a.score < b.score {
			return -1
		} else if a.score > b.score {
			return 1
		}
		return 0
	})
}

func (p *Population) Best() (string, time.Duration, error) {
	execTime, err := testharness.GetSequenceExecutionTime(p.workers.SourceCmd, p.passMan.makeValid(p.members[0]), true)
	if err != nil {
		return "", 0, errors.New("failed assessment of final product\n" + err.Error())
	}

	return p.passMan.makeValid(p.members[0]), execTime, nil
}
