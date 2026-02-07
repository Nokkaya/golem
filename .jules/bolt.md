## 2025-02-18 - [Optimizing Tool File Access]
**Learning:** Tools in this agent architecture often need to read small slices of potentially large files (e.g., logs). Using `os.ReadFile` followed by `strings.Split` forces the entire file into memory twice (bytes + string slice).
**Action:** Prefer streaming file access (checking `\n` manually) and `ReadAt` to extract only the needed segment. This reduces memory pressure from O(File) to O(Buffer). Be careful to replicate `strings.Split` edge cases (e.g., trailing newlines).

## 2026-02-07 - [Optimizing Go Loops with SIMD]
**Learning:** Iterating byte-by-byte in Go (range loop) is significantly slower than using `bytes.IndexByte` and `bytes.Count`, which leverage optimized assembly (SIMD).
**Action:** Replace `for i, b := range chunk` with `bytes.IndexByte` when searching for delimiters in high-throughput paths like file reading tools.
