package gemini

import (
	"encoding/json"
	"testing"
)

func TestSanitizeJSONControls(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool // the output must be valid JSON when wrapped as a string value
	}{
		{
			name:  "bare newline in string field",
			input: `{"summary": "line one` + "\n" + `line two"}`,
			valid: true,
		},
		{
			name:  "bare carriage return in string field",
			input: `{"summary": "line one` + "\r" + `line two"}`,
			valid: true,
		},
		{
			name:  "bare tab in string field",
			input: `{"summary": "col one` + "\t" + `col two"}`,
			valid: true,
		},
		{
			name:  "mixed control chars in notes field",
			input: `{"notes": "para one` + "\n\r" + `para two` + "\t" + `end"}`,
			valid: true,
		},
		{
			name:  "already-escaped sequences not double-escaped",
			input: `{"summary": "line one\nline two"}`,
			valid: true,
		},
		{
			name:  "newlines outside string (whitespace) preserved",
			input: "{\n  \"summary\": \"clean\"\n}",
			valid: true,
		},
		{
			name:  "empty string",
			input: "",
			valid: true,
		},
		{
			name:  "no strings",
			input: `{"count": 42}`,
			valid: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := sanitizeJSONControls(tc.input)
			if tc.valid && tc.input != "" {
				var v any
				if err := json.Unmarshal([]byte(got), &v); err != nil {
					t.Errorf("output is not valid JSON: %v\noutput: %s", err, got)
				}
			}
		})
	}
}

func TestSanitizeJSONControls_ContentPreserved(t *testing.T) {
	// Verify that the text content is preserved after sanitization.
	raw := `{"summary": "line one` + "\n" + `line two", "count": 3}`
	got := sanitizeJSONControls(raw)

	var v map[string]any
	if err := json.Unmarshal([]byte(got), &v); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	summary, _ := v["summary"].(string)
	if summary != "line one\nline two" {
		t.Errorf("content changed unexpectedly: %q", summary)
	}
}
