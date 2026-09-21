// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

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
	postgresOnlyFlag := flag.Bool("postgres-only", false, "Operate on the Postgres connection only (requires POSTGRES_DSN)")
	dbPathFlag := flag.String("db", "", "Path to location.db (default from config); Postgres is seeded too when POSTGRES_DSN is set")

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

Postgres only: seeder --init --postgres-only   (requires POSTGRES_DSN)

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

	var targets []*database.LocationRepository
	if *postgresOnlyFlag {
		pgRepo, err := database.NewPostgresDB(cfg.PostgresDSN, cfg.DBMaxOpenConns)
		if err != nil {
			logger.Fatal().Err(err).Msg("open postgres")
		}
		if pgRepo == nil {
			logger.Fatal().Msg("--postgres-only requires POSTGRES_DSN")
		}
		targets = []*database.LocationRepository{pgRepo}
	} else {
		repo, err := database.NewLocationDB(dbPath, cfg.DBMaxOpenConns)
		if err != nil {
			logger.Fatal().Err(err).Msg("open location db")
		}
		targets = []*database.LocationRepository{repo}

		// Postgres is an additional target when configured.
		pgRepo, err := database.NewPostgresDB(cfg.PostgresDSN, cfg.DBMaxOpenConns)
		if err != nil {
			logger.Fatal().Err(err).Msg("open postgres")
		}
		if pgRepo != nil {
			targets = append(targets, pgRepo)
			logger.Info().Msg("postgres target enabled")
		}
	}

	hasTables, err := targets[0].HasLocationTables(ctx)
	if err != nil {
		logger.Fatal().Err(err).Msg("check tables")
	}

	if *initFlag {
		if hasTables {
			logger.Fatal().Msg("tables already exist, use --drop to recreate")
		}
		for _, t := range targets {
			execLocationSchema(ctx, t)
		}

		if !*postgresOnlyFlag {
			initAddressSchemas(ctx, cfg)
		}
	} else if *dropFlag {
		if !hasTables {
			logger.Fatal().Msg("no tables to drop, use --init for first-time setup")
		}
		if !promptConfirm("This will drop all location tables. Continue? [y/N]: ") {
			fmt.Fprintln(os.Stderr, "aborted")
			os.Exit(1)
		}
		logger.Info().Msg("dropping all tables...")
		for _, t := range targets {
			if err := t.DropAll(ctx); err != nil {
				logger.Fatal().Err(err).Msg("drop all")
			}
		}
		logger.Info().Msg("tables dropped")
		return
	} else if *truncateFlag {
		if !hasTables {
			logger.Fatal().Msg("no tables found, use --init for first-time setup")
		}
		logger.Info().Msg("truncating all data...")
		for _, t := range targets {
			if err := t.TruncateAll(ctx); err != nil {
				logger.Fatal().Err(err).Msg("truncate")
			}
		}
	} else if *normalizeFlag {
		if !hasTables {
			logger.Fatal().Msg("no tables found, use --init for first-time setup")
		}
		logger.Info().Msg("rebuilding lowercase_normalized...")
		for _, t := range targets {
			if err := t.RebuildNormalized(ctx); err != nil {
				logger.Fatal().Err(err).Msg("rebuild normalized")
			}
		}
		logger.Info().Msg("normalize rebuild complete")
		return
	} else if !hasTables {
		fmt.Fprintln(os.Stderr, "Location tables not found. Initialize the schema by running:")
		fmt.Fprintln(os.Stderr, "  bin/seeder --init")
		os.Exit(1)
	}

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

	// Assign explicit, shared ids so SQLite and Postgres store the same ids.
	// The primary target (SQLite, or Postgres under --postgres-only) sets the
	// offset; every target receives the same rows in the same order.
	baseID, err := targets[0].MaxLocationCodeID(ctx)
	if err != nil {
		logger.Fatal().Err(err).Msg("read max location code id")
	}
	for i := range rows {
		rows[i].ID = baseID + int64(i) + 1
	}

	sourceID, err := targets[0].NextLocationSourceID(ctx)
	if err != nil {
		logger.Fatal().Err(err).Msg("read next source id")
	}

	batchSize := 500
	total := len(rows)
	for _, t := range targets {
		storedSourceID, err := t.InsertLocationSource(ctx, sourceID, *sourceCode, *sourceVersion, *sourceName, *sourceDate, *sourceDesc)
		if err != nil {
			logger.Fatal().Err(err).Msg("insert source")
		}
		if storedSourceID != sourceID {
			logger.Warn().Int64("expected", sourceID).Int64("stored", storedSourceID).Msg("source id differs across targets")
		}
		logger.Info().Int64("location_source_id", storedSourceID).Bool("postgres", t.IsPostgres()).Str("code", *sourceCode).Str("version", *sourceVersion).Msg("source created")

		for i := 0; i < total; i += batchSize {
			end := i + batchSize
			if end > total {
				end = total
			}
			if err := t.InsertLocationCodeBatch(ctx, storedSourceID, rows[i:end]); err != nil {
				logger.Fatal().Err(err).Int("start", i).Int("end", end).Msg("batch insert")
			}
			logger.Info().Int("inserted", end).Int("total", total).Msg("batch progress")
		}

		logger.Info().Msg("rebuilding location hierarchy...")
		if err := t.RebuildLocationHierarchy(ctx, storedSourceID); err != nil {
			logger.Fatal().Err(err).Msg("rebuild hierarchy")
		}
		logger.Info().Msg("hierarchy rebuild complete")

		logger.Info().Msg("rebuilding city priority lookup...")
		if err := t.RebuildCityPriority(ctx, storedSourceID); err != nil {
			logger.Fatal().Err(err).Msg("rebuild city priority")
		}
		logger.Info().Msg("city priority rebuild complete")

		if err := t.ResetSequences(ctx); err != nil {
			logger.Fatal().Err(err).Msg("reset sequences")
		}
	}

	logger.Info().Msg("seeding complete")
}

