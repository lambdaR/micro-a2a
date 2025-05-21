package main

import (
	"time"

	"github.com/google/uuid"
	"github.com/micro/micro-a2a/pkg/a2a"
	"go-micro.dev/v5/store"
)

type MyAgentHandlers struct {
	// Add your related handler stuff here
}

func (a *MyAgentHandlers) TaskHandler(req a2a.JSONRPCRequest) a2a.JSONRPCResponse {
	prompt := req.Params.(a2a.TaskSendParams).Message.Parts[0].(a2a.TextPart).Text

	return a2a.JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: &a2a.Task{
			ID:        uuid.New().String(),
			ContextID: uuid.New().String(),
			Status: a2a.TaskStatus{
				State: a2a.TaskStateCompleted,
			},
			Artifacts: []a2a.Artifact{
				{
					Name: prompt,
					Parts: []a2a.Part{
						a2a.TextPart{
							Kind: a2a.PartTypeText,
							Text: time.Now().String(),
						},
					},
				},
			},
		},
		Error: nil,
	}
}

func (a *MyAgentHandlers) StreamHandler(req a2a.JSONRPCRequest, res chan a2a.JSONRPCResponse) {
	tickChan := time.NewTicker(time.Second * 2)
	defer tickChan.Stop()

	timeout := time.After(time.Second * 60)

	for {
		select {
		case <-tickChan.C:
			res <- a2a.JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: &a2a.TaskArtifactUpdateEvent{
					ID: uuid.New().String(),
					Artifact: a2a.Artifact{
						Name: "time ticks every 1 second",
						Parts: []a2a.Part{
							a2a.TextPart{
								Kind: a2a.PartTypeText,
								Text: time.Now().String(),
							},
						},
					},
				},
				Error: nil,
			}
		case <-timeout:
			res <- a2a.JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: &a2a.TaskStatusUpdateEvent{
					ID:    uuid.New().String(),
					Final: true,
					Status: a2a.TaskStatus{
						State: a2a.TaskStateCompleted,
					},
				},
			}

			close(res)
			return
		}
	}
}

func main() {
	agentCard := new(a2a.AgentCard)
	agentCard.Name = "Agent Debo"
	agentCard.URL = ":8081"
	agentCard.Capabilities = &a2a.AgentCapabilities{} // Initialize Capabilities struct
	agentCard.Capabilities.Streaming = true

	agentHandlers := new(MyAgentHandlers)

	agent := a2a.NewAgent(
		*agentCard,
		a2a.WithStore(store.NewMemoryStore()),
		a2a.WithAgentHandler(agentHandlers),
		a2a.WithAgentStreamHandler(agentHandlers),
	)

	agent.SwitchOn()
}
