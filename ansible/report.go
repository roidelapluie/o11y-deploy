package ansible

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type HostStatus struct {
	Processed int `json:"processed"`
	Failures  int `json:"failures"`
	Ok        int `json:"ok"`
	Dark      int `json:"dark"`
	Changed   int `json:"changed"`
	Skipped   int `json:"skipped"`
	Rescued   int `json:"rescued"`
	Ignored   int `json:"ignored"`
}

func readAndPrintJSONReport(filename string) error {
	// Read the JSON file
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	// Unmarshal the JSON data into a map
	var statuses map[string]HostStatus
	if err := json.Unmarshal(data, &statuses); err != nil {
		return err
	}

	// ANSI escape codes for colors
	red := "\033[31m"
	green := "\033[32m"
	reset := "\033[0m"

	// Print header
	fmt.Printf("%-25s  %-10s  %-10s  %-10s  %-12s  %-10s  %-10s  %-10s %-10s\n",
		"Host", "Processed", "Failures", "Ok", "Unreachable", "Changed", "Skipped", "Rescued", "Ignored")

	// Iterate over each host and print its status
	for host, status := range statuses {
		fmt.Printf("%-25s  ", host)
		fmt.Printf("%-10d  ", status.Processed)
		if status.Failures > 0 {
			fmt.Printf("%s%-10d%s  ", red, status.Failures, reset)
		} else {
			fmt.Printf("%-10d  ", status.Failures)
		}
		if status.Ok > 0 {
			fmt.Printf("%s%-10d%s  ", green, status.Ok, reset)
		} else {
			fmt.Printf("%-10d  ", status.Ok)
		}
		if status.Dark > 0 {
			fmt.Printf("%s%-12d%s  ", red, status.Dark, reset)
		} else {
			fmt.Printf("%-12d  ", status.Dark)
		}
		fmt.Printf("%-10d  ", status.Changed)
		fmt.Printf("%-10d  ", status.Skipped)
		fmt.Printf("%-10d  ", status.Rescued)
		fmt.Printf("%-10d\n", status.Ignored)
	}

	return nil
}
