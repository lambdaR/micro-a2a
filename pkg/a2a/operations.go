package a2a

import "encoding/json"

// (TasksSend, TaskSendParams)
//
// (TasksSendSubscrib, TaskSendParams)
//
// (TasksGet, TaskQueryParams)
//
// (TasksCancel, TaskIDParams)
//
// (TasksResubscribe, TaskIDParams)

// MethodToParamsType defines a mapping between each Method and its corresponding Params type.
// This map is used for validation and type checking when processing requests.
// Each method is associated with an empty instance of its expected parameter type.
var MethodToParamsType = map[Method]Params{
	TasksSend:          TaskSendParams{},  // Regular task sending uses TaskSendParams
	TasksSendSubscribe: TaskSendParams{},  // Streaming task sending also uses TaskSendParams
	TasksGet:           TaskQueryParams{}, // Task retrieval uses TaskQueryParams
	TasksCancel:        TaskIDParams{},    // Task cancellation uses TaskIDParams
	TasksResubscribe:   TaskIDParams{},    // Task resubscription uses TaskIDParams
}

// Method represents an A2A API method name.
// These are the operations that can be performed on an A2A-compatible agent.
type Method string

const (
	// Task management methods
	TasksSend          Method = "tasks/send"          // Send a task to an agent
	TasksSendSubscribe Method = "tasks/sendSubscribe" // Send a task and subscribe to streaming updates
	TasksGet           Method = "tasks/get"           // Get information about a task
	TasksCancel        Method = "tasks/cancel"        // Cancel a running task
	TasksResubscribe   Method = "tasks/resubscribe"   // Resubscribe to updates for an existing task
	
	// Push notification methods
	TasksPushNotificationGet Method = "tasks/pushNotification/get" // Get push notification configuration
	TasksPushNotificationSet Method = "tasks/pushNotification/set" // Set push notification configuration
)

// SSEResponse represents a Server-Sent Event response containing a JSON-RPC response.
// This is used when receiving streaming updates from an agent.
type SSEResponse struct {
	// Data contains the JSON-RPC response embedded in the SSE event
	Data JSONRPCResponse `json:"data"`
}

// Params is an interface that all parameter types must implement.
// It serves as a marker interface to group different parameter types
// that can be used with A2A methods.
type Params interface {
	// paramGlue is a marker method that doesn't do anything but
	// ensures type safety when working with different parameter types
	paramGlue()
}

// TaskIdParams represents parameters for task ID-based requests
type TaskIDParams struct {
	ID       string         `json:"id"` // Task ID
	Metadata map[string]any `json:"metadata,omitempty"`
}

func (t TaskIDParams) paramGlue() {}

// TaskQueryParams represents parameters for querying tasks
type TaskQueryParams struct {
	ID            string         `json:"id"`                      // Required task ID
	HistoryLength int            `json:"historyLength,omitempty"` // Optional limit for history entries
	Metadata      map[string]any `json:"metadata,omitempty"`
}

func (t TaskQueryParams) paramGlue() {}

// TaskSendParams is sent by the client to create, continue, or restart a task
type TaskSendParams struct {
	ID               string                  `json:"id"`                         // Task identifier
	SessionID        string                  `json:"sessionId,omitempty"`        // optional session ID
	Message          Message                 `json:"message"`                    // Task message
	HistoryLength    int                     `json:"historyLength,omitempty"`    // number of recent messages to retrieve
	PushNotification *PushNotificationConfig `json:"pushNotification,omitempty"` // notification config
	Metadata         map[string]any          `json:"metadata,omitempty"`         // extension metadata
}

func (t TaskSendParams) paramGlue() {}

type RequestWrapper struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  Method          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCRequest represents a JSON-RPC 2.0 request for sending a task with or without Streaming
type JSONRPCRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id,omitempty"` // Can be string, int, or nil
	Method  Method `json:"method"`
	Params  Params `json:"params,omitempty"`
}

func (r JSONRPCRequest) MarshalJSON() ([]byte, error) {
	wrapper := RequestWrapper{
		JSONRPC: r.JSONRPC,
		ID:      r.ID,
		Method:  r.Method,
	}

	if r.Params != nil {
		params, err := json.Marshal(r.Params)
		if err != nil {
			return nil, err
		}
		wrapper.Params = params
	}

	return json.Marshal(wrapper)
}

func (r *JSONRPCRequest) UnmarshalJSON(data []byte) error {
	var temp RequestWrapper
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	r.JSONRPC = temp.JSONRPC
	r.ID = temp.ID
	r.Method = temp.Method

	switch r.Method {
	case TasksCancel, TasksResubscribe:
		var v TaskIDParams
		if err := json.Unmarshal(temp.Params, &v); err != nil {
			return err
		}
		r.Params = v

	case TasksGet:
		var v TaskQueryParams
		if err := json.Unmarshal(temp.Params, &v); err != nil {
			return err
		}
		r.Params = v
	case TasksSend, TasksSendSubscribe:
		var v TaskSendParams
		if err := json.Unmarshal(temp.Params, &v); err != nil {
			return err
		}
		r.Params = v
	default:
		return nil
	}
	return nil
}

// Result is an interface that all result types must implement.
// It serves as a marker interface to group different result types
// that can be returned from A2A methods.
//
// Implementations include:
// - Task: Represents a complete task with all its details
// - TaskStatusUpdateEvent: Represents an update to a task's status
// - TaskArtifactUpdateEvent: Represents a new artifact produced by a task
type Result interface {
	// resultGlue is a marker method that doesn't do anything but
	// ensures type safety when working with different result types
	resultGlue()
}

