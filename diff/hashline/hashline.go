package hashline

import (
	"hash/fnv"
	"sort"

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

func GenerateEdDiffFromLines(from []string, to []string) (diff.EdDiff, error) {
	type HashPoint struct {
		Hash           uint64
		FromLineNumber int
	}
	// Map from Hash to list of occurrences in 'from'
	m := make(map[uint64][]int, len(from))
	for linePos, line := range from {
		h := hashBytes([]byte(line))
		m[h] = append(m[h], linePos)
	}

	type ContinuousRun struct {
		FromStart int
		ToStart   int
		Length    int
	}

	var runs []ContinuousRun

	// Scan 'to' lines to identify runs
	// We track runs ending at current line.
	// activeRuns maps FromIndex -> RunLength ending at FromIndex
	// When we process 'to' line j, if we match 'from' line i, we look if there was a match at i-1 in activeRuns.
	// This is effectively identifying matching diagonals.
	// activeRuns[i] = length means there is a run of length ending at from[i] and to[j].
	// This uses O(min(N,M)) space if sparse? No, could be O(N).

	// Better approach for finding all runs:
	// Use a map for "current runs": Map[int]int : FromIndex -> StartOfRun
	// Wait, we need to find MAXIMAL runs.

	// Simple greedy identification:
	// For each 'to' index j:
	//   For each 'from' index i matching to[j]:
	//     Extend run if possible.
	// This is O(N*M) worst case.

	// Optimized approach using the Hash Map:
	// Iterate through 'to'.
	// Keep track of active runs.
	// currentRuns: map[int]int  (FromIndex -> Length of run ending at FromIndex)

	currentRuns := make(map[int]int)

	for j, line := range to {
		h := hashBytes([]byte(line))
		nextRuns := make(map[int]int)

		if indices, ok := m[h]; ok {
			for _, i := range indices {
				// Check for string equality to be safe
				if from[i] == to[j] {
					length := 1
					if prevLen, ok := currentRuns[i-1]; ok {
						length = prevLen + 1
					}
					nextRuns[i] = length

					// If a run ended (was not extended), we might have added it already?
					// No, we need to capture ALL maximal runs.
					// A run is maximal if it cannot be extended left or right.
					// We are extending right.
					// We add to 'runs' list only when a run TERMINATES or at the end.
					// But simplified: Just collect all candidate runs, then filter.
					// Or just collect long runs?
					// Let's store every run segment >= 1 for now? No, too many.
					// We only care about the *end* of runs.
					// If nextRuns[i] exists, the run ending at i-1 is extended to i.
					// We implicitly "move" the run.
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
		currentRuns = nextRuns
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
	// User said: "Length of match, 2. Distance from origin."
	// Assuming "Distance from origin" means standard Euclidean or Manhattan distance of the start point (0,0).
	// i.e., prefer matches closer to start? Or matches closer to diagonal?
	// Usually greedy diff prefers matches closer to the start of the file.
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

	// To check conflicts efficiently, we need to ensure new run doesn't overlap with any selected run.
	// Overlap check:
	// A run covers from [f, f+len) and to [t, t+len).
	// A new run [nf, nf+nlen), [nt, nt+nlen) conflicts if ranges overlap OR if relative ordering is violated (crossing).
	// Wait, standard diff allows non-crossing matches.
	// If we pick a run A, then run B is valid only if B is strictly "before" or strictly "after" A in BOTH dimensions.
	// i.e. (B.end <= A.start) OR (B.start >= A.end).
	// AND relative order must be preserved: if B is before A in 'from', it must be before A in 'to'.

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
			// Adds are inserted AFTER currFrom (which is now effectively handled)
			// Wait, adds happen at the point.
			// EdDiff format: Add at line X means insert AFTER line X.
			// If we deleted gapFromLen lines starting at currFrom+1, we are at line currFrom.
			// So insertions should be at currFrom?
			// Or rather:
			// If we replaced (deleted and added), we delete lines [currFrom+1, run.FromStart].
			// And we add lines [currTo, run.ToStart].
			// The Add command should be at `currFrom`.

			// Extract lines to add
			linesToAdd := to[currTo : currTo+gapToLen]

			// Append Add command
			// Note: If we had a deletion, we typically delete first.
			// Add pos is currFrom.
			// BUT if we deleted lines, does the insertion point change?
			// EdDiff uses ORIGINAL line numbers for Delete.
			// Add uses index in the *current* state? No, standard `ed` adds after line N.
			// If we deleted lines currFrom+1..run.FromStart, we are sitting at currFrom.
			// So Add at currFrom is correct.
			// Check grouping: if we have multiple adds at same spot?
			// We handle gap as one block.

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
