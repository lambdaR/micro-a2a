package main

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/micro/micro-a2a/pkg/a2a"
)

func main() {
	client := a2a.NewA2AClient()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	params := a2a.TaskSendParams{
		ID: uuid.NewString(),
		Message: a2a.Message{
			Role: a2a.MessageRoleUser,
			Parts: []a2a.Part{
				&a2a.TextPart{
					Type: a2a.PartTypeText,
					Text: "what is the current time",
				},
			},
		},
	}

	res, err := client.SendReq(ctx, a2a.TasksSend, params, "http://localhost:8081/AgentDebo")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Response: %+v\n", res)
	fmt.Println("")

	ctxStream, cancelStream := context.WithCancel(context.Background())
	defer cancelStream()

	paramsStream := a2a.TaskSendParams{
		ID:        uuid.NewString(),
		SessionID: uuid.NewString(),
		Message: a2a.Message{
			Role: a2a.MessageRoleUser,
			Parts: []a2a.Part{
				&a2a.TextPart{
					Type: a2a.PartTypeText,
					Text: "time ticks every 5 seconds",
				},
			},
		},
	}

	resChan, err := client.SendReqStream(ctxStream, a2a.TasksSendSubscribe, paramsStream, "http://localhost:8081/AgentDebo/stream")
	if err != nil {
		fmt.Println("Stream error:", err)
		return
	}

	for {
		select {
		case e, ok := <-resChan:
			if !ok {
				fmt.Println("Channel closed, exiting...")
				return
			}
			fmt.Printf("ID: %v\nType: %v\nData: %v\n", e.LastEventID, e.Type, e.Data)
		case <-ctxStream.Done():
			fmt.Println("Context cancelled, exiting...")
			return
		}
	}
}
