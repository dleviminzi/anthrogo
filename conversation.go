package anthrogo

import (
	"fmt"
)

// Conversation is a structure holding all messages of the conversation.
type Conversation struct {
	Messages []Message
}

// NewConversation creates and returns a new Conversation.
func NewConversation() *Conversation {
	return &Conversation{}
}

// AddMessage appends a new message to the conversation. Role indicates if the message is from the "Assistant" or the "User".
func (c *Conversation) AddMessage(role Role, content string) {
	c.Messages = append(c.Messages, Message{Role: role, Content: content})
}

// GeneratePrompt formats the conversation into a string which can be used as a prompt for the assistant.
// The prompt ends with an empty Assistant message, indicating where the assistant's next message should go.
func (c *Conversation) GeneratePrompt() string {
	prompt := ""

	for _, msg := range c.Messages {
		prompt += fmt.Sprintf("\n\n%s: %s", msg.Role, msg.Content)
	}

	// Always ends with an empty Assistant message for the next completion
	prompt += fmt.Sprintf("\n\nAssistant:")

	return prompt
}
