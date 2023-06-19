package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/tmc/langchaingo/prompts"
)

func setupSysPrompt(ai *AI, dbs *DBs) string {
	setupPrompt, _ := dbs.identity.Get("generate")
	philosophyTemplate, _ := dbs.identity.Get("philosophy")
	t := prompts.NewPromptTemplate(philosophyTemplate, nil)
	philosophyPrompt, _ := t.Format(map[string]any{"lang": ai.lang})
	return setupPrompt + "\nUseful to know:\n" + philosophyPrompt
}

func simpleGen(ai *AI, dbs *DBs) []openai.ChatCompletionMessage {
	// Run the AI on the main prompt and save the results
	userPrompt, _ := dbs.input.Get("main_prompt")
	messages := ai.Start(
		setupSysPrompt(ai, dbs),
		userPrompt,
	)
	// Save the results to workspace
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

func genSpec(ai *AI, dbs *DBs) []openai.ChatCompletionMessage {
	// Generate a spec from the main prompt + clarifications and save the results to the workspace
	userPrompt, _ := dbs.input.Get("main_prompt")
	messages := []openai.ChatCompletionMessage{
		ai.SystemMessage(setupSysPrompt(ai, dbs)),
		ai.SystemMessage("Instructions: " + userPrompt),
	}

	specPrompt, _ := dbs.identity.Get("spec")
	messages = ai.Next(messages, specPrompt)
	dbs.memory.Set("specification", messages[len(messages)-1].Content)
	return messages
}

func respec(ai *AI, dbs *DBs) []openai.ChatCompletionMessage {
	specMessages, _ := dbs.logs.Get("genSpec")
	var messages []openai.ChatCompletionMessage
	json.Unmarshal([]byte(specMessages), &messages)
	respecPrompt, _ := dbs.identity.Get("respec")
	messages = append(messages, ai.SystemMessage(respecPrompt))

	messages = ai.Next(messages, "")
	messages = ai.Next(messages, "Based on the conversation so far, please reiterate the specification for the program."+
		"If there are things that can be improved, please incorporate the improvements."+
		"If you are satisfied with the specification, just write out the specification word by word again.")

	dbs.memory.Set("specification", messages[len(messages)-1].Content)
	return messages
}

func genUnitTests(ai *AI, dbs *DBs) []openai.ChatCompletionMessage {
	//  Generate unit tests based on the specification, that should work.
	userPrompt, _ := dbs.input.Get("main_prompt")
	specPrompt, _ := dbs.memory.Get("specification")
	messages := []openai.ChatCompletionMessage{
		ai.SystemMessage(setupSysPrompt(ai, dbs)),
		ai.UserMessage("Instructions: " + userPrompt),
		ai.UserMessage("Specification: " + specPrompt),
	}

	unitTestPrompt, _ := dbs.identity.Get("unit_tests")
	messages = ai.Next(messages, unitTestPrompt)
	unitTestContent := messages[len(messages)-1].Content
	dbs.memory.Set("unit_tests", unitTestContent)
	toFiles(unitTestContent, dbs.workspace)

	return messages
}

func genClarifiedCode(ai *AI, dbs *DBs) []openai.ChatCompletionMessage {
	// get the messages from previous step
	clarifyContent, _ := dbs.logs.Get("clarify")
	var messages []openai.ChatCompletionMessage
	json.Unmarshal([]byte(clarifyContent), &messages)

	messages[0] = ai.SystemMessage(setupSysPrompt(ai, dbs))
	useQaPrompt, _ := dbs.identity.Get("use_qa")
	messages = ai.Next(messages, useQaPrompt)

	toFiles(messages[len(messages)-1].Content, dbs.workspace)

	return messages
}

func genCode(ai *AI, dbs *DBs) []openai.ChatCompletionMessage {
	// get the messages from previous step
	userPrompt, _ := dbs.input.Get("main_prompt")
	specPrompt, _ := dbs.memory.Get("specification")
	unitTestPrompt, _ := dbs.memory.Get("unit_tests")
	messages := []openai.ChatCompletionMessage{
		ai.SystemMessage(setupSysPrompt(ai, dbs)),
		ai.UserMessage("Instructions: " + userPrompt),
		ai.UserMessage("Specification: " + specPrompt),
		ai.UserMessage("Unit tests: " + unitTestPrompt),
	}

	useQaPrompt, _ := dbs.identity.Get("use_qa")
	messages = ai.Next(messages, useQaPrompt)

	toFiles(messages[len(messages)-1].Content, dbs.workspace)

	return messages
}

func executeEntrypoint(ai *AI, dbs *DBs) []openai.ChatCompletionMessage {
	command, _ := dbs.workspace.Get("run.sh")

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Do you want to execute this code?")
	fmt.Println()
	fmt.Println(command)
	fmt.Println()
	fmt.Println("If yes, press enter, Otherwise, type \"no\"")
	fmt.Println()
	text, _ := reader.ReadString('\n')
	if text == "no" {
		fmt.Println("ok, not executing the code")
		return []openai.ChatCompletionMessage{}
	}
	fmt.Println("Executing the code...")
	fmt.Println()

	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error executing the code")
		fmt.Println(err)
	}
	return []openai.ChatCompletionMessage{}
}

