package a2a

import (
	"encoding/json"
	"fmt"
	"time"
)

// AgentCard conveys key information about an agent
type AgentCard struct {
	// Human readable name of the agent
	Name string `json:"name"`

	// Human-readable description of the agent
	Description string `json:"description"`

	// URL to the address the agent is hosted at
	URL string `json:"url"`

	// The service provider of the agent
	Provider *AgentProvider `json:"provider,omitempty"`

	// The version of the agent - format is up to the provider
	Version string `json:"version"`

	// URL to documentation for the agent
	DocumentationURL string `json:"documentationUrl,omitempty"`

	// Optional capabilities supported by the agent
	Capabilities *AgentCapabilities `json:"capabilities"`

	// Authentication requirements for the agent
	Authentication *AgentAuthentication `json:"authentication"`

	// Supported mime types for input
	DefaultInputModes []string `json:"defaultInputModes"`

	// Supported mime types for output
	DefaultOutputModes []string `json:"defaultOutputModes"`

	// Skills are a unit of capability that an agent can perform
	Skills []AgentSkill `json:"skills"`
}

// Provider represents the service provider information
type AgentProvider struct {
	Organization string `json:"organization"`
	URL          string `json:"url"`
}

// AgentCapabilities describes optional capabilities supported by the agent
type AgentCapabilities struct {
	// True if the agent supports SSE
	Streaming bool `json:"streaming,omitempty"`

	// True if the agent can notify updates to client
	PushNotifications bool `json:"pushNotifications,omitempty"`

	// True if the agent exposes status change history for tasks
	StateTransitionHistory bool `json:"stateTransitionHistory,omitempty"`
}

type AuthSchema string

// supported Auth schemas
const (
	Basic  AuthSchema = "Basic"
	Bearer AuthSchema = "Bearer"
)

// AgentAuthentication describes authentication requirements
type AgentAuthentication struct {
	// Supported schemes (e.g. Basic, Bearer)
	Schemes []AuthSchema `json:"schemes"`

	// Credentials a client should use for private cards
	Credentials string `json:"credentials,omitempty"`
}

// AgentSkill represents a unit of capability that an agent can perform
type AgentSkill struct {
	// Unique identifier for the agent's skill
	ID string `json:"id"`

	// Human readable name of the skill
	Name string `json:"name"`

	// Description of the skill
	Description string `json:"description"`

	// Tagwords describing classes of capabilities
	Tags []string `json:"tags"`

	// Example scenarios that the skill can perform
	Examples []string `json:"examples,omitempty"`

	// Supported mime types for input (if different than default)
	InputModes []string `json:"inputModes,omitempty"`

	// Supported mime types for output (if different than default)
	OutputModes []string `json:"outputModes,omitempty"`
}

// TaskStatus represents the state of a task with additional context
type TaskStatus struct {
	State     TaskState  `json:"state"`               // current state of the task
	Message   *Message   `json:"message,omitempty"`   // additional status updates
	Timestamp *time.Time `json:"timestamp,omitempty"` // ISO datetime value
}

// TaskState represents possible states of a task
type TaskState string

const (
	TaskStateSubmitted     TaskState = "submitted"
	TaskStateWorking       TaskState = "working"
	TaskStateInputRequired TaskState = "input-required"
	TaskStateCompleted     TaskState = "completed"
	TaskStateCanceled      TaskState = "canceled"
	TaskStateFailed        TaskState = "failed"
	TaskStateUnknown       TaskState = "unknown"
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
)

// Message represents a communication between user and agent
type Message struct {
	Role     MessageRole    `json:"role"`
	Parts    []Part         `json:"parts"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type MessageWrapper struct {
	Role     MessageRole       `json:"role"`
	Parts    []json.RawMessage `json:"parts"`
	Metadata map[string]any    `json:"metadata,omitempty"`
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

	wrapper.Parts = make([]json.RawMessage, len(m.Parts))
	for i, part := range m.Parts {
		var err error
		switch p := part.(type) {
		case TextPart:
			wrapper.Parts[i], err = json.Marshal(p)
		case FilePart:
			wrapper.Parts[i], err = json.Marshal(p)
		case DataPart:
			wrapper.Parts[i], err = json.Marshal(p)
		case *TextPart:
			wrapper.Parts[i], err = json.Marshal(p)
		case *FilePart:
			wrapper.Parts[i], err = json.Marshal(p)
		case *DataPart:
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

	// Initialize the message structure
	m.Role = temp.Role
	m.Metadata = temp.Metadata
	m.Parts = make([]Part, 0, len(temp.Parts))

	// Process each part in the message
	for _, rawPart := range temp.Parts {
		// First unmarshal to get the type
		var partMap map[string]interface{}
		if err := json.Unmarshal(rawPart, &partMap); err != nil {
			return fmt.Errorf("failed to parse part: %w", err)
		}

		// Get the type value
		typeVal, ok := partMap["type"]
		if !ok {
			return fmt.Errorf("part missing required 'type' field")
		}

		typeStr, ok := typeVal.(string)
		if !ok {
			return fmt.Errorf("part 'type' field must be a string")
		}

		partType := PartType(typeStr)

		// Based on the type, create the appropriate part
		switch partType {
		case PartTypeText:
			var textPart TextPart
			if err := json.Unmarshal(rawPart, &textPart); err != nil {
				return fmt.Errorf("failed to unmarshal text part: %w", err)
			}
			m.Parts = append(m.Parts, textPart)

		case PartTypeFile:
			var filePart FilePart
			if err := json.Unmarshal(rawPart, &filePart); err != nil {
				return fmt.Errorf("failed to unmarshal file part: %w", err)
			}
			m.Parts = append(m.Parts, filePart)

		case PartTypeData:
			var dataPart DataPart
			if err := json.Unmarshal(rawPart, &dataPart); err != nil {
				return fmt.Errorf("failed to unmarshal data part: %w", err)
			}
			m.Parts = append(m.Parts, dataPart)

		default:
			return fmt.Errorf("unknown part type: %s", partType)
		}
	}

	return nil
}

// PartType defines the type of part
type PartType string

const (
	PartTypeText PartType = "text"
	PartTypeFile PartType = "file"
	PartTypeData PartType = "data"
)

// Part represents a message part which can be text, file, or data
type Part interface {
	// some trivial method to glue types together
	partGlue()
}

func (p TextPart) partGlue() {}

type TextPart struct {
	Type     PartType       `json:"type"` // default value should be "text"
	Text     string         `json:"text"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type FilePart struct {
	Type     PartType       `json:"type"` // default value should be "file"
	File     FileContent    `json:"file"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

func (p FilePart) partGlue() {}

type DataPart struct {
	Type     PartType       `json:"type"` // default value should be "data"
	Data     map[string]any `json:"data"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

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
