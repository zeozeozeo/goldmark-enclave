package object

import "testing"

func TestFormalizeImageSize(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "number only",
			input:    "100",
			expected: "100px",
		},
		{
			name:     "number with px",
			input:    "120px",
			expected: "120px",
		},
		{
			name:     "number with rem",
			input:    "10rem",
			expected: "10rem",
		},
		{
			name:     "number with percentage",
			input:    "50%",
			expected: "50%",
		},
		{
			name:     "no number, only unit px",
			input:    "px",
			expected: "auto",
		},
		{
			name:     "no number, only unit rem",
			input:    "rem",
			expected: "auto",
		},
		{
			name:     "no number, only unit percentage",
			input:    "%",
			expected: "auto",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "auto",
		},
		{
			name:     "unsupported unit",
			input:    "100em",
			expected: "100px",
		},
		{
			name:     "just text",
			input:    "abc",
			expected: "auto",
		},
		{
			name:     "zero",
			input:    "0",
			expected: "auto",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := formalizeImageSize(tc.input)
			if actual != tc.expected {
				t.Errorf("expected %s, but got %s", tc.expected, actual)
			}
		})
	}
}