func genEntrypoint(ai *AI, dbs *DBs) []openai.ChatCompletionMessage {
	outputTxt, _ := dbs.workspace.Get("all_output.txt")
	messages := ai.Start(
		`You will get information about a codebase that is currently on disk in the current folder.\n`+
			`From this you will answer with code blocks that includes all the necessary unix terminal commands to`+
			`a) install dependencies `+
			`b) run all necessary parts of the codebase (in parallell if necessary).\n`+
			`Do not install globally. `+
			`Do not use sudo.\n`+
			`Do not explain the code, just give the commands.\n`,
		`Information about the codebase:\n\n`+outputTxt,
	)
	fmt.Println()

	regex := regexp.MustCompile("(?s)```\\S*\\n(.+?)```")
	if matches := regex.FindAllStringSubmatch(messages[len(messages)-1].Content, -1); matches != nil {
		runShContent := make([]string, 0)

		for _, match := range matches {
			// match[0] 包含整个匹配到的文本，但我们只关心第一个子匹配
			if len(match) > 1 {
				runShContent = append(runShContent, match[1])
			}
		}

		dbs.workspace.Set("run.sh", strings.Join(runShContent, "\n"))
	}

	return messages
}

func useFeedback(ai *AI, dbs *DBs) []openai.ChatCompletionMessage {
	userPrompt, _ := dbs.input.Get("main_prompt")
	outputTxt, _ := dbs.workspace.Get("all_output.txt")
	useFeedbackPrompt, _ := dbs.identity.Get("use_feedback")
	messages := []openai.ChatCompletionMessage{
		ai.SystemMessage(setupSysPrompt(ai, dbs)),
		ai.UserMessage("Instructions: " + userPrompt),
		ai.AssistantMessage(outputTxt),
		ai.SystemMessage(useFeedbackPrompt),
	}
	feedback, _ := dbs.memory.Get("feedback")
	messages = ai.Next(messages, feedback)
	toFiles(messages[len(messages)-1].Content, dbs.workspace)
	return messages
}

func fixCode(ai *AI, dbs *DBs) []openai.ChatCompletionMessage {
	genCodes, _ := dbs.logs.Get("gen_code")
	var messages []openai.ChatCompletionMessage
	json.Unmarshal([]byte(genCodes), &messages)
	codeOutput := messages[len(messages)-1].Content
	userPrompt, _ := dbs.input.Get("main_prompt")
	fixCodePrompt, _ := dbs.identity.Get("fix_code")
	messages = []openai.ChatCompletionMessage{
		ai.SystemMessage(setupSysPrompt(ai, dbs)),
		ai.UserMessage("Instructions: " + userPrompt),
		ai.UserMessage(codeOutput),
		ai.SystemMessage(fixCodePrompt),
	}
	messages = ai.Next(messages, "Please fix any errors in the code above.")
	toFiles(messages[len(messages)-1].Content, dbs.workspace)
	return messages
}

var STEPS = map[string][]func(*AI, *DBs) []openai.ChatCompletionMessage{
	"default": {
		genSpec,
		genUnitTests,
		genCode,
		genEntrypoint,
		executeEntrypoint,
	},
	"benchmark": {
		genSpec,
		genUnitTests,
		genCode,
		fixCode,
		genEntrypoint,
	},
	"simple": {
		simpleGen,
		genEntrypoint,
		executeEntrypoint,
	},
	"clarify": {
		clarify,
		genClarifiedCode,
		genEntrypoint,
		executeEntrypoint,
	},
	"respec": {
		genSpec,
		respec,
		genUnitTests,
		genCode,
		genEntrypoint,
		executeEntrypoint,
	},
	"execute_only": {
		executeEntrypoint,
	},
	"use_feedback": {
		useFeedback,
	},
}
