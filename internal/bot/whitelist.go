package bot

import (
	"strconv"
	"strings"
	"sync"

	"telegram-reminder/internal/logger"
)

var (
	wlMu        sync.RWMutex
	whitelistID []int64
)

// LoadWhitelist returns the list of whitelisted chat IDs.
// In-memory storage is used, so no errors are expected.
func LoadWhitelist() ([]int64, error) {
	wlMu.RLock()
	ids := append([]int64(nil), whitelistID...)
	wlMu.RUnlock()
	return ids, nil
}

// AddIDToWhitelist stores the chat ID if it is not already present.
// Duplicate IDs are ignored silently.
func AddIDToWhitelist(id int64) error {
	wlMu.Lock()
	for _, v := range whitelistID {
		if v == id {
			wlMu.Unlock()
			logger.L.Debug("whitelist exists", "id", id)
			return nil
		}
	}
	whitelistID = append(whitelistID, id)
	wlMu.Unlock()
	logger.L.Debug("whitelist add", "id", id)
	return nil
}

// RemoveIDFromWhitelist removes the chat ID from the list.
// If the ID is not present, the operation succeeds silently.
func RemoveIDFromWhitelist(id int64) error {
	wlMu.Lock()
	out := whitelistID[:0]
	for _, v := range whitelistID {
		if v != id {
			out = append(out, v)
		}
	}
	whitelistID = out
	wlMu.Unlock()
	logger.L.Debug("whitelist remove", "id", id)
	return nil
}

// ResetWhitelist clears the in-memory list. Used in tests.
func ResetWhitelist() {
	wlMu.Lock()
	whitelistID = nil
	wlMu.Unlock()
}

// FormatWhitelist returns the IDs as a newline separated string.
// Returns an empty string if the list is empty.
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
