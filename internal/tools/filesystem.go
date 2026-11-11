package tools

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// validatePath 验证给定路径是否在允许的工作区范围内，防止路径穿越攻击。
// 返回值是调用方必须用于实际 I/O 的绝对路径。
// 相对路径相对于工作区解析，而不是进程的当前工作目录。
// 悬空符号链接会被拒绝：os.Stat 会把它看成不存在，但 os.WriteFile 会跟随链接，
// 从而可能在工作区外创建文件。
// 如果 workspacePath 为空，则跳过沙箱检查（仅用于向后兼容），仍返回清理后的绝对路径。
func validatePath(path, workspacePath string) (string, error) {
	if workspacePath == "" {
		abs, err := filepath.Abs(path)
		if err != nil {
			return "", fmt.Errorf("failed to resolve path: %w", err)
		}
		return filepath.Clean(abs), nil
	}

	workspaceAbs, err := filepath.Abs(workspacePath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve workspace path: %w", err)
	}
	workspaceAbs = filepath.Clean(workspaceAbs)

	workspaceResolved, err := resolvePathBestEffort(workspaceAbs)
	if err != nil {
		return "", fmt.Errorf("failed to resolve workspace path: %w", err)
	}

	if !filepath.IsAbs(path) {
		path = filepath.Join(workspaceAbs, path)
	}

	targetAbs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}
	targetAbs = filepath.Clean(targetAbs)

	// os.Stat 会跟随符号链接，悬空链接看起来像文件不存在。若只规范化父目录，
	// 父目录在工作区内，但随后的写入会跟随链接落到（可能在工作区外的）目标上。
	// Lstat 不跟随最后一段，因此在回退到父目录规范化之前先拒绝断链。
	if info, lstatErr := os.Lstat(targetAbs); lstatErr == nil && info.Mode()&os.ModeSymlink != 0 {
		if _, statErr := os.Stat(targetAbs); statErr != nil {
			if os.IsNotExist(statErr) {
				return "", fmt.Errorf("access denied: path %q is a broken symlink", path)
			}
			return "", fmt.Errorf("failed to stat path: %w", statErr)
		}
	}

	targetResolved, err := resolvePathBestEffort(targetAbs)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}

	if !isWithinWorkspace(targetResolved, workspaceResolved) {
		return "", fmt.Errorf("access denied: path %q is outside workspace %q", targetResolved, workspaceResolved)
	}
	return targetResolved, nil
}

func isWithinWorkspace(target, workspace string) bool {
	rel, err := filepath.Rel(workspace, target)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

// resolvePathBestEffort 尽力解析路径中已存在部分的符号链接。
// 对于尚不存在的写入目标，它会解析最长存在的父目录前缀。
func resolvePathBestEffort(path string) (string, error) {
	path = filepath.Clean(path)

	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		return filepath.Clean(resolved), nil
	} else if !os.IsNotExist(err) {
		return "", err
	}

	current := path
	var missing []string
	for {
		if _, err := os.Lstat(current); err == nil {
			resolvedPrefix, err := filepath.EvalSymlinks(current)
			if err != nil {
				return "", err
			}
			resolved := filepath.Clean(resolvedPrefix)
			for i := len(missing) - 1; i >= 0; i-- {
				resolved = filepath.Join(resolved, missing[i])
			}
			return filepath.Clean(resolved), nil
		} else if !os.IsNotExist(err) {
			return "", err
		}

		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("failed to find existing parent for path: %s", path)
		}
		missing = append(missing, filepath.Base(current))
		current = parent
	}
}

// ReadFileInput 定义了 read_file 工具的输入参数。
type ReadFileInput struct {
	Path   string `json:"path" jsonschema:"required,description=Absolute path to the file"`
	Offset int    `json:"offset" jsonschema:"description=Starting line number (0-based)"`
	Limit  int    `json:"limit" jsonschema:"description=Maximum number of lines to read"`
}

// ReadFileOutput 定义了 read_file 工具的执行结果。
type ReadFileOutput struct {
	Content    string `json:"content"`     // 文件内容
	TotalLines int    `json:"total_lines"` // 文件总行数
}

