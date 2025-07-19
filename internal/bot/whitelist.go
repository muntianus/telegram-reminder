package bot

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"sync"
)

// WhitelistFile is the path to the JSON file that stores chat IDs.
// It can be overridden in tests.
var WhitelistFile = envDefault("WHITELIST_FILE", "whitelist.json")
var wlMu sync.Mutex

// loadWhitelist reads the whitelist file and returns the list of chat IDs.
// This is an internal function that handles the actual file I/O operations.
// Returns an empty slice if the file doesn't exist.
//
// Returns:
//   - []int64: Array of whitelisted chat IDs
//   - error: Any error that occurred during file reading or parsing
func loadWhitelist() ([]int64, error) {
	wlMu.Lock()
	defer wlMu.Unlock()

	data, err := os.ReadFile(WhitelistFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []int64{}, nil
		}
		return nil, err
	}
	var ids []int64
	if len(data) == 0 {
		return []int64{}, nil
	}
	if err := json.Unmarshal(data, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// saveWhitelist writes the list of chat IDs to the whitelist file.
// This is an internal function that handles the actual file I/O operations.
//
// Parameters:
//   - ids: Array of chat IDs to save
//
// Returns:
//   - error: Any error that occurred during file writing
func saveWhitelist(ids []int64) error {
	wlMu.Lock()
	defer wlMu.Unlock()

	data, err := json.Marshal(ids)
	if err != nil {
		return err
	}
	return os.WriteFile(WhitelistFile, data, 0644)
}

// LoadWhitelist returns the list of whitelisted chat IDs.
// This is the public interface for reading the whitelist.
//
// Returns:
//   - []int64: Array of whitelisted chat IDs
//   - error: Any error that occurred during loading
func LoadWhitelist() ([]int64, error) {
	return loadWhitelist()
}

// AddIDToWhitelist stores the chat ID if it is not already present.
// Duplicate IDs are ignored silently.
//
// Parameters:
//   - id: Chat ID to add to the whitelist
//
// Returns:
//   - error: Any error that occurred during the operation
func AddIDToWhitelist(id int64) error {
	ids, err := loadWhitelist()
	if err != nil {
		return err
	}
	for _, v := range ids {
		if v == id {
			return nil
		}
	}
	ids = append(ids, id)
	return saveWhitelist(ids)
}

// RemoveIDFromWhitelist removes the chat ID from the list.
// If the ID is not present, the operation succeeds silently.
//
// Parameters:
//   - id: Chat ID to remove from the whitelist
//
// Returns:
//   - error: Any error that occurred during the operation
func RemoveIDFromWhitelist(id int64) error {
	ids, err := loadWhitelist()
	if err != nil {
		return err
	}
	out := ids[:0]
	for _, v := range ids {
		if v != id {
			out = append(out, v)
		}
	}
	return saveWhitelist(out)
}

// FormatWhitelist returns the IDs as a newline separated string.
// Returns an empty string if the list is empty.
//
// Parameters:
//   - ids: Array of chat IDs to format
//
// Returns:
//   - string: Newline-separated list of chat IDs
func FormatWhitelist(ids []int64) string {
	if len(ids) == 0 {
		return ""
	}
	strs := make([]string, len(ids))
	for i, v := range ids {
		strs[i] = strconv.FormatInt(v, 10)
	}
	return strings.Join(strs, "\n")
}
