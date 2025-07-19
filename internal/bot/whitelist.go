package bot

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"sync"
)

// WhitelistFile — путь к JSON-файлу со списком чатов (может быть переопределён в тестах).
var WhitelistFile = envDefault("WHITELIST_FILE", "whitelist.json")

// wlMu — мьютекс для потокобезопасной работы с файлом whitelist.
var wlMu sync.Mutex

// loadWhitelist читает список ID из файла whitelist.
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

// saveWhitelist сохраняет список ID в файл whitelist.
func saveWhitelist(ids []int64) error {
	wlMu.Lock()
	defer wlMu.Unlock()

	data, err := json.Marshal(ids)
	if err != nil {
		return err
	}
	return os.WriteFile(WhitelistFile, data, 0644)
}

// LoadWhitelist возвращает список whitelisted chat ID.
func LoadWhitelist() ([]int64, error) {
	return loadWhitelist()
}

// AddIDToWhitelist добавляет ID в whitelist, если его там нет.
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

// RemoveIDFromWhitelist удаляет ID из whitelist.
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

// FormatWhitelist возвращает список ID в виде строки с переводами строк.
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
