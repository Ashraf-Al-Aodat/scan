package prompts

import (
	"context"
	"encoding/json"
	"log"

	"github.com/tmc/langchaingo/llms"
)

// updateMessageHistory updates the message history with the assistant's
// response and requested tool calls.
func UpdateMessageHistory(messageHistory []llms.MessageContent, resp *llms.ContentResponse) []llms.MessageContent {
	respchoice := resp.Choices[0]

	assistantResponse := llms.TextParts(llms.ChatMessageTypeAI, respchoice.Content)
	for _, tc := range respchoice.ToolCalls {
		assistantResponse.Parts = append(assistantResponse.Parts, tc)
	}
	return append(messageHistory, assistantResponse)
}

// executeToolCalls executes the tool calls in the response and returns the
// updated message history.
func ExecuteToolCalls(ctx context.Context, llm llms.Model, messageHistory []llms.MessageContent, resp *llms.ContentResponse) []llms.MessageContent {
	for _, toolCall := range resp.Choices[0].ToolCalls {
		switch toolCall.FunctionCall.Name {
		case "flagFile":
			var args struct {
				Flag int `json:"flag"`
			}
			if err := json.Unmarshal([]byte(toolCall.FunctionCall.Arguments), &args); err != nil {
				log.Fatal(err)
			}

			response := flagFile(args.Flag)

			flagCallResponse := llms.MessageContent{
				Role: llms.ChatMessageTypeTool,
				Parts: []llms.ContentPart{
					llms.ToolCallResponse{
						ToolCallID: toolCall.ID,
						Name:       toolCall.FunctionCall.Name,
						Content:    response,
					},
				},
			}
			messageHistory = append(messageHistory, flagCallResponse)
		default:
			log.Fatalf("Unsupported tool: %s", toolCall.FunctionCall.Name)
		}
	}

	return messageHistory
}

const (
	Flag = iota
	Description
)

// availableTools simulates the tools/functions we're making available for the model.
var availableTools = map[int8]llms.Tool{
	Flag: {
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "flagFile",
			Description: "flag a file if it has any sensitive data like passwords, API keys, etc",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"flag": map[string]any{
						"type":        "integer",
						"enum":        []int{0, 1},
						"description": "0: means the file is safe, 1: means the file has some sensitive data",
					},
				},
				"required": []string{"flag"},
			},
		},
	},
}

func Tools(tool int8) llms.CallOption {
	return llms.WithTools([]llms.Tool{availableTools[tool]})
}

const (
	safe = iota
	leak
)

// +llmfunc: a tool used to flag a file if it has any sensitive data like passwords, API keys, etc.
// +param:flag:description: the flag, enum:[0,1]
func flagFile(flag int) string {
	switch flag {
	case safe:
		return "safe"
	case leak:
		return "leak"
	default:
		return "unknowon flag"
	}
}
