package hashline

import (
	"hash/fnv"
	"runtime"
	"sort"
	"sync"

	"github.com/arran4/golang-rcs/diff"
)

func init() {
	diff.Register("hashline", GenerateEdDiffFromLines)
}

func hashBytes(b []byte) uint64 {
	h := fnv.New64a()
	var err error
	_, err = h.Write(b)
	if err != nil {
		panic(err)
	}
	return h.Sum64()
}

func parallelHash(lines []string) []uint64 {
	n := len(lines)
	hashes := make([]uint64, n)
	numCPU := runtime.NumCPU()
	if n < 1000 || numCPU == 1 {
		for i, line := range lines {
			hashes[i] = hashBytes([]byte(line))
		}
		return hashes
	}

	var wg sync.WaitGroup
	chunkSize := (n + numCPU - 1) / numCPU
	for i := 0; i < numCPU; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if start >= n {
			break
		}
		if end > n {
			end = n
		}
		wg.Add(1)
		go func(s, e int) {
			defer wg.Done()
			for k := s; k < e; k++ {
				hashes[k] = hashBytes([]byte(lines[k]))
			}
		}(start, end)
	}
	wg.Wait()
	return hashes
}

func GenerateEdDiffFromLines(from []string, to []string) (diff.EdDiff, error) {
	// 1. Parallel Hash 'from' lines
	fromHashes := parallelHash(from)

	// Map from Hash to list of occurrences in 'from'
	m := make(map[uint64][]int, len(from))
	for linePos, h := range fromHashes {
		m[h] = append(m[h], linePos)
	}

	type ContinuousRun struct {
		FromStart int
		ToStart   int
		Length    int
	}

	var runs []ContinuousRun

	// Optimized approach using the Hash Map:
	// Iterate through 'to'.
	// Keep track of active runs.
	// currentRuns: map[int]int  (FromIndex -> Length of run ending at FromIndex)

	// We use two maps and swap them to avoid allocation per iteration
	map1 := make(map[int]int)
	map2 := make(map[int]int)
	currentRuns := map1

	// 'nextRuns' will point to map2 initially in the loop logic below (via swap logic)
	// But actually we need to designate one map as 'next' buffer.
	nextRunsBuf := map2

	// 2. Parallel Hash 'to' lines (optional, but good for large files)
	toHashes := parallelHash(to)

	for j, h := range toHashes {
		nextRuns := nextRunsBuf
		// Clear nextRuns for reuse
		clear(nextRuns)

		if indices, ok := m[h]; ok {
			for _, i := range indices {
				// Check for string equality to be safe
				if from[i] == to[j] {
					length := 1
					if prevLen, ok := currentRuns[i-1]; ok {
						length = prevLen + 1
					}
					nextRuns[i] = length
				}
			}
		}

		// Any run in currentRuns that is NOT in nextRuns has terminated.
		for i, length := range currentRuns {
			if _, extended := nextRuns[i+1]; !extended {
				// Run ended at i, j-1
				// Start was i - length + 1, j - 1 - length + 1 => j - length
				runs = append(runs, ContinuousRun{
					FromStart: i - length + 1,
					ToStart:   j - length,
					Length:    length,
				})
			}
		}

		// Swap maps
		// currentRuns becomes the one we just populated (nextRuns)
		// nextRunsBuf becomes the old currentRuns (to be cleared next time)
		tmp := currentRuns
		currentRuns = nextRuns
		nextRunsBuf = tmp
	}

	// Add remaining runs at the end of 'to'
	for i, length := range currentRuns {
		runs = append(runs, ContinuousRun{
			FromStart: i - length + 1,
			ToStart:   len(to) - length,
			Length:    length,
		})
	}

	// Sort runs
	// 1. Length (descending)
	// 2. Distance from origin? i.e. (i+j) ascending? Or closeness to diagonal?
	sort.Slice(runs, func(i, j int) bool {
		if runs[i].Length != runs[j].Length {
			return runs[i].Length > runs[j].Length
		}
		// Tie-breaker: Distance from origin (FromStart + ToStart)
		distI := runs[i].FromStart + runs[i].ToStart
		distJ := runs[j].FromStart + runs[j].ToStart
		return distI < distJ
	})

	// Greedy selection
	var selectedRuns []ContinuousRun

	for _, run := range runs {
		conflict := false
		for _, s := range selectedRuns {
			// Check if run is strictly before or after s
			isBefore := (run.FromStart+run.Length <= s.FromStart) && (run.ToStart+run.Length <= s.ToStart)
			isAfter := (run.FromStart >= s.FromStart+s.Length) && (run.ToStart >= s.ToStart+s.Length)

			if !isBefore && !isAfter {
				conflict = true
				break
			}
		}
		if !conflict {
			selectedRuns = append(selectedRuns, run)
		}
	}

	// Sort selected runs by position (FromStart) to generate linear diff
	sort.Slice(selectedRuns, func(i, j int) bool {
		return selectedRuns[i].FromStart < selectedRuns[j].FromStart
	})

	// Generate diffs for gaps
	var result diff.EdDiff

	currFrom := 0
	currTo := 0

	for _, run := range selectedRuns {
		// Gap between curr and run
		gapFromLen := run.FromStart - currFrom
		gapToLen := run.ToStart - currTo

		// Handle Gap
		if gapFromLen > 0 {
			// Delete from gap
			result = append(result, diff.Delete{currFrom + 1, gapFromLen})
		}
		if gapToLen > 0 {
			// Add to gap
			// Extract lines to add
			linesToAdd := to[currTo : currTo+gapToLen]
			result = append(result, diff.Add{LineStart: currFrom, Lines: linesToAdd})
		}

		// Advance current
		currFrom = run.FromStart + run.Length
		currTo = run.ToStart + run.Length
	}

	// Final gap after last run
	gapFromLen := len(from) - currFrom
	gapToLen := len(to) - currTo

	if gapFromLen > 0 {
		result = append(result, diff.Delete{currFrom + 1, gapFromLen})
	}
	if gapToLen > 0 {
		result = append(result, diff.Add{LineStart: currFrom, Lines: to[currTo:]})
	}

	return result, nil
}
