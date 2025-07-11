package bbolthelper

import (
	"bytes"
	"encoding/csv"
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/agnivade/levenshtein"
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

// FindSimilar searches for words with a similar spelling to the input word.
// It uses the Levenshtein distance to measure similarity and includes performance optimizations.
// The logic is as follows:
// 1. Find all words with a Levenshtein distance of 1.
// 2. Stop searching if more than 10 suggestions are found.
// 3. Sort suggestions: primarily by frequency (desc), secondarily by length (desc).
// 4. If more than 3 suggestions are found, return the top 3. Otherwise, return all.
func (s *DBStore) FindSimilar(word string, maxDistance int) ([]string, error) {
	// suggestion struct holds data for sorting candidates.
	type suggestion struct {
		word string
		freq int
		len  int
	}
	var suggestions []suggestion

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(s.bucketName))
		if b == nil {
			return fmt.Errorf("bucket '%s' not found during FindSimilar operation", s.bucketName)
		}

		c := b.Cursor()
		inputLen := len(word)

		for k, v := c.First(); k != nil; k, v = c.Next() {
			// Stop searching if we have enough candidates.
			if len(suggestions) > 10 {
				break
			}

			dbWord := string(k)

			// Length pruning: if the length difference is greater than the max distance,
			// the Levenshtein distance must also be greater.
			if abs(len(dbWord)-inputLen) > maxDistance {
				continue
			}

			dist := levenshtein.ComputeDistance(word, dbWord)

			if dist > 0 && dist <= maxDistance {
				// Deserialize to get frequency.
				valueMap, err := Deserialize(v)
				if err != nil {
					s.logger.Warn("Failed to deserialize value for suggestion, skipping.", zap.String("word", dbWord), zap.Error(err))
					continue
				}

				freqStr, _ := valueMap["frq"]
				freq, _ := strconv.Atoi(freqStr) // Atoi returns 0 on error, which is acceptable here.

				suggestions = append(suggestions, suggestion{
					word: dbWord,
					freq: freq,
					len:  len(dbWord),
				})
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort the suggestions.
	sort.Slice(suggestions, func(i, j int) bool {
		if suggestions[i].freq != suggestions[j].freq {
			return suggestions[i].freq < suggestions[j].freq // Lower frq value first (higher frequency)
		}
		return suggestions[i].len > suggestions[j].len // Longer word first for ties
	})

	// Limit the number of results.
	if len(suggestions) > 3 {
		suggestions = suggestions[:3]
	}

	// Extract just the words to return.
	resultWords := make([]string, len(suggestions))
	for i, sug := range suggestions {
		resultWords[i] = sug.word
	}

	return resultWords, nil
}

// abs returns the absolute value of an integer.
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
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

			key := strings.ToLower(record[0])
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
		tempDB.Close()                                                  // Ensure tempDB is closed.
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
	if err := tempDB.Close(); err != nil { // Close tempDB after successful copy.
		return fmt.Errorf("failed to close temp DB '%s' after copy: %w. The DBStore is now closed.", tempDBPath, err)
	}
	if err := originalDBReadOnly.Close(); err != nil { // Close original read-only DB.
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
