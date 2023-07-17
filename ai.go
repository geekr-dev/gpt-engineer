package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

var defaultModel = "gpt-4"
var defaultTemperature = 0.1
var defaultLang = "English"

type AI struct {
	model       string
	temperature float64
	lang        string
	client      *openai.Client
}

func NewAI(model string, temperature float64, lang string) *AI {
	var clientConfig openai.ClientConfig
	apiKey := os.Getenv("OPENAI_API_KEY")
	apiBase := os.Getenv("OPENAI_API_BASE")
	if apiBase == "" {
		clientConfig = openai.DefaultConfig(apiKey)
	} else {
		clientConfig = openai.DefaultAzureConfig(apiKey, apiBase)
	}
	client := openai.NewClientWithConfig(clientConfig)
	ai := &AI{model: model, client: client, temperature: temperature, lang: lang}

	// Azure uses deployments and is not exposed in the openai api so model is assumed to be okay
	if clientConfig.APIType == openai.APITypeAzure {
		return ai
	}

	_, err := client.GetModel(context.Background(), model)
	if err != nil {
		fmt.Println("Model gpt-4 not available for provided api key reverting to gpt-3.5.turbo. Sign up for the gpt-4 wait list here: https://openai.com/waitlist/gpt-4-api")
		ai.model = "gpt-3.5-turbo"
	}
	return ai
}

func (ai *AI) Start(system, user string) []openai.ChatCompletionMessage {
	messages := []openai.ChatCompletionMessage{
		ai.SystemMessage(system),
		ai.UserMessage(user),
	}

	return ai.Next(messages, "")
}

func (ai *AI) SystemMessage(message string) openai.ChatCompletionMessage {
	return openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: message,
	}
}

func (ai *AI) UserMessage(message string) openai.ChatCompletionMessage {
	return openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: message,
	}
}

func (ai *AI) AssistantMessage(message string) openai.ChatCompletionMessage {
	return openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: message,
	}
}

func (ai *AI) Next(messages []openai.ChatCompletionMessage, prompt string) []openai.ChatCompletionMessage {
	if prompt != "" {
		messages = append(messages, ai.UserMessage(prompt))
	}

	request := openai.ChatCompletionRequest{
		Model:    ai.model,
		Messages: messages,
	}
	stream, err := ai.client.CreateChatCompletionStream(context.Background(), request)
	if err != nil {
		panic(err)
	}
	defer stream.Close()

	var contents []string
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			log.Println("\nStream finished")
			break
		}

		if err != nil {
			fmt.Printf("\nStream error: %v\n", err)
			break
		}

		content := response.Choices[0].Delta.Content
		fmt.Printf("%s", content)
		contents = append(contents, content)
	}

	return append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: strings.Join(contents, ""),
	})
}
