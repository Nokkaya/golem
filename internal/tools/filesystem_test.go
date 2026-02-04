package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadFile(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")
	content := "line1\nline2\nline3\nline4\nline5"
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Test cases
	tests := []struct {
		name      string
		input     *ReadFileInput
		wantLines []string
		wantTotal int
	}{
		{
			name:      "read all",
			input:     &ReadFileInput{Path: filePath},
			wantLines: []string{"line1", "line2", "line3", "line4", "line5"},
			wantTotal: 5,
		},
		{
			name:      "offset 1",
			input:     &ReadFileInput{Path: filePath, Offset: 1},
			wantLines: []string{"line2", "line3", "line4", "line5"},
			wantTotal: 5,
		},
		{
			name:      "limit 2",
			input:     &ReadFileInput{Path: filePath, Limit: 2},
			wantLines: []string{"line1", "line2"},
			wantTotal: 5,
		},
		{
			name:      "offset 1 limit 2",
			input:     &ReadFileInput{Path: filePath, Offset: 1, Limit: 2},
			wantLines: []string{"line2", "line3"},
			wantTotal: 5,
		},
		{
			name:      "offset out of bounds",
			input:     &ReadFileInput{Path: filePath, Offset: 10},
			wantLines: []string{},
			wantTotal: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := readFile(context.Background(), tt.input)
			if err != nil {
				t.Fatalf("readFile failed: %v", err)
			}
			if out.TotalLines != tt.wantTotal {
				t.Errorf("got TotalLines %d, want %d", out.TotalLines, tt.wantTotal)
			}

			var gotLines []string
			if out.Content == "" {
				gotLines = []string{}
			} else {
				gotLines = strings.Split(out.Content, "\n")
			}

			// Handle empty content resulting in []string{""} from Split
			if len(gotLines) == 1 && gotLines[0] == "" && len(tt.wantLines) == 0 {
				gotLines = []string{}
			}

			if len(gotLines) != len(tt.wantLines) {
				t.Errorf("got %d lines, want %d. Content: %q", len(gotLines), len(tt.wantLines), out.Content)
				return
			}
			for i, line := range gotLines {
				if line != tt.wantLines[i] {
					t.Errorf("line %d: got %q, want %q", i, line, tt.wantLines[i])
				}
			}
		})
	}
}

func TestReadFile_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "empty.txt")
	err := os.WriteFile(filePath, []byte(""), 0644)
	if err != nil {
		t.Fatal(err)
	}

	out, err := readFile(context.Background(), &ReadFileInput{Path: filePath})
	if err != nil {
		t.Fatal(err)
	}

	// Check current behavior for empty file
	t.Logf("Empty file TotalLines: %d", out.TotalLines)
	t.Logf("Empty file Content: %q", out.Content)
}

func TestReadFile_TrailingNewline(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "trailing.txt")
	content := "line1\n"
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	out, err := readFile(context.Background(), &ReadFileInput{Path: filePath})
	if err != nil {
		t.Fatal(err)
	}

	// Check current behavior for trailing newline
	t.Logf("Trailing newline TotalLines: %d", out.TotalLines)
	t.Logf("Trailing newline Content: %q", out.Content)
}

func BenchmarkReadFile_LargeFile_Partial(b *testing.B) {
	tmpDir := b.TempDir()
	filePath := filepath.Join(tmpDir, "bench.txt")
	f, err := os.Create(filePath)
	if err != nil {
		b.Fatal(err)
	}
	// Write 100,000 lines
	line := "this is a test line for benchmarking\n"
	for i := 0; i < 100000; i++ {
		f.WriteString(line)
	}
	f.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := readFile(context.Background(), &ReadFileInput{
			Path:   filePath,
			Offset: 50000,
			Limit:  10,
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}
