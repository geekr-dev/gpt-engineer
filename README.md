# GPT Engineer

[中文文档](https://github.com/geekr-dev/gpt-engineer/blob/master/README_zh.md)

> Golang implement version of [AntonOsika/gpt-engineer](https://github.com/AntonOsika/gpt-engineer), With support to switch language by `-lang` argument.

**Specify what you want it to build, the AI asks for clarification, and then builds it.**

GPT Engineer is made to be easy to adapt, extend, and make your agent learn how you want your code to look. It generates an entire codebase based on a prompt. 

## Project philosophy
- Simple to get value
- Flexible and easy to add new own "AI steps". See `steps.go`.
- Incrementally build towards a user experience of:
  1. high level prompting
  2. giving feedback to the AI that it will remember over time
- Fast handovers back and forth between AI and human
- Simplicity, all computation is "resumable" and persisted to the filesystem


## Usage

### Setup:

**OpenAI**
- `go mod tidy`
- `export OPENAI_API_KEY=[your api key]` with a key that has GPT4 access, if you don't have GPT4 access, then it will use GPT-3.5 as a fallback.

**Azure**
- `export OPENAI_BASE=[your Azure OpenAI endpoint]` should be _https://[deployment name].openai.azure.com/_

### Run:
- Edit `example/main_prompt` to specify what you want to build
- Run `go run .`, default language is English, if you want to use other language, and `-lang` argument like `go run . -lang=Chinese`

**(optional) Azure**
- use the argument `-model [your deployment name]`

### Results:
- Check the generated files in `example/workspace`

## Limitations

Implementing additional chain of thought prompting, e.g. [Reflexion](https://github.com/noahshinn024/reflexion), should be able to make it more reliable and not miss requested functionality in the main prompt.

## Features

You can specify the "identity" of the AI agent by editing the files in the `identity` folder.

Editing the identity, and evolving the main_prompt, is currently how you make the agent remember things between projects.

Each step in `steps.go` will have its communication history with GPT4 stored in the `logs` folder.

