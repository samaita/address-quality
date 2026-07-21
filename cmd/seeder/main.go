// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2025 Samaita

package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"

	"address-quality/internal/config"
	"address-quality/internal/database"
	"address-quality/internal/logger"
)

var tupleRe = regexp.MustCompile(`\('((?:[^']|'')*)',\s*'((?:[^']|'')*)'\)`)

func main() {
	sourceCode := flag.String("source-code", "kemendagri", "Source code identifier")
	sourceVersion := flag.String("source-version", "2025", "Dataset version tag")
	sourceName := flag.String("source-name", "Kepmendagri No 300.2.2-2138", "Human-readable source name")
	sourceDate := flag.String("source-date", "", "Effective date of the codes")
	sourceDesc := flag.String("source-desc", "", "Description of the source dataset")
	dropFlag := flag.Bool("drop", false, "Drop all tables (requires confirmation)")
	initFlag := flag.Bool("init", false, "Create schema from db/location.sql (only when no tables exist)")
	truncateFlag := flag.Bool("truncate", false, "Truncate all data rows (keep schema) before seeding")
	normalizeFlag := flag.Bool("normalize", false, "Rebuild lowercase_normalized column for all existing location_codes")
	dbPathFlag := flag.String("db", "", "Path to location.db (default from config)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: seeder [flags]

Seeds location.db from db/source/wilayah.sql and db/source/wilayah_kodepos.sql.

The source files are MySQL dumps of Indonesian administrative regions and postal codes.
The seeder parses them, determines hierarchy levels from kode patterns, normalizes names,
and batch-inserts into location_codes.

First run:   seeder --init
Reset:        seeder --drop && seeder --init
Retry:        seeder --truncate
Recalc:       seeder --normalize
Update data:  seeder            (tables must already exist)

Flags:
`)
		flag.PrintDefaults()
	}

	flag.Parse()

	flags := 0
	if *dropFlag {
		flags++
	}
	if *initFlag {
		flags++
	}
	if *truncateFlag {
		flags++
	}
	if *normalizeFlag {
		flags++
	}
	if flags > 1 {
		fmt.Fprintln(os.Stderr, "error: --drop, --init, --truncate, and --normalize are mutually exclusive")
		os.Exit(1)
	}

	cfg := config.Load()
	logger.Init(cfg.LogLevel)

	dbPath := *dbPathFlag
	if dbPath == "" {
		dbPath = cfg.LocationDBPath
	}

	ctx := context.Background()

	repo, err := database.NewLocationDB(dbPath, cfg.DBMaxOpenConns)
	if err != nil {
		logger.Fatal().Err(err).Msg("open location db")
	}

	hasTables, err := repo.HasLocationTables(ctx)
	if err != nil {
		logger.Fatal().Err(err).Msg("check tables")
	}

	if *initFlag {
		if hasTables {
			logger.Fatal().Msg("tables already exist, use --drop to recreate")
		}
		logger.Info().Msg("running db/location.sql...")
		schema, err := os.ReadFile("db/location.sql")
		if err != nil {
			logger.Fatal().Err(err).Msg("read db/location.sql")
		}
		if err := repo.ExecSchema(ctx, string(schema)); err != nil {
			logger.Fatal().Err(err).Msg("exec schema")
		}
		logger.Info().Msg("schema created")
	} else if *dropFlag {
		if !hasTables {
			logger.Fatal().Msg("no tables to drop, use --init for first-time setup")
		}
		if !promptConfirm("This will drop all location tables. Continue? [y/N]: ") {
			fmt.Fprintln(os.Stderr, "aborted")
			os.Exit(1)
		}
		logger.Info().Msg("dropping all tables...")
		if err := repo.DropAll(ctx); err != nil {
			logger.Fatal().Err(err).Msg("drop all")
		}
		logger.Info().Msg("tables dropped")
		return
	} else if *truncateFlag {
		if !hasTables {
			logger.Fatal().Msg("no tables found, use --init for first-time setup")
		}
		logger.Info().Msg("truncating all data...")
		if err := repo.TruncateAll(ctx); err != nil {
			logger.Fatal().Err(err).Msg("truncate")
		}
	} else if *normalizeFlag {
		if !hasTables {
			logger.Fatal().Msg("no tables found, use --init for first-time setup")
		}
		logger.Info().Msg("rebuilding lowercase_normalized...")
		if err := repo.RebuildNormalized(ctx); err != nil {
			logger.Fatal().Err(err).Msg("rebuild normalized")
		}
		logger.Info().Msg("normalize rebuild complete")
		return
	} else if !hasTables {
		fmt.Fprintln(os.Stderr, "Location tables not found. Initialize the schema by running:")
		fmt.Fprintln(os.Stderr, "  bin/seeder --init")
		os.Exit(1)
	}

	sourceID, err := repo.InsertLocationSource(ctx, *sourceCode, *sourceVersion, *sourceName, *sourceDate, *sourceDesc)
	if err != nil {
		logger.Fatal().Err(err).Msg("insert source")
	}
	logger.Info().Int64("location_source_id", sourceID).Str("code", *sourceCode).Str("version", *sourceVersion).Msg("source created")

	logger.Info().Msg("parsing db/source/wilayah.sql...")
	rows, err := parseWilayah("db/source/wilayah.sql")
	if err != nil {
		logger.Fatal().Err(err).Msg("parse wilayah")
	}
	logger.Info().Int("count", len(rows)).Msg("parsed wilayah rows")

	logger.Info().Msg("parsing db/source/wilayah_kodepos.sql...")
	kodeposMap, err := parseKodepos("db/source/wilayah_kodepos.sql")
	if err != nil {
		logger.Fatal().Err(err).Msg("parse kodepos")
	}
	logger.Info().Int("count", len(kodeposMap)).Msg("parsed kodepos entries")

	logger.Info().Msg("joining postal codes...")
	postalCount := 0
	for i := range rows {
		if pc, ok := kodeposMap[rows[i].Kode]; ok {
			rows[i].PostalCode = pc
			postalCount++
		}
	}
	logger.Info().Int("count", postalCount).Msg("joined postal codes")

	batchSize := 500
	total := len(rows)
	for i := 0; i < total; i += batchSize {
		end := i + batchSize
		if end > total {
			end = total
		}
		if err := repo.InsertLocationCodeBatch(ctx, sourceID, rows[i:end]); err != nil {
			logger.Fatal().Err(err).Int("start", i).Int("end", end).Msg("batch insert")
		}
		logger.Info().Int("inserted", end).Int("total", total).Msg("batch progress")
	}

	logger.Info().Msg("seeding complete")

	logger.Info().Msg("rebuilding location hierarchy...")
	if err := repo.RebuildLocationHierarchy(ctx, sourceID); err != nil {
		logger.Fatal().Err(err).Msg("rebuild hierarchy")
	}
	logger.Info().Msg("hierarchy rebuild complete")
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

func promptConfirm(prompt string) bool {
	fmt.Fprint(os.Stderr, prompt)
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		resp := strings.TrimSpace(strings.ToLower(scanner.Text()))
		return resp == "y" || resp == "yes"
	}
	return false
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
