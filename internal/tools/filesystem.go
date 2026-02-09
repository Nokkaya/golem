package tools

import (
	"context"
	"fmt"
	"io"
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

	canonicalWorkspace, err := filepath.EvalSymlinks(absWorkspace)
	if err != nil {
		// If workspace path doesn't exist, we can't enforce containment strictly.
		// However, for security, we should probably fail if we can't resolve the workspace path,
		// unless the workspace directory itself is expected to be created later.
		// Given this is a security validation, failing secure is better.
		// But in some setups (e.g. tests), the workspace might not exist yet?
		// Actually, if workspace doesn't exist, we can't really operate in it.
		// Let's assume for now that if EvalSymlinks fails because it doesn't exist, we fall back to Abs.
		if os.IsNotExist(err) {
			canonicalWorkspace = absWorkspace
		} else {
			return "", fmt.Errorf("failed to resolve workspace symlinks: %w", err)
		}
	}

	var absTarget string
	if filepath.IsAbs(target) {
		absTarget = filepath.Clean(target)
	} else {
		absTarget = filepath.Join(canonicalWorkspace, target)
	}

	// Resolve symlinks in the target path to ensure it doesn't point outside
	canonicalTarget, err := filepath.EvalSymlinks(absTarget)
	if err != nil {
		// If the target file doesn't exist (e.g. new file for write_file),
		// we check the parent directory.
		if os.IsNotExist(err) {
			parent := filepath.Dir(absTarget)
			canonicalParent, err := filepath.EvalSymlinks(parent)
			if err != nil {
				// If parent also doesn't exist, we can't easily validate.
				// But write_file usually requires the directory to exist (os.WriteFile doesn't mkdir -p).
				// So if parent doesn't exist, the operation will likely fail later anyway.
				// We return error here to be safe.
				return "", fmt.Errorf("failed to resolve parent path of target: %w", err)
			}
			canonicalTarget = filepath.Join(canonicalParent, filepath.Base(absTarget))
		} else {
			return "", fmt.Errorf("failed to resolve target path symlinks: %w", err)
		}
	}

	rel, err := filepath.Rel(canonicalWorkspace, canonicalTarget)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path relative to workspace: %w", err)
	}

	if strings.HasPrefix(rel, "..") || rel == ".." {
		return "", fmt.Errorf("access denied: path %q is outside workspace", target)
	}

	return canonicalTarget, nil
}

// NewReadFileTool creates the read_file tool
func NewReadFileTool(workspacePath string) (tool.InvokableTool, error) {
	run := func(ctx context.Context, input *ReadFileInput) (*ReadFileOutput, error) {
		path, err := validatePath(workspacePath, input.Path)
		if err != nil {
			return nil, err
		}

		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		// buffer for reading
		buf := make([]byte, 32*1024)

		var totalLines int = 0
		var currentNewlineCount int = 0

		var startOffset int64 = -1
		var endOffset int64 = -1

		// If Offset is 0, startOffset is 0.
		if input.Offset == 0 {
			startOffset = 0
		}

		var currentPos int64 = 0

		for {
			n, err := file.Read(buf)
			if n > 0 {
				chunk := buf[:n]

				// Count newlines in this chunk
				for i, b := range chunk {
					if b == '\n' {
						currentNewlineCount++
						// Found a newline.
						// If this is the Offset-th newline, start reading from next byte.
						if input.Offset > 0 && currentNewlineCount == input.Offset {
							startOffset = currentPos + int64(i) + 1
						}

						// If this is the (Offset + Limit)-th newline, stop reading (endOffset).
						if input.Limit > 0 && currentNewlineCount == (input.Offset+input.Limit) {
							endOffset = currentPos + int64(i)
						}
					}
				}
				currentPos += int64(n)
			}

			if err != nil {
				if err == io.EOF {
					break
				}
				return nil, err
			}
		}

		// Calculate TotalLines
		// Mimic strings.Split behavior:
		// "a\n" -> 2 lines. "a" -> 1 line. "" -> 1 line.
		totalLines = currentNewlineCount + 1

		// Determine offsets
		if startOffset == -1 {
			// Requested offset is beyond file
			startOffset = currentPos // EOF
		}

		if endOffset == -1 {
			// Limit not reached or Limit=0
			endOffset = currentPos
		}

		// Read content
		if startOffset >= endOffset {
			return &ReadFileOutput{
				Content:    "",
				TotalLines: totalLines,
			}, nil
		}

		contentSize := endOffset - startOffset
		// Protection against negative size if logic fails (shouldn't happen)
		if contentSize < 0 {
			contentSize = 0
		}

		contentBuf := make([]byte, contentSize)

		_, err = file.ReadAt(contentBuf, startOffset)
		if err != nil && err != io.EOF {
			return nil, err
		}

		return &ReadFileOutput{
			Content:    string(contentBuf),
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
