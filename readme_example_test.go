package rcs_test

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/arran4/golang-rcs"
)

// NOTE TO DEVS/AGENTS:
// This file contains examples that are used in the README.md.
// If you modify this file, PLEASE UPDATE README.md accordingly.
// The code in the README should be kept in sync with this file to ensure accuracy.

func Example_readme() {
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
		log.Fatalf("Error parsing RCS file: %s", err)
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
	if strings.Contains(output, "Updated description") {
		fmt.Println("Description updated successfully")
	}
	if strings.Contains(output, "jules:1.2") {
		fmt.Println("Lock added successfully")
	}

	// Output:
	// Head: 1.1
	// Description: Initial description
	// Revision 1.1
	//   Date:   2022-03-23T13:18:09Z
	//   Author: arran
	//   State:  Exp
	// Description updated successfully
	// Lock added successfully
}
