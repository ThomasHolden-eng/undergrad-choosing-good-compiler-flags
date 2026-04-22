package util

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"os"
)

// SaveSequence saves a passed sequence to the file specified, as a new line.
func SaveSequence(filename string, sequence []int) error {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(sequence); err != nil {
		return err
	}

	return nil
}

// LoadSequences returns up to n unique, random sequences from a specified file.
func LoadSequences(filename string, n int) ([][]int, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	lines := bytes.Split(bytes.TrimSpace(data), []byte("\n"))

	rand.Shuffle(len(lines), func(i, j int) { lines[i], lines[j] = lines[j], lines[i] })

	count := n
	if len(lines) < n {
		count = len(lines)
	}

	var results [][]int
	for i := 0; i < count; i++ {
		if len(lines[i]) == 0 { // blank line
			continue
		}

		var seq []int
		if err := json.Unmarshal(lines[i], &seq); err != nil {
			return nil, err
		}
		results = append(results, seq)
	}

	return results, nil
}
