package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type TestEvent struct {
	Time    time.Time `json:"Time"`
	Action  string    `json:"Action"`
	Package string    `json:"Package"`
	Test    string    `json:"Test"`
	Elapsed float64   `json:"Elapsed"`
	Output  string    `json:"Output"`
}

func main() {
	cmd := exec.Command("go", "test", "./test/...", "-json")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Run()

	lines := strings.Split(out.String(), "\n")

	reportPath := "doc/test_report.md"
	f, err := os.Create(reportPath)
	if err != nil {
		fmt.Println("Error creating report:", err)
		return
	}
	defer f.Close()

	fmt.Fprintln(f, "# FileShare Test Report")
	fmt.Fprintf(f, "Generated on: %s\n\n", time.Now().Format(time.RFC1123))

	fmt.Fprintln(f, "## Test Results")
	fmt.Fprintln(f, "| Package | Test Name | Status | Duration |")
	fmt.Fprintln(f, "|---------|-----------|--------|----------|")

	var outputs []string

	for _, line := range lines {
		if line == "" {
			continue
		}
		var evt TestEvent
		if err := json.Unmarshal([]byte(line), &evt); err == nil {
			if evt.Action == "pass" && evt.Test != "" {
				fmt.Fprintf(f, "| `%s` | `%s` | ✅ PASS | %.3fs |\n", evt.Package, evt.Test, evt.Elapsed)
			} else if evt.Action == "fail" && evt.Test != "" {
				fmt.Fprintf(f, "| `%s` | `%s` | ❌ FAIL | %.3fs |\n", evt.Package, evt.Test, evt.Elapsed)
			}

			if evt.Action == "output" {
			    outputs = append(outputs, evt.Output)
			}
		}
	}

	fmt.Fprintln(f, "\n## Test Logs")
	fmt.Fprintln(f, "```text")
	for _, out := range outputs {
	    fmt.Fprint(f, out)
	}
	fmt.Fprintln(f, "```")

	fmt.Printf("Test report generated at %s\n", reportPath)
}
