# Golang RCS

`go-rcs` is a library for parsing and interacting with [RCS (Revision Control System)](https://en.wikipedia.org/wiki/Revision_Control_System) files in Go. It allows you to read RCS files (typically ending in `,v`), inspect their revision history, and access metadata and content.

This project was created to fill a gap in the Go ecosystem for handling RCS files. It supports parsing headers, revision metadata, and content, and provides utilities for managing revision histories.

## Features

- **Parse RCS Files:** Read RCS files into structured Go objects.
- **Inspect Revisions:** Access revision metadata like author, date, state, and commit messages.
- **Read Content:** Retrieve the log messages and raw text content of revisions.
- **Handle Metadata:** Parse headers, descriptions, locks, strict locking, access lists, symbols, and other RCS metadata.

## Installation

### Library

To install the library for use in your own Go programs:

```bash
go get github.com/arran4/golang-rcs
```

### Tool

To install the `gorcs` tool for use on the command line:

```bash
go install github.com/arran4/golang-rcs/cmd/gorcs@latest
```

Alternatively, you can download the latest binary from the [Releases](https://github.com/arran4/golang-rcs/releases) page. Binaries are available for Windows, macOS (Darwin), and Linux (amd64, arm, arm64, 386). Packages are also available in formats: apk, deb, rpm, termux.deb, and archlinux.

## Usage

Here is a simple example of how to parse an RCS file and list its revision history:

```go
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/arran4/golang-rcs"
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

	fmt.Printf("Head: %s\n", rcsFile.Head)
	fmt.Printf("Description: %s\n", rcsFile.Description)

	// Iterate over revision headers
	for _, rh := range rcsFile.RevisionHeads {
		fmt.Printf("Revision %s\n", rh.Revision)

		// Parse DateTime string to time.Time
		date, err := rh.Date.DateTime()
		if err != nil {
			log.Printf("Error parsing date: %s", err)
		} else {
			fmt.Printf("  Date:   %s\n", date.Format(time.RFC3339))
		}

		fmt.Printf("  Author: %s\n", rh.Author)
		fmt.Printf("  State:  %s\n", rh.State)
	}
}
```

## Modifying RCS Files

You can also modify the parsed structure and serialize it back to an RCS file string.

```go
	// Modify description
	rcsFile.Description = "Updated description via golang-rcs"

	// Add a new lock
	rcsFile.Locks = append(rcsFile.Locks, &rcs.Lock{
		User:     "jules",
		Revision: "1.2",
	})

	// Print back to stdout (or file)
	fmt.Println(rcsFile.String())
```

## Data Structures

The library exposes several key structures that represent the contents of an RCS file.

### `File`
The top-level structure representing a parsed RCS file.

| Field | Type | Description |
| :--- | :--- | :--- |
| `Head` | `string` | The revision number of the head revision. |
| `Branch` | `string` | The default branch (if any). |
| `Access` | `bool` | Whether the access list is present. |
| `AccessUsers` | `[]string` | List of users in the access list. |
| `Symbols` | `[]*Symbol` | List of symbolic names. |
| `Locks` | `[]*Lock` | List of locks held on the file. |
| `Strict` | `bool` | Whether strict locking is enabled. |
| `Integrity` | `string` | Integrity configuration string. |
| `Comment` | `string` | Comment prefix string. |
| `Expand` | `string` | Keyword expansion mode (e.g., `@kv@`). |
| `Description` | `string` | The description of the file. |
| `RevisionHeads` | `[]*RevisionHead` | Metadata for each revision in the file. |
| `RevisionContents` | `[]*RevisionContent` | The actual content (log and text) for each revision. |

### `Lock`
Represents a lock on a revision.

| Field | Type | Description |
| :--- | :--- | :--- |
| `User` | `string` | The user holding the lock. |
| `Revision` | `string` | The revision locked. |
| `Strict` | `bool` | Whether it is a strict lock. |

### `RevisionHead`
Contains metadata about a specific revision.

| Field | Type | Description |
| :--- | :--- | :--- |
| `Revision` | `Num` | The revision number (e.g., "1.1"). |
| `Date` | `DateTime` | The date and time string of the revision. |
| `Author` | `ID` | The username of the author. |
| `State` | `ID` | The state of the revision (e.g., "Exp"). |
| `Branches` | `[]Num` | List of branches starting from this revision. |
| `NextRevision` | `Num` | The revision number of the next revision in the sequence. |
| `CommitID` | `Sym` | The Commit ID of the revision (if present). |

### Custom Types

The library uses several custom types which are underlying `string` types. They are provided to improve code readability and implement the `fmt.Stringer` interface. You can cast them to `string` if needed, or use them directly in contexts that accept `fmt.Stringer` (like `fmt.Printf`).

*   **`Num`** (underlying `string`): Represents a revision number (e.g., "1.1", "1.2.3.4").
*   **`ID`** (underlying `string`): Represents an identifier, such as an author name or state.
*   **`Sym`** (underlying `string`): Represents a symbolic name or commit ID.
*   **`DateTime`** (underlying `string`): Represents a raw RCS date string (e.g., "2022.03.23.13.18.09"). It provides a method `.DateTime()` which returns `(time.Time, error)` to parse the string into a standard Go `time.Time` object.

### `RevisionContent`
Contains the log message and text content for a revision.

| Field | Type | Description |
| :--- | :--- | :--- |
| `Revision` | `string` | The revision number. |
| `Log` | `string` | The commit log message. |
| `Text` | `string` | The raw text content of the revision. |

## Program: gorcs

This repository includes a utility program `gorcs` with subcommands.

### `gorcs branches default set`

Sets the default branch header in one or more RCS files. Provide the branch revision (for example `1.1.1.1`), and `gorcs` stores the default branch form (`1.1.1`).

**Usage:**

```shell
gorcs branches default set <name> [file1 file2 ...]
```

**Example:**

```shell
gorcs branches default set 1.1.1.1 file.txt
```

### `gorcs list-heads`

A simple tool to list revisions in specified RCS files.

**Usage:**

```shell
gorcs list-heads [file1,v file2,v ...]
```

**Example:**

```shell
> gorcs list-heads testinput.go,v
Parsing:  testinput.go,v
1.6 on 2022-03-23 13:22:51 +1100 AEDT by arran
1.5 on 2022-03-23 13:22:34 +1100 AEDT by arran
1.4 on 2022-03-23 13:22:03 +1100 AEDT by arran
1.3 on 2022-03-23 13:21:35 +1100 AEDT by arran
1.2 on 2022-03-23 13:20:39 +1100 AEDT by arran
1.1 on 2022-03-23 13:18:09 +1100 AEDT by arran
```

### `gorcs normalize-revisions`

A utility to align revision numbers across multiple files based on timestamps. This is useful for analyzing related files where revisions might be out of sync numerically but synchronous in time. It sorts revisions by date and renumbers them starting from 1.0 (implicitly, as 1.x) to match across files.

**Usage:**

```shell
gorcs normalize-revisions [-pad-commits] [file1,v file2,v ...]
```

**Flags:**

- `-pad-commits`: If set, when a file is missing a revision at a specific timestamp (which exists in other files), a dummy revision is created to keep the sequence aligned.

**Example Scenario:**

Imagine these two files:

```shell
> gorcs list-heads file1.go,v
Parsing:  file1.go,v file2.go,v
1.2 on 2022-03-23 15:01:01 +1100 AEDT by arran
1.1 on 2022-03-23 13:01:01 +1100 AEDT by arran
> gorcs list-heads file2.go,v
Parsing:  file2.go,v
1.3 on 2022-03-23 15:01:01 +1100 AEDT by arran
1.2 on 2022-03-23 14:01:01 +1100 AEDT by arran
1.1 on 2022-03-23 13:01:01 +1100 AEDT by arran
```

Notice how revision 1.2 in `file1.go,v` occurs at 15:01:01, while revision 1.3 in `file2.go,v` occurs at the same time.

The `gorcs normalize-revisions` program will align the revision numbers to match based on their timestamps:

```shell
> gorcs normalize-revisions file1.go,v file2.go,v
> gorcs list-heads file1.go,v
Parsing:  file1.go,v file2.go,v
1.3 on 2022-03-23 15:01:01 +1100 AEDT by arran
1.1 on 2022-03-23 13:01:01 +1100 AEDT by arran
> gorcs list-heads file2.go,v
Parsing:  file2.go,v
1.3 on 2022-03-23 15:01:01 +1100 AEDT by arran
1.2 on 2022-03-23 14:01:01 +1100 AEDT by arran
1.1 on 2022-03-23 13:01:01 +1100 AEDT by arran
```

### `gorcs to-json`

Parses one or more RCS files and outputs their contents as JSON.

**Usage:**

```shell
gorcs to-json [-o output_file] [-f] [file1,v ...]
```

- **Output:** By default, creates a `.json` file for each input file (e.g., `file.v` -> `file.v.json`).
- `-o`: Specify output file (only valid with a single input file).
- `-f`: Force overwrite if output file exists.
- `-` as input file reads from stdin (outputs to stdout unless `-o` is used).

Example:

```shell
cat file1.go,v | gorcs to-json - > file1.json
```

### `gorcs from-json`

Parses one or more JSON files (generated by `to-json`) and outputs them as RCS files.

**Usage:**

```shell
gorcs from-json [-o output_file] [-f] [file1.json ...]
```

- **Output:** By default, removes the `.json` extension (e.g., `file.v.json` -> `file.v`).
- `-o`: Specify output file (only valid with a single input file).
- `-f`: Force overwrite if output file exists.
- `-` as input file reads from stdin (outputs to stdout unless `-o` is used).

Example:

```shell
cat file1.json | gorcs from-json - > file1.go,v
```

### `gorcs to-markdown`

Parses one or more RCS files and outputs their contents as a simplified Markdown representation.

**Usage:**

```shell
gorcs to-markdown [-o output_file] [-f] [file1,v ...]
```

- **Output:** By default, creates a `.md` file for each input file (e.g., `file.v` -> `file.v.md`).
- `-o`: Specify output file (only valid with a single input file).
- `-f`: Force overwrite if output file exists.
- `-` as input file reads from stdin.

Example:

```shell
cat file1.go,v | gorcs to-markdown - > file1.md
```

### `gorcs from-markdown`

Parses one or more Markdown files (generated by `to-markdown`) and outputs them as RCS files.

**Usage:**

```shell
gorcs from-markdown [-o output_file] [-f] [file1.md ...]
```

- **Output:** By default, removes the `.md` extension (e.g., `file.v.md` -> `file.v`).
- `-o`: Specify output file (only valid with a single input file).
- `-f`: Force overwrite if output file exists.
- `-` as input file reads from stdin.

Example:

```shell
cat file1.md | gorcs from-markdown - > file1.go,v
```

### `gorcs format`

Reads one or more RCS files and outputs them in RCS format. This is useful for normalizing file formatting or verifying parser round-trips.

**Usage:**

```shell
gorcs format [-o output_file] [-w] [-s] [-f] [file1,v ...]
```

- **Output:** By default, outputs to stdout. If multiple files are provided and output is stdout, uses `txtar` format.
- `-o`: Specify output file (only valid with a single input file).
- `-w`, `--overwrite`: Overwrite the input file with the formatted output.
- `-s`, `--stdout`: Force output to stdout (even if other flags might imply otherwise).
- `-f`, `--force`: Force overwrite if output file exists.
- `-` as input file reads from stdin.

### `gorcs validate`

Reads one or more RCS files, parses them, and re-serializes them to ensure validity. Currently functionally identical to `format`.

**Usage:**

```shell
gorcs validate [-o output_file] [-w] [-s] [-f] [file1,v ...]
```

- **Output:** By default, outputs to stdout. If multiple files are provided and output is stdout, uses `txtar` format.
- `-o`: Specify output file (only valid with a single input file).
- `-w`, `--overwrite`: Overwrite the input file.
- `-s`, `--stdout`: Force output to stdout.
- `-f`, `--force`: Force overwrite if output file exists.
- `-` as input file reads from stdin.

### `gorcs co`

Checks out a revision from an RCS file and optionally updates lock state.

```bash
gorcs co [-q] [-rREV | -l[REV] | -u[REV]] [-dDATE] [-zZONE] [-wUSER] [file ...]
```

- `-rREV`: Check out a specific revision.
- `-l[REV]`: Check out and lock the given revision (or head when omitted).
- `-u[REV]`: Check out and unlock the given revision (or head when omitted).
- `-dDATE`: Check out the latest revision on the default branch (or trunk) that is on or before the specified date.
- `-zZONE`: Specify the timezone for the date parsing (e.g., "LT", "UTC", "-0700", "America/New_York"). Defaults to UTC. See [List of tz database time zones](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones) for valid zone names.
- `-wUSER`: User to apply lock changes for (defaults to current logged in user).
- `-q`: Quiet mode.

### `gorcs locks`

Manage locks and strict locking mode on RCS files.

**Usage:**

```shell
gorcs locks <subcommand> [flags] [files...]
```

**Subcommands:**

*   `lock`: Sets a lock for the current user on a specific revision.
*   `unlock`: Removes a lock for the current user on a specific revision.
*   `clean` (or `clear`): Unlocks the revision if the working file is unmodified.
*   `strict`: Enables strict locking mode.
*   `nonstrict`: Disables strict locking mode.

**Flags:**

*   `-revision`: The revision number to operate on (required for `lock` and `unlock`).

**Examples:**

```shell
# Lock revision 1.1
gorcs locks lock -revision 1.1 file.txt

# Unlock revision 1.1
gorcs locks unlock -revision 1.1 file.txt

# Enable strict locking
gorcs locks strict file.txt

# Disable strict locking
gorcs locks nonstrict file.txt

# Unlock if working file is unmodified (like rcsclean -u)
gorcs locks clean file.txt
```

### `gorcs init`

Creates and initializes a new RCS file.

```bash
gorcs init [-t[description]] [file ...]
```

- `-t[description]`: Set the initial description. If the argument starts with `-`, the rest is taken as the description string. Otherwise, it is treated as a file name to read the description from.


**Example:**

```shell
gorcs init -t-"Initial Description" file.txt
```

### `gorcs access-list copy`

Copies the access list from one RCS file to one or more other RCS files. This replaces the access list of the target files with the one from the source file.

**Usage:**

```shell
gorcs access-list copy -from <source_file,v> <target_file1,v> [target_file2,v ...]
```

- `-from`: The source RCS file to read the access list from.

**Example:**

```shell
gorcs access-list copy -from template.txt,v file1.txt,v file2.txt,v
```

### `gorcs access-list append`

Appends the access list from one RCS file to one or more other RCS files, avoiding duplicates.

**Usage:**

```shell
gorcs access-list append -from <source_file,v> <target_file1,v> [target_file2,v ...]
```

- `-from`: The source RCS file to read the access list from.

**Example:**

```shell
gorcs access-list append -from new_users.txt,v file1.txt,v
```

### `gorcs log message change`

Changes the log message for a specific revision in one or more RCS files.

**Usage:**

```shell
gorcs log message change -rev <revision> -m <message> [file1,v ...]
```

- `-rev`: The revision number to update.
- `-m`: The new log message.

**Example:**

```shell
gorcs log message change -rev 1.2 -m "Updated commit message" file.v
```

### `gorcs log message print`

Prints the log message for a specific revision in one or more RCS files.

**Usage:**

```shell
gorcs log message print -rev <revision> [file1,v ...]
```

- `-rev`: The revision number to query.

**Example:**

```shell
gorcs log message print -rev 1.2 file.v
```

### `gorcs log message list`

Lists all log messages for one or more RCS files.

**Usage:**

```shell
gorcs log message list [file1,v ...]
```

**Example:**

```shell
gorcs log message list file.v
```

### `gorcs log`

Prints information about the RCS file, similar to `rlog`. It supports filtering revisions by state and other criteria.

**Usage:**

```shell
gorcs log [-sState] [-F filter_expression] [file1,v ...]
```

- `-s`: Filter by state. Comma-separated states are supported (e.g., `-sRel,Prod`).
- `-F`, `--filter`: Filter by expression. Supported syntax:
  - `state=<value>` or `s=<value>`
  - `state in (<value1> <value2> ...)`
  - `expression OR expression` or `expression || expression`

**Example:**

```shell
gorcs log -sRel file.v
gorcs log --filter "state=Exp || state=Prod" file.v
```

## License

MIT.
