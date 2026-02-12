package rcs

import (
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
