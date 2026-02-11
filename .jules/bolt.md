## 2025-02-18 - [Optimizing Tool File Access]
**Learning:** Tools in this agent architecture often need to read small slices of potentially large files (e.g., logs). Using `os.ReadFile` followed by `strings.Split` forces the entire file into memory twice (bytes + string slice).
**Action:** Prefer streaming file access (checking `\n` manually) and `ReadAt` to extract only the needed segment. This reduces memory pressure from O(File) to O(Buffer). Be careful to replicate `strings.Split` edge cases (e.g., trailing newlines).

## 2026-02-11 - [Optimizing Byte Scanning in Go]
**Learning:** Naive `for range` loops over byte slices to find delimiters (like `\n`) are significantly slower (2x-5x) than `bytes.IndexByte` and `bytes.Count` which use assembly/SIMD.
**Action:** When scanning large buffers, always use `bytes` package functions for finding and counting delimiters instead of manual iteration.
