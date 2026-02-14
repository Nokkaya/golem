## 2025-02-04 - Path Traversal in Filesystem Tools
**Vulnerability:** The `read_file`, `write_file`, and `list_dir` tools accepted absolute paths without any validation against a workspace root, allowing arbitrary file system access.
**Learning:** Tools in an agentic system must be sandboxed by default. Relying on "safe mode" flags or user trust is insufficient for file operations.
**Prevention:** Always wrap file operations in a path validation helper that resolves paths relative to a strict workspace root using `filepath.Rel` and checking for `..` components.

## 2025-02-18 - Shell Execution Path Traversal
**Vulnerability:** The `exec` tool honored `WorkingDir` parameter even when `restrictToWorkspace` was enabled, allowing execution of commands in arbitrary directories.
**Learning:** Configuration flags like `restrictToWorkspace` must be enforced on all input parameters that affect file system access, not just the command itself.
**Prevention:** Validate `WorkingDir` against the workspace root using the same path validation logic as filesystem tools.

## 2026-02-14 - Symlink Path Traversal in Filesystem Tools
**Vulnerability:** The `validatePath` function used `filepath.Clean` which resolved `..` components but failed to resolve symbolic links. This allowed accessing files outside the workspace via symlinks created inside it.
**Learning:** Path validation logic must always canonicalize paths by resolving all symbolic links. `filepath.Clean` is insufficient for security boundaries when the filesystem supports symlinks.
**Prevention:** Use `filepath.EvalSymlinks` on both the workspace root and the target path to obtain their canonical absolute paths before performing containment checks.
