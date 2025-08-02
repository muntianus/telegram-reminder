package domain

// DigestType represents different types of digests
type DigestType string

const (
	CryptoDigest     DigestType = "crypto"
	TechDigest       DigestType = "tech"
	RealEstateDigest DigestType = "realestate"
	BusinessDigest   DigestType = "business"
	InvestmentDigest DigestType = "investment"
	StartupDigest    DigestType = "startup"
	GlobalDigest     DigestType = "global"
)

// DigestConfig contains configuration for a digest type
type DigestConfig struct {
	Type        DigestType
	Name        string
	CommandName string
	Prompt      string
}

// GetDigestConfigs returns all available digest configurations
func GetDigestConfigs() map[DigestType]DigestConfig {
	return map[DigestType]DigestConfig{
		CryptoDigest: {
			Type:        CryptoDigest,
			Name:        "Криптовалютный дайджест",
			CommandName: "crypto",
			Prompt:      getCryptoDigestPrompt(),
		},
		TechDigest: {
			Type:        TechDigest,
			Name:        "Технологический дайджест",
			CommandName: "tech",
			Prompt:      getTechDigestPrompt(),
		},
		RealEstateDigest: {
			Type:        RealEstateDigest,
			Name:        "Дайджест недвижимости",
			CommandName: "realestate",
			Prompt:      getRealEstateDigestPrompt(),
		},
		BusinessDigest: {
			Type:        BusinessDigest,
			Name:        "Бизнес-дайджест",
			CommandName: "business",
			Prompt:      getBusinessDigestPrompt(),
		},
		InvestmentDigest: {
			Type:        InvestmentDigest,
			Name:        "Инвестиционный дайджест",
			CommandName: "investment",
			Prompt:      getInvestmentDigestPrompt(),
		},
		StartupDigest: {
			Type:        StartupDigest,
			Name:        "Стартап-дайджест",
			CommandName: "startup",
			Prompt:      getStartupDigestPrompt(),
		},
		GlobalDigest: {
			Type:        GlobalDigest,
			Name:        "Глобальный дайджест",
			CommandName: "global",
			Prompt:      getGlobalDigestPrompt(),
		},
	}
}

// Task represents a scheduled task
type Task struct {
	Name   string `json:"name" yaml:"name"`
	Prompt string `json:"prompt" yaml:"prompt"`
	Time   string `json:"time,omitempty" yaml:"time,omitempty"`
	Cron   string `json:"cron,omitempty" yaml:"cron,omitempty"`
	Model  string `json:"model,omitempty" yaml:"model,omitempty"`
}

// User represents a user in the system
type User struct {
	ID       int64
	Username string
	IsActive bool
}
