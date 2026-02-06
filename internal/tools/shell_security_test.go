package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExecTool_Security_RestrictToWorkspace(t *testing.T) {
	// Setup a workspace
	workspace := t.TempDir()

	// Create a tool with restrictToWorkspace = true
	tool, err := NewExecTool(10, true, workspace)
	if err != nil {
		t.Fatalf("Failed to create tool: %v", err)
	}

	// 1. Malicious Case: Try to execute with WorkingDir outside workspace
	targetDir := filepath.Dir(workspace) // Should be outside workspace

	input := ExecInput{
		Command:    "ls",
		WorkingDir: targetDir,
	}

	inputJson, _ := json.Marshal(input)

	ctx := context.Background()
	outputJson, err := tool.InvokableRun(ctx, string(inputJson))
	if err != nil {
		t.Fatalf("Tool execution failed: %v", err)
	}

	var output ExecOutput
	json.Unmarshal([]byte(outputJson), &output)

	if output.ExitCode != 0 {
		if strings.Contains(output.Stderr, "Access denied") || strings.Contains(output.Stderr, "outside workspace") {
			// Good
		} else {
			t.Logf("Command failed (maybe for other reasons): %s", output.Stderr)
		}
	} else {
		t.Fatalf("SECURITY FAILURE: Command succeeded in %s (should have been blocked): %s", targetDir, output.Stdout)
	}

	// 2. Valid Case: Execute with WorkingDir inside workspace
	subdir := filepath.Join(workspace, "subdir")
	if err := os.Mkdir(subdir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	inputValid := ExecInput{
		Command:    "ls", // Or pwd
		WorkingDir: subdir,
	}
	inputValidJson, _ := json.Marshal(inputValid)
	outputValidJson, err := tool.InvokableRun(ctx, string(inputValidJson))
	if err != nil {
		t.Fatalf("Tool execution failed: %v", err)
	}
	var outputValid ExecOutput
	json.Unmarshal([]byte(outputValidJson), &outputValid)

	if outputValid.ExitCode != 0 {
		t.Fatalf("Valid command failed: %s", outputValid.Stderr)
	}

    // 3. Valid Case: Default WorkingDir (empty)
    inputDefault := ExecInput{
        Command: "ls",
    }
    inputDefaultJson, _ := json.Marshal(inputDefault)
    outputDefaultJson, err := tool.InvokableRun(ctx, string(inputDefaultJson))
    if err != nil {
        t.Fatalf("Tool execution failed: %v", err)
    }
    var outputDefault ExecOutput
    json.Unmarshal([]byte(outputDefaultJson), &outputDefault)
    if outputDefault.ExitCode != 0 {
        t.Fatalf("Default command failed: %s", outputDefault.Stderr)
    }
}
