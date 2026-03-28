package znkr_diff

import (
	rcsdiff "github.com/arran4/golang-rcs/diff"
	znkrdiff "znkr.io/diff"
)

func init() {
	rcsdiff.Register("znkr", GenerateEdDiffFromLines)
}

func GenerateEdDiffFromLines(from []string, to []string) (rcsdiff.EdDiff, error) {
	edits := znkrdiff.Edits(from, to)

	var res rcsdiff.EdDiff

	delStart := 0
	delCount := 0

	addStart := 0
	var addLines []string

	inDel := false
	inAdd := false

	commitDel := func() {
		if inDel {
			res = append(res, rcsdiff.Delete{delStart, delCount})
			inDel = false
			delCount = 0
		}
	}

	commitAdd := func() {
		if inAdd {
			res = append(res, rcsdiff.Add{LineStart: addStart, Lines: addLines})
			inAdd = false
			addLines = nil
		}
	}

	currFromPos := 0

	for _, edit := range edits {
		switch edit.Op {
		case znkrdiff.Delete:
			if !inDel {
				delStart = currFromPos + 1
				inDel = true
			}
			delCount++
			currFromPos++

		case znkrdiff.Insert:
			if !inAdd {
				addStart = currFromPos
				inAdd = true
			}
			addLines = append(addLines, edit.Y)

		case znkrdiff.Match:
			// Order is important for test assertions (Add then Delete)
			// But for znkr, it doesn't matter since Apply will just take them.
			commitAdd()
			commitDel()
			currFromPos++
		}
	}

	// Order of commits at end
	commitAdd()
	commitDel()

	return res, nil
}
