# Golang RCS

Is a serializer and deserializer library for the RCS version control system https://en.wikipedia.org/wiki/Revision_Control_System

I couldn't find anything which would do what I wanted. Currently the code is usecase specific, but I'm happy
to accept PRs.

Note; there is a lot of work that needs to be done here to get it to the point it can be considered anything close to generally applicable. -- I will accept most PRs at the stage.

# Usage

The library is simple to use:
```go
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
```

See the godoc for the structure.

# Programs

I have produced 2 programs as part of this

## List-heads

It's a simple program that lists the revisions etc, in the files specified:

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

There are currently no arguments

## normalize-revisions

The purpose of this program is that it lines up all the date times with revisions in multiple files. Ie;

Imagine these two files:

```
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
Notice how revision 1.2 occurs at 14:00 while, revision 1.3 in file2 and 1.2 in file2 occur at the same time.

The idea is this program will align the revision numbers to match: 
```
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

There is no real purpose or reason for doing this.

# License 

MIT. 
