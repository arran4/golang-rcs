package hashline

import (
	"hash/fnv"

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
	// This is a simplified "hashline" implementation based on the user's previous skeleton.
	// Since I don't have the original code, I will implement a basic diff algorithm that uses line hashes.
	// A simple approach is:
	// 1. Hash all lines.
	// 2. Find longest common subsequence of HASHES (faster comparisons).
	// 3. Reconstruct diff from hash LCS.
	// This is effectively the same as LCS but operating on uint64 hashes instead of string comparisons.
	// This satisfies the "hashline" name and potential performance benefit.

	fromHashes := make([]uint64, len(from))
	for i, line := range from {
		fromHashes[i] = hashBytes([]byte(line))
	}

	toHashes := make([]uint64, len(to))
	for i, line := range to {
		toHashes[i] = hashBytes([]byte(line))
	}

	m := len(from)
	n := len(to)
	lcs := make([][]int, m+1)
	for i := range lcs {
		lcs[i] = make([]int, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if fromHashes[i-1] == toHashes[j-1] && from[i-1] == to[j-1] {
				// Verify collision if needed, or assume strong hash.
				// For correctness, we should compare strings if hashes match.
				// But given the name "hashline", maybe it relies on hashes?
				// Let's verify string equality to be safe.
				lcs[i][j] = lcs[i-1][j-1] + 1
			} else {
				if lcs[i-1][j] >= lcs[i][j-1] {
					lcs[i][j] = lcs[i-1][j]
				} else {
					lcs[i][j] = lcs[i][j-1]
				}
			}
		}
	}

	type actionType int
	const (
		actMatch actionType = iota
		actAdd
		actDelete
	)

	type action struct {
		kind actionType
		text string
	}

	var actions []action
	i, j := m, n
	for i > 0 || j > 0 {
		if i > 0 && j > 0 && fromHashes[i-1] == toHashes[j-1] && from[i-1] == to[j-1] {
			actions = append(actions, action{kind: actMatch})
			i--
			j--
		} else if i > 0 && (j == 0 || lcs[i-1][j] >= lcs[i][j-1]) {
			actions = append(actions, action{kind: actDelete})
			i--
		} else {
			actions = append(actions, action{kind: actAdd, text: to[j-1]})
			j--
		}
	}

	for k := 0; k < len(actions)/2; k++ {
		actions[k], actions[len(actions)-1-k] = actions[len(actions)-1-k], actions[k]
	}

	var result diff.EdDiff
	currentLine := 0

	for k := 0; k < len(actions); k++ {
		a := actions[k]
		switch a.kind {
		case actMatch:
			currentLine++
		case actDelete:
			if len(result) > 0 {
				if del, ok := result[len(result)-1].(diff.Delete); ok {
					if del[0]+del[1] == currentLine+1 {
						result[len(result)-1] = diff.Delete{del[0], del[1] + 1}
						currentLine++
						continue
					}
				}
			}
			result = append(result, diff.Delete{currentLine + 1, 1})
			currentLine++

		case actAdd:
			if len(result) > 0 {
				if add, ok := result[len(result)-1].(diff.Add); ok {
					if add.LineStart == currentLine {
						add.Lines = append(add.Lines, a.text)
						result[len(result)-1] = add
						continue
					}
				}
			}
			result = append(result, diff.Add{LineStart: currentLine, Lines: []string{a.text}})
		}
	}

	return result, nil
}
