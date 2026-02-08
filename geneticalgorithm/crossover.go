package geneticalgorithm

import (
	"math/rand"
	"slices"
)

// reproduce enacts crossover between two population members (pass sequences).
func reproduce(a, b []int, maxLen int) ([]int, []int) {
	if len(a) == 0 || len(b) == 0 { // avoid splitting at 0
		return a, b
	}

	newA := slices.Clone(a)
	newB := slices.Clone(b)

	newA, newB = twoPointCrossover(newA, newB)

	if len(newA) > maxLen {
		newA = slices.Delete(newA, maxLen, len(newA))
	}
	if len(newB) > maxLen {
		newB = slices.Delete(newB, maxLen, len(newB))
	}

	return newA, newB
}

// onePointCrossover enacts one point crossover between two population members.
func onePointCrossover(a, b []int) ([]int, []int) {
	var aPtr int
	var bPtr int

	aPtr = rand.Int() % (len(a))
	bPtr = rand.Int() % (len(b))

	aHead, aTail := slices.Clone(a[:aPtr]), slices.Clone(a[aPtr:])
	bHead, bTail := slices.Clone(b[:bPtr]), slices.Clone(b[bPtr:])

	newA := append(aHead, bTail...)
	newB := append(bHead, aTail...)

	return newA, newB
}

// twoPointCrossover enacts two point crossover between two population members.
func twoPointCrossover(a, b []int) ([]int, []int) {
	var aPtrStart int
	var aPtrEnd int
	var bPtrStart int
	var bPtrEnd int

	aPtrStart = rand.Int() % (len(a))
	aPtrEnd = rand.Int()%(len(a)-aPtrStart) + aPtrStart

	bPtrStart = rand.Int() % (len(b))
	bPtrEnd = rand.Int()%(len(b)-bPtrStart) + bPtrStart

	aHead := slices.Clone(a[:aPtrStart])
	aMid := slices.Clone(a[aPtrStart:aPtrEnd])
	aTail := slices.Clone(a[aPtrEnd:])

	bHead := slices.Clone(b[:bPtrStart])
	bMid := slices.Clone(b[bPtrStart:bPtrEnd])
	bTail := slices.Clone(b[bPtrEnd:])

	newA := append(append(aHead, bMid...), aTail...)
	newB := append(append(bHead, aMid...), bTail...)

	return newA, newB
}
