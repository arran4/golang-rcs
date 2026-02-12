package rcs

import (
	"strings"
	"testing"
)

func TestParseHeaderExpandIntegrity(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantExpand    string
		wantIntegrity string
		wantErr       bool
	}{
		{
			name: "Expand and Integrity with quotes",
			input: `head	1.1;
access;
symbols;
locks; strict;
expand	@kv@;
integrity	@int123@;
comment	@# @;


1.1
date	2022.01.01.00.00.00;	author arran;	state Exp;
branches;
next	;


desc
@@


1.1
log
@@
text
@@
`,
			wantExpand:    "kv",
			wantIntegrity: "int123",
			wantErr:       false,
		},
		{
			name: "Expand without quotes",
			input: `head	1.1;
access;
symbols;
locks; strict;
expand	kv;
comment	@# @;


1.1
date	2022.01.01.00.00.00;	author arran;	state Exp;
branches;
next	;


desc
@@


1.1
log
@@
text
@@
`,
			wantExpand:    "kv",
			wantIntegrity: "",
			wantErr:       false,
		},
		{
			name: "Integrity unquoted should fail",
			input: `head	1.1;
integrity	unquoted;
comment	@# @;


1.1
date	2022.01.01.00.00.00;	author arran;	state Exp;
branches;
next	;


desc
@@


1.1
log
@@
text
@@
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := ParseFile(strings.NewReader(tt.input))
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseFile() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if f.Expand != tt.wantExpand {
				t.Errorf("Expand = %q, want %q", f.Expand, tt.wantExpand)
			}
			if f.Integrity != tt.wantIntegrity {
				t.Errorf("Integrity = %q, want %q", f.Integrity, tt.wantIntegrity)
			}

			gotString := f.String()
			f2, err := ParseFile(strings.NewReader(gotString))
			if err != nil {
				t.Errorf("ParseFile(f.String()) error = %v", err)
			} else {
				if f2.Expand != f.Expand {
					t.Errorf("RoundTrip Expand = %q, want %q", f2.Expand, f.Expand)
				}
				if f2.Integrity != f.Integrity {
					t.Errorf("RoundTrip Integrity = %q, want %q", f2.Integrity, f.Integrity)
				}
			}
		})
	}
}
