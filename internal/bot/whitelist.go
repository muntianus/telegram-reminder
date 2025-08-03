package bot

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	tb "gopkg.in/telebot.v3"
	"telegram-reminder/internal/logger"
)

// ChatInfo stores information about a whitelisted chat
type ChatInfo struct {
	ID       int64     `json:"id"`
	Type     string    `json:"type"`     // "private", "group", "supergroup", "channel"
	Title    string    `json:"title"`    // Chat title for groups, username for private
	Username string    `json:"username"` // Username if available
	AddedAt  time.Time `json:"added_at"`
	Active   bool      `json:"active"` // Whether chat is active for broadcasts
}

var (
	wlMu         sync.RWMutex
	whitelistID  []int64             // Legacy support
	chatRegistry map[int64]*ChatInfo // Enhanced chat management
)

func init() {
	chatRegistry = make(map[int64]*ChatInfo)
	// Migration: convert existing whitelist entries to new format
	migrateExistingChats()
}

// migrateExistingChats converts legacy whitelist entries to new chat registry
func migrateExistingChats() {
	// This will be called during startup to migrate any existing data
	for _, id := range whitelistID {
		if _, exists := chatRegistry[id]; !exists {
			chatRegistry[id] = &ChatInfo{
				ID:      id,
				Type:    "private", // Default assumption for legacy entries
				Title:   fmt.Sprintf("Legacy Chat %d", id),
				AddedAt: time.Now(),
				Active:  true,
			}
		}
	}
}

// LoadWhitelist returns the list of whitelisted chat IDs.
// In-memory storage is used, so no errors are expected.
func LoadWhitelist() ([]int64, error) {
	wlMu.RLock()
	ids := append([]int64(nil), whitelistID...)
	wlMu.RUnlock()
	return ids, nil
}

// AddChatToWhitelist adds a chat with full information to the whitelist
func AddChatToWhitelist(chat *tb.Chat) error {
	wlMu.Lock()
	defer wlMu.Unlock()

	chatType := getChatTypeString(chat.Type)
	title := getChatTitle(chat)
	username := getChatUsername(chat)

	// Check if chat already exists
	if existing, exists := chatRegistry[chat.ID]; exists {
		// Update existing chat info
		existing.Type = chatType
		existing.Title = title
		existing.Username = username
		existing.Active = true
		logger.L.Info("chat updated", "id", chat.ID, "type", chatType, "title", title)
		return nil
	}

	// Add new chat
	chatInfo := &ChatInfo{
		ID:       chat.ID,
		Type:     chatType,
		Title:    title,
		Username: username,
		AddedAt:  time.Now(),
		Active:   true,
	}

	chatRegistry[chat.ID] = chatInfo

	// Legacy support - add to old whitelist array
	for _, v := range whitelistID {
		if v == chat.ID {
			logger.L.Info("chat added to registry", "id", chat.ID, "type", chatType, "title", title)
			return nil
		}
	}
	whitelistID = append(whitelistID, chat.ID)

	logger.L.Info("chat added", "id", chat.ID, "type", chatType, "title", title)
	return nil
}

// AddIDToWhitelist stores the chat ID if it is not already present (legacy support)
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

	// Add to new registry if not exists
	if _, exists := chatRegistry[id]; !exists {
		chatRegistry[id] = &ChatInfo{
			ID:      id,
			Type:    "private", // Default assumption for legacy
			Title:   fmt.Sprintf("Chat %d", id),
			AddedAt: time.Now(),
			Active:  true,
		}
	}

	wlMu.Unlock()
	logger.L.Debug("whitelist add", "id", id)
	return nil
}

