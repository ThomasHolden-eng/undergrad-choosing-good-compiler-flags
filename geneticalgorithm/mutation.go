package geneticalgorithm

import (
	"math/rand"
	"slices"
)

// mutate applies a random mutation to a single population member.
func mutate(a []int, maxVal, maxLen int) {
	if len(a) == 0 {
		newPass := rand.Int() % maxVal
		a[0] = newPass
	}
	point := rand.Int() % len(a)

	randf := rand.Float64()
	if randf < 0.33 { // Replace pass
		newPass := rand.Int() % maxVal
		a[point] = newPass
	} else if randf > 0.66 && len(a) < maxLen { // Add pass
		newPass := rand.Int() % maxVal
		a = slices.Insert(a, point, newPass)

	} else { // Remove pass
		if point == 0 {
			point = 1
		}
		a = append(a[:point-1], a[point:]...)
	}
}
