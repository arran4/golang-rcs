package rcs_test

import (
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/arran4/golang-rcs"
)

func TestReadmeExamples(t *testing.T) {
	rcsContent := `head     1.1;
branch   ;
access   ;
symbols  ;
locks    ; strict;
comment  @# @;


1.1
date     2022.03.23.13.18.09;  author arran;  state Exp;
branches ;
next     ;


desc
@Initial description@


1.1
log
@Initial revision
@
text
@package main
@`

	// Example 1: Parsing and Reading
	f := strings.NewReader(rcsContent)
	rcsFile, err := rcs.ParseFile(f)
	if err != nil {
		t.Fatalf("Error parsing RCS file: %s", err)
	}

	fmt.Printf("Head: %s\n", rcsFile.Head)
	fmt.Printf("Description: %s\n", rcsFile.Description)

	for _, rh := range rcsFile.RevisionHeads {
		fmt.Printf("Revision %s\n", rh.Revision)
		// Correct way to parse date
		date, err := rh.Date.DateTime()
		if err != nil {
			log.Printf("Error parsing date: %s", err)
		} else {
			fmt.Printf("  Date:   %s\n", date.Format(time.RFC3339))
		}
		fmt.Printf("  Author: %s\n", rh.Author)
		fmt.Printf("  State:  %s\n", rh.State)
	}

	// Example 2: Modifying
	rcsFile.Description = "Updated description via golang-rcs"

	rcsFile.Locks = append(rcsFile.Locks, &rcs.Lock{
		User:     "jules",
		Revision: "1.2",
	})

	output := rcsFile.String()
	if !strings.Contains(output, "Updated description") {
		t.Errorf("Description not updated")
	}
	if !strings.Contains(output, "jules:1.2") {
		t.Errorf("Lock not added")
	}
}
