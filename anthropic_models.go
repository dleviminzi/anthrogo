package anthrogo

// AnthropicModel is the model to be used for the completion request.
// TODO: model may not include completions or messages API support (add flags)
type AnthropicModel string

const (
	ModelClaude3Opus   AnthropicModel = "claude-3-opus-20240229"
	ModelClaude3Sonnet AnthropicModel = "claude-3-sonnet-20240229"
	ModelClaude3Haiku  AnthropicModel = "claude-3-haiku-20240307"

	ModelClaude2     AnthropicModel = "claude-2.0"
	ModelClaude2Dot1 AnthropicModel = "claude-2.1"

	ModelClaudeInstant1Dot2 AnthropicModel = "claude-instant-1.2"
)
