package config

import (
	"fmt"
	"os"
	"strconv"

	yaml "gopkg.in/yaml.v3"
)

const (
	DefaultBlockchainAPI = "https://api.blockchain.info/stats"
	DefaultWhitelistFile = "whitelist.json"
	DefaultLogLevel      = "info"
	DefaultConfigFile    = "config.yml"
)

// Config holds environment configuration values.
type Config struct {
	TelegramToken string
	ChatID        int64
	OpenAIKey     string
	OpenAIModel   string
	BlockchainAPI string
	LogLevel      string
	WhitelistFile string
}

type fileConfig struct {
	OpenAIModel   string `yaml:"openai_model"`
	BlockchainAPI string `yaml:"blockchain_api"`
	LogLevel      string `yaml:"log_level"`
	WhitelistFile string `yaml:"whitelist_file"`
}

// Load reads environment variables and validates them.
func Load() (Config, error) {
	var cfg Config
	cfg.BlockchainAPI = DefaultBlockchainAPI
	cfg.LogLevel = DefaultLogLevel
	cfg.WhitelistFile = DefaultWhitelistFile

	fc := fileConfig{}
	cfgPath := os.Getenv("CONFIG_FILE")
	if cfgPath == "" {
		cfgPath = DefaultConfigFile
	}
	if data, err := os.ReadFile(cfgPath); err == nil {
		_ = yaml.Unmarshal(data, &fc)
	}

	telegramToken := os.Getenv("TELEGRAM_TOKEN")
	chatIDStr := os.Getenv("CHAT_ID")
	openaiKey := os.Getenv("OPENAI_API_KEY")

	if telegramToken == "" || openaiKey == "" {
		return cfg, fmt.Errorf("missing required env vars")
	}

	var chatID int64
	if chatIDStr != "" {
		var err error
		chatID, err = strconv.ParseInt(chatIDStr, 10, 64)
		if err != nil {
			return cfg, fmt.Errorf("invalid CHAT_ID: %w", err)
		}
	}

	openaiModel := firstNonEmpty(os.Getenv("OPENAI_MODEL"), fc.OpenAIModel)
	blockchainAPI := firstNonEmpty(os.Getenv("BLOCKCHAIN_API"), fc.BlockchainAPI, DefaultBlockchainAPI)
	logLevel := firstNonEmpty(os.Getenv("LOG_LEVEL"), fc.LogLevel, DefaultLogLevel)
	whitelistFile := firstNonEmpty(os.Getenv("WHITELIST_FILE"), fc.WhitelistFile, DefaultWhitelistFile)

	cfg = Config{
		TelegramToken: telegramToken,
		ChatID:        chatID,
		OpenAIKey:     openaiKey,
		OpenAIModel:   openaiModel,
		BlockchainAPI: blockchainAPI,
		LogLevel:      logLevel,
		WhitelistFile: whitelistFile,
	}

	return cfg, nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
