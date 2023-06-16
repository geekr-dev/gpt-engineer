# GPT Engineer

> [AntonOsika/gpt-engineer](https://github.com/AntonOsika/gpt-engineer) 的 Go 语言实现版本, 支持通过 `-lang` 参数切换交互语言。

**通过 Prompt 一句话描述你的需求，AI 如果不清楚会要求你进行澄清，需求明确后 AI 会自动为你编写实现代码构建应用**

GPT Engineer 的设计旨在易于适应、扩展以及让你的智能代理按照指定的代码风格学习。它基于提示（Prompt）生成整个代码库。

## 项目理念
- 轻松获得价值
- 灵活且易于添加新的自定义“AI步骤”。请参见`steps.go`。
- 逐步建立以下用户体验：
  1. 高层次的提示
  2. 使 AI 能够记住随时间推移的反馈
- 人工智能与人类之间快速而顺畅的交替
- 简单性，所有计算都是“可恢复的”并持久化到文件系统

## 快速入门

**初始化**:
- `go mod tidy`
- `export OPENAI_API_KEY=[your api key]` 如果你的 KEY 没有访问 GPT-4 的权限，会降级到使用 GPT-3.5

**运行**:
- 编辑 `example/main_prompt` 通过一句话 Prompt 指定你想要构建的应用
- 运行 `go run .`（注意后面的 `.` 不能省略）, 默认与 AI 交互的语言是英语, 你可以通过 `-lang` 参数进行切换，比如想要使用中文，则通过 `go run . -lang=Chinese` 启动应用即可

**结果**:
- 检查 `example/workspace/all_output.txt` 中生成的内容

### 局限性

实现额外的思维链提示，例如 [Reflexion](https://github.com/noahshinn024/reflexion)，应该能让它更可靠，不会错过主提示中请求的功能。

## 功能

您可以通过编辑 `identity` 文件夹中的文件来指定 AI 代理的“身份”。

目前，编辑身份和发展主提示是让代理在项目之间记住事物的方法。

`steps.go` 中的每个步骤都将与 GPT-4 的通信历史记录在 `logs` 文件夹中。