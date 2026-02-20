package rcs

import (
	"fmt"
	"testing"
	"time"
)

func TestParseKeywordSubstitution(t *testing.T) {
	tests := []struct {
		input    string
		expected KeywordSubstitution
		hasError bool
	}{
		{"kv", KV, false},
		{"kvl", KVL, false},
		{"k", K, false},
		{"o", O, false},
		{"b", B, false},
		{"v", V, false},
		{"invalid", KV, true},
		{"", KV, true},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("input=%q", tc.input), func(t *testing.T) {
			got, err := ParseKeywordSubstitution(tc.input)
			if tc.hasError {
				if err == nil {
					t.Errorf("ParseKeywordSubstitution(%q) expected error, got nil", tc.input)
				}
			} else {
				if err != nil {
					t.Errorf("ParseKeywordSubstitution(%q) unexpected error: %v", tc.input, err)
				}
				if got != tc.expected {
					t.Errorf("ParseKeywordSubstitution(%q) = %v, want %v", tc.input, got, tc.expected)
				}
			}
		})
	}
}

func TestExpandKeywords(t *testing.T) {
	fixedTime := time.Date(2023, 10, 27, 10, 0, 0, 0, time.UTC)
	data := KeywordData{
		Revision: "1.2",
		Date:     fixedTime,
		Author:   "jules",
		State:    "Exp",
		Locker:   "someone",
		Log:      "Initial commit",
		RCSFile:  "file.v",
		Source:   "/path/to/file.v",
	}

	tests := []struct {
		name     string
		mode     KeywordSubstitution
		input    string
		expected string
	}{
		{
			name:     "KV Revision",
			mode:     KV,
			input:    "$Revision$",
			expected: "$Revision: 1.2 $",
		},
		{
			name:     "K Revision",
			mode:     K,
			input:    "$Revision: 1.1 $",
			expected: "$Revision$",
		},
		{
			name:     "O Revision",
			mode:     O,
			input:    "$Revision: 1.1 $",
			expected: "$Revision: 1.1 $",
		},
		{
			name:     "B Revision",
			mode:     B,
			input:    "$Revision: 1.1 $",
			expected: "$Revision: 1.1 $",
		},
		{
			name:     "V Revision",
			mode:     V,
			input:    "$Revision$",
			expected: "1.2",
		},
		{
			name:     "KV Log",
			mode:     KV,
			input:    "$Log$",
			expected: "$Log: file.v $\nRevision 1.2  2023/10/27 10:00:00  jules\nInitial commit",
		},
		{
			name:     "KV Header",
			mode:     KV,
			input:    "$Header$",
			expected: "$Header: /path/to/file.v 1.2 2023/10/27 10:00:00 jules Exp someone $",
		},
		{
			name:     "KV Id",
			mode:     KV,
			input:    "$Id$",
			expected: "$Id: file.v 1.2 2023/10/27 10:00:00 jules Exp someone $",
		},
		{
			name:     "No Keyword",
			mode:     KV,
			input:    "No keyword here",
			expected: "No keyword here",
		},
		{
			name:     "Multiple Keywords",
			mode:     KV,
			input:    "$Revision$ and $Date$",
			expected: "$Revision: 1.2 $ and $Date: 2023/10/27 10:00:00 $",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ExpandKeywords(tc.input, data, tc.mode)
			if got != tc.expected {
				t.Errorf("ExpandKeywords(%q, mode=%v) = %q, want %q", tc.input, tc.mode, got, tc.expected)
			}
		})
	}
}