type readFileToolImpl struct {
	workspacePath string
}

func (t *readFileToolImpl) execute(ctx context.Context, input *ReadFileInput) (*ReadFileOutput, error) {
	path, err := validatePath(input.Path, t.workspacePath)
	if err != nil {
		return nil, err
	}
	return readFileRange(path, input.Offset, input.Limit)
}

// readFileRange 流式扫描换行并只读出所需要的片段。
// bytes.Count 走运行时的快速计数路径，避免把整个大文件先读进内存再 Split。
// TotalLines 与 strings.Split(content, "\n") 的行数保持一致。
func readFileRange(path string, offset, limit int) (*ReadFileOutput, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buf := make([]byte, 32*1024)
	var currentNewlineCount int
	startOffset := int64(-1)
	endOffset := int64(-1)
	if offset == 0 {
		startOffset = 0
	}

	var currentPos int64
	newlineSep := []byte{'\n'}

	for {
		n, err := file.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			// 起止偏移都已经确定后，只需要统计剩余换行，不必逐字节扫描。
			if (offset == 0 || startOffset != -1) && (limit == 0 || endOffset != -1) {
				currentNewlineCount += bytes.Count(chunk, newlineSep)
			} else {
				for i, b := range chunk {
					if b != '\n' {
						continue
					}
					currentNewlineCount++
					if offset > 0 && currentNewlineCount == offset {
						startOffset = currentPos + int64(i) + 1
					}
					if limit > 0 && currentNewlineCount == offset+limit {
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

	totalLines := currentNewlineCount + 1
	if startOffset == -1 {
		startOffset = currentPos
	}
	if endOffset == -1 {
		endOffset = currentPos
	}
	if startOffset >= endOffset {
		return &ReadFileOutput{Content: "", TotalLines: totalLines}, nil
	}

	contentBuf := make([]byte, endOffset-startOffset)
	if _, err := file.ReadAt(contentBuf, startOffset); err != nil && err != io.EOF {
		return nil, err
	}
	return &ReadFileOutput{
		Content:    string(contentBuf),
		TotalLines: totalLines,
	}, nil
}

// NewReadFileTool 创建 read_file 工具实例，用于读取工作区内的文件内容。
func NewReadFileTool(workspacePath string) (tool.InvokableTool, error) {
	impl := &readFileToolImpl{workspacePath: workspacePath}
	return utils.InferTool("read_file", "Read the contents of a file", impl.execute)
}

// WriteFileInput 定义了 write_file 工具的输入参数。
type WriteFileInput struct {
	Path    string `json:"path" jsonschema:"required,description=Absolute path to the file"`
	Content string `json:"content" jsonschema:"required,description=Content to write"`
}

type writeFileToolImpl struct {
	workspacePath string
}

func (t *writeFileToolImpl) execute(ctx context.Context, input *WriteFileInput) (string, error) {
	path, err := validatePath(input.Path, t.workspacePath)
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(path, []byte(input.Content), 0644); err != nil {
		return "", err
	}
	return "File written successfully", nil
}

// NewWriteFileTool 创建 write_file 工具实例，用于在工作区内写入新文件或覆盖已有文件。
func NewWriteFileTool(workspacePath string) (tool.InvokableTool, error) {
	impl := &writeFileToolImpl{workspacePath: workspacePath}
	return utils.InferTool("write_file", "Write content to a file", impl.execute)
}

// ListDirInput 定义了 list_dir 工具的输入参数。
type ListDirInput struct {
	Path string `json:"path" jsonschema:"required,description=Directory path to list"`
}

type listDirToolImpl struct {
	workspacePath string
}

func (t *listDirToolImpl) execute(ctx context.Context, input *ListDirInput) ([]string, error) {
	path, err := validatePath(input.Path, t.workspacePath)
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

// NewListDirTool 创建 list_dir 工具实例，用于列出目录下的所有文件和文件夹。
func NewListDirTool(workspacePath string) (tool.InvokableTool, error) {
	impl := &listDirToolImpl{workspacePath: workspacePath}
	return utils.InferTool("list_dir", "List contents of a directory", impl.execute)
}
