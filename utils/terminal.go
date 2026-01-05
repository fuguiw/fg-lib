package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/nsf/jsondiff"
)

// ShowJsonDiff compares two objects as JSON and returns the difference.
func ShowJsonDiff(oldJson interface{}, newJson interface{}) (string, error) {
	oldBytes, err := json.Marshal(oldJson)
	if err != nil {
		return "", fmt.Errorf("failed to marshal oldJson: %w", err)
	}
	newBytes, err := json.Marshal(newJson)
	if err != nil {
		return "", fmt.Errorf("failed to marshal newJson: %w", err)
	}

	o := jsondiff.DefaultConsoleOptions()
	o.SkipMatches = true
	d, diff := jsondiff.Compare(oldBytes, newBytes, &o)
	if d == jsondiff.FullMatch {
		return "json no diff", nil
	}
	return diff, nil
}

// PrintTable prints a table to stdout.
func PrintTable(header []interface{}, rows [][]interface{}) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(header)
	for _, row := range rows {
		t.AppendRow(row)
	}
	t.SetStyle(table.StyleLight) // Use a lighter style by default
	t.Render()
}

// EditInTempFile opens the input in a text editor (EDITOR env var or vi) and returns the modified content.
func EditInTempFile(prefix string, input []byte) ([]byte, error) {
	f, err := os.CreateTemp(os.TempDir(), prefix+"_*.temp")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() {
		_ = os.Remove(f.Name())
	}()

	if err := os.WriteFile(f.Name(), input, 0640); err != nil {
		return nil, fmt.Errorf("failed to write to temp file: %w", err)
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	var cmd *exec.Cmd
	if editor == "vi" || editor == "vim" {
		// Use special flags for vi/vim to set tab width
		cmd = exec.Command(editor, "--cmd", "set expandtab tabstop=2", f.Name())
	} else {
		cmd = exec.Command(editor, f.Name())
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("editor command failed: %w", err)
	}

	return os.ReadFile(f.Name())
}
