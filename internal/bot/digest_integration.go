package bot

import (
	"telegram-reminder/internal/container"
	"telegram-reminder/internal/domain"
	"telegram-reminder/internal/handlers"
	"telegram-reminder/internal/services"

	tb "gopkg.in/telebot.v3"
)

// DigestIntegration manages the integration of new digest architecture with existing bot
type DigestIntegration struct {
	container     *container.Container
	digestHandler *handlers.DigestHandler
}

// NewDigestIntegration creates a new digest integration
func NewDigestIntegration(client ChatCompleter, errorHandler *ErrorHandler) (*DigestIntegration, error) {
	// Create DI container
	diContainer := container.NewContainer()

	// Register dependencies
	aiAdapter := services.NewOpenAIAdapter(client)
	diContainer.RegisterService(container.AIClientName, aiAdapter)
	diContainer.RegisterService(container.ErrorHandlerName, &ErrorHandlerAdapter{handler: errorHandler})
	diContainer.RegisterConfig(container.AITimeoutConfig, OpenAITimeout)

	// Build services
	builder := container.NewServiceBuilder(diContainer)
	digestHandler, err := builder.BuildDigestHandler()
	if err != nil {
		return nil, err
	}

	return &DigestIntegration{
		container:     diContainer,
		digestHandler: digestHandler,
	}, nil
}

// RegisterHandlers registers all digest handlers with the bot
func (di *DigestIntegration) RegisterHandlers(bot *tb.Bot) {
	// Register new digest handlers
	di.digestHandler.RegisterDigestHandlers(&BotAdapter{bot: bot})
}

// BotAdapter adapts tb.Bot to our TelegramBot interface
type BotAdapter struct {
	bot *tb.Bot
}

func (ba *BotAdapter) Handle(endpoint interface{}, handler interface{}, middlewares ...tb.MiddlewareFunc) {
	if handlerFunc, ok := handler.(func(tb.Context) error); ok {
		ba.bot.Handle(endpoint, handlerFunc, middlewares...)
	}
}

// ErrorHandlerAdapter adapts our ErrorHandler to the handlers interface
type ErrorHandlerAdapter struct {
	handler *ErrorHandler
}

func (ea *ErrorHandlerAdapter) HandleOpenAIError(err error, model string) string {
	return ea.handler.HandleOpenAIError(err, model)
}

// ReplaceDigestHandlers replaces old digest handlers with new ones
func (di *DigestIntegration) ReplaceDigestHandlers(bot *tb.Bot, client ChatCompleter) {
	// Get all digest configs
	configs := domain.GetDigestConfigs()
	
	// Create handlers using new architecture
	for digestType, config := range configs {
		// Create a closure to capture the digest type
		capturedType := digestType
		handler := di.digestHandler.HandleDigest(capturedType)
		
		// Register with bot
		bot.Handle("/"+config.CommandName, handler)
	}
}

// GetDigestTypes returns all available digest types
func (di *DigestIntegration) GetDigestTypes() []string {
	configs := domain.GetDigestConfigs()
	types := make([]string, 0, len(configs))
	
	for _, config := range configs {
		types = append(types, config.CommandName)
	}
	
	return types
}