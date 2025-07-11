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
				"ipa":        "həˈloʊ",
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

func TestDBStore_FindSimilar(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "bbolthelper_findsimilar_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test_findsimilar.db")
	store, err := NewDBStore(Config{
		DBPath:     dbPath,
		BucketName: "TestFindSimilarBucket",
		Logger:     zap.NewNop(),
	})
	if err != nil {
		t.Fatalf("NewDBStore() failed: %v", err)
	}
	defer store.Close()

	// Populate with test data including frequency and length variations.
	testData := []struct {
		word string
		freq string
	}{
		{"develop", "100"},      // len 7
		{"development", "80"},   // len 11
		{"developer", "90"},     // len 9
		{"devel", "70"},         // len 5, freq lower than develop
		{"test", "200"},
		{"testing", "150"},
		{"apple", "300"},
		{"apply", "250"},
		// For testing tie-breaking and limits
		{"cat", "50"},
		{"bat", "50"},
		{"rat", "50"},
		{"mat", "50"},
	}
	for _, item := range testData {
		if err := store.Put(item.word, map[string]string{"frq": item.freq}); err != nil {
			t.Fatalf("Failed to put test word '%s': %v", item.word, err)
		}
	}

	tests := []struct {
		name        string
		inputWord   string
		maxDistance int
		want        []string
		wantErr     bool
	}{
		{
			name:        "finds substitution, frq/len sort",
			inputWord:   "develp", // dist 1 to develop, devel
			maxDistance: 1,
			// "devel" (70) < "develop" (100), so "devel" comes first.
			want:    []string{"devel", "develop"},
			wantErr: false,
		},
		{
			name:        "finds single char deletion",
			inputWord:   "deveoper", // dist 1 to developer
			maxDistance: 1,
			want:        []string{"developer"},
			wantErr:     false,
		},
		{
			name:        "no match found within distance",
			inputWord:   "xyz",
			maxDistance: 1,
			want:        []string{}, // Expect empty slice
			wantErr:     false,
		},
		{
			name:        "exact match is distance 0, should not be returned",
			inputWord:   "apple",
			maxDistance: 1,
			want:        []string{"apply"}, // "apply" is dist 1, "apple" (dist 0) is excluded
			wantErr:     false,
		},
		{
			name:        "more than 3 results, should be truncated to top 3",
			inputWord:   "dat", // dist 1 to cat, bat, rat, mat
			maxDistance: 1,
			// All have same freq (50) and len (3), so order is not guaranteed.
			// We just check that we get 3 results.
			want:    []string{"cat", "bat", "rat"}, // Order might vary, will check length
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.FindSimilar(tt.inputWord, tt.maxDistance)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindSimilar() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// For the case with more than 3 results, we only check the length
			// because the order of items with identical sort keys is not guaranteed.
			if tt.name == "more than 3 results, should be truncated to top 3" {
				if len(got) != 3 {
					t.Errorf("FindSimilar() for '%s' got %d results, want 3", tt.inputWord, len(got))
				}
				// Create a map for quick lookups
				wantMap := make(map[string]bool)
				for _, w := range []string{"cat", "bat", "rat", "mat"} {
					wantMap[w] = true
				}
				for _, g := range got {
					if !wantMap[g] {
						t.Errorf("FindSimilar() got unexpected word '%s'", g)
					}
				}
			} else {
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("FindSimilar() got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
