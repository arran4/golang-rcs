package diff

func init() {
	Register("lcs", GenerateEdDiffFromLines)
}

func GenerateEdDiffFromLines(from []string, to []string) (EdDiff, error) {
	m := len(from)
	n := len(to)
	lcs := make([][]int, m+1)
	for i := range lcs {
		lcs[i] = make([]int, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if from[i-1] == to[j-1] {
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
		text string // for add
		// for delete, we don't strictly need text but useful for debug
		// for match, we don't need anything
	}

	var actions []action
	i, j := m, n
	for i > 0 || j > 0 {
		if i > 0 && j > 0 && from[i-1] == to[j-1] {
			actions = append(actions, action{kind: actMatch})
			i--
			j--
		} else if i > 0 && (j == 0 || lcs[i-1][j] >= lcs[i][j-1]) {
			// Prefer Delete (move i)
			actions = append(actions, action{kind: actDelete})
			i--
		} else {
			// Prefer Add (move j)
			actions = append(actions, action{kind: actAdd, text: to[j-1]})
			j--
		}
	}

	// Reverse actions to get forward order
	for k := 0; k < len(actions)/2; k++ {
		actions[k], actions[len(actions)-1-k] = actions[len(actions)-1-k], actions[k]
	}

	var result EdDiff
	currentLine := 0 // 0-based index of original file processed

	for k := 0; k < len(actions); k++ {
		a := actions[k]
		switch a.kind {
		case actMatch:
			currentLine++
		case actDelete:
			// Check if we can extend previous delete
			if len(result) > 0 {
				if del, ok := result[len(result)-1].(Delete); ok {
					// Check if this delete is contiguous
					// del[0] is start line (1-based), del[1] is count
					// The range deleted is [del[0], del[0] + del[1] - 1]
					// Next line to be deleted would be del[0] + del[1]
					// currentLine is 0-based index of line being processed.
					// Since we are at `actDelete`, we are about to delete `currentLine` (0-based) => `currentLine+1` (1-based).
					// So if `del[0] + del[1] == currentLine + 1`, it's contiguous.

					if del[0]+del[1] == currentLine+1 {
						// Extend
						result[len(result)-1] = Delete{del[0], del[1] + 1}
						currentLine++
						continue
					}
				}
			}

			result = append(result, Delete{currentLine + 1, 1})
			currentLine++

		case actAdd:
			// Check if we can extend previous add
			if len(result) > 0 {
				if add, ok := result[len(result)-1].(Add); ok {
					// Check if this add is at same position
					// add.LineStart is the line number *after* which we insert.
					// If we are still at the same insertion point (currentLine), extend.
					if add.LineStart == currentLine {
						// Extend
						add.Lines = append(add.Lines, a.text)
						result[len(result)-1] = add
						continue
					}
				}
			}

			result = append(result, Add{LineStart: currentLine, Lines: []string{a.text}})
			// Do not increment currentLine
		}
	}

	return result, nil
}
