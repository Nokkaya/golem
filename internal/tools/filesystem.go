package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// ReadFileInput parameters for read_file tool
type ReadFileInput struct {
	Path   string `json:"path" jsonschema:"required,description=Absolute path to the file"`
	Offset int    `json:"offset" jsonschema:"description=Starting line number (0-based)"`
	Limit  int    `json:"limit" jsonschema:"description=Maximum number of lines to read"`
}

// ReadFileOutput result of read_file tool
type ReadFileOutput struct {
	Content    string `json:"content"`
	TotalLines int    `json:"total_lines"`
}

// validatePath ensures the target path is within the workspace
func validatePath(workspace, target string) (string, error) {
	if workspace == "" {
		return target, nil
	}

	absWorkspace, err := filepath.Abs(workspace)
	if err != nil {
		return "", fmt.Errorf("failed to resolve workspace path: %w", err)
	}

	var absTarget string
	if filepath.IsAbs(target) {
		absTarget = filepath.Clean(target)
	} else {
		absTarget = filepath.Join(absWorkspace, target)
	}

	rel, err := filepath.Rel(absWorkspace, absTarget)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path relative to workspace: %w", err)
	}

	if strings.HasPrefix(rel, "..") || rel == ".." {
		return "", fmt.Errorf("access denied: path %q is outside workspace", target)
	}

	return absTarget, nil
}

// NewReadFileTool creates the read_file tool
func NewReadFileTool(workspacePath string) (tool.InvokableTool, error) {
	run := func(ctx context.Context, input *ReadFileInput) (*ReadFileOutput, error) {
		path, err := validatePath(workspacePath, input.Path)
		if err != nil {
			return nil, err
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		content := string(data)
		lines := strings.Split(content, "\n")
		totalLines := len(lines)

		if input.Offset > 0 {
			if input.Offset >= len(lines) {
				lines = []string{}
			} else {
				lines = lines[input.Offset:]
			}
		}

		if input.Limit > 0 && input.Limit < len(lines) {
			lines = lines[:input.Limit]
		}

		return &ReadFileOutput{
			Content:    strings.Join(lines, "\n"),
			TotalLines: totalLines,
		}, nil
	}
	return utils.InferTool("read_file", "Read the contents of a file", run)
}

// WriteFileInput parameters for write_file tool
type WriteFileInput struct {
	Path    string `json:"path" jsonschema:"required,description=Absolute path to the file"`
	Content string `json:"content" jsonschema:"required,description=Content to write"`
}

// NewWriteFileTool creates the write_file tool
func NewWriteFileTool(workspacePath string) (tool.InvokableTool, error) {
	run := func(ctx context.Context, input *WriteFileInput) (string, error) {
		path, err := validatePath(workspacePath, input.Path)
		if err != nil {
			return "", err
		}

		err = os.WriteFile(path, []byte(input.Content), 0644)
		if err != nil {
			return "", err
		}
		return "File written successfully", nil
	}
	return utils.InferTool("write_file", "Write content to a file", run)
}

// ListDirInput parameters for list_dir tool
type ListDirInput struct {
	Path string `json:"path" jsonschema:"required,description=Directory path to list"`
}

// NewListDirTool creates the list_dir tool
func NewListDirTool(workspacePath string) (tool.InvokableTool, error) {
	run := func(ctx context.Context, input *ListDirInput) ([]string, error) {
		path, err := validatePath(workspacePath, input.Path)
		if err != nil {
			return nil, err
		}

		entries, err := os.ReadDir(path)
		if err != nil {
			return nil, err
		}

		var result []string
		for _, entry := range entries {
			name := entry.Name()
			if entry.IsDir() {
				name += "/"
			}
			result = append(result, name)
		}
		return result, nil
	}
	return utils.InferTool("list_dir", "List contents of a directory", run)
}
