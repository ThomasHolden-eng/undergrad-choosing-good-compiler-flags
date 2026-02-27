package geneticalgorithm

import (
	"slices"
	"testing"
)

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
		passString := passMan.GetInverse(i)
		if passString == nil {
			t.Errorf("Expected pass %d to be loaded", i)
		} else if passMan.Get(*passString) == nil {
			t.Errorf("Expected pass value for %s to be loaded", *passString)
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
		if !passMan.moduleRange.Within(i) &&
			!passMan.cgsccRange.Within(i) &&
			!passMan.functionRange.Within(i) &&
			!passMan.loopRange.Within(i) &&
			!slices.Contains(o3String, *passMan.GetInverse(i)) { // O3 module passes are excluded
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
	o3Int := passMan.GetO3Int()
	o3String := passMan.makeValid(o3Int)
	if o3String == "" {
		t.Fatalf("Expected a valid O3 string, got none")
	}

	wellFormedCount := 0
	for _, c := range o3String {
		switch c {
		case '(', '[', '{':
			wellFormedCount++
		case ')', ']', '}':
			wellFormedCount--
		default:
			continue
		}
	}
	if wellFormedCount != 0 {
		t.Fatalf("Expected O3 string to be well-formed, got %s", o3String)
	}
}
