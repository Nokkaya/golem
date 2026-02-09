package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkReadFileTool(b *testing.B) {
	// Setup
	tmpDir := b.TempDir()
	fileName := "large_file.txt"
	filePath := filepath.Join(tmpDir, fileName)

	// Create a large file (e.g., ~10MB) with random content and newlines
	f, err := os.Create(filePath)
	if err != nil {
		b.Fatal(err)
	}

	// specific seed for reproducibility
	r := rand.New(rand.NewSource(42))

	content := make([]byte, 1024)
	for i := 0; i < 10000; i++ {
		r.Read(content)
		// Inject some newlines roughly every 100 bytes
		for j := 50; j < len(content); j += 100 {
			content[j] = '\n'
		}
		f.Write(content)
	}
	f.Close()

	tool, err := NewReadFileTool(tmpDir)
	if err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()
	// Scan the whole file by setting a large offset.
    // This forces the tool to read everything to count lines and find the offset,
    // but returns empty content, avoiding JSON serialization overhead.
	input := fmt.Sprintf(`{"path": "%s", "offset": 10000000}`, fileName)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		outStr, err := tool.InvokableRun(ctx, input)
		if err != nil {
			b.Fatal(err)
		}

		if i == 0 {
			var out ReadFileOutput
			if err := json.Unmarshal([]byte(outStr), &out); err != nil {
				b.Fatal(err)
			}
            // We expect it to have counted lines
			if out.TotalLines == 0 {
				b.Fatal("Expected > 0 lines")
			}
            // And empty content
            if len(out.Content) > 0 {
                b.Fatal("Expected empty content")
            }
		}
	}
}
