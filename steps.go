package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/tmc/langchaingo/prompts"
)

func setupSysPrompt(ai *AI, dbs *DBs) string {
	setupPrompt, _ := dbs.identity.Get("setup")
	philosophyTemplate, _ := dbs.identity.Get("philosophy")
	t := prompts.NewPromptTemplate(philosophyTemplate, nil)
	philosophyPrompt, _ := t.Format(map[string]any{"lang": ai.lang})
	return setupPrompt + "\nUseful to know:\n" + philosophyPrompt
}

func run(ai *AI, dbs *DBs) []openai.ChatCompletionMessage {
	// Run the AI on the main prompt and save the results
	userPrompt, _ := dbs.input.Get("main_prompt")
	messages := ai.Start(
		setupSysPrompt(ai, dbs),
		userPrompt,
	)
	toFiles(messages[len(messages)-1].Content, dbs.workspace)
	return messages
}

// Ask the user if they want to clarify anything and save the results to the workspace
func clarify(ai *AI, dbs *DBs) []openai.ChatCompletionMessage {
	// 1. system prompt
	qaTemplate, _ := dbs.identity.Get("qa")
	t := prompts.NewPromptTemplate(qaTemplate, nil)
	qaPrompt, _ := t.Format(map[string]any{"lang": ai.lang})
	messages := []openai.ChatCompletionMessage{
		ai.SystemMessage(qaPrompt),
	}
	// 2. user specify what want to build
	userPrompt, _ := dbs.input.Get("main_prompt")
	// 3. ask if anything is unclear, then user clarify, repetitive, util everything is clear
	for {
		messages = ai.Next(messages, userPrompt)
		// everything is clear, break
		if strings.ToLower(strings.TrimSpace(messages[len(messages)-1].Content)) == "no" {
			break
		}

		// user clarify to system
		fmt.Print("\n(answer in text, or \"q\" to move on)\n")
		fmt.Scanln(&userPrompt)
		// user input q or nothing, break
		if userPrompt == "" || userPrompt == "q" {
			break
		}
		// after clarify ask system if anything is unclear
		userPrompt += "\n\nIs anything else unclear? " +
			"If yes, only answer in the form:\n" +
			"{remaining unclear areas} remaining questions.\n" +
			"{Next question}\n" +
			"If everything is sufficiently clear, only answer \"no\"."
	}
	fmt.Println()
	return messages
}

// after clarify, run the AI on the main prompt and save the results
func runClarified(ai *AI, dbs *DBs) []openai.ChatCompletionMessage {
	// get the messages from previous step
	var messages []openai.ChatCompletionMessage
	clarifyPrompt, _ := dbs.logs.Get("clarify")
	json.Unmarshal([]byte(clarifyPrompt), &messages)

	// replace system prompt from clarify to execute
	messages[0] = ai.SystemMessage(setupSysPrompt(ai, dbs))
	// run the AI on the main prompt
	mainPrompt, _ := dbs.input.Get("main_prompt")
	messages = ai.Next(messages, mainPrompt)
	// save the results to workspace
	toFiles(messages[len(messages)-1].Content, dbs.workspace)

	return messages
}

var STEPS = []func(*AI, *DBs) []openai.ChatCompletionMessage{
	clarify,
	runClarified,
}
