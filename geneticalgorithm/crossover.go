package geneticalgorithm

import (
	"math/rand"
	"slices"
)

// reproduce enacts crossover between two population members (pass sequences).
func reproduce(a, b []int, maxVal, maxLen int) ([]int, []int) {
	var newA, newB []int
	randNum := rand.Int() % 5

	if len(a) == 0 || len(b) == 0 { // avoid splitting at 0
		newA, newB = pairMutate(a, b, maxVal, maxLen)
	} else if randNum < 3 && !slices.Equal(a, b) {
		newA, newB = lcsCrossover(a, b)
	} else if randNum == 3 || slices.Equal(a, b) {
		newA, newB = twoPointCrossover(a, b)
	} else {
		newA, newB = pairMutate(a, b, maxVal, maxLen)
	}

	if len(newA) > maxLen {
		newA = slices.Delete(newA, maxLen, len(newA))
	}
	if len(newB) > maxLen {
		newB = slices.Delete(newB, maxLen, len(newB))
	}

	return newA, newB
}

// pairMutate returns mutated versions of the input slices.
func pairMutate(a, b []int, maxVal, maxLen int) ([]int, []int) {
	newA := slices.Clone(a)
	newB := slices.Clone(b)
	newA = mutate(newA, maxVal, maxLen)
	newB = mutate(newB, maxVal, maxLen)

	return newA, newB
}

// onePointCrossover enacts one point crossover between two population members.
func onePointCrossover(a, b []int) ([]int, []int) {
	if len(a) == 0 || len(b) == 0 {
		return a, b
	}

	var aPtr int
	var bPtr int
	var newA, newB []int

	aPtr = rand.Int()%(len(a)-2) + 1
	bPtr = rand.Int()%(len(b)-2) + 1

	newA = append(newA, a[:aPtr]...)
	newA = append(newA, b[bPtr:]...)

	newB = append(newB, b[:bPtr]...)
	newB = append(newB, a[aPtr:]...)

	return newA, newB
}

// twoPointCrossover enacts two point crossover between two population members.
func twoPointCrossover(a, b []int) ([]int, []int) {
	if len(a) == 0 || len(b) == 0 {
		return a, b
	}

	var aPtrStart int
	var aPtrEnd int
	var bPtrStart int
	var bPtrEnd int
	var newA, newB []int

	aPtrStart = rand.Int()%(len(a)-2) + 1
	aPtrEnd = rand.Int()%(len(a)-aPtrStart-1) + 1 + aPtrStart

	bPtrStart = rand.Int()%(len(b)-2) + 1
	bPtrEnd = rand.Int()%(len(b)-bPtrStart-1) + 1 + bPtrStart

	newA = append(newA, a[:aPtrStart]...)
	newA = append(newA, b[bPtrStart:bPtrEnd]...)
	newA = append(newA, a[aPtrEnd:]...)

	newB = append(newB, b[:bPtrStart]...)
	newB = append(newB, a[aPtrStart:aPtrEnd]...)
	newB = append(newB, b[bPtrEnd:]...)

	return newA, newB
}

// lcsCrossover takes the longest common subsequence of two population members.
// Following this, it swaps the non-matching subsequences randomly to generate
// two new members with the common subsequence preserved.
//
// https://en.wikipedia.org/wiki/Longest_common_subsequence
func lcsCrossover(a, b []int) ([]int, []int) {
	if len(a) == 0 || len(b) == 0 {
		return a, b
	}

	matchedIdxsA, matchedIdxsB := lcsPairs(a, b)
	if len(matchedIdxsA) == 0 || len(matchedIdxsB) == 0 {
		return twoPointCrossover(a, b)
	}
	var newA, newB []int

	var prevA, prevB, currA, currB int
	// One extra loop to handle the tail without code duplication
	for k := range len(matchedIdxsA) + 1 {
		currA, currB = len(a), len(b)
		if k < len(matchedIdxsA) {
			currA, currB = matchedIdxsA[k], matchedIdxsB[k]
		}

		segA, segB := a[prevA:currA], b[prevB:currB]

		if rand.Int()%2 != 0 {
			segA, segB = segB, segA // Swaps slice headers
		}
		newA = append(newA, segA...)
		newB = append(newB, segB...)

		// Append the match itself
		if k < len(matchedIdxsA) {
			val := a[currA]
			newA = append(newA, val)
			newB = append(newB, val)
			prevA, prevB = currA+1, currB+1
		}
	}

	return newA, newB
}

// lcsPairs returns two slices listing the LCS idxs of the first sequence and
// the matching idxs of the second sequence respectively.
func lcsPairs[T comparable](seqA, seqB []T) ([]int, []int) {
	n, m := len(seqA), len(seqB)
	dpTable := make([][]int, n+1)
	for i := range dpTable {
		dpTable[i] = make([]int, m+1)
	}

	// Constructs the dp table of LCS lengths for every prefix of the seqs
	for i := 1; i <= n; i++ {
		for j := 1; j <= m; j++ {
			if seqA[i-1] == seqB[j-1] {
				dpTable[i][j] = dpTable[i-1][j-1] + 1
			} else {
				dpTable[i][j] = max(dpTable[i-1][j], dpTable[i][j-1])
			}
		}
	}

	var matchedIdxsA []int
	var matchedIdxsB []int

	// Backtracks the dp table to find the matches of the longest lengths
	for n > 0 && m > 0 {
		if seqA[n-1] == seqB[m-1] {
			matchedIdxsA = append(matchedIdxsA, n-1)
			matchedIdxsB = append(matchedIdxsB, m-1)
			n--
			m--
		} else if dpTable[n-1][m] > dpTable[n][m-1] {
			n--
		} else {
			m--
		}
	}

	slices.Reverse(matchedIdxsA)
	slices.Reverse(matchedIdxsB)

	return matchedIdxsA, matchedIdxsB
}
