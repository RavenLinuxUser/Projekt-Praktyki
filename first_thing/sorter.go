package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
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


func readCSV(path string) (headers []string, rows []map[string]string) {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("Failed to open CSV %q: %v", path, err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		log.Fatalf("Failed to parse CSV %q: %v", path, err)
	}
	if len(records) == 0 {
		log.Fatalf("CSV %q appears empty", path)
	}

	headers = records[0]

	for _, rec := range records[1:] {
		rows = append(rows, rowAsMap(headers, rec))
	}
	return headers, rows
}


func findCheapest(rows []map[string]string, colName string) map[string]string {
	if colName == "" || len(rows) == 0 {
		return rows[0]
	}

	var cheapest map[string]string
	minVal := 0.0
	first := true

	for _, row := range rows {
		valStr, ok := row[colName]
		if !ok {
			continue 
		}
		val, err := strconv.ParseFloat(valStr, 64)
		if err != nil {
			continue
		}
		if first || val < minVal {
			minVal = val
			cheapest = row
			first = false
		}
	}
	if cheapest == nil {
		
		return rows[0]
	}
	return cheapest
}

func main() {
	
	csvPaths := flag.String("csv", "", "Comma‑separated list of CSV file paths (required)")
	field := flag.String("field", "", "Column name to filter on (optional)")
	value := flag.String("value", "", "Value to match for the given column (optional)")
	cheapest := flag.String("cheapest", "", "Numeric column name to pick the cheapest row (optional)")
	flag.Parse()

	if *csvPaths == "" {
		log.Fatalf("Missing required -csv flag")
	}

	
	files := splitAndTrim(*csvPaths, ",")

	var allHeaders []string
	var allRows []map[string]string

	
	for i, path := range files {
		headers, rows := readCSV(path)

		
		if i == 0 {
			allHeaders = headers
		} else if !sameHeaders(allHeaders, headers) {
			log.Fatalf("Header mismatch between %q and %q", files[0], path)
		}
		allRows = append(allRows, rows...)
	}

	
	if *field != "" && *value != "" {
		filtered := make([]map[string]string, 0, len(allRows))
		for _, row := range allRows {
			if v, ok := row[*field]; ok && v == *value {
				filtered = append(filtered, row)
			}
		}
		allRows = filtered
	}

	
	if *cheapest != "" && len(allRows) > 0 {
		cheapestRow := findCheapest(allRows, *cheapest)
		printRow(allHeaders, cheapestRow)
		return
	}

	
	printRow(allHeaders, rowAsMap(allHeaders, allHeaders))

	for _, row := range allRows {
		printRow(allHeaders, row)
	}
}




func splitAndTrim(s, sep string) []string {
	raw := strings.Split(s, sep)
	out := make([]string, 0, len(raw))
	for _, part := range raw {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}


func sameHeaders(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
