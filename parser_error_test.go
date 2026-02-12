package rcs

import (
	"strings"
	"testing"
)

func TestParseFile_ErrorPropagation(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantErrString []string
	}{
		{
			name:  "Invalid header - missing head",
			input: "invalid",
			wantErrString: []string{
				"parsing",
				"looking for \"head\"",
			},
		},
		{
			name:  "Invalid property in header",
			input: "head invalid",
			wantErrString: []string{
				"parsing",
				"looking for",
			},
		},
		{
			name:  "Invalid revision header",
			input: "head 1.1;\n\ninvalid\n",
			wantErrString: []string{
				"parsing",
				"finding revision header field",
			},
		},
		{
			name:  "Invalid description",
			input: "head 1.1;\n\ndesc\ninvalid",
			wantErrString: []string{
				"parsing",
				"quote string",
			},
		},
		{
			name:  "Invalid revision content",
			input: "head 1.1;\n\ndesc\n@@\n\ninvalid\ninvalid",
			wantErrString: []string{
				"parsing",
				"looking for",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			_, err := ParseFile(r)
			if err == nil {
				t.Error("ParseFile() error = nil, wantErr true")
				return
			}
			for _, s := range tt.wantErrString {
				if !strings.Contains(err.Error(), s) {
					t.Errorf("ParseFile() error = %v, want to contain %v", err, s)
				}
			}
		})
	}
}
