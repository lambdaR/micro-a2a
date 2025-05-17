package a2a

import (
	"go-micro.dev/v5/logger"
	"go-micro.dev/v5/store"
)

type AgentOptions struct {
	AgentCard AgentCard
	Logger    logger.Logger
	Store     store.Store
	// use for an agent that replies with one response
	AgentHandler *AgentHandler
	// use for an agent that replies with multiple data objects
	AgentStreamHandler *AgentStreamHandler
}

type AgentOption func(ao *AgentOptions)

func WithLogger(logger logger.Logger) AgentOption {
	return func(ao *AgentOptions) {
		ao.Logger = logger
	}
}

func WithStore(store store.Store) AgentOption {
	return func(ao *AgentOptions) {
		ao.Store = store
	}
}

func WithAgentHandler(handler AgentHandler) AgentOption {
	return func(ao *AgentOptions) {
		ao.AgentHandler = &handler
	}
}

func WithAgentStreamHandler(streamHandler AgentStreamHandler) AgentOption {
	return func(ao *AgentOptions) {
		ao.AgentStreamHandler = &streamHandler
	}
}