// RemoveIDFromWhitelist removes the chat ID from the list.
// If the ID is not present, the operation succeeds silently.
func RemoveIDFromWhitelist(id int64) error {
	return DeactivateChat(id)
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

// FormatChatList returns a formatted list of all registered chats with details
func FormatChatList() string {
	wlMu.RLock()
	defer wlMu.RUnlock()

	if len(chatRegistry) == 0 {
		return "📭 Список чатов пуст"
	}

	var result strings.Builder
	result.WriteString("📋 Активные чаты:\n\n")

	for _, chat := range chatRegistry {
		if !chat.Active {
			continue
		}

		var icon string
		switch chat.Type {
		case "private":
			icon = "👤"
		case "group":
			icon = "👥"
		case "supergroup":
			icon = "🏢"
		case "channel":
			icon = "📢"
		default:
			icon = "💬"
		}

		title := chat.Title
		if title == "" {
			title = fmt.Sprintf("Chat %d", chat.ID)
		}

		result.WriteString(fmt.Sprintf("%s %s\n", icon, title))
		result.WriteString(fmt.Sprintf("   ID: <code>%d</code>\n", chat.ID))
		result.WriteString(fmt.Sprintf("   Тип: %s\n", getChatTypeRussian(chat.Type)))
		if chat.Username != "" {
			result.WriteString(fmt.Sprintf("   @%s\n", chat.Username))
		}
		result.WriteString(fmt.Sprintf("   Добавлен: %s\n\n", chat.AddedAt.Format("02.01.2006 15:04")))
	}

	return result.String()
}

// GetActiveChats returns all active chat IDs for broadcasting
func GetActiveChats() ([]int64, error) {
	wlMu.RLock()
	defer wlMu.RUnlock()

	var activeIDs []int64
	for _, chat := range chatRegistry {
		if chat.Active {
			activeIDs = append(activeIDs, chat.ID)
		}
	}

	// Fallback to legacy whitelist if registry is empty
	if len(activeIDs) == 0 {
		return append([]int64(nil), whitelistID...), nil
	}

	return activeIDs, nil
}

// Helper functions
func getChatTypeString(chatType tb.ChatType) string {
	switch chatType {
	case tb.ChatPrivate:
		return "private"
	case tb.ChatGroup:
		return "group"
	case tb.ChatSuperGroup:
		return "supergroup"
	case tb.ChatChannel:
		return "channel"
	default:
		return "unknown"
	}
}

func getChatTitle(chat *tb.Chat) string {
	if chat.Title != "" {
		return chat.Title
	}
	if chat.FirstName != "" || chat.LastName != "" {
		return strings.TrimSpace(chat.FirstName + " " + chat.LastName)
	}
	if chat.Username != "" {
		return "@" + chat.Username
	}
	return fmt.Sprintf("Chat %d", chat.ID)
}

func getChatUsername(chat *tb.Chat) string {
	return chat.Username
}

func getChatTypeRussian(chatType string) string {
	switch chatType {
	case "private":
		return "Личный"
	case "group":
		return "Группа"
	case "supergroup":
		return "Супергруппа"
	case "channel":
		return "Канал"
	default:
		return "Неизвестный"
	}
}

// DeactivateChat marks a chat as inactive without removing it
func DeactivateChat(id int64) error {
	wlMu.Lock()
	defer wlMu.Unlock()

	if chat, exists := chatRegistry[id]; exists {
		chat.Active = false
		logger.L.Info("chat deactivated", "id", id, "title", chat.Title)
	}

	// Also remove from legacy whitelist
	out := whitelistID[:0]
	for _, v := range whitelistID {
		if v != id {
			out = append(out, v)
		}
	}
	whitelistID = out

	return nil
}

// GetChatStats returns statistics about registered chats
func GetChatStats() map[string]int {
	wlMu.RLock()
	defer wlMu.RUnlock()

	stats := map[string]int{
		"total":      0,
		"active":     0,
		"private":    0,
		"group":      0,
		"supergroup": 0,
		"channel":    0,
	}

	for _, chat := range chatRegistry {
		stats["total"]++
		if chat.Active {
			stats["active"]++
		}
		stats[chat.Type]++
	}

	return stats
}
