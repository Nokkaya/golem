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

func TestExecTool_Security_SymlinkWorkingDirOutside(t *testing.T) {
	tmp := t.TempDir()
	workspace := filepath.Join(tmp, "workspace")
	outside := filepath.Join(tmp, "outside")
	if err := os.Mkdir(workspace, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(outside, 0755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(workspace, "escape")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatal(err)
	}

	tool, err := NewExecTool(10, true, workspace)
	if err != nil {
		t.Fatal(err)
	}
	input, _ := json.Marshal(ExecInput{Command: "pwd", WorkingDir: link})
	outputJSON, err := tool.InvokableRun(context.Background(), string(input))
	if err != nil {
		t.Fatal(err)
	}
	var output ExecOutput
	if err := json.Unmarshal([]byte(outputJSON), &output); err != nil {
		t.Fatal(err)
	}
	if output.ExitCode == 0 {
		t.Fatalf("SECURITY FAILURE: command ran via symlink working dir: %s", output.Stdout)
	}
	if !strings.Contains(output.Stderr, "access denied") && !strings.Contains(output.Stderr, "outside") {
		t.Fatalf("expected access denied, got %s", output.Stderr)
	}
}

func TestExecTool_Security_RelativeWorkingDirInside(t *testing.T) {
	workspace := t.TempDir()
	sub := filepath.Join(workspace, "sub")
	if err := os.Mkdir(sub, 0755); err != nil {
		t.Fatal(err)
	}
	resolved, err := filepath.EvalSymlinks(sub)
	if err != nil {
		t.Fatal(err)
	}

	tool, err := NewExecTool(10, true, workspace)
	if err != nil {
		t.Fatal(err)
	}
	input, _ := json.Marshal(ExecInput{Command: "pwd", WorkingDir: "sub"})
	outputJSON, err := tool.InvokableRun(context.Background(), string(input))
	if err != nil {
		t.Fatal(err)
	}
	var output ExecOutput
	if err := json.Unmarshal([]byte(outputJSON), &output); err != nil {
		t.Fatal(err)
	}
	if output.ExitCode != 0 {
		t.Fatalf("relative working dir inside workspace failed: %s", output.Stderr)
	}
	if !strings.Contains(output.Stdout, resolved) {
		t.Fatalf("pwd = %q, want path under %s", output.Stdout, resolved)
	}
}
