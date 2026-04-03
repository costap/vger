package youtube

import (
	"testing"
)

func TestExtractVideoID(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "watch url",
			input: "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			want:  "dQw4w9WgXcQ",
		},
		{
			name:  "short url",
			input: "https://youtu.be/dQw4w9WgXcQ",
			want:  "dQw4w9WgXcQ",
		},
		{
			name:  "embed url",
			input: "https://www.youtube.com/embed/dQw4w9WgXcQ",
			want:  "dQw4w9WgXcQ",
		},
		{
			name:  "live url",
			input: "https://www.youtube.com/live/dQw4w9WgXcQ",
			want:  "dQw4w9WgXcQ",
		},
		{
			name:  "mobile url",
			input: "https://m.youtube.com/watch?v=dQw4w9WgXcQ",
			want:  "dQw4w9WgXcQ",
		},
		{
			name:    "not a youtube url",
			input:   "https://vimeo.com/123456",
			wantErr: true,
		},
		{
			name:    "empty url",
			input:   "",
			wantErr: true,
		},
		{
			name:    "youtu.be with no id",
			input:   "https://youtu.be/",
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := extractVideoID(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error, got id=%q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestParseISO8601Duration(t *testing.T) {
	cases := []struct {
		input string
		want  int
	}{
		{"PT45M30S", 45*60 + 30},
		{"PT1H15M30S", 1*3600 + 15*60 + 30},
		{"PT1H", 3600},
		{"PT30S", 30},
		{"PT0S", 0},
		{"", 0},
		{"invalid", 0},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got := parseISO8601Duration(tc.input)
			if got != tc.want {
				t.Errorf("parseISO8601Duration(%q) = %d, want %d", tc.input, got, tc.want)
			}
		})
	}
}
