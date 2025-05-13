package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"
	"go.uber.org/zap"

	"github.com/suchasplus/ne/internal/bbolthelper"
)

const (
	defaultCsvDir  = "./assets"
	defaultCsvFile = "ecdict.csv"

	progressReportInterval = 50000
)

func main() {
	logger := zap.NewExample()
	defer logger.Sync() // flushes buffer, if any

	var csvPathFlag string
	var dbPathFlag string
	var bucketNameFlag string

	cmd := &cli.Command{
		Name:  "kvbuilder-importer",
		Usage: "Imports data from a CSV file into a bbolt key-value store.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "csv",
				Aliases:     []string{"c"},
				Usage:       fmt.Sprintf("Load CSV from `FILE_PATH`. Defaults to %s/%s or %s", defaultCsvDir, defaultCsvFile, defaultCsvFile),
				Destination: &csvPathFlag,
			},
			&cli.StringFlag{
				Name:        "dbpath",
				Aliases:     []string{"d"},
				Usage:       fmt.Sprintf("Path to bbolt DB. If unset, searches PATH. If not found, defaults to $HOME/.cache/ne/%s (will be created if needed).", bbolthelper.DefaultDBPath),
				Destination: &dbPathFlag,
			},
			&cli.StringFlag{
				Name:        "bucket",
				Aliases:     []string{"b"},
				Usage:       fmt.Sprintf("Name of the bucket within the bbolt database. Defaults to '%s'", bbolthelper.DefaultBucketName),
				Destination: &bucketNameFlag,
			},
		},
		Action: func(ctx context.Context, cCtx *cli.Command) error {
			// Determine actual CSV path
			actualCsvPath := csvPathFlag
			if actualCsvPath == "" {
				path1 := filepath.Join(defaultCsvDir, defaultCsvFile)
				if _, err := os.Stat(path1); err == nil {
					actualCsvPath = path1
				} else {
					path2 := defaultCsvFile
					if _, err := os.Stat(path2); err == nil {
						actualCsvPath = path2
					} else {
						return fmt.Errorf("default CSV file not found in '%s' or current directory, and --csv flag not provided", defaultCsvDir)
					}
				}
			} else {
				if _, err := os.Stat(actualCsvPath); err != nil {
					return fmt.Errorf("specified CSV file '%s' not found or not accessible: %w", actualCsvPath, err)
				}
			}
			logger.Info("Using CSV file", zap.String("path", actualCsvPath))

			// Determine DB path and bucket name
			actualDBPath := dbPathFlag
			if actualDBPath == "" {
				resolvedPath, err := resolveDefaultDBPathForKvBuilder(bbolthelper.DefaultDBPath, logger)
				if err != nil {
					logger.Error("Failed to resolve or prepare default database path", zap.Error(err))
					fmt.Fprintf(os.Stderr, "Error resolving DB path: %v\n", err)
					return err
				}
				actualDBPath = resolvedPath
				logger.Info("Using database path", zap.String("path", actualDBPath))
			}

			actualBucketName := bucketNameFlag
			if actualBucketName == "" {
				actualBucketName = bbolthelper.DefaultBucketName
			}

			logger.Info("Target database settings",
				zap.String("dbPath", actualDBPath),
				zap.String("bucketName", actualBucketName),
			)

			storeConfig := bbolthelper.Config{
				DBPath:     actualDBPath,
				BucketName: actualBucketName,
				Logger:     logger,
				// FileMode will use DefaultDBFileMode from bbolthelper
				// ReadOnly will be false by default
			}
			store, err := bbolthelper.NewDBStore(storeConfig)
			if err != nil {
				return fmt.Errorf("failed to initialize db store: %w", err)
			}

			// NewDBStore already opens the database, so no explicit store.Open() is needed.
			defer store.Close() // Ensure DB is closed even if subsequent steps fail

			logger.Info("Starting import process...")
			recordsProcessed, err := store.ImportFromCSV(actualCsvPath, progressReportInterval)
			if err != nil {
				return fmt.Errorf("failed to import data from CSV '%s': %w", actualCsvPath, err)
			}
			logger.Info("Import process completed successfully.",
				zap.Int("recordsProcessed", recordsProcessed),
				zap.String("outputDB", actualDBPath),
			)

			logger.Info("Starting database compaction...")
			if err := store.Compact(bbolthelper.DefaultTempDBPath); err != nil {
				return fmt.Errorf("failed to compact database: %w", err)
			}
			logger.Info("Database compaction completed.")
			logger.Info("Process completed successfully.")
			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		// Logger might not be initialized if error is from CLI parsing.
		fmt.Fprintf(os.Stderr, "Error running kvbuilder-importer: %v\n", err)
		os.Exit(1)
	}
}

// resolveDefaultDBPathForKvBuilder searches for the database file in PATH first.
// If not found, it defaults to $HOME/.cache/ne/DB_NAME, ensuring the directory exists.
func resolveDefaultDBPathForKvBuilder(dbName string, logger *zap.Logger) (string, error) {
	// 1. Check directories in PATH
	pathEnv := os.Getenv("PATH")
	for _, dir := range filepath.SplitList(pathEnv) {
		dbPath := filepath.Join(dir, dbName)
		if fi, err := os.Stat(dbPath); err == nil && !fi.IsDir() { // Check if it exists and is a file
			logger.Info("Found existing database in PATH", zap.String("path", dbPath))
			return dbPath, nil
		}
	}

	// 2. Default to ~/.cache/ne/DB_NAME
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	cacheDir := filepath.Join(homeDir, ".cache", "ne")
	dbPathInCache := filepath.Join(cacheDir, dbName)

	logger.Info("Database not found in PATH, defaulting to cache directory.", zap.String("path", dbPathInCache))

	// Ensure the cache directory exists
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory '%s': %w", cacheDir, err)
	}

	return dbPathInCache, nil
}
