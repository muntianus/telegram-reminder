package container

import (
	"sync"
	"time"

	"telegram-reminder/internal/handlers"
	"telegram-reminder/internal/services"
)

// Container manages dependency injection
type Container struct {
	mu       sync.RWMutex
	services map[string]interface{}
	configs  map[string]interface{}
}

// NewContainer creates a new DI container
func NewContainer() *Container {
	return &Container{
		services: make(map[string]interface{}),
		configs:  make(map[string]interface{}),
	}
}

// RegisterService registers a service in the container
func (c *Container) RegisterService(name string, service interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.services[name] = service
}

// GetService retrieves a service from the container
func (c *Container) GetService(name string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	service, exists := c.services[name]
	return service, exists
}

// RegisterConfig registers configuration in the container
func (c *Container) RegisterConfig(name string, config interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.configs[name] = config
}

// GetConfig retrieves configuration from the container
func (c *Container) GetConfig(name string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	config, exists := c.configs[name]
	return config, exists
}

// Service name constants
const (
	DigestServiceName = "digest_service"
	DigestHandlerName = "digest_handler"
	ErrorHandlerName  = "error_handler"
	AIClientName      = "ai_client"
)

// Config name constants
const (
	AITimeoutConfig = "ai_timeout"
	RuntimeConfig   = "runtime_config"
)

// ServiceBuilder helps build services with dependencies
type ServiceBuilder struct {
	container *Container
}

// NewServiceBuilder creates a new service builder
func NewServiceBuilder(container *Container) *ServiceBuilder {
	return &ServiceBuilder{container: container}
}

// BuildDigestService builds the digest service with its dependencies
func (sb *ServiceBuilder) BuildDigestService() (*services.DigestService, error) {
	aiClient, exists := sb.container.GetService(AIClientName)
	if !exists {
		return nil, ErrServiceNotFound{ServiceName: AIClientName}
	}

	timeout, exists := sb.container.GetConfig(AITimeoutConfig)
	if !exists {
		timeout = 5 * time.Minute // default timeout
	}

	timeoutDuration, ok := timeout.(time.Duration)
	if !ok {
		timeoutDuration = 5 * time.Minute // default timeout
	}

	aiClientTyped, ok := aiClient.(services.AIClient)
	if !ok {
		return nil, ErrInvalidServiceType{ServiceName: AIClientName, ExpectedType: "services.AIClient"}
	}

	digestService := services.NewDigestService(aiClientTyped, timeoutDuration)
	sb.container.RegisterService(DigestServiceName, digestService)

	return digestService, nil
}

// BuildDigestHandler builds the digest handler with its dependencies
func (sb *ServiceBuilder) BuildDigestHandler() (*handlers.DigestHandler, error) {
	digestService, exists := sb.container.GetService(DigestServiceName)
	if !exists {
		// Try to build digest service first
		ds, err := sb.BuildDigestService()
		if err != nil {
			return nil, err
		}
		digestService = ds
	}

	errorHandler, exists := sb.container.GetService(ErrorHandlerName)
	if !exists {
		return nil, ErrServiceNotFound{ServiceName: ErrorHandlerName}
	}

	digestServiceTyped, ok := digestService.(*services.DigestService)
	if !ok {
		return nil, ErrInvalidServiceType{ServiceName: DigestServiceName, ExpectedType: "*services.DigestService"}
	}

	errorHandlerTyped, ok := errorHandler.(handlers.ErrorHandler)
	if !ok {
		return nil, ErrInvalidServiceType{ServiceName: ErrorHandlerName, ExpectedType: "handlers.ErrorHandler"}
	}

	digestHandler := handlers.NewDigestHandler(digestServiceTyped, errorHandlerTyped)
	sb.container.RegisterService(DigestHandlerName, digestHandler)

	return digestHandler, nil
}

// AutoWire automatically wires all services
func (sb *ServiceBuilder) AutoWire() error {
	// Build services in dependency order
	_, err := sb.BuildDigestService()
	if err != nil {
		return err
	}

	_, err = sb.BuildDigestHandler()
	if err != nil {
		return err
	}

	return nil
}

// Errors
type ErrServiceNotFound struct {
	ServiceName string
}

func (e ErrServiceNotFound) Error() string {
	return "service not found: " + e.ServiceName
}

type ErrInvalidServiceType struct {
	ServiceName  string
	ExpectedType string
}

func (e ErrInvalidServiceType) Error() string {
	return "invalid service type for " + e.ServiceName + ", expected: " + e.ExpectedType
}
