# anthropic go (anthrogo)
[![Go Reference](https://pkg.go.dev/badge/github.com/dleviminzi/go-anthropic.svg)](https://pkg.go.dev/github.com/dleviminzi/go-anthropic)
[![Go Report Card](https://goreportcard.com/badge/github.com/dleviminzi/go-anthropic)](https://goreportcard.com/report/github.com/dleviminzi/go-anthropic)
[![codecov](https://codecov.io/gh/dleviminzi/go-anthropic/branch/main/graph/badge.svg?token=OP2W7ENYN5)](https://codecov.io/gh/dleviminzi/go-anthropic)

This is a simple client for using Anthropic's api to get claude completions. It is not an official client. Contributions are welcome!

## Installation
```
go get github.com/dleviminzi/anthrogo
```

## Basic usage 

### Message API
```go
func main() {
	c, err := anthrogo.NewClient()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	systemPrompt := "you are an expert in all things bananas"

	// Read user input for the prompt
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your prompt: ")
	userPrompt, _ := reader.ReadString('\n')
	userPrompt = strings.TrimSuffix(userPrompt, "\n")

	resp, err := c.MessageRequest(context.Background(), anthrogo.MessagePayload{
		Model: anthrogo.ModelClaude3Opus,
		Messages: []anthrogo.Message{{
			Role: anthrogo.RoleTypeUser,
			Content: []anthrogo.MessageContent{{
				Type: anthrogo.ContentTypeText,
				Text: &userPrompt,
			}},
		}},
		System:    &systemPrompt,
		MaxTokens: 1000,
	})
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	fmt.Println(resp.Content[0].Text)
}
```

### Completion API
```go
func main() {
	c, err := anthrogo.NewClient()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// Read user input for the prompt
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your prompt: ")
	userPrompt, _ := reader.ReadString('\n')
	userPrompt = strings.TrimSuffix(userPrompt, "\n")

	// Create conversation with user input
	conversation := anthrogo.NewConversation()
	conversation.AddMessage(anthrogo.RoleHuman, userPrompt)

	// Set up the payload and send completion stream request
	resp, err := c.CompletionRequest(context.Background(), anthrogo.CompletionPayload{
		MaxTokensToSample: 256,
		Model:             anthrogo.ModelClaude2,
		Prompt:            conversation.GeneratePrompt(),
	})
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	fmt.Println(resp.Completion)

	// Add claude's response to conversation for further prompting...
	conversation.AddMessage(anthrogo.RoleAssistant, resp.Completion)
}
```
## Message Streaming (coming soon)

## Completion Streaming
[streaming-completion-example (trimmed).webm](https://github.com/dleviminzi/go-anthropic/assets/51272568/14f80831-a53b-47bd-a8e3-67fe4c279df6)
<details>
<summary>Code</summary>	
	
```go
func main() {
	// Create a new client
	// optionally provide api key otherwise we will look for it in ANTHROPIC_API_KEY variable
	c, err := anthrogo.NewClient()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// Read user input for the prompt
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your prompt: ")
	userPrompt, _ := reader.ReadString('\n')
	userPrompt = strings.TrimSuffix(userPrompt, "\n")

	// Create conversation with user input
	conversation := anthrogo.NewConversation()
	conversation.AddMessage(anthrogo.RoleHuman, userPrompt)

	// Set up the payload and send completion stream request
	completeStreamResp, _ := c.StreamingCompletionRequest(context.Background(), anthrogo.CompletionPayload{
		MaxTokensToSample: 256,
		Model:             anthrogo.ModelClaude2,
		Prompt:            conversation.GeneratePrompt(),
		CompleteOptions: anthrogo.CompleteOptions{
			Stream:      true,
			Temperature: 1,
		},
	})

	// Ensure that the request is canceled after timeout (default 1 minute)
	defer completeStreamResp.Cancel()

	// Ensure that the stream response body is closed when the function returns
	defer completeStreamResp.Close()

	// Continually read from the response until an error or EOF is encountered
	for {
		event, err := completeStreamResp.Decode()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		if event != nil {
			fmt.Printf("%s", event.Data.Completion)
		}
	}
}
```
</details>

