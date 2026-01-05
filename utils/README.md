# Utils Library

`utils` provides a collection of handy terminal and utility functions for Go applications.

## Features

- **JSON Diff**: Compare two objects and display a colorized JSON difference in the terminal.
- **Table Printing**: Render data as ASCII tables with customizable headers and rows.
- **Interactive Editing**: Open content in the system's default text editor (e.g., `vi`, `nano`) for interactive editing.

## Usage

```go
package main

import (
	"fmt"
	"github.com/fuguiw/fg-lib/utils"
)

func main() {
	// 1. JSON Diff
	oldObj := map[string]interface{}{"name": "Alice", "age": 30}
	newObj := map[string]interface{}{"name": "Alice", "age": 31}
	
	diff, err := utils.ShowJsonDiff(oldObj, newObj)
	if err != nil {
		panic(err)
	}
	fmt.Println(diff)

	// 2. Print Table
	header := []interface{}{"ID", "Name", "Role"}
	rows := [][]interface{}{
		{1, "Alice", "Admin"},
		{2, "Bob", "User"},
	}
	utils.PrintTable(header, rows)

	// 3. Edit in Temp File
	content := []byte("Initial content")
	modified, err := utils.EditInTempFile("example", content)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Modified content: %s\n", modified)
}
```
