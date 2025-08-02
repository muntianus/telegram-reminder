package bot

// Testing utilities for accessing and modifying runtime configuration

// GetCurrentModel returns the current model for testing
func GetCurrentModel() string {
	return getRuntimeConfig().CurrentModel
}

// SetCurrentModel sets the current model for testing
func SetCurrentModel(model string) {
	updateRuntimeConfig(func(cfg *RuntimeConfig) {
		cfg.CurrentModel = model
	})
}

// GetEnableWebSearch returns the web search setting for testing
func GetEnableWebSearch() bool {
	return getRuntimeConfig().EnableWebSearch
}

// SetEnableWebSearch sets the web search setting for testing
func SetEnableWebSearch(enabled bool) {
	updateRuntimeConfig(func(cfg *RuntimeConfig) {
		cfg.EnableWebSearch = enabled
	})
}

// GetMaxTokens returns the max tokens setting for testing
func GetMaxTokens() int {
	return getRuntimeConfig().MaxTokens
}

// SetMaxTokens sets the max tokens setting for testing
func SetMaxTokens(tokens int) {
	updateRuntimeConfig(func(cfg *RuntimeConfig) {
		cfg.MaxTokens = tokens
	})
}

// GetToolChoice returns the tool choice setting for testing
func GetToolChoice() string {
	return getRuntimeConfig().ToolChoice
}

// SetToolChoice sets the tool choice setting for testing
func SetToolChoice(choice string) {
	updateRuntimeConfig(func(cfg *RuntimeConfig) {
		cfg.ToolChoice = choice
	})
}

// ResetRuntimeConfig resets runtime configuration to defaults for testing
func ResetRuntimeConfig() {
	updateRuntimeConfig(func(cfg *RuntimeConfig) {
		cfg.CurrentModel = "gpt-4.1"
		cfg.MaxTokens = 600
		cfg.EnableWebSearch = true
		cfg.ToolChoice = "auto"
		cfg.BasePrompt = ""
		cfg.ServiceTier = ""
		cfg.ReasoningEffort = ""
	})
}
