package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/dleviminzi/anthrogo"
)

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
