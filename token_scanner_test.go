package rcs

import (
	"fmt"
	"strings"
	"testing"
)

func TestScanTokenNum(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"Simple number", "123;", "123", false},
		{"Decimal", "1.2;", "1.2", false},
		{"Multiple dots", "1.2.3.4;", "1.2.3.4", false},
		{"Ends with dot", "1.;", "1.", false},
		{"Starts with dot", ".1;", ".1", false},
		{"Only dots", "...;", "...", false},
		{"Stops at letter", "123a", "123", false},
		{"Stops at space", "123 ", "123", false},
		{"Stops at special", "123;", "123", false},
		{"Empty", ";", "", true}, // Minimum 1 char
		{"Non-digit", "a;", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScanner(strings.NewReader(tt.input))
			got, err := ScanTokenNum(s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ScanTokenNum() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ScanTokenNum() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScanTokenPhrase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		want     string // Raw value
		wantType string // "SimpleString" or "QuotedString"
		wantErr  bool
	}{
		{"Quoted string", "@foo@", "foo", "QuotedString", false},
		{"Quoted with @", "@foo@@bar@", "foo@bar", "QuotedString", false},
		{"Unquoted ID", "foo;", "foo", "SimpleString", false},
		{"Colon", ":", ":", "SimpleString", false},
		{"Empty string", "@@", "", "QuotedString", false},
		{"Empty ID", ";", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScanner(strings.NewReader(tt.input))
			got, err := ScanTokenPhrase(s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ScanTokenPhrase() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Raw() != tt.want {
					t.Errorf("ScanTokenPhrase() raw = %q, want %q", got.Raw(), tt.want)
				}
				var gotType string
				switch got.(type) {
				case SimpleString:
					gotType = "SimpleString"
				case QuotedString:
					gotType = "QuotedString"
				default:
					gotType = fmt.Sprintf("%T", got)
				}
				if gotType != tt.wantType {
					t.Errorf("ScanTokenPhrase() type = %s, want %s", gotType, tt.wantType)
				}
			}
		})
	}
}

func TestPhraseValuesFormat(t *testing.T) {
	tests := []struct {
		name  string
		input PhraseValues
		want  PhraseValues // Expected types/values after Format
	}{
		{
			"All valid SimpleStrings",
			PhraseValues{SimpleString("foo"), SimpleString("bar")},
			PhraseValues{SimpleString("foo"), SimpleString("bar")},
		},
		{
			"Invalid SimpleString to Quoted",
			PhraseValues{SimpleString("foo bar"), SimpleString("baz")},
			PhraseValues{QuotedString("foo bar"), SimpleString("baz")},
		},
		{
			"Valid QuotedString to Simple",
			PhraseValues{QuotedString("foo"), QuotedString("bar")},
			PhraseValues{SimpleString("foo"), SimpleString("bar")},
		},
		{
			"Invalid QuotedString stays Quoted",
			PhraseValues{QuotedString("foo bar")},
			PhraseValues{QuotedString("foo bar")},
		},
		{
			"Mixed",
			PhraseValues{SimpleString("a b"), QuotedString("c"), SimpleString("d")},
			PhraseValues{QuotedString("a b"), SimpleString("c"), SimpleString("d")},
		},
		{
			"With dot",
			PhraseValues{SimpleString("foo.bar"), QuotedString("baz.qux")},
			PhraseValues{SimpleString("foo.bar"), SimpleString("baz.qux")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy to modify
			input := make(PhraseValues, len(tt.input))
			copy(input, tt.input)
			input.Format()

			if len(input) != len(tt.want) {
				t.Fatalf("Format() len = %d, want %d", len(input), len(tt.want))
			}
			for i, v := range input {
				w := tt.want[i]
				if v != w {
					t.Errorf("Format()[%d] = %#v, want %#v", i, v, w)
				}
			}
		})
	}
}

