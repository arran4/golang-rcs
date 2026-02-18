package rcs

import (
	"errors"
	"testing"
	"time"
)

func TestParseDate(t *testing.T) {
	// Set up reference time: 2023-10-25 12:00:00 UTC
	now := time.Date(2023, 10, 25, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		input       string
		defaultZone *time.Location // Defaults to UTC if nil
		want        time.Time
		checkZone   bool // If true, check strict equality including zone
	}{
		{
			name:  "Time only (8:00 pm)",
			input: "8:00 pm",
			// Expect: 2023-10-25 20:00:00 UTC
			want:      time.Date(2023, 10, 25, 20, 0, 0, 0, time.UTC),
			checkZone: true,
		},
		{
			name:  "Time with LT (8:00 pm lt)",
			input: "8:00 pm lt",
			// Expect: 2023-10-25 20:00:00 Local
			want:      time.Date(2023, 10, 25, 20, 0, 0, 0, time.Local),
			checkZone: true,
		},
		{
			name:  "Date and Time (4:00 AM, Jan. 12, 1990)",
			input: "4:00 AM, Jan. 12, 1990",
			// Expect: 1990-01-12 04:00:00 UTC
			want:      time.Date(1990, 1, 12, 4, 0, 0, 0, time.UTC),
			checkZone: true,
		},
		{
			name:  "ISO 8601 UTC (1990-01-12 04:00:00+00)",
			input: "1990-01-12 04:00:00+00",
			// Expect: 1990-01-12 04:00:00 UTC
			want:      time.Date(1990, 1, 12, 4, 0, 0, 0, time.UTC),
			checkZone: true,
		},
		{
			name:  "ISO 8601 Offset (1990-01-11 20:00:00-08)",
			input: "1990-01-11 20:00:00-08",
			// Expect: 1990-01-11 20:00:00 -0800
			// Equivalent to 1990-01-12 04:00:00 UTC
			want:      time.Date(1990, 1, 11, 20, 0, 0, 0, time.FixedZone("", -8*3600)),
			checkZone: false, // Offset names might differ
		},
		{
			name:  "Traditional RCS (1990/01/12 04:00:00)",
			input: "1990/01/12 04:00:00",
			// Expect: 1990-01-12 04:00:00 UTC
			want:      time.Date(1990, 1, 12, 4, 0, 0, 0, time.UTC),
			checkZone: true,
		},
		{
			name:  "ctime + LT (Thu Jan 11 20:00:00 1990 LT)",
			input: "Thu Jan 11 20:00:00 1990 LT",
			// Expect: 1990-01-11 20:00:00 Local
			want:      time.Date(1990, 1, 11, 20, 0, 0, 0, time.Local),
			checkZone: true,
		},
		{
			name:  "GMT (Fri Jan 12 04:00:00 GMT 1990)",
			input: "Fri Jan 12 04:00:00 GMT 1990",
			// Expect: 1990-01-12 04:00:00 UTC (GMT is UTC)
			// time.Parse usually handles GMT as UTC.
			want:      time.Date(1990, 1, 12, 4, 0, 0, 0, time.UTC),
			checkZone: false,
		},
		{
			name:  "RFC 822 (Thu, 11 Jan 1990 20:00:00 -0800)",
			input: "Thu, 11 Jan 1990 20:00:00 -0800",
			// Expect: 1990-01-11 20:00:00 -0800
			want:      time.Date(1990, 1, 11, 20, 0, 0, 0, time.FixedZone("", -8*3600)),
			checkZone: false,
		},
		{
			name:  "Day, Time (20, 10:30)",
			input: "20, 10:30",
			// Expect: 2023-10-20 10:30:00 UTC
			want:      time.Date(2023, 10, 20, 10, 30, 0, 0, time.UTC),
			checkZone: true,
		},
		{
			name:  "YEAR-DOY (2018-110)",
			input: "2018-110",
			// Expect: 2018, 110th day. April 20.
			want:      time.Date(2018, 4, 20, 0, 0, 0, 0, time.UTC),
			checkZone: true,
		},
		{
			name:  "YEAR-wWEEK-DOW (2018-w16-5)",
			input: "2018-w16-5",
			// 2018. Week 16. Day 5 (Friday). April 20.
			want:      time.Date(2018, 4, 20, 0, 0, 0, 0, time.UTC),
			checkZone: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDate(tt.input, now, tt.defaultZone)
			if err != nil {
				t.Fatalf("ParseDate() error = %v", err)
			}

			// Compare timestamps (ignoring zone name if checkZone is false, but time instant should match)
			if !got.Equal(tt.want) {
				t.Errorf("ParseDate() = %v, want %v", got, tt.want)
			}

			if tt.checkZone {
				// Check zone offset and name if possible
				_, o1 := got.Zone()
				_, o2 := tt.want.Zone()
				if o1 != o2 {
					t.Errorf("ParseDate() zone offset = %v, want %v", o1, o2)
				}
			}
		})
	}
}

