package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/dleviminzi/anthrogo"
)

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
	resp, err := c.Complete(context.Background(), anthrogo.CompletePayload{
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
