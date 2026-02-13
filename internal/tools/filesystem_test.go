package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPathTraversal_ReadFile(t *testing.T) {
	tmpDir := t.TempDir()
	outsideFile := filepath.Join(tmpDir, "outside.txt")
	err := os.WriteFile(outsideFile, []byte("secret content"), 0644)
	if err != nil {
		t.Fatalf("failed to create outside file: %v", err)
	}

	workspaceDir := filepath.Join(tmpDir, "workspace")
	err = os.Mkdir(workspaceDir, 0755)
	if err != nil {
		t.Fatalf("failed to create workspace dir: %v", err)
	}

	// Now we pass workspaceDir to the tool
	tool, err := NewReadFileTool(workspaceDir)
	if err != nil {
		t.Fatalf("NewReadFileTool error: %v", err)
	}

	ctx := context.Background()
	// Attempt path traversal using absolute path to outside file
	argsJSON := fmt.Sprintf(`{"path": %q}`, outsideFile)

	result, err := tool.InvokableRun(ctx, argsJSON)

	// Now we expect an ERROR because we fixed it.
	if err == nil {
		t.Fatalf("SECURITY FAILURE: Was able to read file outside workspace! Result: %s", result)
	}

	if !strings.Contains(err.Error(), "access denied") {
		t.Errorf("Expected 'access denied' error, got: %v", err)
	}
}

func TestPathTraversal_RelativePath(t *testing.T) {
	tmpDir := t.TempDir()
	outsideFile := filepath.Join(tmpDir, "outside.txt")
	err := os.WriteFile(outsideFile, []byte("secret content"), 0644)
	if err != nil {
		t.Fatalf("failed to create outside file: %v", err)
	}

	workspaceDir := filepath.Join(tmpDir, "workspace")
	err = os.Mkdir(workspaceDir, 0755)
	if err != nil {
		t.Fatalf("failed to create workspace dir: %v", err)
	}

	tool, err := NewReadFileTool(workspaceDir)
	if err != nil {
		t.Fatalf("NewReadFileTool error: %v", err)
	}

	ctx := context.Background()
	// Attempt path traversal using relative path
	argsJSON := fmt.Sprintf(`{"path": "../outside.txt"}`)

	result, err := tool.InvokableRun(ctx, argsJSON)
	if err == nil {
		t.Fatalf("SECURITY FAILURE: Was able to read file outside workspace via relative path! Result: %s", result)
	}
	if !strings.Contains(err.Error(), "access denied") {
		t.Errorf("Expected 'access denied' error, got: %v", err)
	}
}

func TestPathTraversal_Symlink(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir := t.TempDir()

	// Create a "secret" file outside the workspace
	secretFile := filepath.Join(tmpDir, "secret.txt")
	err := os.WriteFile(secretFile, []byte("super secret data"), 0644)
	if err != nil {
		t.Fatalf("failed to create secret file: %v", err)
	}

	// Create a workspace directory
	workspaceDir := filepath.Join(tmpDir, "workspace")
	err = os.Mkdir(workspaceDir, 0755)
	if err != nil {
		t.Fatalf("failed to create workspace dir: %v", err)
	}

	// Create a symlink inside the workspace that points to the secret file
	symlinkPath := filepath.Join(workspaceDir, "link_to_secret")
	err = os.Symlink(secretFile, symlinkPath)
	if err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	// Initialize the read_file tool restricted to the workspace
	tool, err := NewReadFileTool(workspaceDir)
	if err != nil {
		t.Fatalf("NewReadFileTool error: %v", err)
	}

	ctx := context.Background()
	// Attempt to read the secret file via the symlink
	// The path passed to the tool is just "link_to_secret", which is seemingly inside the workspace
	argsJSON := fmt.Sprintf(`{"path": "link_to_secret"}`)

	result, err := tool.InvokableRun(ctx, argsJSON)

	// If the tool is secure, it should DENY access or fail to read content that is outside the workspace
	// However, without EvalSymlinks, it will likely succeed.

	if err == nil {
		// If we get here, it means the read succeeded.
		// Check if we actually read the secret data.
		if strings.Contains(result, "super secret data") {
			t.Fatalf("SECURITY FAILURE: Was able to read file outside workspace via symlink! Result: %s", result)
		}
	} else {
        // If it failed, that's good! But let's check the error message.
        if !strings.Contains(err.Error(), "access denied") && !strings.Contains(err.Error(), "outside workspace") {
             t.Logf("Read failed with unexpected error: %v", err)
        }
    }
}

func TestReadFileEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	tool, err := NewReadFileTool(tmpDir)
	if err != nil {
		t.Fatalf("failed to create tool: %v", err)
	}
	ctx := context.Background()

	tests := []struct {
		name     string
		content  string
		offset   int
		limit    int
		expected string
		lines    int
	}{
		{
			name:     "Empty file",
			content:  "",
			offset:   0,
			limit:    0,
			expected: "",
			lines:    1,
		},
		{
			name:     "One line no newline",
			content:  "hello",
			offset:   0,
			limit:    0,
			expected: "hello",
			lines:    1,
		},
		{
			name:     "One line with newline",
			content:  "hello\n",
			offset:   0,
			limit:    0,
			expected: "hello\n",
			lines:    2,
		},
		{
			name:     "Two lines",
			content:  "hello\nworld",
			offset:   0,
			limit:    0,
			expected: "hello\nworld",
			lines:    2,
		},
		{
			name:     "Two lines with trailing newline",
			content:  "hello\nworld\n",
			offset:   0,
			limit:    0,
			expected: "hello\nworld\n",
			lines:    3,
		},
		{
			name:     "Offset 1",
			content:  "line1\nline2\nline3",
			offset:   1,
			limit:    1,
			expected: "line2",
			lines:    3,
		},
		{
			name:     "Offset 1 Limit 0 (rest of file)",
			content:  "line1\nline2\nline3",
			offset:   1,
			limit:    0,
			expected: "line2\nline3",
			lines:    3,
		},
		{
			name:     "Offset out of bounds",
			content:  "line1",
			offset:   5,
			limit:    1,
			expected: "",
			lines:    1,
		},
		{
			name:     "Limit spans newline",
			content:  "line1\nline2\nline3",
			offset:   0,
			limit:    2,
			expected: "line1\nline2",
			lines:    3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fname := filepath.Join(tmpDir, tc.name+".txt")
			os.WriteFile(fname, []byte(tc.content), 0644)

			jsonPath := fmt.Sprintf("%s.txt", tc.name)
			input := fmt.Sprintf(`{"path": "%s", "offset": %d, "limit": %d}`, jsonPath, tc.offset, tc.limit)
			outStr, err := tool.InvokableRun(ctx, input)
			if err != nil {
				t.Fatalf("run failed: %v", err)
			}

			var out ReadFileOutput
			if err := json.Unmarshal([]byte(outStr), &out); err != nil {
				t.Fatalf("failed to unmarshal output: %v", err)
			}

			if out.Content != tc.expected {
				t.Errorf("expected content %q, got %q", tc.expected, out.Content)
			}
			if out.TotalLines != tc.lines {
				t.Errorf("expected total lines %d, got %d", tc.lines, out.TotalLines)
			}
		})
	}
}
