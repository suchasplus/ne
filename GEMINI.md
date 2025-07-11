# Gemini Project Context

This file provides context for the Gemini AI assistant to understand the `ne` project.

## Project Overview

`ne` (stands for "玩转 English" - "Mastering English") is a command-line dictionary tool written in Go. It uses a local [BoltDB](https://github.com/etcd-io/bbolt) database for fast word lookups. The dictionary data is built from a source CSV file using a separate `kvbuilder` tool.

The core functionalities are split between two binaries:
1.  **`kvbuilder`**: A tool to create or update the BoltDB database from a large CSV file (`ecdict.csv`). It also compacts the database for optimal size.
2.  **`ne`**: The main CLI tool to query the database for a word and display its definition, translation, and other details in a formatted table.

## Tech Stack

-   **Language**: Go
-   **Database**: BoltDB (via `go.etcd.io/bbolt`) for a local key-value store.
-   **CLI Framework**: `github.com/urfave/cli/v3`
-   **Logging**: `go.uber.org/zap`
-   **UI/Output Formatting**:
    -   `github.com/charmbracelet/lipgloss/table` for creating formatted tables in the terminal.

## Project Structure

```
.
├── assets/
│   ├── ecdict.csv.xz   # The compressed dictionary data source. MUST be decompressed before use.
│   └── README.md       # Instructions on how to decompress the data file.
├── cmd/
│   ├── kvbuilder/      # The source for the 'kvbuilder' tool, which builds the database.
│   └── ne/             # The source for the 'ne' dictionary lookup tool.
└── internal/
    └── bbolthelper/    # A helper package to abstract BoltDB operations (open, close, get, put, import, compact).
```

-   **`cmd/kvbuilder/main.go`**: The entry point for the database builder tool.
-   **`cmd/ne/main.go`**: The main entry point for the dictionary lookup application. It takes the search term as a direct argument.
-   **`internal/bbolthelper/bbolthelper.go`**: Contains the `DBStore` struct and all logic for interacting with the BoltDB database.
-   **`assets/ecdict.csv.xz`**: The primary data source. It is compressed with `xz`.

## Development Workflow

### 1. Data Setup (One-time)

The dictionary data must be decompressed before it can be used to build the database.

```bash
# Navigate to the assets directory
cd assets

# Decompress the file
xz -d ecdict.csv.xz

# Navigate back to the project root
cd ..
```

### 2. Building the Applications

The project produces two executables.

```bash
# Build the database builder tool
go build -o kvbuilder ./cmd/kvbuilder

# Build the dictionary lookup tool
go build -o ne ./cmd/ne
```

### 3. Building the Database

Once the `kvbuilder` is built and the CSV is decompressed, create the BoltDB database.

**Usage:** `./kvbuilder [global options]`

**Example:**
```bash
# This command reads from assets/ecdict.csv and creates ecdict.bbolt
./kvbuilder --csv assets/ecdict.csv
```

**Global Options:**
-   `--csv FILE_PATH`, `-c FILE_PATH`: Load CSV from `FILE_PATH`.
-   `--dbpath string`, `-d string`: Path to bbolt DB.
-   `--bucket string`, `-b string`: Name of the bucket within the database.

The `kvbuilder` will automatically run a compaction operation upon completion.

### 4. Running the Application (Looking up words)

The `ne` tool takes the search term directly as an argument.

**Usage:** `./ne [global options] <term>`

**Example:**
```bash
# Look up a word using the 'ne' tool
./ne hello
```

**Global Options:**
-   `--verbose`, `-v`: Enable verbose logging output.
-   `--json`, `-j`, `-q`: Output result as JSON.
-   `--full`, `-f`: Show full map output in plain text (if not JSON).
-   `--dbpath string`, `-d string`: Path to the bbolt database file.
-   `--bucket string`, `-b string`: Name of the bucket within the database.

### 5. Running Tests

Standard Go test command.

```bash
go test ./...
```

---
---

## Project Replication Plan for Gemini

This section outlines a series of prompts and instructions for a Gemini CLI agent to replicate this project from scratch.

### Phase 1: Project Initialization

**Prompt 1.1: Create project structure**
"Create a new directory named `ne`, change into it, initialize a Go module, and create the necessary subdirectories."
```bash
mkdir ne && cd ne && go mod init github.com/suchasplus/ne && mkdir -p cmd/ne cmd/kvbuilder internal/bbolthelper assets
```

**Prompt 1.2: Create `.gitignore` file**
"Create a `.gitignore` file to exclude build artifacts and the local database file."
```bash
# Command to execute via write_file tool
# Filepath: .gitignore
# Content:
ne
kvbuilder
ecdict.bbolt
ecdict.bbolt.tmp
```

### Phase 2: Core Logic (`internal/bbolthelper`)

**Prompt 2.1: Write the database helper**
"Write the core database helper logic to `internal/bbolthelper/bbolthelper.go`."
```go
// Command to execute via write_file tool
// Filepath: internal/bbolthelper/bbolthelper.go
// Content:
package bbolthelper

import (
	"bytes"
	"encoding/csv"
	"encoding/gob"
	"fmt"
	"io"
	"os"

	bolt "go.etcd.io/bbolt"
	"go.uber.org/zap"
)

const (
	DefaultDBPath     = "ecdict.bbolt"
	DefaultTempDBPath = "ecdict.bbolt.tmp" // For compaction
	DefaultBucketName = "EcdictBucket"
	DefaultDBFileMode = os.FileMode(0644)
)

// DBStore manages interactions with a BoltDB database.
type DBStore struct {
	db         *bolt.DB
	logger     *zap.Logger
	dbPath     string
	bucketName string
	dbFileMode os.FileMode
}

// Config holds configuration for the DBStore.
type Config struct {
	DBPath     string
	BucketName string
	FileMode   os.FileMode
	ReadOnly   bool
	Logger     *zap.Logger
}

// NewDBStore creates or opens a BoltDB database and returns a DBStore instance.
func NewDBStore(cfg Config) (*DBStore, error) {
	if cfg.Logger == nil {
		// If no logger is provided, use a no-op logger to avoid nil panics.
		// Consumers can provide a configured zap.Logger if logging is desired.
		cfg.Logger = zap.NewNop()
	}
	if cfg.DBPath == "" {
		cfg.DBPath = DefaultDBPath
	}
	if cfg.BucketName == "" {
		cfg.BucketName = DefaultBucketName
	}
	if cfg.FileMode == 0 {
		cfg.FileMode = DefaultDBFileMode
	}

	opts := &bolt.Options{ReadOnly: cfg.ReadOnly}
	// Ensure Timeout is set if necessary, e.g., for NFS mounts, though not typically needed for local files.
	// opts.Timeout = 1 * time.Second

	db, err := bolt.Open(cfg.DBPath, cfg.FileMode, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open bbolt database '%s': %w", cfg.DBPath, err)
	}

	store := &DBStore{
		db:         db,
		logger:     cfg.Logger,
		dbPath:     cfg.DBPath,
		bucketName: cfg.BucketName,
		dbFileMode: cfg.FileMode,
	}

	// Ensure the bucket exists if not in read-only mode
	if !cfg.ReadOnly {
		err = db.Update(func(tx *bolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists([]byte(cfg.BucketName))
			if err != nil {
				return fmt.Errorf("failed to create bucket '%s': %w", cfg.BucketName, err)
			}
			return nil
		})
		if err != nil {
			db.Close() // Close DB if bucket creation fails
			return nil, fmt.Errorf("failed to ensure bucket '%s' exists: %w", cfg.BucketName, err)
		}
	}

	store.logger.Debug("DBStore initialized", zap.String("dbPath", store.dbPath), zap.String("bucketName", store.bucketName), zap.Bool("readOnly", cfg.ReadOnly))
	return store, nil
}

// Close closes the BoltDB database.
func (s *DBStore) Close() error {
	if s.db == nil {
		s.logger.Debug("Attempted to close an already nil DBStore.db")
		return nil
	}
	s.logger.Debug("Closing DBStore", zap.String("dbPath", s.dbPath))
	return s.db.Close()
}

// Serialize converts a map[string]string to a byte slice using gob.
func Serialize(data map[string]string) ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(data); err != nil {
		return nil, fmt.Errorf("failed to serialize data: %w", err)
	}
	return buf.Bytes(), nil
}

// Deserialize converts a byte slice back to a map[string]string using gob.
func Deserialize(data []byte) (map[string]string, error) {
	var result map[string]string
	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to deserialize data: %w", err)
	}
	return result, nil
}

// Get retrieves a value by key from the database.
// Returns the deserialized map, a boolean indicating if the key was found, and an error.
func (s *DBStore) Get(key string) (map[string]string, bool, error) {
	var valueMap map[string]string
	found := false

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(s.bucketName))
		if b == nil {
			return fmt.Errorf("bucket '%s' not found during Get operation", s.bucketName)
		}

		valBytes := b.Get([]byte(key))
		if valBytes == nil {
			return nil // Key not found, not an error for View
		}

		deserialized, err := Deserialize(valBytes)
		if err != nil {
			return fmt.Errorf("failed to deserialize value for key '%s': %w", key, err)
		}
		valueMap = deserialized
		found = true
		return nil
	})

	if err != nil {
		return nil, false, err
	}
	return valueMap, found, nil
}

// putCore performs the actual put operation for a serialized value within an existing transaction.
// It's an unexported method intended for internal use by Put and ImportFromCSV.
func (s *DBStore) putCore(tx *bolt.Tx, key string, serializedValue []byte) error {
	b := tx.Bucket([]byte(s.bucketName))
	if b == nil {
		// This might occur if the bucket was not created properly, though NewDBStore aims to prevent this.
		return fmt.Errorf("bucket '%s' not found during putCore operation", s.bucketName)
	}
	if err := b.Put([]byte(key), serializedValue); err != nil {
		return fmt.Errorf("failed to put key '%s' (serialized) into bucket '%s' in transaction: %w", key, s.bucketName, err)
	}
	return nil
}

// Put stores a key-value (map[string]string) pair into the database.
func (s *DBStore) Put(key string, valueMap map[string]string) error {
	serializedValue, err := Serialize(valueMap)
	if err != nil {
		return fmt.Errorf("failed to serialize value for key '%s' before Put: %w", key, err)
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		return s.putCore(tx, key, serializedValue) // Use the core put logic
	})
}

// ImportFromCSV reads records from a CSV file and stores them in the BoltDB database.
// It returns the number of records processed and an error if any occurred.
func (s *DBStore) ImportFromCSV(csvFilePath string, progressReportInterval int) (int, error) {
	s.logger.Info("Starting CSV import...", zap.String("sourceCsv", csvFilePath))

	csvFile, err := os.Open(csvFilePath)
	if err != nil {
		return 0, fmt.Errorf("failed to open CSV file '%s': %w", csvFilePath, err)
	}
	defer csvFile.Close()

	reader := csv.NewReader(csvFile)
	header, err := reader.Read() // Read the header row
	if err != nil {
		if err == io.EOF {
			return 0, fmt.Errorf("CSV file '%s' is empty or has no header", csvFilePath)
		}
		return 0, fmt.Errorf("failed to read header from CSV '%s': %w", csvFilePath, err)
	}

	if len(header) < 1 {
		return 0, fmt.Errorf("CSV file '%s' header is invalid (too few columns)", csvFilePath)
	}

	s.logger.Info("Processing CSV records...", zap.String("csvPath", csvFilePath))
	var recordsProcessed int

	err = s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(s.bucketName))
		if b == nil {
			// This should ideally not happen if NewDBStore correctly created the bucket.
			return fmt.Errorf("bucket '%s' unexpectedly not found during CSV import", s.bucketName)
		}

		for {
			record, err := reader.Read()
			if err == io.EOF {
				break // End of file
			}
			if err != nil {
				s.logger.Warn("Error reading record from CSV, skipping record.", zap.String("csvPath", csvFilePath), zap.Error(err))
				continue
			}

			if len(record) < 1 {
				s.logger.Warn("Empty record found in CSV, skipping.", zap.String("csvPath", csvFilePath))
				continue
			}

			key := record[0]
			valueMap := make(map[string]string)

			for i := 1; i < len(record); i++ {
				if i < len(header) {
					valueMap[header[i]] = record[i]
				} else {
					s.logger.Warn("Record has more columns than header, extra columns ignored.", zap.String("key", key), zap.String("csvPath", csvFilePath))
				}
			}

			// Serialize the valueMap for the current record
			serializedRecordValue, serErr := Serialize(valueMap)
			if serErr != nil {
				s.logger.Error("Failed to serialize record, skipping", zap.String("key", key), zap.Error(serErr))
				continue // Skip this record
			}

			// Use the DBStore's putCore method with the existing transaction
			if err := s.putCore(tx, key, serializedRecordValue); err != nil {
				// Log the error and decide whether to continue or stop the import.
				// For robustness, we'll log and skip the problematic record.
				// A more critical error (like transaction failure) would be returned by db.Update's main error.
				s.logger.Error("Failed to put record into DB using putCore, record skipped", zap.String("key", key), zap.Error(err))
				continue
			}
			recordsProcessed++
			if progressReportInterval > 0 && recordsProcessed%progressReportInterval == 0 {
				s.logger.Info("Processed records milestone", zap.Int("count", recordsProcessed))
			}
		}
		return nil // Return nil for the transaction func if loop completes without critical error
	})

	if err != nil {
		// This error comes from db.Update if the transaction itself failed (e.g., disk full, permissions)
		return recordsProcessed, fmt.Errorf("failed during bbolt transaction for CSV import: %w", err)
	}

	s.logger.Info("Successfully imported records from CSV.",
		zap.Int("totalRecords", recordsProcessed),
		zap.String("dbPath", s.dbPath),
		zap.String("bucketName", s.bucketName),
	)
	return recordsProcessed, nil
}

// Compact compacts the BoltDB database.
// It requires the DBStore to be re-initialized by the caller after compaction if it was not read-only,
// as this method closes the current DB instance and replaces the file.
// For a read-only DBStore, this operation is not directly applicable as it modifies the DB.
func (s *DBStore) Compact(tempDBPath string) error {
	if s.db == nil {
		return fmt.Errorf("cannot compact a closed or uninitialized DBStore")
	}
	if tempDBPath == "" {
		tempDBPath = DefaultTempDBPath
	}
	s.logger.Info("Starting database compaction",
		zap.String("originalDB", s.dbPath),
		zap.String("tempDB", tempDBPath),
	)

	// Store current configuration to potentially re-open/re-initialize later
	// However, the caller should generally handle re-initialization after compaction.
	originalPath := s.dbPath
	originalFileMode := s.dbFileMode
	// originalBucketName := s.bucketName
	// originalLogger := s.logger

	// 1. Close the current database instance managed by this DBStore.
	// This is crucial because compaction typically involves replacing the database file.
	err := s.db.Close()
	s.db = nil // Mark as closed to prevent further use of the old instance
	if err != nil {
		return fmt.Errorf("failed to close current database instance ('%s') before compaction: %w", originalPath, err)
	}
	s.logger.Info("Current database instance closed for compaction.", zap.String("dbPath", originalPath))

	// 2. Open the original database in read-only mode for copying its contents.
	originalDBReadOnly, err := bolt.Open(originalPath, originalFileMode, &bolt.Options{ReadOnly: true})
	if err != nil {
		// Attempt to reopen the original DB for the store if compaction setup fails
		// This part is tricky, as the state might be inconsistent. Best to return error.
		return fmt.Errorf("failed to open original DB '%s' as read-only for compaction: %w. The DBStore is now closed.", originalPath, err)
	}
	defer originalDBReadOnly.Close()

	// 3. Create/Open the temporary database for writing the compacted data.
	tempDB, err := bolt.Open(tempDBPath, originalFileMode, nil) // Default options for new DB
	if err != nil {
		return fmt.Errorf("failed to open temp DB '%s' for compaction: %w. The DBStore is now closed.", tempDBPath, err)
	}
	defer func() {
		tempDB.Close()
		if _, statErr := os.Stat(tempDBPath); !os.IsNotExist(statErr) { // If tempDB still exists (rename failed/not reached)
			s.logger.Info("Removing temporary database file after compaction attempt.", zap.String("tempDB", tempDBPath))
			os.Remove(tempDBPath)
		}
	}()

	// 4. Copy data from the read-only original to the new temporary database.
	s.logger.Info("Copying data to temporary database for compaction...", zap.String("from", originalPath), zap.String("to", tempDBPath))
	err = originalDBReadOnly.View(func(tx *bolt.Tx) error {
		// tx.CopyFile is bbolt's recommended way to compact.
		// It efficiently copies the database content to the file descriptor of an open bolt.DB.
		return tx.CopyFile(tempDB.Path(), originalFileMode)
	})
	if err != nil {
		return fmt.Errorf("failed to copy data from '%s' to '%s' during compaction: %w. The DBStore is now closed.", originalPath, tempDBPath, err)
	}
	s.logger.Info("Data copy for compaction successful.")

	// 5. Critical: Close both databases before rename/remove operations.
	if err := tempDB.Close(); err != nil {
		return fmt.Errorf("failed to close temp DB '%s' after copy: %w. The DBStore is now closed.", tempDBPath, err)
	}
	if err := originalDBReadOnly.Close(); err != nil {
		s.logger.Warn("Failed to close original read-only DB after copy (this is usually a defer, but checking explicitly)", zap.Error(err))
	}

	// 6. Remove the original (now old) database file.
	s.logger.Info("Removing original database file before replacing with compacted version.", zap.String("originalDB", originalPath))
	if err := os.Remove(originalPath); err != nil {
		return fmt.Errorf("failed to remove original DB '%s' to replace with compacted version: %w. The DBStore is now closed.", originalPath, err)
	}

	// 7. Rename the temporary (now compacted) database to the original database name.
	s.logger.Info("Renaming temporary (compacted) database to original name.", zap.String("tempDB", tempDBPath), zap.String("targetDB", originalPath))
	if err := os.Rename(tempDBPath, originalPath); err != nil {
		return fmt.Errorf("failed to rename temp DB '%s' to '%s': %w. The DBStore is now closed; manual recovery of '%s' might be needed.", tempDBPath, originalPath, err, tempDBPath)
	}

	s.logger.Info("Database compaction completed successfully. The underlying file has been replaced.", zap.String("dbPath", originalPath))
	s.logger.Info("The DBStore instance is now closed. Please re-initialize a new DBStore instance to use the compacted database.")
	// s.db remains nil. The caller is responsible for creating a new DBStore instance.
	return nil
}
```

**Prompt 2.2: Write the test for the helper**
"Write the corresponding test file `internal/bbolthelper/bbolthelper_test.go` to ensure the logic is sound."
```go
// Command to execute via write_file tool
// Filepath: internal/bbolthelper/bbolthelper_test.go
// Content:
package bbolthelper

import (
	"reflect"
	"testing"
	"os"
	"path/filepath"
	"go.uber.org/zap"
)

func TestSerializeDeserialize(t *testing.T) {
	tests := []struct {
		name    string
		input   map[string]string
		wantErr bool
	}{
		{
			name: "simple map",
			input: map[string]string{
				"word":       "hello",
				"definition": "a greeting",
				"ipa":        "hәˈloʊ",
			},
			wantErr: false,
		},
		{
			name:    "empty map",
			input:   map[string]string{},
			wantErr: false,
		},
		{
			name:    "nil map", // serialize should handle nil map gracefully if gob allows, otherwise expect error
			input:   nil,
			wantErr: false, // gob.Encode treats nil map as an empty map
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serialized, err := Serialize(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Serialize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return // If we expected an error, no need to deserialize
			}

			deserialized, err := Deserialize(serialized)
			if err != nil {
				t.Errorf("Deserialize() error = %v", err)
				return
			}

			expectedOutput := tt.input
			if tt.input == nil {
				// gob decodes a nil map (serialized as an empty map) into a non-nil empty map.
				expectedOutput = make(map[string]string)
			}

			if !reflect.DeepEqual(deserialized, expectedOutput) {
				t.Errorf("Deserialize() got = %v, want %v", deserialized, expectedOutput)
			}
		})
	}
}

func TestNewDBStore(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "bbolthelper_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name           string
		config         Config
		expectedDBPath string
		expectedBucket string
		wantErr        bool
	}{
		{
			name: "default config",
			config: Config{ // Logger will be set to NopLogger by NewDBStore if nil
				DBPath: filepath.Join(tempDir, "default.db"), // Use temp dir for test isolation
			},
			expectedDBPath: filepath.Join(tempDir, "default.db"),
			expectedBucket: DefaultBucketName,
			wantErr:        false,
		},
		{
			name: "custom config",
			config: Config{
				DBPath:     filepath.Join(tempDir, "custom.db"),
				BucketName: "MyCustomBucket",
				Logger:     zap.NewNop(), // Provide a Nop logger
			},
			expectedDBPath: filepath.Join(tempDir, "custom.db"),
			expectedBucket: "MyCustomBucket",
			wantErr:        false,
		},
		{
			name: "empty db path (should use default)",
			config: Config{
				DBPath:     "", // Test default path mechanism (will be DefaultDBPath)
				BucketName: "TestBucketForDefaultPath",
				Logger:     zap.NewNop(),
			},
			expectedDBPath: DefaultDBPath, // Actual default, not in tempDir for this case
			expectedBucket: "TestBucketForDefaultPath",
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// If we are testing the default DBPath, ensure we clean it up if it's created
			if tt.config.DBPath == "" {
				// Clean up default db path if it exists from a previous failed test or this test
				// This is a bit of a hack for testing default path creation outside tempDir
				defer os.Remove(DefaultDBPath)
			}

			store, err := NewDBStore(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDBStore() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			defer store.Close()

			if store.dbPath != tt.expectedDBPath {
				t.Errorf("NewDBStore() dbPath got = %v, want %v", store.dbPath, tt.expectedDBPath)
			}
			if store.bucketName != tt.expectedBucket {
				t.Errorf("NewDBStore() bucketName got = %v, want %v", store.bucketName, tt.expectedBucket)
			}

			// Check if db file was created
			dbFileToStat := store.dbPath
			if tt.name == "empty db path (should use default)" {
				dbFileToStat = DefaultDBPath // Use the actual default path for stat check
			}

			if _, statErr := os.Stat(dbFileToStat); os.IsNotExist(statErr) {
				t.Errorf("NewDBStore() did not create db file at %v", dbFileToStat)
			}

			// Special cleanup for the test case that uses the actual DefaultDBPath
			if tt.name == "empty db path (should use default)" {
				store.Close() // Close it before removing
				os.Remove(DefaultDBPath) // Explicitly remove the default db if created by this test case
			}
		})
	}
}

func TestDBStore_PutGet(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "bbolthelper_putget_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test_putget.db")
	store, err := NewDBStore(Config{
		DBPath:     dbPath,
		BucketName: "TestPutGetBucket",
		Logger:     zap.NewNop(),
	})
	if err != nil {
		t.Fatalf("NewDBStore() failed: %v", err)
	}
	defer store.Close()

	key1 := "testKey1"
	value1 := map[string]string{"data": "value1", "type": "string"}

	key2 := "testKey2"
	value2 := map[string]string{"data": "value2", "count": "42"}

	nonExistentKey := "nonExistentKey"

	t.Run("Put and Get existing key", func(t *testing.T) {
		if err := store.Put(key1, value1); err != nil {
			t.Fatalf("Put(%s) error = %v", key1, err)
		}
		if err := store.Put(key2, value2); err != nil {
			t.Fatalf("Put(%s) error = %v", key2, err)
		}

		retrievedVal1, found1, err1 := store.Get(key1)
		if err1 != nil {
			t.Errorf("Get(%s) error = %v", key1, err1)
		}
		if !found1 {
			t.Errorf("Get(%s) key not found, expected to be found", key1)
		}
		if !reflect.DeepEqual(retrievedVal1, value1) {
			t.Errorf("Get(%s) got = %v, want %v", key1, retrievedVal1, value1)
		}

		retrievedVal2, found2, err2 := store.Get(key2)
		if err2 != nil {
			t.Errorf("Get(%s) error = %v", key2, err2)
		}
		if !found2 {
			t.Errorf("Get(%s) key not found, expected to be found", key2)
		}
		if !reflect.DeepEqual(retrievedVal2, value2) {
			t.Errorf("Get(%s) got = %v, want %v", key2, retrievedVal2, value2)
		}
	})

	t.Run("Get non-existent key", func(t *testing.T) {
		retrievedVal, found, err := store.Get(nonExistentKey)
		if err != nil {
			t.Errorf("Get(%s) error = %v for non-existent key", nonExistentKey, err)
		}
		if found {
			t.Errorf("Get(%s) key found, expected not to be found", nonExistentKey)
		}
		if retrievedVal != nil {
			t.Errorf("Get(%s) got = %v, want nil for non-existent key", nonExistentKey, retrievedVal)
		}
	})

	t.Run("Put overwriting existing key", func(t *testing.T) {
		newValue1 := map[string]string{"data": "newValue1", "status": "updated"}
		if err := store.Put(key1, newValue1); err != nil {
			t.Fatalf("Put(%s) for overwrite error = %v", key1, err)
		}

		retrievedVal, found, err := store.Get(key1)
		if err != nil {
			t.Errorf("Get(%s) after overwrite error = %v", key1, err)
		}
		if !found {
			t.Errorf("Get(%s) after overwrite key not found, expected to be found", key1)
		}
		if !reflect.DeepEqual(retrievedVal, newValue1) {
			t.Errorf("Get(%s) after overwrite got = %v, want %v", key1, retrievedVal, newValue1)
		}
	})
}
```

### Phase 3: Application Code

**Prompt 3.1: Write the `kvbuilder` tool**
"Write the code for the database builder tool to `cmd/kvbuilder/main.go`."
```go
// Command to execute via write_file tool
// Filepath: cmd/kvbuilder/main.go
// Content:
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
```

**Prompt 3.2: Write the `ne` lookup tool**
"Write the code for the main dictionary lookup tool to `cmd/ne/main.go`."
```go
// Command to execute via write_file tool
// Filepath: cmd/ne/main.go
// Content:
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
```

### Phase 4: Dependencies and Data

**Prompt 4.1: Tidy Go modules**
"Now that all the Go code is in place, run `go mod tidy` to download all dependencies and clean up the `go.mod` and `go.sum` files."
```bash
go mod tidy
```

**Prompt 4.2: Download and compress dictionary data**
"Download the `ecdict.csv` file from its official source into the `assets` directory, and then compress it using `xz`."
```bash
curl -L -o assets/ecdict.csv https://raw.githubusercontent.com/skywind3000/ECDICT/master/ecdict.csv && xz assets/ecdict.csv
```

### Phase 5: Documentation

**Prompt 5.1: Create README for assets**
"Create a `README.md` in the `assets` directory explaining that the data file is compressed."
```bash
# Command to execute via write_file tool
# Filepath: assets/README.md
# Content:
# Asset Files

The `ecdict.csv` file has been compressed using `xz` to reduce the repository size.

## Decompression

To use the dictionary file, you need to decompress `ecdict.csv.xz` first.

### On macOS or Linux
```bash
xz -d ecdict.csv.xz
```
```

**Prompt 5.2: Create the main project README**
"Finally, create the main `README.md` for the project, explaining its purpose, features, and usage."
```markdown
// Command to execute via write_file tool
// Filepath: README.md
// Content:
# ne (玩转 English)

A blazingly fast command-line dictionary tool powered by Go and BoltDB.

`ne` (stands for "玩转 English" - "Mastering English") provides instant, offline access to a comprehensive English dictionary directly from your terminal.

## Features

-   **Offline First**: All lookups are performed locally. No internet connection required after initial setup.
-   **Extremely Fast**: Built on Go and using BoltDB, a high-performance key-value store, for near-instantaneous lookups.
-   **Simple & Clean UI**: Results are displayed in a clean, readable table format.
-   **Flexible Output**: Supports both human-readable tables and structured `JSON` output for scripting.
-   **Comprehensive Data**: Uses the extensive [ECDICT](https://github.com/skywind3000/ECDICT) dictionary data.

## Getting Started

### Prerequisites

-   **Go**: Version 1.24 or newer.
-   **xz**: A command-line tool for decompressing `.xz` files (e.g., `xz-utils` on Debian/Ubuntu, `xz` on macOS via Homebrew).

### Installation & Setup

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/suchasplus/ne.git
    cd ne
    ```

2.  **Decompress the Dictionary Data:**
    The dictionary source file comes compressed. You must decompress it before building the database.
    ```bash
    # Navigate to the assets directory
    cd assets

    # Decompress the file
    xz -d ecdict.csv.xz

    # Navigate back to the project root
    cd ..
    ```

3.  **Build the Tools:**
    This project uses two separate command-line tools: `kvbuilder` to build the database and `ne` to query it.
    ```bash
    # Build the database builder tool
    go build -o kvbuilder ./cmd/kvbuilder

    # Build the dictionary lookup tool
    go build -o ne ./cmd/ne
    ```
    You can move the `kvbuilder` and `ne` executables to a directory in your `$PATH` (e.g., `/usr/local/bin`) for easy access.

4.  **Build the Database:**
    Now, use the `kvbuilder` tool to create the local BoltDB database from the CSV file.
    ```bash
    # This command reads the CSV and creates the database file (ecdict.bbolt)
    ./kvbuilder --csv assets/ecdict.csv
    ```
    This process may take a minute. `kvbuilder` will create the `ecdict.bbolt` file in your current directory or in `$HOME/.cache/ne/` if it has permissions.

## Usage

To look up a word, simply pass it as an argument to the `ne` command.

**Syntax:**
```bash
./ne [options] <term>
```

**Options:**
-   `--json`, `-j`: Output the result in JSON format.
-   `--full`, `-f`: Show all available data fields for a term.
-   `--dbpath <path>`: Specify a custom path to the `ecdict.bbolt` database file.
-   `--verbose`, `-v`: Enable detailed logging.

## Examples

### Standard Lookup

A standard lookup displays the most common fields in a clean table.

```bash
$ ./ne hello

┌───────────────┬────────────────────────────────────────────────────────────┐
│ term          │ hello                                                      │
├───────────────┼────────────────────────────────────────────────────────────┤
│ translation   │ interj. 喂, 嘿                                             │
├───────────────┼────────────────────────────────────────────────────────────┤
│ definition    │ n. an expression of greeting                               │
├───────────────┼────────────────────────────────────────────────────────────┤
│ exchange      │ s:hellos                                                   │
└───────────────┴────────────────────────────────────────────────────────────┘
```

### JSON Output

For scripting or integration with other tools, you can output the full entry as a JSON object.

```bash
$ ./ne hello --json

{
  "term": "hello",
  "data": {
    "audio": "",
    "bnc": "2319",
    "collins": "3",
    "definition": "n. an expression of greeting",
    "detail": "",
    "exchange": "s:hellos",
    "frq": "2238",
    "oxford": "1",
    "phonetic": "hә'lәu",
    "pos": "",
    "tag": "zk gk",
    "translation": "interj. 喂, 嘿"
  }
}
```

## License

This project is licensed under the Apache License 2.0. See the [LICENSE](LICENSE) file for details.
```

Following these steps will result in a complete and functional replication of the `ne` project.
