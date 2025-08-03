# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Development Commands
- `go run ./cmd/bot` - Run the telegram bot locally
- `go build -o bot ./cmd/bot` - Build the bot binary
- `go build ./...` - Build all packages
- `go test ./...` - Run all tests
- `go test ./tests/...` - Run specific test packages

### Code Quality
- `gofmt -w -s .` - Format Go code (required before commits)
- `go vet ./...` - Run static analysis
- `golangci-lint run` - Run comprehensive linter (install with `go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.63.0`)

### Docker
- `docker build -t telegram-bot .` - Build Docker image
- `docker-compose up -d` - Run with docker-compose

## Architecture

### Core Structure
This is a Telegram bot written in Go that generates AI-powered content digests using OpenAI APIs. The bot operates on a scheduled basis, sending different types of content to subscribed chats.

### Key Components

**Entry Point**: `cmd/bot/main.go` - Simple main function that loads config and starts the bot

**Core Bot Logic**: `internal/bot/` 
- `app.go` - Bot initialization and dependency injection
- `bot.go` - Main bot runner with scheduler setup
- `handlers.go` - Telegram command handlers (/start, /chat, /search, etc.)
- `task.go` - Task execution logic for scheduled content generation
- `openai_*.go` - OpenAI API integration with web search capabilities

**Configuration**: `internal/config/config.go` - Environment variable parsing and validation

**Logging System**: `internal/logger/` - Comprehensive structured logging with multiple outputs (file, Telegram, HTTP)

**Services**: `internal/services/` - Digest generation services for different content types

**Domain Models**: `internal/domain/` - Core business logic and prompt definitions

### Task System
The bot uses a YAML-based task configuration system (`tasks.yml`) that defines:
- Scheduled content generation (crypto, tech, real estate, business digests)
- Time-based execution in Moscow timezone
- OpenAI model selection per task
- Template-based prompt system with variable substitution

### OpenAI Integration
- Supports multiple OpenAI models (gpt-4.1, o3, o3-mini, gpt-4o-mini)
- Built-in web search capabilities using OpenAI's search tools
- Streaming and non-streaming completion modes
- Configurable parameters (max tokens, tool choice, service tier, reasoning effort)

### Whitelist System
Chat subscription management stored in `whitelist.json` with commands to add/remove chats from the broadcast list.

### Testing
Comprehensive test suite in `tests/` covering:
- Command handlers
- OpenAI integration
- Task scheduling
- Configuration loading
- Edge cases and error handling

## Environment Variables

### Required
- `TELEGRAM_TOKEN` - Telegram bot token
- `OPENAI_API_KEY` - OpenAI API key

### Optional Configuration
- `CHAT_ID` - Initial chat ID for testing
- `LOG_CHAT_ID` - Chat ID for log messages
- `OPENAI_MODEL` - Default OpenAI model (default: gpt-4.1)
- `OPENAI_MAX_TOKENS` - Max response tokens (default: 600)
- `OPENAI_TOOL_CHOICE` - Tool usage mode (default: auto)
- `OPENAI_SERVICE_TIER` - OpenAI service tier
- `OPENAI_REASONING_EFFORT` - Reasoning effort level
- `ENABLE_WEB_SEARCH` - Enable web search (default: true)
- `LOG_LEVEL` - Logging level (debug, info, warn, error)
- `TASKS_FILE` - Path to custom tasks YAML file
- `WHITELIST_FILE` - Path to whitelist JSON file (default: whitelist.json)
- `LUNCH_TIME` / `BRIEF_TIME` - Custom scheduling times

## Development Notes

### Code Style
- All code must be formatted with `gofmt -w -s` before commits
- Use structured logging via `logger.L` throughout the codebase
- Follow Go naming conventions and package organization

### Testing
- Tests are located in `tests/` directory
- Use table-driven tests where appropriate
- Mock external dependencies (OpenAI, Telegram API)
- Test both success and error cases

### OpenAI Models Strategy
Different OpenAI models are used for different task types:
- o3/o3-mini: Complex analysis and reasoning tasks
- gpt-4o-mini: Simple content generation and responses
- gpt-4.1: General purpose with web search capabilities

### Deployment
- GitHub Actions workflow handles automated deployment
- Docker multi-arch builds (linux/amd64, linux/arm64)
- VPS deployment via SSH and docker-compose