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
var WhitelistFile = "whitelist.json"
var wlMu sync.Mutex

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
func LoadWhitelist() ([]int64, error) {
	return loadWhitelist()
}

// AddIDToWhitelist stores the chat ID if it is not already present.
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
