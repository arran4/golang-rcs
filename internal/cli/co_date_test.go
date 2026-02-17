package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCoDate(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gorcs_test_co_date")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	rcsFile := filepath.Join(tmpDir, "input.txt,v")
	rcsContent := `head	1.3;
access;
symbols;
locks; strict;
comment	@# @;


1.3
date	2023.01.03.12.00.00;	author jules;	state Exp;
branches;
next	1.2;

1.2
date	2023.01.02.12.00.00;	author jules;	state Exp;
branches;
next	1.1;

1.1
date	2023.01.01.12.00.00;	author jules;	state Exp;
branches;
next	;


desc
@@


1.3
log
@Rev 1.3
@
text
@Content 1.3
@


1.2
log
@Rev 1.2
@
text
@d1 1
a1 1
Content 1.2
@


1.1
log
@Rev 1.1
@
text
@d1 1
a1 1
Content 1.1
@
`
	if err := os.WriteFile(rcsFile, []byte(rcsContent), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		args     []string // revision, date, zone
		expected string
		wantErr  bool
	}{
		{
			name:     "Date between 1.2 and 1.3",
			args:     []string{"", "2023-01-02 13:00:00", ""},
			expected: "Content 1.2\n",
		},
		{
			name:     "Date after 1.3",
			args:     []string{"", "2023-01-04 12:00:00", ""},
			expected: "Content 1.3\n",
		},
		{
			name:     "Date exactly on 1.2 (UTC)",
			args:     []string{"", "2023-01-02 12:00:00", ""},
			expected: "Content 1.2\n",
		},
		{
			name:     "Date before 1.2 but after 1.1",
			args:     []string{"", "2023-01-02 11:00:00", ""},
			expected: "Content 1.1\n",
		},
		{
			name:    "Date before 1.1 (should fail)",
			args:    []string{"", "2020-01-01", ""},
			wantErr: true,
		},
		{
			name: "Date with timezone (PST -0800)",
			// 2023-01-01 21:00 PST = 2023-01-02 05:00 UTC.
			// 1.2 is 12:00 UTC.
			// Target (05:00) < 1.2 (12:00). So verify 1.1.
			// 1.1 is 2023-01-01 12:00 UTC.
			// Target (05:00 Jan 2) > 1.1 (12:00 Jan 1).
			// So expect 1.1.
			args:     []string{"", "2023-01-01 21:00:00", "-0800"},
			expected: "Content 1.1\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workingFile := filepath.Join(tmpDir, "input.txt")
			// Clean working file
			if err := os.Remove(workingFile); err != nil && !os.IsNotExist(err) {
				t.Fatalf("failed to remove working file: %v", err)
			}

			rev, date, zone := tt.args[0], tt.args[1], tt.args[2]
			err := Co(rev, date, zone, false, false, "tester", true, workingFile)

			if (err != nil) != tt.wantErr {
				t.Errorf("Co() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			content, err := os.ReadFile(workingFile)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.expected, string(content)); diff != "" {
				t.Errorf("Co() content mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
