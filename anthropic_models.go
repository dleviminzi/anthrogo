package anthrogo

// AnthropicModel is the model to be used for the completion request.
// TODO: model may not include completions or messages API support (add flags)
type AnthropicModel string

const (
	Claude3Opus   AnthropicModel = "claude-3-opus-20240229"
	Claude3Sonnet AnthropicModel = "claude-3-sonnet-20240229"

	ModelClaude2     AnthropicModel = "claude-2"
	ModelClaude2Dot1 AnthropicModel = "claude-2.1"

	ModelClaudeInstant1     AnthropicModel = "claude-instant-1"
	ModelClaudeInstant1Dot1 AnthropicModel = "claude-instant-1.1"
	ModelClaudeInstant1Dot2 AnthropicModel = "claude-instant-1.2"
)