func TestParseDate_Errors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"Empty string", ""},
		{"Invalid text", "not a date"},
		{"Invalid Year-DOY", "2018-400"},
		{"Invalid Year-Week-Day", "2018-w54-8"},
		{"Invalid Month", "2018-13-01"},
		{"Invalid Day", "2018-01-32"},
		// Additional invalid formats
		{"Invalid ISO 8601", "2006-13-02 15:04:05"},   // Month 13
		{"Invalid Time", "2006-01-02 25:00:00"},       // Hour 25
		{"Invalid Minute", "2006-01-02 15:60:00"},     // Minute 60 (leap seconds usually not supported in standard parsing this way or simply out of range)
		{"Invalid Second", "2006-01-02 15:04:61"},     // Second 61
		{"Invalid Zone", "2006-01-02 15:04:05 +2500"}, // Offset too large
		{"Garbage ISO 8601", "2006-01-02T15:04:05ZGarbage"},
		{"Invalid RFC1123", "Mon, 02 Jan 2006 25:04:05 MST"},
		{"Invalid RFC822", "02 Jan 06 25:04 MST"},
		{"Invalid Traditional RCS", "2006/13/02 15:04:05"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseDate(tt.input, time.Now(), time.UTC)
			if err == nil {
				t.Errorf("ParseDate(%q) expected error, got nil", tt.input)
				return
			}
			if !errors.Is(err, ErrDateParse) {
				t.Errorf("ParseDate(%q) expected error wrapping ErrDateParse, got %v", tt.input, err)
			}
		})
	}
}

func TestParseZone(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantUTC   bool
		wantLocal bool
		wantOffset int
		wantErr   bool
	}{
		{
			name:    "Empty",
			input:   "",
			wantUTC: true,
		},
		{
			name:      "LT",
			input:     "LT",
			wantLocal: true,
		},
		{
			name:    "UTC",
			input:   "UTC",
			wantUTC: true,
		},
		{
			name:       "-0800",
			input:      "-0800",
			wantOffset: -8 * 3600,
		},
		{
			name:       "+05:30",
			input:      "+05:30",
			wantOffset: 5*3600 + 30*60,
		},
		{
			name:    "Unknown",
			input:   "UnknownZone",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc, err := ParseZone(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseZone() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			if tt.wantUTC {
				if loc.String() != "UTC" && loc.String() != "" { // UTC sometimes string is empty or UTC? time.UTC string is UTC.
					// time.LoadLocation("UTC") returns loc with name "UTC".
					// time.UTC has name "UTC".
					// time.FixedZone("UTC", 0) has name "UTC".
					if loc.String() != "UTC" {
						t.Errorf("ParseZone() = %v, want UTC", loc)
					}
				}
			} else if tt.wantLocal {
				if loc != time.Local {
					t.Errorf("ParseZone() = %v, want Local", loc)
				}
			} else if tt.wantOffset != 0 {
				_, offset := time.Now().In(loc).Zone()
				if offset != tt.wantOffset {
					t.Errorf("ParseZone() offset = %v, want %v", offset, tt.wantOffset)
				}
			}
		})
	}
}
