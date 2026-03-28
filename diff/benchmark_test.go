package diff_test

import (
	"fmt"
	"math/rand"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/arran4/golang-rcs/diff"
	_ "github.com/arran4/golang-rcs/diff/hashline"
	_ "github.com/arran4/golang-rcs/diff/znkr_diff"
)

// GenerateRandomLines generates random lines using a local random source for reproducibility in benchmarks.
func GenerateRandomLines(n int, lineLen int) []string {
	// Use a fixed seed for reproducible benchmarks across runs, or time-based if variance is desired.
	// For benchmarks, consistency is usually better.
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)

	lines := make([]string, n)
	for i := 0; i < n; i++ {
		lines[i] = randomString(r, lineLen)
	}
	return lines
}

func randomString(r *rand.Rand, n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(r.Intn(26) + 'a')
	}
	return string(b)
}

func GenerateCodeLines(n int) []string {
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)

	lines := make([]string, n)
	indent := 0
	for i := 0; i < n; i++ {
		if r.Float32() < 0.1 && indent > 0 {
			indent--
		}
		prefix := strings.Repeat("\t", indent)
		lines[i] = prefix + "if (cond) {"
		if r.Float32() < 0.2 {
			indent++
		}
	}
	return lines
}

func GenerateRepetitiveLines(n int, uniqueLines int) []string {
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)

	// Generate uniques first
	uniques := make([]string, uniqueLines)
	for i := 0; i < uniqueLines; i++ {
		uniques[i] = randomString(r, 20)
	}

	lines := make([]string, n)
	for i := 0; i < n; i++ {
		lines[i] = uniques[r.Intn(uniqueLines)]
	}
	return lines
}

func benchmarkDiff(b *testing.B, algoName string, generator func(int) []string, n int) {
	algo, err := diff.GetAlgorithm(algoName)
	if err != nil {
		b.Fatalf("algorithm %s not found: %v", algoName, err)
	}

	// For micro-benchmarks, we might want to generate once outside the loop to measure pure algo time.
	// However, if the algo modifies input (it shouldn't), that would be bad.
	// Assuming algo is read-only.
	lines1 := generator(n)
	lines2 := generator(n)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := algo(lines1, lines2)
		if err != nil {
			b.Fatalf("algorithm failed: %v", err)
		}
	}
}

func BenchmarkDiff_LCS_Random_100(b *testing.B) {
	benchmarkDiff(b, "lcs", func(n int) []string { return GenerateRandomLines(n, 20) }, 100)
}
func BenchmarkDiff_LCS_Random_1000(b *testing.B) {
	benchmarkDiff(b, "lcs", func(n int) []string { return GenerateRandomLines(n, 20) }, 1000)
}
func BenchmarkDiff_LCS_Random_5000(b *testing.B) {
	benchmarkDiff(b, "lcs", func(n int) []string { return GenerateRandomLines(n, 20) }, 5000)
}

func BenchmarkDiff_HashLine_Random_100(b *testing.B) {
	benchmarkDiff(b, "hashline", func(n int) []string { return GenerateRandomLines(n, 20) }, 100)
}
func BenchmarkDiff_HashLine_Random_1000(b *testing.B) {
	benchmarkDiff(b, "hashline", func(n int) []string { return GenerateRandomLines(n, 20) }, 1000)
}
func BenchmarkDiff_HashLine_Random_5000(b *testing.B) {
	benchmarkDiff(b, "hashline", func(n int) []string { return GenerateRandomLines(n, 20) }, 5000)
}
func BenchmarkDiff_HashLine_Random_10000(b *testing.B) {
	benchmarkDiff(b, "hashline", func(n int) []string { return GenerateRandomLines(n, 20) }, 10000)
}

func TestBenchmarkReport(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping benchmark report in short mode")
	}

	algos := []string{"lcs", "hashline", "znkr"}
	sizes := []int{100, 1000, 5000, 10000} // LCS might fail/timeout on 10000 depending on implementation efficiency

	for _, algoName := range algos {
		fmt.Printf("\nAlgorithm: %s\n", algoName)
		fmt.Printf("Size\tTime(ms)\tAlloc(MB)\tEdDiff Size\n")
		algo, err := diff.GetAlgorithm(algoName)
		if err != nil {
			t.Fatalf("algorithm %s not found: %v", algoName, err)
		}

		for _, size := range sizes {
			// Skip LCS for 10000 as it might be too slow/memory intensive (O(N*M))
			if algoName == "lcs" && size > 5000 {
				fmt.Printf("%d\tSKIPPED (O(N^2))\n", size)
				continue
			}

			lines1 := GenerateRandomLines(size, 20)
			lines2 := GenerateRandomLines(size, 20)

			// Force GC before measurement
			runtime.GC()
			var m1, m2 runtime.MemStats
			runtime.ReadMemStats(&m1)

			start := time.Now()
			edDiff, err := algo(lines1, lines2)
			duration := time.Since(start)

			runtime.ReadMemStats(&m2)
			if err != nil {
				t.Fatalf("algorithm failed: %v", err)
			}

			alloc := float64(m2.TotalAlloc-m1.TotalAlloc) / 1024 / 1024

			// Measure output size (approximate number of changes)
			// EdDiff is []diff.EdDiffCommand
			diffSize := len(edDiff)

			fmt.Printf("%d\t%d\t%.2f\t%d\n", size, duration.Milliseconds(), alloc, diffSize)
		}
	}
}

