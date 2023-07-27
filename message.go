package anthrogo

// Role represents the role of a participant in a conversation. It could either be a "Human" or an "Assistant".
type Role string

const (
	RoleHuman     Role = "Human"
	RoleAssistant Role = "Assistant"
)

// Message represents a single message in a conversation. It includes the Role of the sender and the Content of the message.
type Message struct {
	Role    Role
	Content string
}
