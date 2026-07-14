package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"address-quality/internal/config"
	"address-quality/internal/database"
)

var tupleRe = regexp.MustCompile(`\('((?:[^']|'')*)',\s*'((?:[^']|'')*)'\)`)

func main() {
	sourceCode := flag.String("source-code", "kemendagri", "Source code identifier")
	sourceVersion := flag.String("source-version", "2025", "Dataset version tag")
	sourceName := flag.String("source-name", "Kepmendagri No 300.2.2-2138", "Human-readable source name")
	sourceDate := flag.String("source-date", "", "Effective date of the codes")
	sourceDesc := flag.String("source-desc", "", "Description of the source dataset")
	dropFlag := flag.Bool("drop", false, "Drop all tables and recreate from db/location.sql before seeding")
	truncateFlag := flag.Bool("truncate", false, "Truncate all data rows (keep schema) before seeding")
	dbPathFlag := flag.String("db", "", "Path to location.db (default from config)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: seeder [flags]

Seeds location.db from db/source/wilayah.sql and db/source/wilayah_kodepos.sql.

The source files are MySQL dumps of Indonesian administrative regions and postal codes.
The seeder parses them, determines hierarchy levels from kode patterns, normalizes names,
and batch-inserts into location_codes.

Flags:
`)
		flag.PrintDefaults()
	}

	flag.Parse()

	cfg := config.Load()
	dbPath := *dbPathFlag
	if dbPath == "" {
		dbPath = cfg.LocationDBPath
	}

	if *dropFlag && *truncateFlag {
		log.Printf("warning: both --drop and --truncate set, using --drop")
	}

	ctx := context.Background()

	repo, err := database.NewLocationDB(dbPath)
	if err != nil {
		log.Fatalf("open location db: %v", err)
	}

	if *dropFlag {
		log.Print("dropping all tables...")
		if err := repo.DropAll(ctx); err != nil {
			log.Fatalf("drop all: %v", err)
		}

		log.Print("running db/location.sql...")
		schema, err := os.ReadFile("db/location.sql")
		if err != nil {
			log.Fatalf("read db/location.sql: %v", err)
		}
		if err := repo.ExecSchema(ctx, string(schema)); err != nil {
			log.Fatalf("exec schema: %v", err)
		}
		log.Print("schema created")
	} else if *truncateFlag {
		log.Print("truncating all data...")
		if err := repo.TruncateAll(ctx); err != nil {
			log.Fatalf("truncate: %v", err)
		}
	}

	hasTables, err := repo.HasLocationTables(ctx)
	if err != nil {
		log.Fatalf("check tables: %v", err)
	}
	if !hasTables {
		fmt.Fprintln(os.Stderr, "Location tables not found. Initialize the schema by running:")
		fmt.Fprintln(os.Stderr, "  bin/seeder --drop")
		os.Exit(1)
	}

	sourceID, err := repo.InsertLocationSource(ctx, *sourceCode, *sourceVersion, *sourceName, *sourceDate, *sourceDesc)
	if err != nil {
		log.Fatalf("insert source: %v", err)
	}
	log.Printf("location_source_id=%d (%s/%s)", sourceID, *sourceCode, *sourceVersion)

	log.Print("parsing db/source/wilayah.sql...")
	rows, err := parseWilayah("db/source/wilayah.sql")
	if err != nil {
		log.Fatalf("parse wilayah: %v", err)
	}
	log.Printf("parsed %d wilayah rows", len(rows))

	log.Print("parsing db/source/wilayah_kodepos.sql...")
	kodeposMap, err := parseKodepos("db/source/wilayah_kodepos.sql")
	if err != nil {
		log.Fatalf("parse kodepos: %v", err)
	}
	log.Printf("parsed %d kodepos entries", len(kodeposMap))

	log.Print("joining postal codes...")
	postalCount := 0
	for i := range rows {
		if pc, ok := kodeposMap[rows[i].Kode]; ok {
			rows[i].PostalCode = pc
			postalCount++
		}
	}
	log.Printf("joined %d postal codes", postalCount)

	batchSize := 500
	total := len(rows)
	for i := 0; i < total; i += batchSize {
		end := i + batchSize
		if end > total {
			end = total
		}
		if err := repo.InsertLocationCodeBatch(ctx, sourceID, rows[i:end]); err != nil {
			log.Fatalf("batch insert [%d-%d]: %v", i, end, err)
		}
		log.Printf("inserted %d/%d rows", end, total)
	}

	log.Print("seeding complete")
}

func parseWilayah(path string) ([]database.LocationCodeRow, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var rows []database.LocationCodeRow
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		matches := tupleRe.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		kode := strings.ReplaceAll(matches[1], "''", "'")
		name := strings.ReplaceAll(matches[2], "''", "'")
		levelID := strings.Count(kode, ".") + 2

		rows = append(rows, database.LocationCodeRow{
			Kode:    kode,
			Name:    name,
			LevelID: levelID,
		})
	}

	return rows, scanner.Err()
}

func parseKodepos(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	results := make(map[string]string)
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		matches := tupleRe.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		kode := strings.ReplaceAll(matches[1], "''", "'")
		kp := strings.ReplaceAll(matches[2], "''", "'")
		if kp != "" {
			results[kode] = kp
		}
	}

	return results, scanner.Err()
}