// Task represents a unit of work being processed by an agent
type Task struct {
	ID        string         `json:"id"`                  // unique identifier for the task
	SessionID string         `json:"sessionId"`           // client-generated id for the session holding the task
	Status    TaskStatus     `json:"status"`              // current status of the task
	History   []Message      `json:"history,omitempty"`   // optional message history
	Artifacts []Artifact     `json:"artifacts,omitempty"` // optional collection of artifacts
	Metadata  map[string]any `json:"metadata,omitempty"`  // extension metadata
}

func (t Task) resultGlue() {}

// TaskStatusUpdateEvent is sent during sendSubscribe or subscribe requests
type TaskStatusUpdateEvent struct {
	ID       string         `json:"id"`                 // Task id
	Status   TaskStatus     `json:"status"`             // Updated status
	Final    bool           `json:"final"`              // indicates end of event stream
	Metadata map[string]any `json:"metadata,omitempty"` // extension metadata
}

func (t TaskStatusUpdateEvent) resultGlue() {}

// TaskArtifactUpdateEvent is sent during sendSubscribe or subscribe requests
type TaskArtifactUpdateEvent struct {
	ID       string         `json:"id"`                 // Task id
	Artifact Artifact       `json:"artifact"`           // New artifact
	Metadata map[string]any `json:"metadata,omitempty"` // extension metadata
}

func (t TaskArtifactUpdateEvent) resultGlue() {}

type ResponseWrapper struct {
	JSONRPC string          `json:"jsonrpc" default:"2.0"`
	ID      any             `json:"id,omitempty"` // Can be string, int, or nil
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response with task data
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc" default:"2.0"`
	ID      any           `json:"id,omitempty"` // Can be string, int, or nil
	Result  Result        `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

func (r JSONRPCResponse) MarshalJSON() ([]byte, error) {
	wrapper := ResponseWrapper{
		JSONRPC: r.JSONRPC,
		ID:      r.ID,
		Error:   r.Error,
	}

	if r.Result != nil {
		result, err := json.Marshal(r.Result)
		if err != nil {
			return nil, err
		}
		wrapper.Result = result
	}

	return json.Marshal(wrapper)
}

func (r *JSONRPCResponse) UnmarshalJSON(data []byte) error {
	var temp ResponseWrapper
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	r.JSONRPC = temp.JSONRPC
	r.ID = temp.ID
	r.Error = temp.Error

	// If there's no result or it's null, return early
	if len(temp.Result) == 0 || string(temp.Result) == "null" {
		return nil
	}

	// Try to determine the result type based on the content
	var resultMap map[string]interface{}
	if err := json.Unmarshal(temp.Result, &resultMap); err != nil {
		return err
	}

	// Check for fields that would identify the result type
	if _, hasID := resultMap["id"]; hasID {
		if _, hasStatus := resultMap["status"]; hasStatus {
			// This is likely a Task
			var task Task
			if err := json.Unmarshal(temp.Result, &task); err != nil {
				return err
			}
			result := Result(task)
			r.Result = result
		} else if _, hasArtifact := resultMap["artifact"]; hasArtifact {
			// This is likely a TaskArtifactUpdateEvent
			var event TaskArtifactUpdateEvent
			if err := json.Unmarshal(temp.Result, &event); err != nil {
				return err
			}
			result := Result(event)
			r.Result = result
		} else if _, hasFinal := resultMap["final"]; hasFinal {
			// This is likely a TaskStatusUpdateEvent
			var event TaskStatusUpdateEvent
			if err := json.Unmarshal(temp.Result, &event); err != nil {
				return err
			}
			result := Result(event)
			r.Result = result
		}
	}

	return nil
}

// GetTaskPushNotificationRequest represents a JSON-RPC 2.0 request to get push notification config
// type GetTaskPushNotificationRequest struct {
// 	JSONRPC string       `json:"jsonrpc" default:"2.0" const:"2.0"`
// 	ID      any          `json:"id,omitempty"` // Can be string, int, or nil
// 	Method  string       `json:"method" default:"tasks/pushNotification/get" const:"tasks/pushNotification/get"`
// 	Params  TaskIDParams `json:"params"`
// }

// GetTaskPushNotificationResponse represents a JSON-RPC 2.0 response with push notification config
// type GetTaskPushNotificationResponse struct {
// 	JSONRPC string                      `json:"jsonrpc" default:"2.0"`
// 	ID      any                         `json:"id,omitempty"` // Can be string, int, or nil
// 	Result  *TaskPushNotificationConfig `json:"result,omitempty"`
// 	Error   *a2aerr.A2ARPCError         `json:"error,omitempty"`
// }

// SetTaskPushNotificationRequest represents a JSON-RPC 2.0 request to set push notification config
// type SetTaskPushNotificationRequest struct {
// 	JSONRPC string                     `json:"jsonrpc" default:"2.0" const:"2.0"`
// 	ID      any                        `json:"id,omitempty"` // Can be string, int, or nil
// 	Method  string                     `json:"method" default:"tasks/pushNotification/set" const:"tasks/pushNotification/set"`
// 	Params  TaskPushNotificationConfig `json:"params"`
// }

// SetTaskPushNotificationResponse represents a JSON-RPC 2.0 response after setting push notification config
// type SetTaskPushNotificationResponse struct {
// 	JSONRPC string                      `json:"jsonrpc" default:"2.0"`
// 	ID      any                         `json:"id,omitempty"` // Can be string, int, or nil
// 	Result  *TaskPushNotificationConfig `json:"result,omitempty"`
// 	Error   *a2aerr.A2ARPCError         `json:"error,omitempty"`
// }
