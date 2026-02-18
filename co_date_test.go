package rcs

import (
	"strings"
	"testing"
	"time"
)

func TestCheckout_WithDate(t *testing.T) {
	input := `head	1.2;
access;
symbols;
locks; strict;
comment	@# @;


1.2
date	2023.01.01.00.00.00;	author user;	state Exp;
branches;
next	1.1;

1.1
date	2022.01.01.00.00.00;	author user;	state Exp;
branches;
next	;


desc
@@


1.2
log
@rev 1.2@
text
@A
B
@


1.1
log
@rev 1.1@
text
@d2 1
@
`

	f, err := ParseFile(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}

	tests := []struct {
		name     string
		date     time.Time
		wantRev  string
        wantContent string
	}{
		{
			name:    "Before 1.1",
			date:    time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			wantRev: "", // No revision found
		},
		{
			name:    "At 1.1",
			date:    time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
			wantRev: "1.1",
            wantContent: "A\n",
		},
		{
			name:    "Between 1.1 and 1.2",
			date:    time.Date(2022, 6, 1, 0, 0, 0, 0, time.UTC),
			wantRev: "1.1",
            wantContent: "A\n",
		},
		{
			name:    "At 1.2",
			date:    time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			wantRev: "1.2",
            wantContent: "A\nB\n",
		},
		{
			name:    "After 1.2",
			date:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			wantRev: "1.2",
            wantContent: "A\nB\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verdict, err := f.Checkout("user", WithDate(tt.date))
			if tt.wantRev == "" {
				if err == nil {
					t.Errorf("Checkout(date=%v) expected error, got rev %s", tt.date, verdict.Revision)
				}
			} else {
				if err != nil {
					t.Fatalf("Checkout(date=%v) error: %v", tt.date, err)
				}
				if verdict.Revision != tt.wantRev {
					t.Errorf("Checkout(date=%v) rev = %s, want %s", tt.date, verdict.Revision, tt.wantRev)
				}
                if tt.wantContent != "" && verdict.Content != tt.wantContent {
                    t.Errorf("Checkout(date=%v) content = %q, want %q", tt.date, verdict.Content, tt.wantContent)
                }
			}
		})
	}
}
