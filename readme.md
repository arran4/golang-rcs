# Golang RCS

`go-rcs` is a library for parsing and interacting with [RCS (Revision Control System)](https://en.wikipedia.org/wiki/Revision_Control_System) files in Go. It allows you to read RCS files (typically ending in `,v`), inspect their revision history, and access metadata and content.

This project was created to fill a gap in the Go ecosystem for handling RCS files. While it currently supports parsing and basic interaction, it is a work in progress, and contributions are welcome.

## Features

- **Parse RCS Files:** Read RCS files into structured Go objects.
- **Inspect Revisions:** Access revision metadata like author, date, state, and commit messages.
- **Read Content:** Retrieve the log messages and raw text content of revisions.
- **Handle Metadata:** Parse headers, descriptions, locks, and other RCS metadata.

## Installation

```bash
go get github.com/yourusername/go-rcs
```
*(Replace `github.com/yourusername/go-rcs` with the actual import path)*

## Usage

Here is a simple example of how to parse an RCS file and list its revision history:

```go
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/yourusername/go-rcs" // Replace with actual import path
)

func main() {
	fileName := "example.go,v"
	f, err := os.Open(fileName)
	if err != nil {
		log.Panicf("Error opening file: %s", err)
	}
	defer f.Close()

	// Parse the RCS file
	rcsFile, err := rcs.ParseFile(f)
	if err != nil {
		log.Panicf("Error parsing RCS file: %s", err)
	}

	fmt.Printf("File Head: %s\n", rcsFile.Head)
	fmt.Printf("Description: %s\n", rcsFile.Description)

	// Iterate over revision headers
	for _, rh := range rcsFile.RevisionHeads {
		fmt.Printf("Revision %s\n", rh.Revision)
		fmt.Printf("  Date:   %s\n", rh.Date.Format(time.RFC3339))
		fmt.Printf("  Author: %s\n", rh.Author)
		fmt.Printf("  State:  %s\n", rh.State)
	}
}
```

## Data Structures

The library exposes several key structures that represent the contents of an RCS file.

### `File`
The top-level structure representing a parsed RCS file.

| Field | Type | Description |
| :--- | :--- | :--- |
| `Head` | `string` | The revision number of the head revision. |
| `Branch` | `string` | The default branch (if any). |
| `Description` | `string` | The description of the file. |
| `Locks` | `[]*Lock` | List of locks held on the file. |
| `Strict` | `bool` | Whether strict locking is enabled. |
| `RevisionHeads` | `[]*RevisionHead` | Metadata for each revision in the file. |
| `RevisionContents` | `[]*RevisionContent` | The actual content (log and text) for each revision. |

### `RevisionHead`
Contains metadata about a specific revision.

| Field | Type | Description |
| :--- | :--- | :--- |
| `Revision` | `string` | The revision number (e.g., "1.1"). |
| `Date` | `time.Time` | The date and time of the revision. |
| `Author` | `string` | The username of the author. |
| `State` | `string` | The state of the revision (e.g., "Exp"). |
| `Branches` | `[]string` | List of branches starting from this revision. |
| `NextRevision` | `string` | The revision number of the next revision in the sequence. |

### `RevisionContent`
Contains the log message and text content for a revision.

| Field | Type | Description |
| :--- | :--- | :--- |
| `Revision` | `string` | The revision number. |
| `Log` | `string` | The commit log message. |
| `Text` | `string` | The raw text content of the revision. |

## Programs

This repository includes two utility programs.

### `list-heads`

A simple tool to list revisions in specified RCS files.

**Usage:**

```shell
> list-heads testinput.go,v
Parsing:  testinput.go,v
1.6 on 2022-03-23 13:22:51 +1100 AEDT by arran
1.5 on 2022-03-23 13:22:34 +1100 AEDT by arran
1.4 on 2022-03-23 13:22:03 +1100 AEDT by arran
1.3 on 2022-03-23 13:21:35 +1100 AEDT by arran
1.2 on 2022-03-23 13:20:39 +1100 AEDT by arran
1.1 on 2022-03-23 13:18:09 +1100 AEDT by arran
```

### `normalize-revisions`

A utility to align revision numbers across multiple files based on timestamps. This is useful for analyzing related files where revisions might be out of sync numerically but synchronous in time.

**Example Scenario:**

Imagine these two files:

```shell
> list-heads file1.go,v
Parsing:  file1.go,v file2.go,v
1.2 on 2022-03-23 15:01:01 +1100 AEDT by arran
1.1 on 2022-03-23 13:01:01 +1100 AEDT by arran
> list-heads file2.go,v
Parsing:  file2.go,v
1.3 on 2022-03-23 15:01:01 +1100 AEDT by arran
1.2 on 2022-03-23 14:01:01 +1100 AEDT by arran
1.1 on 2022-03-23 13:01:01 +1100 AEDT by arran
```

Notice how revision 1.2 in `file1.go,v` occurs at 15:01:01, while revision 1.3 in `file2.go,v` occurs at the same time.

The `normalize-revisions` program will align the revision numbers to match based on their timestamps:

```shell
> normalize-revisions file1.go,v file2.go,v
> list-heads file1.go,v
Parsing:  file1.go,v file2.go,v
1.3 on 2022-03-23 15:01:01 +1100 AEDT by arran
1.1 on 2022-03-23 13:01:01 +1100 AEDT by arran
> list-heads file2.go,v
Parsing:  file2.go,v
1.3 on 2022-03-23 15:01:01 +1100 AEDT by arran
1.2 on 2022-03-23 14:01:01 +1100 AEDT by arran
1.1 on 2022-03-23 13:01:01 +1100 AEDT by arran
```

## License

MIT.
