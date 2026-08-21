package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	rcs "github.com/arran4/golang-rcs"
)

type CIVerdict struct {
	File            string
	RCSFile         string
	Revision        string
	FileModified    bool
	LockSet         bool
	LockCleared     bool
	Created         bool
	AppendedNewline bool
}

// CiDisposition controls what happens to the working file after checkin.
type CiDisposition int

const (
	CiRemove CiDisposition = iota // default: remove working file
	CiLock                        // -l: keep writable (locked)
	CiUnlock                      // -u: keep read-only (unlocked)
)

// Ci performs checkin operations over one or more working files.
//
// Flags:
//
//	revision: -r explicit revision number
//	disposition: CiRemove (default), CiLock (-l), CiUnlock (-u)
//	user: -w override author
//	quiet: -q suppress status output
//	message: -m log message
//	force: -f force even if unchanged
//	state: -s set revision state
//	date: -d override checkin date
//	zone: -z zone for date parsing
//	files: ... List of working files to process
func Ci(revision string, disposition CiDisposition, user string, quiet bool,
	message string, force bool, state string, ciDate, ciZone string,
	files ...string) error {

	if user == "" {
		user = currentLoggedInUser()
	}
	if len(files) == 0 {
		return fmt.Errorf("no files provided")
	}
	for _, file := range files {
		result, err := ciFile(revision, disposition, user, quiet, message, force, state, ciDate, ciZone, file)
		if err != nil {
			return err
		}
		if result.AppendedNewline {
			fmt.Fprintf(os.Stderr, "gorcs: %s: warning: file did not end with a newline; one was appended\n",
				filepath.Base(result.File))
		}
		if !quiet {
			if result.Created {
				fmt.Printf("%s  <--  %s\ninitial revision: %s\n",
					filepath.Base(result.RCSFile),
					filepath.Base(result.File),
					result.Revision,
				)
			} else {
				fmt.Printf("%s  <--  %s\nnew revision: %s\n",
					filepath.Base(result.RCSFile),
					filepath.Base(result.File),
					result.Revision,
				)
			}
			fmt.Printf("done\n")
		}
	}
	return nil
}

func ciFile(revision string, disposition CiDisposition, user string, quiet bool,
	message string, force bool, state string, ciDate, ciZone string,
	workingFile string) (CIVerdict, error) {

	// Determine RCS archive path.
	rcsFile := workingFile
	if strings.HasSuffix(rcsFile, ",v") {
		// Argument is the archive — derive working file.
		workingFile = strings.TrimSuffix(rcsFile, ",v")
	} else {
		rcsFile += ",v"
	}

	// Read working file content.
	content, err := os.ReadFile(workingFile)
	if err != nil {
		return CIVerdict{}, fmt.Errorf("read %s: %w", workingFile, err)
	}
	text := string(content)

	// Capture the working file's mode so the checkin can preserve its base
	// permissions (execute/group/other bits) and toggle only the write bits,
	// instead of forcing 0644/0444.
	workingPerm := os.FileMode(0644)
	if wfInfo, statErr := os.Stat(workingFile); statErr == nil {
		workingPerm = wfInfo.Mode().Perm()
	}

	// Parse existing archive, or create new file.
	var parsed *rcs.File
	initial := false
	archivePerm := os.FileMode(0)
	haveArchivePerm := false
	if info, err := os.Stat(rcsFile); err == nil {
		archivePerm = info.Mode().Perm()
		haveArchivePerm = true
		f, err := os.Open(rcsFile)
		if err != nil {
			return CIVerdict{}, fmt.Errorf("open %s: %w", rcsFile, err)
		}
		parsed, err = rcs.ParseFile(f)
		_ = f.Close()
		if err != nil {
			return CIVerdict{}, fmt.Errorf("parse %s: %w", rcsFile, err)
		}
	} else if os.IsNotExist(err) {
		parsed = rcs.NewFile()
		initial = true
	} else {
		return CIVerdict{}, fmt.Errorf("stat %s: %w", rcsFile, err)
	}

	// Build checkin options.
	ops := make([]any, 0, 6)
	if revision != "" {
		ops = append(ops, rcs.WithRevision(revision))
	}
	if ciDate != "" {
		zone, err := rcs.ParseZone(ciZone)
		if err != nil {
			return CIVerdict{}, fmt.Errorf("invalid zone %q: %w", ciZone, err)
		}
		t, err := rcs.ParseDate(ciDate, time.Now(), zone)
		if err != nil {
			return CIVerdict{}, fmt.Errorf("invalid date %q: %w", ciDate, err)
		}
		ops = append(ops, rcs.WithDate(t))
	}
	if force {
		ops = append(ops, rcs.WithForce{})
	}
	if state != "" {
		ops = append(ops, rcs.WithState(state))
	}
	if initial {
		ops = append(ops, rcs.WithInitial{})
	}
	switch disposition {
	case CiLock:
		ops = append(ops, rcs.WithSetLock)
	case CiUnlock:
		ops = append(ops, rcs.WithClearLock)
	}

	verdict, err := parsed.Checkin(user, message, text, ops...)
	if err != nil {
		return CIVerdict{}, fmt.Errorf("ci %s: %w", workingFile, err)
	}

	// Write updated archive, preserving base permissions. RCS archives are
	// read-only; derive the base mode from an existing archive, or from the
	// working file for a new one, and clear the write bits. Temporarily add the
	// owner-write bit so an existing read-only archive can be replaced.
	basePerm := workingPerm
	if haveArchivePerm {
		basePerm = archivePerm
	}
	archiveRO := basePerm &^ 0222
	_ = os.Chmod(rcsFile, archiveRO|0200)
	if err := os.WriteFile(rcsFile, []byte(parsed.String()), archiveRO|0200); err != nil {
		return CIVerdict{}, fmt.Errorf("write %s: %w", rcsFile, err)
	}
	if err := os.Chmod(rcsFile, archiveRO); err != nil {
		return CIVerdict{}, fmt.Errorf("chmod %s: %w", rcsFile, err)
	}

	// Handle working file disposition, preserving the working file's base mode
	// bits and toggling only the write bits.
	switch disposition {
	case CiLock:
		// Keep writable.
		if err := os.Chmod(workingFile, workingPerm|0200); err != nil {
			return CIVerdict{}, fmt.Errorf("chmod %s: %w", workingFile, err)
		}
	case CiUnlock:
		// Keep read-only.
		if err := os.Chmod(workingFile, workingPerm&^0222); err != nil {
			return CIVerdict{}, fmt.Errorf("chmod %s: %w", workingFile, err)
		}
	default:
		// Remove working file.
		if err := os.Remove(workingFile); err != nil {
			return CIVerdict{}, fmt.Errorf("remove %s: %w", workingFile, err)
		}
	}

	return CIVerdict{
		File:            workingFile,
		RCSFile:         rcsFile,
		Revision:        verdict.Revision,
		FileModified:    verdict.FileModified,
		LockSet:         verdict.LockSet,
		LockCleared:     verdict.LockCleared,
		Created:         verdict.Created,
		AppendedNewline: verdict.AppendedNewline,
	}, nil
}
