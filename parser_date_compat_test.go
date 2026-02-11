package rcs

import (
	"strings"
	"testing"
	"time"
)

func TestParseRevisionHeaderDateLine_Compat(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedTime time.Time
	}{
		{
			name:         "2-digit year (1997)",
			input:        "date\t97.04.06.08.41.11;\tauthor arran;\tstate Exp;\n",
			expectedTime: time.Date(1997, 4, 6, 8, 41, 11, 0, time.UTC),
		},
		{
			name:         "4-digit year (1997)",
			input:        "date\t1997.04.06.08.41.11;\tauthor arran;\tstate Exp;\n",
			expectedTime: time.Date(1997, 4, 6, 8, 41, 11, 0, time.UTC),
		},
		{
			name:         "2-digit year (2020) - assuming 00-68 maps to 2000-2068",
			input:        "date\t20.01.01.00.00.00;\tauthor arran;\tstate Exp;\n",
			expectedTime: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScanner(strings.NewReader(tt.input))
			rh := &RevisionHead{}
			err := ParseRevisionHeaderDateLine(s, false, rh)
			if err != nil {
				t.Errorf("ParseRevisionHeaderDateLine failed: %v", err)
				return
			}
			if !rh.Date.Equal(tt.expectedTime) {
				t.Errorf("Expected time %v, got %v", tt.expectedTime, rh.Date)
			}
		})
	}
}
