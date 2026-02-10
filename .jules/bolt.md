## 2025-02-18 - [Optimizing Tool File Access]
**Learning:** Tools in this agent architecture often need to read small slices of potentially large files (e.g., logs). Using `os.ReadFile` followed by `strings.Split` forces the entire file into memory twice (bytes + string slice).
**Action:** Prefer streaming file access (checking `\n` manually) and `ReadAt` to extract only the needed segment. This reduces memory pressure from O(File) to O(Buffer). Be careful to replicate `strings.Split` edge cases (e.g., trailing newlines).

## 2025-05-21 - [Go Byte Scanning Performance]
**Learning:** Manual byte-by-byte iteration in Go (`for i, b := range chunk`) to find delimiters is significantly slower than standard library functions which use assembly/SIMD.
**Action:** Use `bytes.IndexByte` and `bytes.Count` for scanning buffers. In benchmarks, replacing a manual loop with `bytes.IndexByte` yielded a ~5x speedup (9ms -> 1.7ms) for 10MB file scanning.
