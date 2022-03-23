package main

import (
	"flag"
	"fmt"
	rcs "github.com/arran4/golang-rcs"
	"log"
	"os"
	"time"
)

func init() {
	log.SetFlags(log.Flags() | log.Lshortfile)
	flag.Parse()
}

func main() {
	for _, f := range flag.Args() {
		ListHeads(f)
	}
}

func ListHeads(fn string) {
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
	for _, rh := range r.RevisionHeads {
		fmt.Printf("%s on %s by %s\n", rh.Revision, rh.Date.In(time.Local), rh.Author)
	}
}
