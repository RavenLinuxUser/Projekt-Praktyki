package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
)

func rowAsMap(headers, record []string) map[string]string {
	m := make(map[string]string, len(headers))
	for i, h := range headers {
		if i < len(record) {
			m[h] = record[i]
		} else {
			m[h] = ""
		}
	}
	return m
}
func printRow(headers []string, row map[string]string) {
	for i, h := range headers {
		if i > 0 {
			fmt.Print("\t")
		}
		fmt.Print(row[h])
	}
	fmt.Println()
}
func main() {
	csvPath := flag.String("csv", "", "Path to the CSV file (required)")
	field := flag.String("field", "", "Column name to filter on (optional)")
	value := flag.String("value", "", "Value to match for the given column (optional)")
	flag.Parse()
	if *csvPath == "" {
		log.Fatalf("Missing required -csv flag")
	}
	f, err := os.Open(*csvPath)
	if err != nil {
		log.Fatalf("Failed to open CSV: %v", err)
	}
	defer f.Close()
	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to parse CSV: %v", err)
	}
	if len(records) == 0 {
		log.Fatalf("CSV appears empty")
	}
	headers := records[0]
	printRow(headers, rowAsMap(headers, headers))
	for _, rec := range records[1:] {
		row := rowAsMap(headers, rec)
		if *field != "" && *value != "" {
			if v, ok := row[*field]; !ok || v != *value {
				continue
			}
		}
		printRow(headers, row)
	}
}
