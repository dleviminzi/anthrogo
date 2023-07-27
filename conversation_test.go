package anthrogo

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddMessage(t *testing.T) {
	conv := NewConversation()

	humanMsg := "Hello, Assistant!"
	assistantMsg := "Hello, User!"

	conv.AddMessage(RoleHuman, humanMsg)
	conv.AddMessage(RoleAssistant, assistantMsg)

	assert.Equal(t, 2, len(conv.Messages), "Expected message count is not correct")

	assert.Equal(t, RoleHuman, conv.Messages[0].Role, "The role of the first message is not correct")
	assert.Equal(t, humanMsg, conv.Messages[0].Content, "The content of the first message is not correct")

	assert.Equal(t, RoleAssistant, conv.Messages[1].Role, "The role of the second message is not correct")
	assert.Equal(t, assistantMsg, conv.Messages[1].Content, "The content of the second message is not correct")
}

func TestGeneratePrompt(t *testing.T) {
	conv := NewConversation()

	humanMsg := "Hello, Assistant!"
	assistantMsg := "Hello, RoleHuman!"

	conv.AddMessage(RoleHuman, humanMsg)
	conv.AddMessage(RoleAssistant, assistantMsg)

	expectedPrompt := fmt.Sprintf("\n\n%s: %s", RoleHuman, humanMsg)
	expectedPrompt += fmt.Sprintf("\n\n%s: %s", RoleAssistant, assistantMsg)
	expectedPrompt += "\n\nAssistant:"

	assert.Equal(t, expectedPrompt, conv.GeneratePrompt(), "The generated prompt is not correct")
}
