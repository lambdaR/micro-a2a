package a2a

import (
	"encoding/json"
	"fmt"
	"time"
)

// TaskStatus represents the state of a task with additional context
type TaskStatus struct {
	State     TaskState  `json:"state"`               // current state of the task
	Message   *Message   `json:"message,omitempty"`   // additional status updates
	Timestamp *time.Time `json:"timestamp,omitempty"` // ISO datetime value
}

// TaskState represents possible states of a task
type TaskState string

const (
	// Task received by server, acknowledged, but processing has not yet actively started.
	TaskStateSubmitted TaskState = "submitted"
	// Task is actively being processed by the agent.
	TaskStateWorking TaskState = "working"
	// Agent requires additional input from the client/user to proceed. (Task is paused)
	TaskStateInputRequired TaskState = "input-required"
	// Task finished successfully. (Terminal state)
	TaskStateCompleted TaskState = "completed"
	// Task was canceled by the client or potentially by the server. (Terminal state)
	TaskStateCanceled TaskState = "canceled"
	// Task terminated due to an error during processing. (Terminal state)
	TaskStateFailed TaskState = "failed"
	// Task has be rejected by the remote agent (Terminal state)
	TaskStateRejected TaskState = "rejected"
	// Authentication required from client/user to proceed. (Task is paused)
	TaskStateAuthRequired TaskState = "auth-required"
	// The state of the task cannot be determined (e.g., task ID invalid or expired). (Effectively a terminal state from client's PoV for that ID)
	TaskStateUnknown TaskState = "unknown"
)

// Artifact represents a piece of output or data produced by an agent
type Artifact struct {
	Name        string         `json:"name,omitempty"`
	Description string         `json:"description,omitempty"`
	Parts       []Part         `json:"parts"` // Required parts of the artifact
	Metadata    map[string]any `json:"metadata,omitempty"`
	Index       int            `json:"index"`
	Append      bool           `json:"append,omitempty"`
	LastChunk   bool           `json:"lastChunk,omitempty"`
}

