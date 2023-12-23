package anthrogo

// AnthropicModel is the model to be used for the completion request.
type AnthropicModel string

const (
	ModelClaude2     AnthropicModel = "claude-2"
	ModelClaude2Dot1 AnthropicModel = "claude-2.1"

	ModelClaudeInstant1     AnthropicModel = "claude-instant-1"
	ModelClaudeInstant1Dot1 AnthropicModel = "claude-instant-1.1"
	ModelClaudeInstant1Dot2 AnthropicModel = "claude-instant-1.2"
)
