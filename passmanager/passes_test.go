package passmanager

import (
	"math/rand/v2"
	"slices"
	"testing"
)

// basicMutate applies a random mutation to a slice for testing.
func basicMutate(a []int, maxVal, maxLen int) []int {
	point := rand.Int() % len(a)

	newPass := a[point]
	for newPass == a[point] {
		newPass = rand.Int() % maxVal
	}
	a[point] = newPass

	return a
}

func TestGetPassManager(t *testing.T) {
	if GetPassManager() == nil {
		t.Fatal("GetPassManager returned nil")
	}
	passMan := GetPassManager()
	if passMan != GetPassManager() {
		t.Errorf("GetPassManager returned different instances")
	}
	if passMan.MaxPass <= 0 {
		t.Errorf("Expected MaxPass > 0, got %v", passMan.MaxPass)
	}
	if passMan.cgsccRange == nil ||
		passMan.moduleRange == nil ||
		passMan.functionRange == nil ||
		passMan.loopRange == nil {
		t.Errorf("Expected non-nil ranges, got nil")
	}
}

func TestLoadPasses(t *testing.T) {
	passMan := &PassManager{}
	passMan.loadPasses()

	for i := range passMan.MaxPass {
		passString, ok := passMan.GetInverse(i)
		if !ok {
			t.Errorf("Expected pass %d to be loaded", i)
		} else if n, ok := passMan.Get(passString); !ok || n != i {
			t.Errorf("Expected pass value for %s to be loaded", passString)
		}
	}
}

func TestInitRanges(t *testing.T) {
	passMan := &PassManager{}
	passMan.loadPasses()
	passMan.initRanges()

	if passMan.moduleRange == nil ||
		passMan.cgsccRange == nil ||
		passMan.functionRange == nil ||
		passMan.loopRange == nil {
		t.Errorf("Expected non-nil ranges, got nil")
	}
	o3String := getO3String()
	for i := range passMan.MaxPass {
		s, ok := passMan.GetInverse(i)
		if !passMan.moduleRange.Within(i) &&
			!passMan.cgsccRange.Within(i) &&
			!passMan.functionRange.Within(i) &&
			!passMan.loopRange.Within(i) &&
			ok && !slices.Contains(o3String, s) { // O3 module passes are excluded
			t.Errorf("Expected ranges to cover pass %d", i)
		}
	}
}

func TestGetPasses(t *testing.T) {
	passes := getPasses()
	if len(passes) == 0 {
		t.Errorf("Expected at least one pass, got none")
	}
}

func TestGetO3String(t *testing.T) {
	o3String := getO3String()
	if len(o3String) == 0 {
		t.Errorf("Expected at least one O3 pass, got none")
	}
}

func TestGetO3Int(t *testing.T) {
	passMan := GetPassManager()
	o3Int := passMan.GetO3Int()
	if len(o3Int) == 0 {
		t.Errorf("Expected at least one O3 pass, got none")
	}
}

func TestPassToInt(t *testing.T) {
	passMan := GetPassManager()
	allPasses := append(getPasses(), getO3String()...)
	if len(allPasses) == 0 {
		t.Errorf("Expected at least one pass, got none")
	}

	for _, pass := range allPasses {
		if passMan.PassToInt(pass) == 0 && pass != allPasses[0] {
			t.Errorf("Expected pass %s to have an int representation", pass)
		}
	}
}

func TestIntToPass(t *testing.T) {
	passMan := GetPassManager()

	for i := range passMan.MaxPass {
		passString := passMan.IntToPass(i)
		if passString == "" {
			t.Errorf("Expected pass %d to be loaded", i)
		}
	}
}

func TestMakeValid(t *testing.T) {
	passMan := GetPassManager()

	mutatedO3function := passMan.GetO3Int()
	for range 100 {
		basicMutate(mutatedO3function, passMan.MaxPass, 100)
	}
	for i := range passMan.MaxPass {
		if passMan.functionRange.Within(i) {
			mutatedO3function[0] = i
		}
	}

	mutatedO3loop := passMan.GetO3Int()
	for range 100 {
		basicMutate(mutatedO3loop, passMan.MaxPass, 100)
	}
	for i := range passMan.MaxPass {
		if passMan.loopRange.Within(i) {
			mutatedO3loop[0] = i
		}
	}

	tests := []struct {
		name string
		seq  []int
	}{
		{
			name: "O3 sequence",
			seq:  passMan.GetO3Int(),
		},
		{
			name: "empty sequence",
			seq:  []int{},
		},
		{
			name: "mutated sequence starting with function pass",
			seq:  mutatedO3function,
		},
		{
			name: "mutated sequence starting with loop pass",
			seq:  mutatedO3loop,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			passString := passMan.MakeValid(test.seq)
			if passString == "" && len(test.seq) > 0 {
				t.Fatalf("Expected a valid string, got none")
			} else if passString != "" && len(test.seq) == 0 {
				t.Fatalf("Expected an empty string, got %s", passString)
			}

			wellFormedCount := 0
			for _, c := range passString {
				switch c {
				case '(', '[', '{', '<':
					wellFormedCount++
				case ')', ']', '}', '>':
					wellFormedCount--
				default:
					continue
				}
			}
			if wellFormedCount != 0 {
				t.Fatalf("Expected string to be well-formed, got %s", passString)
			}
		})
	}
}
