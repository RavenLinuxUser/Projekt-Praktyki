package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type Product struct {
	CompanyID string  `json:"company_id"` // our internal product identifier (unique)
	Kind      string  `json:"kind"`       // category, e.g. "desk lamp"
	Price     float64 `json:"price"`      // numeric price
}

type Store interface {
	Add(p Product)
	All() []Product
	FilterByPrice(op string, value float64) []Product
}

type memStore struct {
	mu      sync.RWMutex
	records []Product
}

func NewMemStore() *memStore { return &memStore{} }

func (s *memStore) Add(p Product) {
	s.mu.Lock()
	s.records = append(s.records, p)
	s.mu.Unlock()
}

func (s *memStore) All() []Product {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cpy := make([]Product, len(s.records))
	copy(cpy, s.records)
	return cpy
}

// op can be "<", "=", ">"
func (s *memStore) FilterByPrice(op string, v float64) []Product {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Product
	for _, p := range s.records {
		switch op {
		case "<":
			if p.Price < v {
				out = append(out, p)
			}
		case "=":
			if p.Price == v {
				out = append(out, p)
			}
		case ">":
			if p.Price > v {
				out = append(out, p)
			}
		}
	}
	return out
}

func loadCSVDir(dir string, st Store) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err // abort walk on error
		}
		if info.IsDir() || !strings.HasSuffix(strings.ToLower(info.Name()), ".csv") {
			return nil // skip non‑CSV files
		}
		return loadSingleCSV(path, st)
	})
}

func loadSingleCSV(filePath string, st Store) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open %s: %w", filePath, err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.TrimLeadingSpace = true

	// Header row → map column name → index
	header, err := r.Read()
	if err != nil {
		return fmt.Errorf("read header %s: %w", filePath, err)
	}
	idx := map[string]int{}
	for i, col := range header {
		idx[strings.TrimSpace(strings.ToLower(col))] = i
	}

	// Required columns – adjust if your CSV uses different headings
	required := []string{"companyid", "kind", "price"}
	for _, col := range required {
		if _, ok := idx[col]; !ok {
			return fmt.Errorf("%s missing required column %q", filePath, col)
		}
	}

	// Parse rows
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read row %s: %w", filePath, err)
		}
		priceStr := strings.ReplaceAll(row[idx["price"]], "$", "")
		priceStr = strings.TrimSpace(priceStr)
		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			// Silently skip rows with bad price values
			continue
		}
		p := Product{
			CompanyID: row[idx["companyid"]],
			Kind:      row[idx["kind"]],
			Price:     price,
		}
		st.Add(p)
	}
	return nil
}

func printTable(products []Product) {
	if len(products) == 0 {
		fmt.Println("No matching products.")
		return
	}
	// Simple fixed‑width table
	fmt.Printf("%-15s %-20s %-15s\n",
		"CompanyID", "Kind", "Price")
	fmt.Println(strings.Repeat("-", 85))
	for _, p := range products {
		fmt.Printf("%-15s %-20s %10.2f\n",
			p.CompanyID, p.Kind, p.Price)
	}
}

func main() {
	// ----- Flags -----
	dirFlag := flag.String("dir", "", "Directory containing CSV files")
	opFlag := flag.String("op", "", "Price comparison operator: <, =, or >")
	priceFlag := flag.Float64("price", 0, "Numeric price to compare against")
	jsonOut := flag.Bool("json", false, "Output matches as JSON instead of a table")
	flag.Parse()

	if *dirFlag == "" {
		log.Fatal("please provide -dir flag pointing at the CSV folder")
	}
	if *opFlag == "" || (*opFlag != "<" && *opFlag != "=" && *opFlag != ">") {
		log.Fatal(`-op must be one of "<", "=", ">"`)
	}
	if *priceFlag <= 0 {
		log.Fatal("-price must be a positive number")
	}

	// ----- Load CSVs -----
	store := NewMemStore()
	if err := loadCSVDir(*dirFlag, store); err != nil {
		log.Fatalf("error loading CSVs: %v", err)
	}
	log.Printf("loaded %d total records", len(store.All()))

	// ----- Apply filter -----
	matches := store.FilterByPrice(*opFlag, *priceFlag)

	// ----- Output -----
	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", " ")
		if err := enc.Encode(matches); err != nil {
			log.Fatalf("JSON encode error: %v", err)
		}
	} else {
		printTable(matches)
	}
}
