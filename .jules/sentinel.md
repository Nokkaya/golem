## 2025-02-04 - Path Traversal in Filesystem Tools
**Vulnerability:** The `read_file`, `write_file`, and `list_dir` tools accepted absolute paths without any validation against a workspace root, allowing arbitrary file system access.
**Learning:** Tools in an agentic system must be sandboxed by default. Relying on "safe mode" flags or user trust is insufficient for file operations.
**Prevention:** Always wrap file operations in a path validation helper that resolves paths relative to a strict workspace root using `filepath.Rel` and checking for `..` components.

## 2025-02-18 - Shell Execution Path Traversal
**Vulnerability:** The `exec` tool honored `WorkingDir` parameter even when `restrictToWorkspace` was enabled, allowing execution of commands in arbitrary directories.
**Learning:** Configuration flags like `restrictToWorkspace` must be enforced on all input parameters that affect file system access, not just the command itself.
**Prevention:** Validate `WorkingDir` against the workspace root using the same path validation logic as filesystem tools.

## 2025-04-18 - Symlink Path Traversal Bypass
**Vulnerability:** Attackers could bypass path validation by creating symlinks inside the workspace pointing to external files (e.g., `ln -s /etc/passwd workspace/link`). Standard `filepath.Rel` checks passed because the symlink path itself was inside the workspace.
**Learning:** `filepath.Rel` works on lexical paths and does not resolve symlinks. A path can be lexically inside a root but physically outside.
**Prevention:** Use `filepath.EvalSymlinks` on both the workspace root and the target path before validation to resolve them to their canonical physical paths.

## 2025-04-18 - Broken Symlink Write Vulnerability
**Vulnerability:** When writing to a new file, `filepath.EvalSymlinks` fails if the target doesn't exist. If the target was a broken symlink pointing outside, falling back to parent directory validation allowed writing through the broken symlink, creating the external file.
**Learning:** "File does not exist" does not mean "Path is safe". It might be a broken symlink.
**Prevention:** In `os.IsNotExist` error handlers, explicitly check `os.Lstat` to see if the path exists as a symlink. If so, deny access or treat it as unsafe.
