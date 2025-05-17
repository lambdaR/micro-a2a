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

type A2AClient struct {
	Client      resty.Client
	EventSource resty.EventSource
}

func NewA2AClient() *A2AClient {
	return &A2AClient{
		Client:      *resty.New(),
		EventSource: *resty.NewEventSource(),
	}
}

// ValidateMethodParams checks if the combination of method and params is valid
// based on the MethodToParamsType map
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

func (c *A2AClient) SendReq(ctx context.Context, method Method, params Params, url string) (*JSONRPCResponse, error) {
	// Validate method and params combination
	if err := validateMethodParams(method, params); err != nil {
		return nil, NewError(ErrorInvalidRequest, err.Error(), nil)
	}

	rpcRes := &JSONRPCResponse{}

	newID := uuid.NewString()

	req := JSONRPCRequest{
		ID:      newID,
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}

	res, err := c.Client.R().SetResult(rpcRes).SetBody(req).Post(url)
	if err != nil {
		return nil, NewError(ErrorInternal, fmt.Sprintf("failed to send request: %v", err), nil)
	}

	defer res.Body.Close()

	return rpcRes, nil
}

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
