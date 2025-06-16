package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/suchasplus/ne/internal/bbolthelper"
	"github.com/urfave/cli/v3"
	"go.uber.org/zap"
)

// JsonResult is used for structuring the JSON output
type JsonResult struct {
	Term  string            `json:"term"`
	Data  map[string]string `json:"data,omitempty"`
	Error string            `json:"error,omitempty"`
}

func main() {
	// Logger will be initialized based on the verbose flag inside the Action func

	var dbPathFlag string
	var bucketNameFlag string
	var verboseFlag bool
	var jsonFlag bool
	var fullOutputFlag bool

	cmd := &cli.Command{
		Name:      "ne",
		Usage:     "Reads a term from a bbolt key-value store using ecdict.",
		ArgsUsage: "<term>",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "verbose",
				Aliases:     []string{"v"},
				Usage:       "Enable verbose logging output",
				Destination: &verboseFlag,
			},
			&cli.BoolFlag{
				Name:        "json",
				Aliases:     []string{"j", "q"},
				Usage:       "Output result as JSON",
				Destination: &jsonFlag,
			},
			&cli.BoolFlag{
				Name:        "full",
				Aliases:     []string{"f"},
				Usage:       "Show full map output in plain text (if not JSON)",
				Destination: &fullOutputFlag,
			},
			&cli.StringFlag{
				Name:        "dbpath",
				Aliases:     []string{"d"},
				Usage:       fmt.Sprintf("Path to the bbolt database file. If not set, searches in PATH, then $HOME/.cache/ne/%s", bbolthelper.DefaultDBPath),
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
			var logger *zap.Logger
			if verboseFlag {
				logger = zap.NewExample()
			} else {
				logger = zap.NewNop()
			}
			defer logger.Sync()

			if cCtx.NArg() == 0 {
				cli.ShowAppHelpAndExit(cCtx, 1)
				return fmt.Errorf("error: search key argument is required")
			}
			searchKey := strings.ToLower(cCtx.Args().First())

			actualDBPath := dbPathFlag
			if actualDBPath == "" {
				resolvedPath, err := resolveDefaultDBPathForNe(bbolthelper.DefaultDBPath)
				if err != nil {
					logger.Error("Failed to find database file", zap.Error(err))
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					return err // Or cli.Exit for cleaner exit code handling
				}
				actualDBPath = resolvedPath
				logger.Info("Using resolved database path", zap.String("path", actualDBPath))
			}

			actualBucketName := bucketNameFlag
			if actualBucketName == "" {
				actualBucketName = bbolthelper.DefaultBucketName
			}

			logger.Info("Attempting to read key from bbolt database",
				zap.String("key", searchKey),
				zap.String("dbPath", actualDBPath),
				zap.String("bucketName", actualBucketName),
			)

			storeConfig := bbolthelper.Config{
				DBPath:     actualDBPath,
				BucketName: actualBucketName,
				FileMode:   bbolthelper.DefaultDBFileMode, // Ensure correct file mode
				ReadOnly:   true,
				Logger:     logger,
			}
			dbStore, err := bbolthelper.NewDBStore(storeConfig)
			if err != nil {
				logger.Error("Failed to open database store", zap.Error(err))
				fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
				return err
			}
			defer dbStore.Close()

			valueMap, found, err := dbStore.Get(searchKey)
			if err != nil {
				msg := "Error retrieving key"
				if jsonFlag {
					jsonResult := JsonResult{Term: searchKey, Error: fmt.Sprintf("%s: %v", msg, err)}
					jsonValue, _ := json.Marshal(jsonResult)
					fmt.Println(string(jsonValue))
				} else {
					fmt.Printf("%s '%s': %v\n", msg, searchKey, err)
				}
				logger.Error(msg, zap.String("key", searchKey), zap.Error(err))
				return err
			}

			if !found {
				msg := "term not found"
				if jsonFlag {
					jsonResult := JsonResult{Term: searchKey, Error: msg}
					jsonValue, _ := json.Marshal(jsonResult)
					fmt.Println(string(jsonValue))
				} else {
					fmt.Printf("%s '%s' in bucket '%s' of database '%s'.\n", msg, searchKey, actualBucketName, actualDBPath)
				}
				logger.Warn(msg, zap.String("key", searchKey), zap.String("dbPath", actualDBPath), zap.String("bucket", actualBucketName))
				return nil // Not an error for the CLI if key simply not found
			}

			if jsonFlag {
				jsonResult := JsonResult{Term: searchKey, Data: valueMap}
				jsonValue, jErr := json.MarshalIndent(jsonResult, "", "  ")
				if jErr != nil {
					// This error is about JSON marshaling, not finding the key
					logger.Error("Failed to marshal JSON output", zap.Error(jErr))
					fmt.Fprintf(os.Stderr, "Error generating JSON: %v\n", jErr)
					return jErr
				}
				fmt.Println(string(jsonValue))
			} else {
				// 2-column table output using lipgloss/table
				const keyColumnWidth = 15
				const valueColumnWidth = 60 // Adjusted for table borders/padding

				t := table.New().
					BorderBottom(true).
					BorderRow(true).
					Width(keyColumnWidth + valueColumnWidth + 3). // Total width approx
					Border(lipgloss.NormalBorder()).              // Use double-line border
					StyleFunc(func(row, col int) lipgloss.Style {
						// Basic padding for cells
						style := lipgloss.NewStyle().Padding(0, 1)
						// The table's Border will handle line drawing.
						// We can apply specific styles for headers or other special cells if needed.
						if col == 0 { // Key column
							return style.Width(keyColumnWidth)
						}
						return style.Width(valueColumnWidth) // Value column
					})

				var rowsData [][]string
				// Prepare data for table
				rowsData = append(rowsData, []string{"term", searchKey})

				displayFields := []string{"translation", "definition", "exchange"}
				if fullOutputFlag {
					// Collect all keys from valueMap and sort them for consistent order
					allKeys := make([]string, 0, len(valueMap))
					for k := range valueMap {
						if k != "term" { // Exclude term if already added, though it's not typically in valueMap here
							allKeys = append(allKeys, k)
						}
					}
					sort.Strings(allKeys) // Sort for consistent output
					displayFields = allKeys
				}

				for _, fieldKey := range displayFields {
					if val, ok := valueMap[fieldKey]; ok {
						processedVal := strings.ReplaceAll(val, "\\n", "\n")
						processedVal = strings.ReplaceAll(processedVal, "\\r", "\r") // Ensure \r is also processed
						processedVal = strings.ReplaceAll(processedVal, "\\t", "\t")
						// Only add to rowsData if the processed value is not empty after trimming whitespace
						if strings.TrimSpace(processedVal) != "" {
							rowsData = append(rowsData, []string{fieldKey, processedVal})
						}
					}
				}

				t.Rows(rowsData...)

				if len(rowsData) > 0 {
					fmt.Println(t.Render())
				} else {
					fmt.Println("No data to display for term after filtering.")
				}
			}
			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		// The logger might not be initialized if error occurs before Action
		// or if the error is from cli parsing itself.
		fmt.Fprintf(os.Stderr, "Error running command: %v\n", err)
		os.Exit(1)
	}
}

// resolveDefaultDBPathForNe searches for the database file in standard locations.
func resolveDefaultDBPathForNe(dbName string) (string, error) {
	// 1. Check directories in PATH
	pathEnv := os.Getenv("PATH")
	for _, dir := range filepath.SplitList(pathEnv) {
		dbPath := filepath.Join(dir, dbName)
		if _, err := os.Stat(dbPath); err == nil {
			return dbPath, nil // Found
		}
	}

	// 2. Check ~/.cache/ne/
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	cachePath := filepath.Join(homeDir, ".cache", "ne", dbName)
	if _, err := os.Stat(cachePath); err == nil {
		return cachePath, nil // Found
	}

	return "", fmt.Errorf("'%s' not found in PATH directories or in %s", dbName, filepath.Join("$HOME", ".cache", "ne"))
}
