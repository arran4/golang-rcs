package cli

import (
	"reflect"
	"testing"
)

func TestParseDelimitedList(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		delims string
		want   []string
	}{
		{
			name:   "single delimiter comma",
			input:  "user1,user2,user3",
			delims: ",",
			want:   []string{"user1", "user2", "user3"},
		},
		{
			name:   "multiple delimiters space comma tab",
			input:  "user1, user2\tuser3",
			delims: " ,\t",
			want:   []string{"user1", "user2", "user3"},
		},
		{
			name:   "adjacent delimiters",
			input:  "user1,, user2  \t  user3",
			delims: " ,\t",
			want:   []string{"user1", "user2", "user3"},
		},
		{
			name:   "leading and trailing delimiters",
			input:  ", user1, user2, ",
			delims: " ,",
			want:   []string{"user1", "user2"},
		},
		{
			name:   "semi and comma for rcs -o",
			input:  "1.1;1.2:1.3,1.4",
			delims: ";,",
			want:   []string{"1.1", "1.2:1.3", "1.4"},
		},
		{
			name:   "empty string",
			input:  "",
			delims: ",",
			want:   []string{},
		},
		{
			name:   "only delimiters",
			input:  ",,,",
			delims: ",",
			want:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseDelimitedList(tt.input, tt.delims)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseDelimitedList() = %v, want %v", got, tt.want)
			}
		})
	}
}