func TestBenchmarkReport_Repetitive(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping benchmark report in short mode")
	}

	algos := []string{"lcs", "hashline", "znkr"}
	sizes := []int{100, 1000, 5000} // Repetitive might be better handled by HashLine

	for _, algoName := range algos {
		fmt.Printf("\nAlgorithm (Repetitive): %s\n", algoName)
		fmt.Printf("Size\tTime(ms)\tAlloc(MB)\tEdDiff Size\n")
		algo, err := diff.GetAlgorithm(algoName)
		if err != nil {
			t.Fatalf("algorithm %s not found: %v", algoName, err)
		}

		for _, size := range sizes {
			// Generate repetitive lines (e.g., 10 unique lines repeated)
			lines1 := GenerateRepetitiveLines(size, 10)

			// Use local rand for modifications too
			src := rand.NewSource(time.Now().UnixNano())
			r := rand.New(src)

			// Modify some lines in lines2 to create diffs but keep repetitive structure
			lines2 := make([]string, size)
			copy(lines2, lines1)
			// Change 10% of lines
			for i := 0; i < size/10; i++ {
				idx := r.Intn(size)
				lines2[idx] = "modified line"
			}

			// Force GC before measurement
			runtime.GC()
			var m1, m2 runtime.MemStats
			runtime.ReadMemStats(&m1)

			start := time.Now()
			edDiff, err := algo(lines1, lines2)
			duration := time.Since(start)

			runtime.ReadMemStats(&m2)
			if err != nil {
				t.Fatalf("algorithm failed: %v", err)
			}

			alloc := float64(m2.TotalAlloc-m1.TotalAlloc) / 1024 / 1024
			diffSize := len(edDiff)

			fmt.Printf("%d\t%d\t%.2f\t%d\n", size, duration.Milliseconds(), alloc, diffSize)
		}
	}
}

// GenerateTrickyLines creates a challenging dataset with many small shifting blocks,
// repetitions, and scattered inserts/deletes to test diff algorithm robustness.
func GenerateTrickyLines(n int) []string {
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)

	// Create a base template with repeating patterns
	pattern := []string{"start", "header", "body1", "body2", "footer", "end"}

	lines := make([]string, 0, n)
	for i := 0; len(lines) < n; i++ {
		lines = append(lines, pattern[i%len(pattern)])
		// Occasionally insert random noise or modify the block
		if r.Float32() < 0.15 {
			lines = append(lines, fmt.Sprintf("noise_%d", r.Intn(100)))
		}
	}

	// Ensure we return exactly n
	if len(lines) > n {
		return lines[:n]
	}
	return lines
}

func TestBenchmarkReport_Tricky(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping benchmark report in short mode")
	}

	algos := []string{"lcs", "hashline", "znkr"}
	sizes := []int{100, 1000, 5000} // Skip 10k due to LCS

	for _, algoName := range algos {
		fmt.Printf("\nAlgorithm (Tricky): %s\n", algoName)
		fmt.Printf("Size\tTime(ms)\tAlloc(MB)\tEdDiff Size\tAdded Lines\n")
		algo, err := diff.GetAlgorithm(algoName)
		if err != nil {
			t.Fatalf("algorithm %s not found: %v", algoName, err)
		}

		for _, size := range sizes {
			lines1 := GenerateTrickyLines(size)

			// Modify 'lines1' heavily to create 'lines2'
			src := rand.NewSource(time.Now().UnixNano())
			r := rand.New(src)

			lines2 := make([]string, 0, size)
			for i := 0; i < len(lines1); i++ {
				// 10% chance to delete
				if r.Float32() < 0.10 {
					continue
				}

				// 10% chance to modify
				if r.Float32() < 0.10 {
					lines2 = append(lines2, "MODIFIED_LINE")
					continue
				}

				lines2 = append(lines2, lines1[i])

				// 5% chance to insert random block
				if r.Float32() < 0.05 {
					for j := 0; j < 5; j++ {
						lines2 = append(lines2, "NEW_BLOCK_DATA")
					}
				}
			}

			// Force GC before measurement
			runtime.GC()
			var m1, m2 runtime.MemStats
			runtime.ReadMemStats(&m1)

			start := time.Now()
			edDiff, err := algo(lines1, lines2)
			duration := time.Since(start)

			runtime.ReadMemStats(&m2)
			if err != nil {
				t.Fatalf("algorithm failed: %v", err)
			}

			alloc := float64(m2.TotalAlloc-m1.TotalAlloc) / 1024 / 1024

			// Measure output size (number of command instructions vs number of strings added)
			diffCmdsSize := len(edDiff)
			addedLines := 0
			for _, cmd := range edDiff {
				if addCmd, ok := cmd.(diff.Add); ok {
					addedLines += len(addCmd.Lines)
				}
			}

			fmt.Printf("%d\t%d\t%.2f\t%d\t\t%d\n", size, duration.Milliseconds(), alloc, diffCmdsSize, addedLines)
		}
	}
}
