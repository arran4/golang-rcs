package cli

import (
	"fmt"
	rcs "github.com/arran4/golang-rcs"
	"os"
	"sort"
	"time"
)

type DateSorter struct {
	dates []time.Time
}

func (d DateSorter) Len() int {
	return len(d.dates)
}

func (d DateSorter) Less(i, j int) bool {
	return d.dates[i].Unix() < d.dates[j].Unix()
}

func (d DateSorter) Swap(i, j int) {
	d.dates[j], d.dates[i] = d.dates[i], d.dates[j]
}

// NormalizeRevisions is a subcommand `gorcs normalize-revisions`
//
// Flags:
//
//	padCommits: -p --pad-commits pad commits with empty commits
//	files: ... List of files to process
func NormalizeRevisions(padCommits bool, files ...string) error {
	type Pair struct {
		Rcs *rcs.File
		FN  string
	}
	var rs []Pair
	datesSet := map[time.Time]struct{}{}
	for _, f := range files {
		r, err := ReadParse(f)
		if err != nil {
			return err
		}
		rs = append(rs, Pair{
			Rcs: r,
			FN:  f,
		})
		for _, rh := range r.RevisionHeads {
			dt, _ := rh.Date.DateTime()
			fmt.Printf("%s on %s by %s\n", rh.Revision, dt.In(time.Local), rh.Author)
			datesSet[dt] = struct{}{}
		}
	}
	var dates []time.Time
	for d := range datesSet {
		dates = append(dates, d)
	}
	sort.Sort(DateSorter{dates})
	dateToRevision := map[time.Time]string{}
	for i, d := range dates {
		r := fmt.Sprintf("1.%d", i)
		dateToRevision[d] = r
		fmt.Println("Updating date ", d.Format(rcs.DateFormat), " to revision: ", r)
	}
	for _, r := range rs {
		fmt.Println("File", r.FN)
		revisionToRevision := map[string]string{}
		for _, rh := range r.Rcs.RevisionHeads {
			dt, _ := rh.Date.DateTime()
			s := dateToRevision[dt]
			revisionToRevision[rh.Revision.String()] = s
			fmt.Println("Updating date ", dt.Format(rcs.DateFormat), " to revision: ", s, "from", rh.Revision)
			rh.Revision = rcs.Num(s)
		}

		type hc struct {
			h *rcs.RevisionHead
			c *rcs.RevisionContent
		}

		for _, rh := range r.Rcs.RevisionHeads {
			rh.NextRevision = rcs.Num(revisionToRevision[rh.NextRevision.String()])
		}

		byDate := map[time.Time]hc{}
		for i, h := range r.Rcs.RevisionHeads {
			if i >= len(r.Rcs.RevisionContents) {
				return fmt.Errorf("file %s has mismatching heads (%d) and contents (%d)", r.FN, len(r.Rcs.RevisionHeads), len(r.Rcs.RevisionContents))
			}
			dt, _ := h.Date.DateTime()
			byDate[dt] = hc{h: h, c: r.Rcs.RevisionContents[i]}
		}
		for _, rc := range r.Rcs.RevisionContents {
			rc.Revision = revisionToRevision[rc.Revision]
		}

		var newHeads []*rcs.RevisionHead
		var newContents []*rcs.RevisionContent

		for i := len(dates) - 1; i >= 0; i-- {
			d := dates[i]
			pair, ok := byDate[d]
			if ok {
				pair.h.Revision = rcs.Num(dateToRevision[d])
				pair.c.Revision = dateToRevision[d]
				newHeads = append(newHeads, pair.h)
				newContents = append(newContents, pair.c)
			} else if padCommits {
				dtStr := d.Format(rcs.DateFormat)
				h := &rcs.RevisionHead{Revision: rcs.Num(dateToRevision[d]), Date: rcs.DateTime(dtStr)}
				c := &rcs.RevisionContent{Revision: dateToRevision[d]}
				newHeads = append(newHeads, h)
				newContents = append(newContents, c)
			}
		}

		for i := 0; i < len(newHeads); i++ {
			if i+1 < len(newHeads) {
				newHeads[i].NextRevision = newHeads[i+1].Revision
			} else {
				newHeads[i].NextRevision = ""
			}
		}

		if len(newHeads) > 0 {
			r.Rcs.Head = newHeads[0].Revision.String()
		}
		r.Rcs.RevisionHeads = newHeads
		r.Rcs.RevisionContents = newContents
	}
	for _, r := range rs {
		if err := WriteFile(r.FN, r.Rcs); err != nil {
			return err
		}
	}
	return nil
}

func WriteFile(fn string, file *rcs.File) error {
	fmt.Println("Saving: ", fn)
	if err := os.WriteFile(fn, []byte(file.String()), 0644); err != nil {
		return fmt.Errorf("error saving file: %s: %w", fn, err)
	}
	return nil
}

func ReadParse(fn string) (*rcs.File, error) {
	f, err := os.Open(fn)
	if err != nil {
		return nil, fmt.Errorf("error with file %s: %w", fn, err)
	}
	defer func() {
		_ = f.Close()
	}()
	fmt.Println("Parsing: ", fn)
	r, err := rcs.ParseFile(f)
	if err != nil {
		return nil, fmt.Errorf("error parsing: %w", err)
	}
	return r, nil
}
