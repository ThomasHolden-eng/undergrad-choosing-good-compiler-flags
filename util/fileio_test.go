package util

import (
	"os"
	"path/filepath"
	"testing"
)

// TestSaveSequence assesses the functionality of util.SaveSequence
func TestSaveSequence(t *testing.T) {
	tempDir := t.TempDir()
	validFile := filepath.Join(tempDir, "sequences.json")
	invalidFile := filepath.Join(tempDir, "nonexistent", "seq.json")

	tests := []struct {
		name     string
		filename string
		sequence []int
		wantErr  bool
	}{
		{
			name:     "Create new file and save successfully",
			filename: validFile,
			sequence: []int{1, 2, 3},
			wantErr:  false,
		},
		{
			name:     "Append to existing file successfully",
			filename: validFile,
			sequence: []int{4, 5, 6},
			wantErr:  false,
		},
		{
			name:     "Fail to save to an invalid directory path",
			filename: invalidFile,
			sequence: []int{7, 8},
			wantErr:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := SaveSequence(test.filename, test.sequence)
			if (err != nil) != test.wantErr {
				t.Errorf("SaveSequence() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

// TestLoadSequence assesses the functionality of util.LoadSequence
func TestLoadSequences(t *testing.T) {
	tempDir := t.TempDir()

	// Helper function to write mock data to temporary files
	writeMockFile := func(name, content string) string {
		path := filepath.Join(tempDir, name)
		_ = os.WriteFile(path, []byte(content), 0644)
		return path
	}

	validFile := writeMockFile("valid.json", "[1,2,3]\n[4,5,6]\n[7,8,9]\n")
	malformedFile := writeMockFile("malformed.json", "[1,2,3]\n[4,5,\n")
	blankLineFile := writeMockFile("blank.json", "[1,2]\n\n[3,4]")

	tests := []struct {
		name     string
		filename string
		n        int
		wantLen  int
		wantErr  bool
	}{
		{
			name:     "Load exact number of available sequences",
			filename: validFile,
			n:        3,
			wantLen:  3,
			wantErr:  false,
		},
		{
			name:     "Load fewer than available sequences",
			filename: validFile,
			n:        2,
			wantLen:  2,
			wantErr:  false,
		},
		{
			name:     "Request more sequences than are available",
			filename: validFile,
			n:        10,
			wantLen:  3,
			wantErr:  false,
		},
		{
			name:     "Return error when file does not exist",
			filename: filepath.Join(tempDir, "missing.json"),
			n:        2,
			wantLen:  0,
			wantErr:  true,
		},
		{
			name:     "Return error on malformed JSON data",
			filename: malformedFile,
			n:        2,
			wantLen:  0,
			wantErr:  true,
		},
		{
			name:     "Skip blank lines properly",
			filename: blankLineFile,
			n:        5,
			wantLen:  2,
			wantErr:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := LoadSequences(test.filename, test.n)

			if (err != nil) != test.wantErr {
				t.Errorf("Error loading sequence: %v", err)
			}

			if !test.wantErr && len(got) != test.wantLen {
				t.Errorf("LoadSequences returned %d sequences, wanted %d", len(got), test.wantLen)
			}
		})
	}
}
