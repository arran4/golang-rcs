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
)

func GenerateRandomLines(n int, lineLen int) []string {
	lines := make([]string, n)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < n; i++ {
		lines[i] = randomString(lineLen)
	}
	return lines
}

func randomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(rand.Intn(26) + 'a')
	}
	return string(b)
}

func GenerateCodeLines(n int) []string {
	lines := make([]string, n)
	rand.Seed(time.Now().UnixNano())
	indent := 0
	for i := 0; i < n; i++ {
		if rand.Float32() < 0.1 && indent > 0 {
			indent--
		}
		prefix := strings.Repeat("\t", indent)
		lines[i] = prefix + "if (cond) {"
		if rand.Float32() < 0.2 {
			indent++
		}
	}
	return lines
}

func GenerateRepetitiveLines(n int, uniqueLines int) []string {
	lines := make([]string, n)
	uniques := GenerateRandomLines(uniqueLines, 20)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < n; i++ {
		lines[i] = uniques[rand.Intn(uniqueLines)]
	}
	return lines
}

func benchmarkDiff(b *testing.B, algoName string, generator func(int) []string, n int) {
	algo, err := diff.GetAlgorithm(algoName)
	if err != nil {
		b.Fatalf("algorithm %s not found: %v", algoName, err)
	}

	lines1 := generator(n)
	lines2 := generator(n) // Two different sets generated similarly

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

	algos := []string{"lcs", "hashline"}
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

	algos := []string{"lcs", "hashline"}
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
			// Modify some lines in lines2 to create diffs but keep repetitive structure
			lines2 := make([]string, size)
			copy(lines2, lines1)
			// Change 10% of lines
			for i := 0; i < size/10; i++ {
				idx := rand.Intn(size)
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
