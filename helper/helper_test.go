package helper

import (
	"reflect"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

func TestGetParagraphs(t *testing.T) {
	tests := []struct {
		name     string
		markdown []byte
		want     []string
	}{
		{
			name:     "single paragraph",
			markdown: []byte("This is a simple paragraph."),
			want:     []string{"This is a simple paragraph."},
		},
		{
			name:     "multiple paragraphs",
			markdown: []byte("First paragraph.\n\nSecond paragraph.\n\nThird paragraph."),
			want:     []string{"First paragraph.", "Second paragraph.", "Third paragraph."},
		},
		{
			name:     "skip blockquote",
			markdown: []byte("Normal paragraph.\n\n> This is in a blockquote\n> Another line in blockquote\n\nAnother normal paragraph."),
			want:     []string{"Normal paragraph.", "Another normal paragraph."},
		},
		{
			name:     "skip code blocks",
			markdown: []byte("Before code.\n\n```\ncode block content\nmore code\n```\n\nAfter code."),
			want:     []string{"Before code.", "After code."},
		},
		{
			name:     "skip tables",
			markdown: []byte("Before table.\n\n| Header1 | Header2 |\n|---------|----------|\n| Cell1   | Cell2   |\n\nAfter table."),
			want:     []string{"Before table.", "After table."},
		},
		{
			name:     "empty input",
			markdown: []byte(""),
			want:     []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetParagraphs(tt.markdown)
			if len(got) != len(tt.want) {
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("GetParagraphs() = %v, %s, want %v, %s, name: %s", got, reflect.TypeOf(got), tt.want, reflect.TypeOf(tt.want), tt.name)
				}
			}
		})
	}
}

// --- Start Edit ---
func TestExtractTextRecursively(t *testing.T) {
	tests := []struct {
		name     string
		markdown []byte
		want     string // Expected alt text
	}{
		{
			name:     "simple alt text",
			markdown: []byte("![Simple alt text](image.png)"),
			want:     "Simple alt text",
		},
		{
			name:     "no image",
			markdown: []byte("Just some text."),
			want:     "", // Expect empty string if no image node is found/passed
		},
	}

	parser := goldmark.DefaultParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := text.NewReader(tt.markdown)
			root := parser.Parse(reader)
			var imgNode ast.Node
			var got string

			// Find the first image node in the parsed AST
			ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
				if img, ok := n.(*ast.Image); ok && entering {
					imgNode = img
					return ast.WalkStop, nil // Stop walking once we find the first image
				}
				return ast.WalkContinue, nil
			})

			// If an image node was found, extract text from it
			if imgNode != nil {
				got = ExtractTextRecursivelyByReader(imgNode, reader)
			} else if tt.name != "no image" {
				// Fail if we expected an image but didn't find one
				t.Errorf("Test case '%s' failed: No image node found in markdown: %s", tt.name, tt.markdown)
				return
			}

			// Compare the extracted text with the expected text
			if got != tt.want {
				t.Errorf("ExtractTextRecursively() for test '%s' = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}