func TestScanTokenId(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"Simple ID", "foo;", "foo", false},
		{"With digits", "foo123;", "foo123", false},
		{"With dot", "foo.bar;", "foo.bar", false},
		{"With dash", "foo-bar;", "foo-bar", false},
		{"With underscore", "foo_bar;", "foo_bar", false},
		{"Stops at space", "foo ", "foo", false},
		{"Stops at colon", "foo:", "foo", false},
		{"Stops at comma", "foo,", "foo", false},
		{"Stops at @", "foo@", "foo", false},
		{"Empty", ";", "", true},
		{"Invalid char start", "$foo;", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScanner(strings.NewReader(tt.input))
			got, err := ScanTokenId(s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ScanTokenId() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ScanTokenId() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScanTokenSym(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"Simple Sym", "foo;", "foo", false},
		{"With digits", "foo123;", "foo123", false},
		{"With dash", "foo-bar;", "foo-bar", false},
		{"Stops at dot", "foo.bar;", "foo", false}, // Sym stops at dot
		{"Stops at space", "foo ", "foo", false},
		{"Stops at colon", "foo:", "foo", false},
		{"Empty", ";", "", true},
		{"Invalid char start", ".foo;", "", true}, // dot is special
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScanner(strings.NewReader(tt.input))
			got, err := ScanTokenSym(s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ScanTokenSym() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ScanTokenSym() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScanTokenString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"Simple string", "@foo@", "foo", false},
		{"Doubled @", "@foo@@bar@", "foo@bar", false},
		{"Empty content", "@@", "", false},
		{"Multiline", "@foo\nbar@", "foo\nbar", false},
		{"Missing end quote", "@foo", "", true},
		{"Missing start quote", "foo@", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScanner(strings.NewReader(tt.input))
			got, err := ScanTokenString(s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ScanTokenString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ScanTokenString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScanTokenIntString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"Simple intstring", "@foo@", "foo", false},
		// The implementation assumes simpler scanning for intstring (scan until @)
		// Doubled @ is NOT handled specially in my implementation (stops at first @)
		{"Stops at first @", "@foo@@bar@", "foo", false},
		{"Empty content", "@@", "", false},
		{"Missing end quote", "@foo", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScanner(strings.NewReader(tt.input))
			got, err := ScanTokenIntString(s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ScanTokenIntString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ScanTokenIntString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScanTokenWord(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"Quoted string", "@foo@", "foo", false},
		{"Unquoted ID", "foo;", "foo", false},
		{"Number with dot", "1.2.3;", "1.2.3", false},
		{"Colon", ":", ":", false},
		{"Stops at space", "foo ", "foo", false},
		{"Stops at semicolon", "foo;", "foo", false},
		{"Empty", ";", "", true},
		{"String", "@foo@", "foo", false},
		{"Id", "foo;", "foo", false},
		{"Id with digits", "foo123;", "foo123", false},
		{"Id with dot", "foo.bar;", "foo.bar", false},
		{"Empty string", "@@", "", false},
		{"Missing start quote ID", "foo@", "foo", false},
		{"Missing end quote String", "@foo", "", true},
		{"Expand Unquoted", "kv;", "kv", false},
		{"Colon", ":", ":", false},
		{"Colon with suffix", ":;", ":", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScanner(strings.NewReader(tt.input))
			got, err := ScanTokenWord(s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ScanTokenWord() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ScanTokenWord() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScanTokenAuthor(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"Quoted author", "@author@", "author", false},
		{"Quoted author with @", "@aut@@hor@", "aut@hor", false},
		{"Unquoted author", "author;", "author", false},
		{"Unquoted author with space", "author name;", "author name", false},
		{"Empty quoted author", "@@", "", false},
		{"Empty unquoted author", ";", "", true},
		{"Missing end quote", "@author", "", true},
		{"Valid unquoted stops at ;", "john.doe;next", "john.doe", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScanner(strings.NewReader(tt.input))
			got, err := ScanTokenAuthor(s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ScanTokenAuthor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ScanTokenAuthor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScanLockIdOrStrings(t *testing.T) {
	type args struct {
		s    *Scanner
		strs []string
	}
	tests := []struct {
		name      string
		args      args
		wantId    string
		wantMatch string
		wantErr   bool
	}{
		{
			name: "Match keyword",
			args: args{
				s:    NewScanner(strings.NewReader("keyword")),
				strs: []string{"keyword"},
			},
			wantId:    "",
			wantMatch: "keyword",
			wantErr:   false,
		},
		{
			name: "Match lock ID",
			args: args{
				s:    NewScanner(strings.NewReader("user:1.1")),
				strs: []string{"keyword"},
			},
			wantId:    "user",
			wantMatch: "",
			wantErr:   false,
		},
		{
			name: "Match nothing",
			args: args{
				s:    NewScanner(strings.NewReader("other")),
				strs: []string{"keyword"},
			},
			wantId:    "",
			wantMatch: "",
			wantErr:   true,
		},
		{
			name: "Empty input",
			args: args{
				s:    NewScanner(strings.NewReader("")),
				strs: []string{"keyword"},
			},
			wantId:    "",
			wantMatch: "",
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotId, gotMatch, err := ScanLockIdOrStrings(tt.args.s, tt.args.strs...)
			if (err != nil) != tt.wantErr {
				t.Errorf("ScanLockIdOrStrings() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotId != tt.wantId {
				t.Errorf("ScanLockIdOrStrings() gotId = %v, want %v", gotId, tt.wantId)
			}
			if gotMatch != tt.wantMatch {
				t.Errorf("ScanLockIdOrStrings() gotMatch = %v, want %v", gotMatch, tt.wantMatch)
			}
		})
	}
}

func TestParseAtQuotedString(t *testing.T) {
	type args struct {
		s *Scanner
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Base @ String",
			args: args{
				s: NewScanner(strings.NewReader("@# @")),
			},
			want:    "# ",
			wantErr: false,
		},
		{
			name: "Double @@ For literal",
			args: args{
				s: NewScanner(strings.NewReader("@ @@ @")),
			},
			want:    " @ ",
			wantErr: false,
		},
		{
			name: "New lines are fine",
			args: args{
				s: NewScanner(strings.NewReader("@Hello\nyou@")),
			},
			want:    "Hello\nyou",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAtQuotedString(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAtQuotedString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseAtQuotedString() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseAtQuotedString_Errors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{
			name:    "Missing start quote",
			input:   "no quote",
			wantErr: "open quote: looking for \"@\"",
		},
		{
			name:    "Missing end quote",
			input:   "@start quote",
			wantErr: "looking for \"@\"", // ScanUntilStrings returns ScanNotFound at EOF
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScanner(strings.NewReader(tt.input))
			_, err := ParseAtQuotedString(s)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("ParseAtQuotedString() error = %v, wantErr containing %q", err, tt.wantErr)
			}
		})
	}
}
