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

	systemPrompt := "you are an expert in all things bananas"

	// Read user input for the prompt
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your prompt: ")
	userPrompt, _ := reader.ReadString('\n')
	userPrompt = strings.TrimSuffix(userPrompt, "\n")

	resp, err := c.MessageRequest(context.Background(), anthrogo.MessagePayload{
		Model: anthrogo.ModelClaude3Dot5Sonnet,
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
