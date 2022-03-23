package rcs

import (
	"github.com/google/go-cmp/cmp"
	"testing"
)

func Test_scanFound(t *testing.T) {
	type args struct {
		found   []byte
		advance int
		pos     *Pos
	}
	tests := []struct {
		name     string
		args     args
		expected *Pos
	}{
		{
			name: "Advance characters no new Line",
			args: args{
				found:   []byte("testing"),
				advance: len("testing"),
				pos:     &Pos{},
			},
			expected: &Pos{
				Line:   0,
				Offset: len("testing"),
			},
		},
		{
			name: "New Line at end zeros advance and increments new line",
			args: args{
				found:   []byte("testing\n"),
				advance: len("testing\n"),
				pos:     &Pos{},
			},
			expected: &Pos{
				Line:   1,
				Offset: 0,
			},
		},
		{
			name: "Trailing content after new line increments advance correctly",
			args: args{
				found:   []byte("testing\ntesting 123"),
				advance: len("testing\n"),
				pos:     &Pos{},
			},
			expected: &Pos{
				Line:   1,
				Offset: len("testing 123"),
			},
		},
		{
			name: "Multiple new lines don't break anything",
			args: args{
				found:   []byte("testing\ntesting 123\nand some more content\noh and more!"),
				advance: len("testing\ntesting 123\nand some more content\noh and more!"),
				pos:     &Pos{},
			},
			expected: &Pos{
				Line:   3,
				Offset: len("oh and more!"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanFound(tt.args.found, tt.args.advance, tt.args.pos)
			if diff := cmp.Diff(tt.expected, tt.args.pos); diff != "" {
				t.Errorf("Failed: %s", diff)
			}
		})
	}
}

func TestPos_String(t *testing.T) {
	type fields struct {
		line   int
		offset int
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "test",
			fields: fields{
				line:   3,
				offset: 34,
			},
			want: "3:34",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Pos{
				Line:   tt.fields.line,
				Offset: tt.fields.offset,
			}
			if got := p.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewScanner(t *testing.T) {
	tests := []struct {
		name string
		want *Scanner
	}{
		{
			name: "Set the pos correctly",
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewScanner(nil)
			if got.pos.Line != 1 {
				t.Errorf("Wrong line number")
			}
		})
	}
}
