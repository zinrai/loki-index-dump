package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	version = "0.1.0"
	usage   = `loki-index-dump - Dump Loki labels and their values to JSON

Usage:
  loki-index-dump [options]

Options:
  -days        Number of days to look back (default: 30)
  -output      Output file path (default: metadata.json)
  -help        Show this help message
  -version     Show version

Examples:
  # Dump last 30 days of labels to metadata.json
  loki-index-dump

  # Dump last 7 days to a specific file
  loki-index-dump -days 7 -output weekly-index.json

Environment:
  LOKI_ADDR    Loki server address (required by logcli)
`
)

type IndexData struct {
	Labels   []string            `json:"labels"`
	Values   map[string][]string `json:"values"`
	DumpedAt time.Time           `json:"dumped_at"`
	Days     int                 `json:"days"`
}

func main() {
	var (
		days        int
		output      string
		showHelp    bool
		showVersion bool
	)

	flag.IntVar(&days, "days", 30, "Number of days to look back")
	flag.StringVar(&output, "output", "metadata.json", "Output file path")
	flag.BoolVar(&showHelp, "help", false, "Show help")
	flag.BoolVar(&showVersion, "version", false, "Show version")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
	}

	flag.Parse()

	if showHelp {
		fmt.Print(usage)
		os.Exit(0)
	}

	if showVersion {
		fmt.Printf("loki-index-dump version %s\n", version)
		os.Exit(0)
	}

	if err := dumpIndex(days, output); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func dumpIndex(days int, outputFile string) error {
	// Check LOKI_ADDR environment variable
	if os.Getenv("LOKI_ADDR") == "" {
		return fmt.Errorf("LOKI_ADDR environment variable is not set")
	}

	fmt.Printf("Dumping Loki index data for the last %d days...\n", days)

	// Check if logcli exists
	if _, err := exec.LookPath("logcli"); err != nil {
		return fmt.Errorf("logcli not found in PATH")
	}

	// Get all labels
	fmt.Println("Fetching labels...")
	labels, err := getLabels(days)
	if err != nil {
		return fmt.Errorf("failed to get labels: %w", err)
	}
	fmt.Printf("Found %d labels\n", len(labels))

	// Get values for each label
	indexData := &IndexData{
		Labels:   labels,
		Values:   make(map[string][]string),
		DumpedAt: time.Now(),
		Days:     days,
	}

	for i, label := range labels {
		fmt.Printf("[%d/%d] Fetching values for label: %s\n", i+1, len(labels), label)

		values, err := getLabelValues(label, days)
		if err != nil {
			// Don't fail entirely if one label fails
			fmt.Printf("  Warning: failed to get values for label %s: %v\n", label, err)
			continue
		}

		indexData.Values[label] = values
		fmt.Printf("  Found %d values\n", len(values))
	}

	// Save to file
	if err := saveIndexData(indexData, outputFile); err != nil {
		return fmt.Errorf("failed to save index data: %w", err)
	}

	fmt.Printf("\nIndex data dumped successfully!\n")
	fmt.Printf("Labels dumped: %d\n", len(indexData.Labels))
	fmt.Printf("Output file: %s\n", outputFile)

	return nil
}

func getLabels(days int) ([]string, error) {
	cmd := exec.Command("logcli", "labels", "--since", fmt.Sprintf("%dd", days))
	fmt.Printf("Executing: %s\n", strings.Join(cmd.Args, " "))

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("logcli labels failed: %w", err)
	}

	labels := []string{}
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "http") {
			labels = append(labels, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to parse labels: %w", err)
	}

	return labels, nil
}

func getLabelValues(label string, days int) ([]string, error) {
	cmd := exec.Command("logcli", "labels", label, "--since", fmt.Sprintf("%dd", days))
	fmt.Printf("  Executing: %s\n", strings.Join(cmd.Args, " "))

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("logcli labels %s failed: %w", label, err)
	}

	values := []string{}
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "http") {
			values = append(values, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to parse label values: %w", err)
	}

	return values, nil
}

func saveIndexData(data *IndexData, filename string) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal index data: %w", err)
	}

	if err := os.WriteFile(filename, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}
