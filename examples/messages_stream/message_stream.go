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

	r, _, err := c.MessageStreamRequest(context.Background(), anthrogo.MessagePayload{
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
	defer r.Close()

	// Create an SSEDecoder
	decoder := anthrogo.NewMessageSSEDecoder(r)
	for {
		message, err := decoder.Decode(anthrogo.DecodeOptions{ContentOnly: true})
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Print(err)
			continue
		}

		if message.Event == "message_stop" {
			break
		}

		fmt.Print(message.Data.Content)
	}
}
