package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/yardbirdsax/bubblewrap"
)

const (
	modelID = "us.anthropic.claude-3-5-sonnet-20241022-v2:0"
)

// Define Message at the package level so both main and queryClaudeWithHistory can use it
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func main() {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile("Joshua-E-Feierman.AdministratorAccess"))
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}
	bedrock := bedrockruntime.NewFromConfig(cfg)

	fmt.Println("Welcome to Claude CLI! Press Ctrl+C to exit.")

	newFunction(ctx, bedrock)
}

func newFunction(ctx context.Context, bedrock *bedrockruntime.Client) {
	const maxHistory = 500
	history := make([]Message, maxHistory)
	historySize := 0

	for {
		prompt, err := bubblewrap.Input("You: ")
		if err != nil {
			log.Fatalf("Error getting prompt from user: %v", err)
		}
		if prompt == "" {
			continue
		}

		fmt.Printf("You: %s\n\n", prompt)

		if historySize == maxHistory {
			// Shift everything left by one (remove oldest message)
			copy(history, history[1:])
			historySize--
		}

		// Add user message
		history[historySize] = Message{Role: "user", Content: prompt}
		historySize++

		response := queryClaudeWithHistory(ctx, bedrock, history[:historySize])

		if historySize == maxHistory {
			// Shift everything left by one (remove oldest message)
			copy(history, history[1:])
			historySize--
		}

		// Add assistant message
		history[historySize] = Message{Role: "assistant", Content: response}
		historySize++

		fmt.Printf("Claude: %s\n\n", response)
	}
}

func queryClaudeWithHistory(ctx context.Context, client *bedrockruntime.Client, history []Message) string {
	type ClaudeRequest struct {
		AnthropicVersion string    `json:"anthropic_version"`
		Messages         []Message `json:"messages"`
		MaxTokens        int       `json:"max_tokens"`
		Temperature      float64   `json:"temperature"`
		StopSequences    []string  `json:"stop_sequences,omitempty"`
	}

	req := ClaudeRequest{
		AnthropicVersion: "bedrock-2023-05-31",
		Messages:         history,
		MaxTokens:        500,
		Temperature:      0.7,
	}

	body, _ := json.Marshal(req)
	res, err := client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(modelID),
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
		Body:        body,
	})
	if err != nil {
		log.Fatalf("Failed to call Claude: %v", err)
	}

	type ClaudeResponseContent struct {
		Text string `json:"text"`
	}

	var output struct {
		Content []ClaudeResponseContent `json:"content"`
	}
	json.Unmarshal(res.Body, &output)

	return output.Content[0].Text
}
