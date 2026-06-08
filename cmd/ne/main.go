package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

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

// isChineseQuery returns true if the term contains any CJK Unicode character.
func isChineseQuery(term string) bool {
	for _, r := range term {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}

func main() {
	// Logger will be initialized based on the verbose flag inside the Action func

	var dbPathFlag string
	var cjkDBPathFlag string
	var bucketNameFlag string
	var verboseFlag bool
	var jsonFlag bool
	var fullOutputFlag bool
	var debugFlag bool

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
			&cli.BoolFlag{
				Name:        "debug",
				Usage:       "Enable fuzzy search debug statistics",
				Destination: &debugFlag,
			},
			&cli.StringFlag{
				Name:        "dbpath",
				Aliases:     []string{"d"},
				Usage:       fmt.Sprintf("Path to the bbolt database file. If not set, searches in PATH, then $HOME/.cache/ne/%s", bbolthelper.DefaultDBPath),
				Destination: &dbPathFlag,
			},
			&cli.StringFlag{
				Name:        "cjkdbpath",
				Usage:       fmt.Sprintf("Path to the CC-CEDICT bbolt database file. If not set, searches in PATH, then $HOME/.cache/ne/%s", bbolthelper.DefaultCedictDBPath),
				Destination: &cjkDBPathFlag,
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

			if isChineseQuery(searchKey) {
				return runChineseQuery(searchKey, cjkDBPathFlag, jsonFlag, fullOutputFlag, logger)
			}
			return runEnglishQuery(searchKey, dbPathFlag, bucketNameFlag, jsonFlag, fullOutputFlag, logger)
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error running command: %v\n", err)
		os.Exit(1)
	}
}

// runEnglishQuery handles English word lookup against ecdict.bbolt.
func runEnglishQuery(searchKey, dbPathFlag, bucketNameFlag string, jsonFlag, fullOutputFlag bool, logger *zap.Logger) error {
	actualDBPath := dbPathFlag
	if actualDBPath == "" {
		resolvedPath, err := resolveDefaultDBPath(bbolthelper.DefaultDBPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return err
		}
		actualDBPath = resolvedPath
		logger.Info("Using resolved database path", zap.String("path", actualDBPath))
	}

	actualBucketName := bucketNameFlag
	if actualBucketName == "" {
		actualBucketName = bbolthelper.DefaultBucketName
	}

	dbStore, err := bbolthelper.NewDBStore(bbolthelper.Config{
		DBPath:     actualDBPath,
		BucketName: actualBucketName,
		FileMode:   bbolthelper.DefaultDBFileMode,
		ReadOnly:   true,
		Logger:     logger,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		return err
	}
	defer dbStore.Close()

	valueMap, found, err := dbStore.Get(searchKey)
	if err != nil {
		return outputError(searchKey, fmt.Sprintf("Error retrieving key: %v", err), jsonFlag)
	}

	if !found {
		if !jsonFlag {
			fmt.Printf("Term '%s' not found. Searching for similar terms...\n", searchKey)
		}

		suggestions, err := dbStore.FindSimilar(searchKey, 1)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error during fuzzy search: %v\n", err)
			return err
		}

		if len(suggestions) == 0 {
			return outputError(searchKey, "term not found", jsonFlag)
		}

		if len(suggestions) > 1 {
			if jsonFlag {
				jsonResult := JsonResult{Term: searchKey, Data: map[string]string{"suggestions": strings.Join(suggestions, ", ")}}
				jsonValue, _ := json.MarshalIndent(jsonResult, "", "  ")
				fmt.Println(string(jsonValue))
			} else {
				fmt.Println("Did you mean one of these?")
				for _, s := range suggestions {
					fmt.Printf(" - %s\n", s)
				}
			}
			return nil
		}

		bestMatch := suggestions[0]
		if !jsonFlag {
			fmt.Printf("Did you mean '%s'?\n\n", bestMatch)
		}
		valueMap, found, err = dbStore.Get(bestMatch)
		if err != nil || !found {
			return outputError(bestMatch, "could not retrieve suggestion", jsonFlag)
		}
		searchKey = bestMatch
	}

	if jsonFlag {
		return printJSON(searchKey, valueMap)
	}

	displayFields := []string{"translation", "definition", "exchange"}
	if fullOutputFlag {
		allKeys := make([]string, 0, len(valueMap))
		for k := range valueMap {
			allKeys = append(allKeys, k)
		}
		sort.Strings(allKeys)
		displayFields = allKeys
	}
	printTable(searchKey, valueMap, displayFields)
	return nil
}

// runChineseQuery handles Chinese word lookup against cedict.bbolt.
func runChineseQuery(searchKey, cjkDBPathFlag string, jsonFlag, fullOutputFlag bool, logger *zap.Logger) error {
	actualDBPath := cjkDBPathFlag
	if actualDBPath == "" {
		resolvedPath, err := resolveDefaultDBPath(bbolthelper.DefaultCedictDBPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return err
		}
		actualDBPath = resolvedPath
		logger.Info("Using resolved cedict database path", zap.String("path", actualDBPath))
	}

	dbStore, err := bbolthelper.NewDBStore(bbolthelper.Config{
		DBPath:     actualDBPath,
		BucketName: bbolthelper.DefaultCedictBucketName,
		FileMode:   bbolthelper.DefaultDBFileMode,
		ReadOnly:   true,
		Logger:     logger,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening cedict database: %v\n", err)
		return err
	}
	defer dbStore.Close()

	valueMap, found, err := dbStore.Get(searchKey)
	if err != nil {
		return outputError(searchKey, fmt.Sprintf("Error retrieving key: %v", err), jsonFlag)
	}

	if !found {
		msg := fmt.Sprintf("Term '%s' not found in Chinese dictionary.", searchKey)
		if jsonFlag {
			return outputError(searchKey, "term not found", jsonFlag)
		}
		fmt.Println(msg)
		return nil
	}

	if jsonFlag {
		return printJSON(searchKey, valueMap)
	}

	// For Chinese results, show traditional only if different from simplified
	displayFields := []string{"pinyin", "definitions"}
	if fullOutputFlag {
		displayFields = []string{"traditional", "simplified", "pinyin", "definitions"}
	} else {
		if t, ok := valueMap["traditional"]; ok && t != valueMap["simplified"] {
			displayFields = append([]string{"traditional"}, displayFields...)
		}
	}
	printTable(searchKey, valueMap, displayFields)
	return nil
}

// outputError prints an error in json or plain text format.
func outputError(term, msg string, jsonFlag bool) error {
	if jsonFlag {
		jsonResult := JsonResult{Term: term, Error: msg}
		jsonValue, _ := json.Marshal(jsonResult)
		fmt.Println(string(jsonValue))
	} else {
		fmt.Printf("%s\n", msg)
	}
	return nil
}

// printJSON marshals and prints the result as indented JSON.
func printJSON(term string, valueMap map[string]string) error {
	jsonResult := JsonResult{Term: term, Data: valueMap}
	jsonValue, jErr := json.MarshalIndent(jsonResult, "", "  ")
	if jErr != nil {
		fmt.Fprintf(os.Stderr, "Error generating JSON: %v\n", jErr)
		return jErr
	}
	fmt.Println(string(jsonValue))
	return nil
}

// printTable renders a 2-column lipgloss table for the given fields.
func printTable(term string, valueMap map[string]string, displayFields []string) {
	const keyColumnWidth = 15
	const valueColumnWidth = 60

	t := table.New().
		BorderBottom(true).
		BorderRow(true).
		Width(keyColumnWidth + valueColumnWidth + 3).
		Border(lipgloss.NormalBorder()).
		StyleFunc(func(row, col int) lipgloss.Style {
			style := lipgloss.NewStyle().Padding(0, 1)
			if col == 0 {
				return style.Width(keyColumnWidth)
			}
			return style.Width(valueColumnWidth)
		})

	var rowsData [][]string
	rowsData = append(rowsData, []string{"term", term})

	for _, fieldKey := range displayFields {
		if val, ok := valueMap[fieldKey]; ok {
			processedVal := strings.ReplaceAll(val, "\\n", "\n")
			processedVal = strings.ReplaceAll(processedVal, "\\r", "\r")
			processedVal = strings.ReplaceAll(processedVal, "\\t", "\t")
			// For definitions field, replace "/" separators with newlines for readability
			if fieldKey == "definitions" {
				processedVal = strings.ReplaceAll(processedVal, "/", "\n")
			}
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

// resolveDefaultDBPath searches for the database file in standard locations.
func resolveDefaultDBPath(dbName string) (string, error) {
	// 1. Check directories in PATH
	pathEnv := os.Getenv("PATH")
	for _, dir := range filepath.SplitList(pathEnv) {
		dbPath := filepath.Join(dir, dbName)
		if _, err := os.Stat(dbPath); err == nil {
			return dbPath, nil
		}
	}

	// 2. Check ~/.cache/ne/
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	cachePath := filepath.Join(homeDir, ".cache", "ne", dbName)
	if _, err := os.Stat(cachePath); err == nil {
		return cachePath, nil
	}

	return "", fmt.Errorf("'%s' not found in PATH directories or in %s", dbName, filepath.Join("$HOME", ".cache", "ne"))
}