// UnmarshalJSON implements custom JSON unmarshaling for Artifact
func (a *Artifact) UnmarshalJSON(data []byte) error {
	type ArtifactAlias Artifact
	temp := struct {
		*ArtifactAlias
		Parts []json.RawMessage `json:"parts"`
	}{
		ArtifactAlias: (*ArtifactAlias)(a),
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Initialize parts slice
	a.Parts = make([]Part, 0, len(temp.Parts))

	// Process each part
	for _, rawPart := range temp.Parts {
		var partMap map[string]interface{}
		if err := json.Unmarshal(rawPart, &partMap); err != nil {
			return fmt.Errorf("failed to parse part: %w", err)
		}

		typeVal, ok := partMap["type"]
		if !ok {
			return fmt.Errorf("part missing required 'type' field")
		}

		typeStr, ok := typeVal.(string)
		if !ok {
			return fmt.Errorf("part 'type' field must be a string")
		}

		partType := PartType(typeStr)

		switch partType {
		case PartTypeText:
			var textPart TextPart
			if err := json.Unmarshal(rawPart, &textPart); err != nil {
				return fmt.Errorf("failed to unmarshal text part: %w", err)
			}
			a.Parts = append(a.Parts, textPart)

		case PartTypeFile:
			var filePart FilePart
			if err := json.Unmarshal(rawPart, &filePart); err != nil {
				return fmt.Errorf("failed to unmarshal file part: %w", err)
			}
			a.Parts = append(a.Parts, filePart)

		case PartTypeData:
			var dataPart DataPart
			if err := json.Unmarshal(rawPart, &dataPart); err != nil {
				return fmt.Errorf("failed to unmarshal data part: %w", err)
			}
			a.Parts = append(a.Parts, dataPart)

		default:
			return fmt.Errorf("unknown part type: %s", partType)
		}
	}

	return nil
}

type AuthenticationInfo struct {
	Schemes     []string `json:"schemes"`               // Required array of strings
	Credentials string   `json:"credentials,omitempty"` // Optional string
}

// MessageRole defines the possible roles for a message
type MessageRole string

const (
	MessageRoleUser  MessageRole = "user"
	MessageRoleAgent MessageRole = "agent"

	// Message kind constant
	MessageKind string = "message"
)

// Message represents a communication between user and agent
type Message struct {
	Kind             string         `json:"kind"`                       // Event type, const "message"
	MessageId        string         `json:"messageId"`                  // Identifier created by the message creator
	Role             MessageRole    `json:"role"`                       // Message sender's role
	Parts            []Part         `json:"parts"`                      // Message content
	Metadata         map[string]any `json:"metadata,omitempty"`         // Extension metadata
	TaskId           string         `json:"taskId,omitempty"`           // Identifier of task the message is related to
	ContextId        string         `json:"contextId,omitempty"`        // The context the message is associated with
	ReferenceTaskIds []string       `json:"referenceTaskIds,omitempty"` // List of tasks referenced as context by this message
}

type MessageWrapper struct {
	Kind             string            `json:"kind"`
	MessageId        string            `json:"messageId"`
	Role             MessageRole       `json:"role"`
	Parts            []json.RawMessage `json:"parts"`
	Metadata         map[string]any    `json:"metadata,omitempty"`
	TaskId           string            `json:"taskId,omitempty"`
	ContextId        string            `json:"contextId,omitempty"`
	ReferenceTaskIds []string          `json:"referenceTaskIds,omitempty"`
}

// MarshalJSON implements the json.Marshaler interface for Message
func (m Message) MarshalJSON() ([]byte, error) {
	type MessageAlias Message
	wrapper := struct {
		MessageAlias
		Parts []json.RawMessage `json:"parts"`
	}{
		MessageAlias: MessageAlias(m),
	}

	// Set default kind if not specified
	if wrapper.Kind == "" {
		wrapper.Kind = MessageKind
	}

	// Ensure required fields for spec compliance
	if wrapper.MessageId == "" {
		return nil, fmt.Errorf("messageId is required")
	}

	wrapper.Parts = make([]json.RawMessage, len(m.Parts))
	for i, part := range m.Parts {
		var err error
		switch p := part.(type) {
		case TextPart:
			// Ensure kind is set
			if p.Kind == "" {
				p.Kind = PartTypeText
			}
			wrapper.Parts[i], err = json.Marshal(p)
		case FilePart:
			if p.Kind == "" {
				p.Kind = PartTypeFile
			}
			wrapper.Parts[i], err = json.Marshal(p)
		case DataPart:
			if p.Kind == "" {
				p.Kind = PartTypeData
			}
			wrapper.Parts[i], err = json.Marshal(p)
		case *TextPart:
			if p.Kind == "" {
				p.Kind = PartTypeText
			}
			wrapper.Parts[i], err = json.Marshal(p)
		case *FilePart:
			if p.Kind == "" {
				p.Kind = PartTypeFile
			}
			wrapper.Parts[i], err = json.Marshal(p)
		case *DataPart:
			if p.Kind == "" {
				p.Kind = PartTypeData
			}
			wrapper.Parts[i], err = json.Marshal(p)
		default:
			return nil, fmt.Errorf("unknown part type: %T", part)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to marshal part: %w", err)
		}
	}

	return json.Marshal(wrapper)
}

func (m *Message) UnmarshalJSON(data []byte) error {
	var temp MessageWrapper
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Ensure required fields for spec compliance
	if temp.MessageId == "" {
		return fmt.Errorf("messageId is required")
	}

	// Copy all fields from the wrapper
	m.Kind = temp.Kind
	m.MessageId = temp.MessageId
	m.Role = temp.Role
	m.Metadata = temp.Metadata
	m.TaskId = temp.TaskId
	m.ContextId = temp.ContextId
	m.ReferenceTaskIds = temp.ReferenceTaskIds

	// Set default kind if not specified
	if m.Kind == "" {
		m.Kind = MessageKind
	}

	// Initialize parts slice with capacity
	m.Parts = make([]Part, 0, len(temp.Parts))

	// Process each part in the message
	for _, rawPart := range temp.Parts {
		var partMap map[string]interface{}
		if err := json.Unmarshal(rawPart, &partMap); err != nil {
			return fmt.Errorf("failed to parse part: %w", err)
		}

		// Get the kind value (previously called "type")
		kindVal, ok := partMap["kind"]
		if !ok {
			// Try legacy "type" field for backward compatibility
			kindVal, ok = partMap["type"]
			if !ok {
				return fmt.Errorf("part missing required 'kind' field")
			}
		}

		kindStr, ok := kindVal.(string)
		if !ok {
			return fmt.Errorf("part 'kind' field must be a string")
		}

		// Create the appropriate part based on kind
		switch kindStr {
		case string(PartTypeText):
			var textPart TextPart
			if err := json.Unmarshal(rawPart, &textPart); err != nil {
				return fmt.Errorf("failed to unmarshal text part: %w", err)
			}
			m.Parts = append(m.Parts, textPart)

		case string(PartTypeFile):
			var filePart FilePart
			if err := json.Unmarshal(rawPart, &filePart); err != nil {
				return fmt.Errorf("failed to unmarshal file part: %w", err)
			}
			m.Parts = append(m.Parts, filePart)

		case string(PartTypeData):
			var dataPart DataPart
			if err := json.Unmarshal(rawPart, &dataPart); err != nil {
				return fmt.Errorf("failed to unmarshal data part: %w", err)
			}
			m.Parts = append(m.Parts, dataPart)

		default:
			return fmt.Errorf("unknown part kind: %s", kindStr)
		}
	}

	return nil
}

// PartType defines the type of content in a message part.
// This is used to determine how to interpret the part's content.
type PartType string

const (
	PartTypeText PartType = "text" // Text content (plain text, markdown, etc.)
	PartTypeFile PartType = "file" // File content (binary data, documents, images, etc.)
	PartTypeData PartType = "data" // Structured data (JSON, etc.)
)

// Part represents a message part which can contain different types of content.
// It's implemented as an interface to allow for different concrete types
// while maintaining a common interface for handling message parts.
//
// Implementations:
// - TextPart: Contains text content
// - FilePart: Contains file content (binary data or references)
// - DataPart: Contains structured data
type Part interface {
	// partGlue is a marker method that doesn't do anything but
	// ensures type safety when working with different part types
	partGlue()
}

// TextPart represents a message part containing text content.
// It implements the Part interface.
type TextPart struct {
	// Kind indicates this is a text part (should always be "text")
	Kind PartType `json:"kind"`

	// Text contains the actual text content
	Text string `json:"text"`

	// Metadata contains optional additional information about this part
	Metadata map[string]any `json:"metadata,omitempty"`
}

// partGlue implements the Part interface for TextPart
func (p TextPart) partGlue() {}

// FilePart represents a message part containing file content.
// It implements the Part interface.
type FilePart struct {
	// Kind indicates this is a file part (should always be "file")
	Kind PartType `json:"kind"`

	// File contains the file content or reference
	File FileContent `json:"file"`

	// Metadata contains optional additional information about this part
	Metadata map[string]any `json:"metadata,omitempty"`
}

// partGlue implements the Part interface for FilePart
func (p FilePart) partGlue() {}

// DataPart represents a message part containing structured data.
// It implements the Part interface.
type DataPart struct {
	// Kind indicates this is a data part (should always be "data")
	Kind PartType `json:"kind"`

	// Data contains the structured data as a map
	Data map[string]any `json:"data"`

	// Metadata contains optional additional information about this part
	Metadata map[string]any `json:"metadata,omitempty"`
}

// partGlue implements the Part interface for DataPart
func (p DataPart) partGlue() {}

// TextContent represents a text part
type TextContent struct {
	Text string `json:"text"`
}

// FileContent represents a file part
type FileContent struct {
	Name     string `json:"name,omitempty"`
	MimeType string `json:"mimeType,omitempty"`
	Bytes    string `json:"bytes,omitempty"` // base64 encoded content
	URI      string `json:"uri,omitempty"`
}

// // NewTextPart creates a new text part
// func NewTextPart(text string, metadata map[string]any) Part {
// 	return Part{
// 		Type:        PartTypeText,
// 		TextContent: &TextContent{Text: text},
// 		Metadata:    metadata,
// 	}
// }
//
// // NewFilePart creates a new file part
// func NewFilePart(file FileContent, metadata map[string]any) Part {
// 	return Part{
// 		Type:        PartTypeFile,
// 		FileContent: &file,
// 		Metadata:    metadata,
// 	}
// }
//
// // NewDataPart creates a new data part
// func NewDataPart(data map[string]any, metadata map[string]any) Part {
// 	return Part{
// 		Type:        PartTypeData,
// 		DataContent: data,
// 		Metadata:    metadata,
// 	}
// }
//
// // UnmarshalJSON custom unmarshaler for Part to handle the union type
// func (p *Part) UnmarshalJSON(data []byte) error {
// 	type Alias Part
// 	aux := &struct {
// 		*Alias
// 	}{
// 		Alias: (*Alias)(p),
// 	}
//
// 	if err := json.Unmarshal(data, &aux); err != nil {
// 		return err
// 	}
//
// 	// Validate that only the correct content is set based on type
// 	switch p.Type {
// 	case PartTypeText:
// 		if p.TextContent == nil {
// 			return json.Unmarshal(data, &p.TextContent)
// 		}
// 	case PartTypeFile:
// 		if p.FileContent == nil {
// 			return json.Unmarshal(data, &p.FileContent)
// 		}
// 	case PartTypeData:
// 		if p.DataContent == nil {
// 			return json.Unmarshal(data, &p.DataContent)
// 		}
// 	default:
// 		return nil
// 	}
//
// 	return nil
// }

// PushNotificationConfig defines configuration for push notifications
type PushNotificationConfig struct {
	URL            string              `json:"url"`
	Token          string              `json:"token,omitempty"` // token unique to this task/session
	Authentication *AuthenticationInfo `json:"authentication,omitempty"`
}

// AuthConfig defines authentication schemes for push notifications
type AuthConfig struct {
	Schemes     []string `json:"schemes"`
	Credentials *string  `json:"credentials,omitempty"`
}

// TaskPushNotificationConfig associates a push config with a specific task
type TaskPushNotificationConfig struct {
	ID                     string                 `json:"id"` // task id
	PushNotificationConfig PushNotificationConfig `json:"pushNotificationConfig"`
}
