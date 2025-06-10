package main

import (
	"flag"
	"fmt"
	rcs "github.com/arran4/golang-rcs"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"time"
)

func init() {
	log.SetFlags(log.Flags() | log.Lshortfile)
}

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

func main() {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	padCommits := fs.Bool("pad-commits", false, "pad commits with empty commits")
	fs.Parse(os.Args[1:])

	type Pair struct {
		Rcs *rcs.File
		FN  string
	}
	var rs []Pair
	datesSet := map[time.Time]struct{}{}
	for _, f := range fs.Args() {
		r := ReadParse(f)
		rs = append(rs, Pair{
			Rcs: r,
			FN:  f,
		})
		for _, rh := range r.RevisionHeads {
			fmt.Printf("%s on %s by %s\n", rh.Revision, rh.Date.In(time.Local), rh.Author)
			datesSet[rh.Date] = struct{}{}
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
		type hc struct {
			h *rcs.RevisionHead
			c *rcs.RevisionContent
		}
		byDate := map[time.Time]hc{}
		for i, h := range r.Rcs.RevisionHeads {
			byDate[h.Date] = hc{h: h, c: r.Rcs.RevisionContents[i]}
		}

		var newHeads []*rcs.RevisionHead
		var newContents []*rcs.RevisionContent

		for i := len(dates) - 1; i >= 0; i-- {
			d := dates[i]
			pair, ok := byDate[d]
			if ok {
				pair.h.Revision = dateToRevision[d]
				pair.c.Revision = dateToRevision[d]
				newHeads = append(newHeads, pair.h)
				newContents = append(newContents, pair.c)
			} else if *padCommits {
				h := &rcs.RevisionHead{Revision: dateToRevision[d], Date: d}
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
			r.Rcs.Head = newHeads[0].Revision
		}
		r.Rcs.RevisionHeads = newHeads
		r.Rcs.RevisionContents = newContents

	}
	for _, r := range rs {
		WriteFile(r.FN, r.Rcs)
	}
}

func WriteFile(fn string, file *rcs.File) {
	fmt.Println("Saving: ", fn)
	if err := ioutil.WriteFile(fn, []byte(file.String()), 0644); err != nil {
		log.Panicf("Error saving file: %s: %s", fn, err)
	}
}

func ReadParse(fn string) *rcs.File {
	f, err := os.Open(fn)
	if err != nil {
		log.Panicf("Error with file %s: %s", fn, err)
	}
	defer func() {
		if err = f.Close(); err != nil {
			log.Panicf("Error closing file; %s: %s", fn, err)
		}
	}()
	fmt.Println("Parsing: ", fn)
	r, err := rcs.ParseFile(f)
	if err != nil {
		log.Panicf("Error parsing: %s", err)
	}
	return r
}
