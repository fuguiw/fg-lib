package utils

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShowJsonDiff(t *testing.T) {
	tests := []struct {
		name    string
		oldJson interface{}
		newJson interface{}
		want    string
		wantErr bool
	}{
		{
			name:    "No Diff",
			oldJson: map[string]string{"a": "b"},
			newJson: map[string]string{"a": "b"},
			want:    "json no diff",
			wantErr: false,
		},
		{
			name:    "Has Diff",
			oldJson: map[string]string{"a": "b"},
			newJson: map[string]string{"a": "c"},
			// The exact diff output depends on jsondiff, but it should contain "c"
			want:    "c",
			wantErr: false,
		},
		{
			name:    "Error Marshal",
			oldJson: make(chan int),
			newJson: map[string]string{"a": "b"},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ShowJsonDiff(tt.oldJson, tt.newJson)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.want == "json no diff" {
					assert.Equal(t, tt.want, got)
				} else {
					assert.Contains(t, got, tt.want)
				}
			}
		})
	}
}

func TestPrintTable(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	header := []interface{}{"Name", "Age"}
	rows := [][]interface{}{
		{"Alice", 30},
		{"Bob", 25},
	}

	PrintTable(header, rows)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	assert.Contains(t, output, "Alice")
	assert.Contains(t, output, "Bob")
	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "AGE")
}

func TestEditInTempFile(t *testing.T) {
	// Create a dummy editor script
	editorContent := `#!/bin/sh
echo "modified" > "$1"
`
	editorFile, err := os.CreateTemp("", "dummy_editor_*.sh")
	assert.NoError(t, err)
	defer os.Remove(editorFile.Name())

	_, err = editorFile.WriteString(editorContent)
	assert.NoError(t, err)
	err = editorFile.Chmod(0755)
	assert.NoError(t, err)
	editorFile.Close()

	// Set EDITOR env var
	os.Setenv("EDITOR", editorFile.Name())
	defer os.Unsetenv("EDITOR")

	input := []byte("original")
	got, err := EditInTempFile("test", input)
	assert.NoError(t, err)
	assert.Equal(t, "modified\n", string(got))
}
