# go-claude

This is a simple client for using Anthropic's api to get claude completions. It is not an official client. Contributions are welcome!

```
go get github.com/dleviminzi/go-claude
```

## Basic usage

## Streaming completion usage

## Project structure
```
├── client.go                   \\ client definition and options
├── complete.go                 \\ basic completions
├── complete_stream.go          \\ stream completions
├── complete_stream_test.go
├── complete_test.go
├── conversation.go             \\ build conversations with chains of prompts 
├── conversation_test.go
├── examples
│   └── stream
│       └── main.go
├── go.mod
├── go.sum
├── message.go                  \\ messages are the building blocks of conversations
├── mocks
│   └── mocks.go
├── README.md
├── sse_decoder.go              \\ server sent event decoder for streaming completions
└── sse_decoder_test.go
```

## To-do
- [ ] Actually retry and follow limit
- [ ] Use timeout limit 
- [ ] Improve documentation