// execLocationSchema runs the location schema for the target's dialect.
func execLocationSchema(ctx context.Context, t *database.LocationRepository) {
	file := "db/location.sql"
	if t.IsPostgres() {
		file = "db/location_postgres.sql"
	}
	logger.Info().Str("file", file).Msg("running location schema...")
	schema, err := os.ReadFile(file)
	if err != nil {
		logger.Fatal().Err(err).Str("file", file).Msg("read location schema")
	}
	if err := t.ExecSchema(ctx, string(schema)); err != nil {
		logger.Fatal().Err(err).Str("file", file).Msg("exec location schema")
	}
	logger.Info().Str("file", file).Msg("location schema created")
}

// initAddressSchemas creates the SQLite address and google_maps schemas.
func initAddressSchemas(ctx context.Context, cfg *config.Config) {
	addressRepo, err := database.New(cfg.AddressDBPath, cfg.DBMaxOpenConns)
	if err != nil {
		logger.Fatal().Err(err).Msg("open address db")
	}

	hasAddressTables, err := addressRepo.HasAddressTables(ctx)
	if err != nil {
		logger.Fatal().Err(err).Msg("check address tables")
	}
	if !hasAddressTables {
		logger.Info().Msg("running db/address.sql...")
		addressSchema, err := os.ReadFile("db/address.sql")
		if err != nil {
			logger.Fatal().Err(err).Msg("read db/address.sql")
		}
		if err := addressRepo.ExecSchema(ctx, string(addressSchema)); err != nil {
			logger.Fatal().Err(err).Msg("exec address schema")
		}
		logger.Info().Msg("address schema created")
	} else {
		logger.Info().Msg("address tables already exist")
	}

	hasGoogleMapsTable, err := addressRepo.HasGoogleMapsTable(ctx)
	if err != nil {
		logger.Fatal().Err(err).Msg("check google maps table")
	}
	if !hasGoogleMapsTable {
		logger.Info().Msg("running db/google_maps.sql...")
		googleMapsSchema, err := os.ReadFile("db/google_maps.sql")
		if err != nil {
			logger.Fatal().Err(err).Msg("read db/google_maps.sql")
		}
		if err := addressRepo.ExecSchema(ctx, string(googleMapsSchema)); err != nil {
			logger.Fatal().Err(err).Msg("exec google maps schema")
		}
		logger.Info().Msg("google maps schema created")
	} else {
		logger.Info().Msg("google maps table already exists")
	}
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
