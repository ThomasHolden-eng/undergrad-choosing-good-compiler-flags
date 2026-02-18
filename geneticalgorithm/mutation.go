package geneticalgorithm

import (
	"math/rand"
	"slices"
)

// mutate applies a random mutation to a single population member.
func mutate(a []int, maxVal, maxLen int) []int {
	if len(a) == 0 {
		newPass := rand.Int() % maxVal
		a = append(a, newPass)
		return a
	}
	point := rand.Int() % len(a)

	randf := rand.Float64()
	if randf < 0.33 { // Replace pass
		newPass := a[point]
		for newPass == a[point] {
			newPass = rand.Int() % maxVal
		}
		a[point] = newPass
	} else if randf > 0.66 && len(a) < maxLen { // Add pass
		newPass := rand.Int() % maxVal
		a = slices.Insert(a, point, newPass)

	} else { // Remove pass
		a = append(a[:point], a[point+1:]...)
	}

	return a
}
