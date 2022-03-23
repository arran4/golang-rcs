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
	flag.Parse()
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
	type Pair struct {
		Rcs *rcs.File
		FN  string
	}
	var rs []Pair
	datesSet := map[time.Time]struct{}{}
	for _, f := range flag.Args() {
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
		dateToRevision[d] = fmt.Sprintf("1.%d", i)
	}
	for _, r := range rs {
		revisionToRevision := map[string]string{}
		for _, rh := range r.Rcs.RevisionHeads {
			s := dateToRevision[rh.Date]
			revisionToRevision[rh.Revision] = s
			rh.Revision = s
		}
		for _, rh := range r.Rcs.RevisionHeads {
			rh.NextRevision = revisionToRevision[rh.NextRevision]
		}
		for _, rc := range r.Rcs.RevisionContents {
			rc.Revision = revisionToRevision[rc.Revision]
		}
	}
	for _, r := range rs {
		WriteFile(r.FN, r.Rcs)
	}
}

func WriteFile(fn string, file *rcs.File) {
	if err := ioutil.WriteFile(fn, []byte(file.String()), 0644); err != nil {
		log.Panicf("Error saving file: %s: %s", fn, err)
	}
}

func ReadParse(fn string) *rcs.File {
	f, err := os.Open(fn)
	if err != nil {
		log.Panicf("Error with file: %s", err)
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
