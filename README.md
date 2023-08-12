# anthropic go (anthrogo)
[![Go Reference](https://pkg.go.dev/badge/github.com/dleviminzi/go-anthropic.svg)](https://pkg.go.dev/github.com/dleviminzi/go-anthropic)
[![Go Report Card](https://goreportcard.com/badge/github.com/dleviminzi/go-anthropic)](https://goreportcard.com/report/github.com/dleviminzi/go-anthropic)
[![codecov](https://codecov.io/gh/dleviminzi/go-anthropic/branch/main/graph/badge.svg?token=OP2W7ENYN5)](https://codecov.io/gh/dleviminzi/go-anthropic)

This is a simple client for using Anthropic's api to get claude completions. It is not an official client. Contributions are welcome!

## Installation
```
go get github.com/dleviminzi/go-anthropic
```

## Basic usage
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
	resp, err := c.Complete(&anthrogo.CompletePayload{
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
## Streaming completion usage
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
	completeStreamResp, err := c.CompleteStream(&anthrogo.CompletePayload{
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


## Project structure
```
├── anthropic_models.go         <- available anthropic models 
├── client.go                   <- client definition, options, and request formatting
├── complete.go                 <- basic and stream completion requests/responses
├── complete_test.go
├── conversation.go             <- build conversations with chains of prompts  
├── conversation_test.go
├── message.go                  <- messages are the building blocks of conversations
├── sse_decoder.go              <- server sent event decoder for streaming completions
├── sse_decoder_test.go
├── go.mod
├── go.sum
├── README.md
├── LICENSE
├── examples
│   ├── completion
│   │   └── basic_example.go
│   └── stream
│       └── stream_example.go
└── mocks
    └── mocks.go
```

## To-do
- [ ] Tokenizer
