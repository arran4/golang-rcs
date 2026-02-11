package rcs

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ParseDate parses a date string according to RCS date formats.
// It accepts a reference time 'now' to fill in missing fields.
// If 'now' is zero, it uses time.Now().
// If 'defaultZone' is nil, it uses UTC.
func ParseDate(input string, now time.Time, defaultZone *time.Location) (time.Time, error) {
	if now.IsZero() {
		now = time.Now()
	}
	if defaultZone == nil {
		defaultZone = time.UTC
	}

	input = strings.TrimSpace(input)

	// Handle LT (Local Time)
	// RCS doc says "The special value ‘LT’ stands for the “local time zone”."
	// We'll interpret this as forcing the use of time.Local (or the system's local zone).
	forceLocal := false
	if strings.HasSuffix(strings.ToLower(input), " lt") {
		forceLocal = true
		input = input[:len(input)-3]
		input = strings.TrimSpace(input)
	} else if strings.HasSuffix(strings.ToLower(input), "lt") {
		// handle case where no space? Documentation examples always have space or comma.
		// "8:00 pm lt"
		// "Thu Jan 11 20:00:00 1990 LT"
		forceLocal = true
		input = input[:len(input)-2]
		input = strings.TrimSpace(input)
	}

	targetZone := defaultZone
	if forceLocal {
		targetZone = time.Local
	}

	// Try specific regex formats first (YEAR-DOY, YEAR-wWEEK-DOW)
	if t, ok := parseYearDOY(input, targetZone); ok {
		return t, nil
	}
	if t, ok := parseYearWeekDow(input, targetZone); ok {
		return t, nil
	}

	// Try standard layouts
	for _, layout := range dateLayouts {
		// Try parsing in targetZone
		t, err := time.ParseInLocation(layout.layout, input, targetZone)
		if err == nil {
			return applyDefaults(t, now, layout.fields, targetZone), nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", input)
}

const (
	FieldYear = 1 << iota
	FieldMonth
	FieldDay
	FieldHour
	FieldMinute
	FieldSecond
	FieldZone
)

type dateLayout struct {
	layout string
	fields int
}

var dateLayouts = []dateLayout{
	// ISO 8601 variations
	{"2006-01-02 15:04:05-07", FieldYear | FieldMonth | FieldDay | FieldHour | FieldMinute | FieldSecond | FieldZone},
	{"2006-01-02 15:04:05-0700", FieldYear | FieldMonth | FieldDay | FieldHour | FieldMinute | FieldSecond | FieldZone},
	{"2006-01-02 15:04:05", FieldYear | FieldMonth | FieldDay | FieldHour | FieldMinute | FieldSecond},

	// Traditional RCS format
	{"2006/01/02 15:04:05", FieldYear | FieldMonth | FieldDay | FieldHour | FieldMinute | FieldSecond},

	// RFC 1123 / RFC 822
	{time.RFC1123Z, FieldYear | FieldMonth | FieldDay | FieldHour | FieldMinute | FieldSecond | FieldZone},
	{time.RFC1123, FieldYear | FieldMonth | FieldDay | FieldHour | FieldMinute | FieldSecond | FieldZone},
	{time.RFC822Z, FieldYear | FieldMonth | FieldDay | FieldHour | FieldMinute | FieldSecond | FieldZone},
	{time.RFC822, FieldYear | FieldMonth | FieldDay | FieldHour | FieldMinute | FieldSecond | FieldZone},
	{"Mon, 2 Jan 2006 15:04:05 -0700", FieldYear | FieldMonth | FieldDay | FieldHour | FieldMinute | FieldSecond | FieldZone},

	// ctime style
	{time.ANSIC, FieldYear | FieldMonth | FieldDay | FieldHour | FieldMinute | FieldSecond}, // "Mon Jan _2 15:04:05 2006"

	// "Thu Jan 11 20:00:00 PST 1990" - Unix date command
	{time.UnixDate, FieldYear | FieldMonth | FieldDay | FieldHour | FieldMinute | FieldSecond | FieldZone}, // "Mon Jan _2 15:04:05 MST 2006"

	// Flexible formats
	{"3:04 pm", FieldHour | FieldMinute},
	{"3:04 PM", FieldHour | FieldMinute},
	{"15:04", FieldHour | FieldMinute},
	{"3:04 AM, Jan. 2, 2006", FieldYear | FieldMonth | FieldDay | FieldHour | FieldMinute},
	{"3:04 AM, Jan. 02, 2006", FieldYear | FieldMonth | FieldDay | FieldHour | FieldMinute},

	// "20, 10:30"
	{"02, 15:04", FieldDay | FieldHour | FieldMinute},
	{"2, 15:04", FieldDay | FieldHour | FieldMinute},

	// "12-January-1990, 04:00"
	{"02-January-2006, 15:04", FieldYear | FieldMonth | FieldDay | FieldHour | FieldMinute},
	{"2-January-2006, 15:04", FieldYear | FieldMonth | FieldDay | FieldHour | FieldMinute},

	// With Zone (MST) - explicit layouts because time.Parse with MST requires it to be known or handled?
	// Actually ParseInLocation handles known abbreviations for that location, but arbitrary ones like WET might fail if not known.
	// We will trust time.Parse to do its best.
	{"02-January-2006, 15:04 MST", FieldYear | FieldMonth | FieldDay | FieldHour | FieldMinute | FieldZone},
	{"2-January-2006, 15:04 MST", FieldYear | FieldMonth | FieldDay | FieldHour | FieldMinute | FieldZone},

	// date(1) variations
	{"Fri Jan 02 15:04:05 MST 2006", FieldYear | FieldMonth | FieldDay | FieldHour | FieldMinute | FieldSecond | FieldZone},
	{"Fri Jan 2 15:04:05 MST 2006", FieldYear | FieldMonth | FieldDay | FieldHour | FieldMinute | FieldSecond | FieldZone},
}

func applyDefaults(t time.Time, now time.Time, fields int, targetZone *time.Location) time.Time {
	// Identify highest provided field
	highest := 0
	if fields&FieldYear != 0 {
		highest = FieldYear
	} else if fields&FieldMonth != 0 {
		highest = FieldMonth
	} else if fields&FieldDay != 0 {
		highest = FieldDay
	} else if fields&FieldHour != 0 {
		highest = FieldHour
	} else if fields&FieldMinute != 0 {
		highest = FieldMinute
	} else if fields&FieldSecond != 0 {
		highest = FieldSecond
	}

	// Helper to get current values in target zone
	// Wait, "For omitted fields that are of higher significance than the highest provided field, the time zone’s current values are assumed."
	// This implies we should use 'now' converted to 'targetZone'.
	nowInZone := now.In(targetZone)

	year, month, day := t.Date()
	hour, min, sec := t.Clock()

	// Defaults
	dYear, dMonth, dDay := nowInZone.Date()
	// dHour, dMin, dSec := nowInZone.Clock() // Not used for higher significance?
	// "For all other omitted fields, the lowest possible values are assumed."

	// Apply higher significance defaults
	if highest > FieldYear {
		year = dYear
	}
	if highest > FieldMonth {
		month = dMonth
	}
	if highest > FieldDay {
		day = dDay
	}

	// Apply lower significance defaults (lowest possible values)
	// time.Parse already defaults to 0 (or 1 for month/day)
	// But we need to be careful. time.Parse defaults year to 0, month to 1, day to 1.
	// If FieldYear is missing, we overwrite with dYear.
	// If FieldMonth is missing, we overwrite with dMonth ONLY if highest > FieldMonth.
	// If highest <= FieldMonth (e.g. Year was provided), then missing Month means we use lowest possible (January).
	// time.Parse already gives us January (1).

	// However, if we parsed "10:30" (highest=Minute), then Year, Month, Day are missing and higher.
	// So we overwrite them with dYear, dMonth, dDay.

	// If we parsed "Jan 2" (highest=Day), then Year is missing and higher.
	// We overwrite Year with dYear.
	// Hour, Minute, Second are missing and lower. We leave them as 0 (from time.Parse).

	finalZone := targetZone
	if fields&FieldZone != 0 {
		finalZone = t.Location()
	}

	return time.Date(year, month, day, hour, min, sec, 0, finalZone)
}

// parseYearDOY handles format YEAR-DOY (e.g., 2018-110)
func parseYearDOY(input string, loc *time.Location) (time.Time, bool) {
	re := regexp.MustCompile(`^(\d{4})-(\d{1,3})$`)
	matches := re.FindStringSubmatch(input)
	if matches == nil {
		return time.Time{}, false
	}
	year, _ := strconv.Atoi(matches[1])
	doy, _ := strconv.Atoi(matches[2])

	if doy < 1 || doy > 366 {
		return time.Time{}, false
	}

	t := time.Date(year, 1, 1, 0, 0, 0, 0, loc).AddDate(0, 0, doy-1)
	return t, true
}

// parseYearWeekDow handles format YEAR-wWEEK-DOW (e.g., 2018-w16-5)
func parseYearWeekDow(input string, loc *time.Location) (time.Time, bool) {
	re := regexp.MustCompile(`^(\d{4})-w(\d{1,2})-(\d)$`)
	matches := re.FindStringSubmatch(input)
	if matches == nil {
		return time.Time{}, false
	}
	year, _ := strconv.Atoi(matches[1])
	week, _ := strconv.Atoi(matches[2])
	dow, _ := strconv.Atoi(matches[3]) // 1=Monday, 7=Sunday

	if week < 0 || week > 53 || dow < 1 || dow > 7 {
		return time.Time{}, false
	}

	// ISO Week Date calculation
	// Start of the year
	t := time.Date(year, 1, 1, 0, 0, 0, 0, loc)
	// Find the first Thursday of the year
	// ISO weeks start on Monday. The first week of the year is the one that contains the first Thursday.
	for t.Weekday() != time.Thursday {
		t = t.AddDate(0, 0, 1)
	}
	// t is now the first Thursday.
	// The ISO week 1 starts on the Monday before this Thursday.
	week1Start := t.AddDate(0, 0, -3) // Monday of week 1

	// Target week start
	targetWeekStart := week1Start.AddDate(0, 0, (week-1)*7)

	// Target day
	// dow: 1=Monday -> offset 0
	targetDay := targetWeekStart.AddDate(0, 0, dow-1)

	return targetDay, true
}
