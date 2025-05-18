package a2a

import (
	// two possible alternatives to resty is
	// https://github.com/r3labs/sse
	// https://github.com/tmaxmax/go-sse

	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	go_sse "github.com/tmaxmax/go-sse"

	"resty.dev/v3"
)

// A2AClient is a client for interacting with A2A-compatible agents.
// It provides methods for sending requests and establishing streaming connections.
type A2AClient struct {
	// Client is the HTTP client used for making requests
	Client resty.Client

	// EventSource is used for Server-Sent Events (SSE) connections
	EventSource resty.EventSource
}

// NewA2AClient creates a new A2A client with default configuration.
// It initializes both the HTTP client and the EventSource for SSE connections.
//
// Returns:
//   - A pointer to a new A2AClient instance ready for use
func NewA2AClient() *A2AClient {
	return &A2AClient{
		Client:      *resty.New(),
		EventSource: *resty.NewEventSource(),
	}
}

// validateMethodParams checks if the combination of method and params is valid
// based on the MethodToParamsType map. It ensures that the method is supported
// and that the params are of the correct type for the method.
//
// Parameters:
//   - method: The A2A method to validate
//   - params: The parameters to validate against the method
//
// Returns:
//   - An error if the validation fails, or nil if the combination is valid
//
// Validation checks:
//   - Method exists in the MethodToParamsType map
//   - Params is not nil
//   - Params is of the expected type for the method
func validateMethodParams(method Method, params Params) error {
	expectedType, exists := MethodToParamsType[method]
	if !exists {
		return fmt.Errorf("unsupported method: %s", method)
	}

	if params == nil {
		return fmt.Errorf("params cannot be nil for method: %s", method)
	}

	// Check if the params type matches the expected type for this method
	expectedTypeName := fmt.Sprintf("%T", expectedType)
	actualTypeName := fmt.Sprintf("%T", params)

	if expectedTypeName != actualTypeName {
		return fmt.Errorf("invalid params type for method %s: expected %s, got %s",
			method, expectedTypeName, actualTypeName)
	}

	return nil
}

// SendReq sends a JSON-RPC request to an A2A-compatible agent and returns the response.
//
// Parameters:
//   - ctx: Context for the request, which can be used for cancellation
//   - method: The A2A method to call (e.g., TasksSend, TasksGet)
//   - params: The parameters for the method, must match the expected type for the method
//   - url: The URL of the A2A agent endpoint
//
// Returns:
//   - JSONRPCResponse: The response from the agent
//   - error: An error if the request failed, or nil if successful
//
// The method performs the following steps:
//  1. Validates that the method and params combination is valid
//  2. Creates a JSON-RPC request with a new UUID
//  3. Sends the request to the specified URL
//  4. Returns the response or an error
func (c *A2AClient) SendReq(ctx context.Context, method Method, params Params, url string) (JSONRPCResponse, error) {
	// Validate method and params combination
	if err := validateMethodParams(method, params); err != nil {
		return JSONRPCResponse{}, NewError(ErrorInvalidRequest, err.Error(), nil)
	}

	rpcRes := JSONRPCResponse{}

	newID := uuid.NewString()

	req := JSONRPCRequest{
		ID:      newID,
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}

	res, err := c.Client.R().SetResult(&rpcRes).SetBody(req).Post(url)
	if err != nil {
		return JSONRPCResponse{}, NewError(ErrorInternal, fmt.Sprintf("failed to send request: %v", err), nil)
	}

	defer res.Body.Close()

	return rpcRes, nil
}

// SendReqStream sends a JSON-RPC request to an A2A-compatible agent and establishes
// a Server-Sent Events (SSE) connection for streaming updates.
//
// Parameters:
//   - ctx: Context for the request, which can be used for cancellation
//   - method: The A2A method to call (typically TasksSendSubscribe)
//   - params: The parameters for the method, must match the expected type for the method
//   - addr: The URL of the A2A agent streaming endpoint
//
// Returns:
//   - chan go_sse.Event: A channel that will receive SSE events from the agent
//   - error: An error if the request or connection setup failed, or nil if successful
//
// The method performs the following steps:
//  1. Validates that the method and params combination is valid
//  2. Creates a JSON-RPC request with a new UUID
//  3. Sends the initial request to establish the streaming connection
//  4. Sets up an SSE connection to receive streaming updates
//  5. Returns a channel for receiving events
//
// Note: Currently only TasksSendSubscribe is fully implemented for streaming.
// Other methods (TasksGet, TasksCancel, TasksResubscribe) are placeholders.
func (c *A2AClient) SendReqStream(ctx context.Context, method Method, params Params, addr string) (chan go_sse.Event, error) {
	// Validate method and params combination
	if err := validateMethodParams(method, params); err != nil {
		return nil, NewError(ErrorInvalidRequest, err.Error(), nil)
	}

	resChan := make(chan go_sse.Event, 100)
	id := uuid.NewString()

	req := JSONRPCRequest{
		ID:      id,
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}

	switch method {
	// Initiation
	case TasksSendSubscribe:
		// first, sent initial request
		res, err := c.Client.R().SetBody(req).Post(addr)
		if err != nil {
			return resChan, NewError(ErrorInternal, fmt.Sprintf("failed to send request: %v", err), nil)
		}

		defer res.Body.Close()

		if res.IsError() {
			return resChan, NewError(ErrorInternal, fmt.Sprintf("server returned: %v", res.StatusCode()), nil)
		}

		// second, subscribe to events
		newAddr := fmt.Sprintf("%v?id=%v", addr, id)

		c.EventSource.SetURL(newAddr).
			OnOpen(func(url string) {
				log.Printf("connection with [%v] established", url)
			}).
			OnError(func(err error) {
				log.Println(err)
			}).
			OnMessage(func(e any) {
				e = e.(*resty.Event)
				fmt.Println(e)
			}, nil)

		err = c.EventSource.Get()
		if err != nil {
			return resChan, NewError(ErrorInternal, fmt.Sprintf("failed to establish a connection with [%v]: %v", newAddr, err), nil)
		}

	case TasksGet:
	case TasksCancel:
	case TasksResubscribe:
	}

	return resChan, nil
}

// switch (response.Data.Result).(type) {
// case *TaskStatusUpdateEvent:
// 	st.StreamStatus = append(st.StreamStatus, (response.Data.Result).(TaskStatusUpdateEvent))
// case *TaskArtifactUpdateEvent:
// 	st.StreamArtifact = append(st.StreamArtifact, (response.Data.Result).(TaskArtifactUpdateEvent))
// }
