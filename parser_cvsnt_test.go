package rcs

import (
	"github.com/google/go-cmp/cmp"
	"strings"
	"testing"
)

func TestParseRevisionHeader_CVSNT(t *testing.T) {
	// Input with multiple values for hardlinks, owner, etc.
	// As per user request: hardlinks README @install.txt@ @Installation Notes@;
	input := `1.2
date	99.01.12.14.05.31;	author lhecking;	state dead;
branches;
next	1.1;
owner	640;
group	15;
permissions	644;
hardlinks	README @install.txt@ @Installation Notes@;
deltatype	text;
kopt	kv;
mergepoint	1.1.1.1;
filename	readme.txt;
username	user1;
newfield	value1 @value 2@;

`

	s := NewScanner(strings.NewReader(input))
	rh, _, _, err := ParseRevisionHeader(s)
	if err != nil {
		t.Fatalf("ParseRevisionHeader returned error: %v", err)
	}

	// Verify Hardlinks parsed as multiple values
	expectedHardlinks := []string{"README", "install.txt", "Installation Notes"}
	if diff := cmp.Diff(rh.Hardlinks, expectedHardlinks); diff != "" {
		t.Errorf("Hardlinks mismatch (-got +want):\n%s", diff)
	}

	// Verify other fields
	if len(rh.Deltatype) != 1 || rh.Deltatype[0] != "text" {
		t.Errorf("Deltatype mismatch: %v", rh.Deltatype)
	}
	if len(rh.Kopt) != 1 || rh.Kopt[0] != "kv" {
		t.Errorf("Kopt mismatch: %v", rh.Kopt)
	}
	if len(rh.Mergepoint) != 1 || rh.Mergepoint[0] != "1.1.1.1" {
		t.Errorf("Mergepoint mismatch: %v", rh.Mergepoint)
	}
	if len(rh.Filename) != 1 || rh.Filename[0] != "readme.txt" {
		t.Errorf("Filename mismatch: %v", rh.Filename)
	}
	if len(rh.Username) != 1 || rh.Username[0] != "user1" {
		t.Errorf("Username mismatch: %v", rh.Username)
	}

	// Verify NewPhrases
	if len(rh.NewPhrases) != 1 {
		t.Fatalf("Expected 1 NewPhrase, got %d", len(rh.NewPhrases))
	}
	if rh.NewPhrases[0].Key != "newfield" {
		t.Errorf("NewPhrase Key mismatch: %s", rh.NewPhrases[0].Key)
	}
	expectedNewFieldValues := []string{"value1", "value 2"}
	if diff := cmp.Diff(rh.NewPhrases[0].Value, expectedNewFieldValues); diff != "" {
		t.Errorf("NewPhrase Value mismatch (-got +want):\n%s", diff)
	}

	// Verify Round Trip String()
	// Note: formatting might change quotes.
	// README -> README (ID)
	// install.txt -> install.txt (ID)
	// Installation Notes -> @Installation Notes@ (quoted string)
	// value1 -> value1
	// value 2 -> @value 2@

	expectedString := `1.2
date	99.01.12.14.05.31;	author lhecking;	state dead;
branches;
next	1.1;
owner	640;
group	15;
permissions	644;
hardlinks	README install.txt @Installation Notes@;
deltatype	text;
kopt	kv;
mergepoint	1.1.1.1;
filename	readme.txt;
username	user1;
newfield	value1 @value 2@;
`
	gotString := rh.String()
	if diff := cmp.Diff(gotString, expectedString); diff != "" {
		t.Errorf("String() mismatch (-got +want):\n%s", diff)
	}
}
