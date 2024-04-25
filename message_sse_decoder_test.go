package anthrogo

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessageSSEDecoder_Decode(t *testing.T) {
	testCases := []struct {
		name           string
		input          string
		expectedEvents []*MessageEventPayload
		expectedError  error
		options        DecodeOptions
	}{
		{
			name: "valid events",
			input: `event: message_start
data: {"type": "message_start", "message": {"id": "1", "type": "text_completion", "role": "assistant", "content": [], "model": "claude", "stop_reason": "stop_sequence", "stop_sequence": null, "usage": {"input_tokens": 5, "output_tokens": 0}}}

event: content_block_start
data: {"type": "content_block_start", "index": 0, "content_block": {"type": "text", "text": "Hello"}}

event: content_block_delta
data: {"type": "content_block_delta", "index": 0, "delta": {"type": "text", "text": " world!"}}

event: message_stop
data: {"type": "message_stop"}
`,
			expectedEvents: []*MessageEventPayload{
				{
					Event: "message_start",
					Data: EventData{
						Data: MessageStart{
							Type: "message_start",
							Message: struct {
								ID           string   "json:\"id\""
								Type         string   "json:\"type\""
								Role         string   "json:\"role\""
								Content      []string "json:\"content\""
								Model        string   "json:\"model\""
								StopReason   string   "json:\"stop_reason\""
								StopSequence string   "json:\"stop_sequence\""
								Usage        struct {
									InputTokens  int "json:\"input_tokens\""
									OutputTokens int "json:\"output_tokens\""
								} "json:\"usage\""
							}{
								ID:           "1",
								Type:         "text_completion",
								Role:         "assistant",
								Content:      []string{},
								Model:        "claude",
								StopReason:   "stop_sequence",
								StopSequence: "",
								Usage: struct {
									InputTokens  int "json:\"input_tokens\""
									OutputTokens int "json:\"output_tokens\""
								}{InputTokens: 5, OutputTokens: 0},
							},
						},
					},
				},
				{
					Event: "content_block_start",
					Data: EventData{
						Content: "Hello",
						Data: ContentBlockStart{
							Type:  "content_block_start",
							Index: 0,
							ContentBlock: struct {
								Type string "json:\"type\""
								Text string "json:\"text\""
							}{
								Type: "text",
								Text: "Hello",
							},
						},
					},
				},
				{
					Event: "content_block_delta",
					Data: EventData{
						Content: " world!",
						Data: ContentBlockDelta{
							Type:  "content_block_delta",
							Index: 0,
							Delta: struct {
								Type string "json:\"type\""
								Text string "json:\"text\""
							}{
								Type: "text",
								Text: " world!",
							},
						},
					},
				},
				{
					Event: "message_stop",
					Data: EventData{
						Data: MessageStopData{
							Type: "message_stop",
						},
					},
				},
			},
		},
		{
			name: "error event",
			input: `event: error
data: {"type": "error", "error": {"type": "invalid_request_error", "message": "Invalid model"}}
`,
			expectedError: errors.New("error(invalid_request_error) -  Invalid model"),
		},
		{
			name:  "empty input",
			input: "",
		},
		{
			name: "invalid SSE format",
			input: `event
data: {"type": "message_start"}
`,
			expectedError: errors.New("invalid SSE format"),
		},
		{
			name: "content only",
			input: `event: message_start
		data: {"type": "message_start", "message": {"id": "1", "type": "text_completion", "role": "assistant", "content": [], "model": "claude", "stop_reason": "stop_sequence", "stop_sequence": null, "usage": {"input_tokens": 5, "output_tokens": 0}}}
		
		event: content_block_start
		data: {"type": "content_block_start", "index": 0, "content_block": {"type": "text", "text": "Hello"}}
		
		event: content_block_delta
		data: {"type": "content_block_delta", "index": 0, "delta": {"type": "text", "text": " world!"}}
		
		event: message_stop
		data: {"type": "message_stop"}
		`,
			options: DecodeOptions{
				ContentOnly: true,
			},
			expectedEvents: []*MessageEventPayload{
				{
					Event: "content_block_start",
					Data: EventData{
						Content: "Hello",
						Data: ContentBlockStart{
							Type:  "content_block_start",
							Index: 0,
							ContentBlock: struct {
								Type string "json:\"type\""
								Text string "json:\"text\""
							}{
								Type: "text",
								Text: "Hello",
							},
						},
					},
				},
				{
					Event: "content_block_delta",
					Data: EventData{
						Content: " world!",
						Data: ContentBlockDelta{
							Type:  "content_block_delta",
							Index: 0,
							Delta: struct {
								Type string "json:\"type\""
								Text string "json:\"text\""
							}{
								Type: "text",
								Text: " world!",
							},
						},
					},
				},
				{
					Event: "message_stop",
					Data: EventData{
						Data: MessageStopData{
							Type: "message_stop",
						},
					},
				},
			},
		},
		{
			name: "content_block_stop event",
			input: `event: message_start
		data: {"type": "message_start", "message": {"id": "1", "type": "text_completion", "role": "assistant", "content": [], "model": "claude", "stop_reason": "stop_sequence", "stop_sequence": null, "usage": {"input_tokens": 5, "output_tokens": 0}}}
		
		event: content_block_start
		data: {"type": "content_block_start", "index": 0, "content_block": {"type": "text", "text": "Hello"}}
		
		event: content_block_stop
		data: {"type": "content_block_stop", "index": 0}
		
		event: message_stop
		data: {"type": "message_stop"}
		`,
			expectedEvents: []*MessageEventPayload{
				{
					Event: "message_start",
					Data: EventData{
						Data: MessageStart{
							Type: "message_start",
							Message: struct {
								ID           string   "json:\"id\""
								Type         string   "json:\"type\""
								Role         string   "json:\"role\""
								Content      []string "json:\"content\""
								Model        string   "json:\"model\""
								StopReason   string   "json:\"stop_reason\""
								StopSequence string   "json:\"stop_sequence\""
								Usage        struct {
									InputTokens  int "json:\"input_tokens\""
									OutputTokens int "json:\"output_tokens\""
								} "json:\"usage\""
							}{
								ID:           "1",
								Type:         "text_completion",
								Role:         "assistant",
								Content:      []string{},
								Model:        "claude",
								StopReason:   "stop_sequence",
								StopSequence: "",
								Usage: struct {
									InputTokens  int "json:\"input_tokens\""
									OutputTokens int "json:\"output_tokens\""
								}{
									InputTokens:  5,
									OutputTokens: 0,
								},
							},
						},
					},
				},
				{
					Event: "content_block_start",
					Data: EventData{
						Content: "Hello",
						Data: ContentBlockStart{
							Type:  "content_block_start",
							Index: 0,
							ContentBlock: struct {
								Type string "json:\"type\""
								Text string "json:\"text\""
							}{
								Type: "text",
								Text: "Hello",
							},
						},
					},
				},
				{
					Event: "content_block_stop",
					Data: EventData{
						Data: ContentBlockStop{
							Type:  "content_block_stop",
							Index: 0,
						},
					},
				},
				{
					Event: "message_stop",
					Data: EventData{
						Data: MessageStopData{
							Type: "message_stop",
						},
					},
				},
			},
		},
		{
			name: "message_delta event",
			input: `event: message_start
		data: {"type": "message_start", "message": {"id": "1", "type": "text_completion", "role": "assistant", "content": [], "model": "claude", "stop_reason": "stop_sequence", "stop_sequence": null, "usage": {"input_tokens": 5, "output_tokens": 0}}}
		
		event: message_delta
		data: {"type": "message_delta", "delta": {"output_tokens": 10}, "usage": {"output_tokens": 10}}
		
		event: message_stop
		data: {"type": "message_stop"}
		`,
			expectedEvents: []*MessageEventPayload{
				{
					Event: "message_start",
					Data: EventData{
						Data: MessageStart{
							Type: "message_start",
							Message: struct {
								ID           string   "json:\"id\""
								Type         string   "json:\"type\""
								Role         string   "json:\"role\""
								Content      []string "json:\"content\""
								Model        string   "json:\"model\""
								StopReason   string   "json:\"stop_reason\""
								StopSequence string   "json:\"stop_sequence\""
								Usage        struct {
									InputTokens  int "json:\"input_tokens\""
									OutputTokens int "json:\"output_tokens\""
								} "json:\"usage\""
							}{
								ID:           "1",
								Type:         "text_completion",
								Role:         "assistant",
								Content:      []string{},
								Model:        "claude",
								StopReason:   "stop_sequence",
								StopSequence: "",
								Usage: struct {
									InputTokens  int "json:\"input_tokens\""
									OutputTokens int "json:\"output_tokens\""
								}{
									InputTokens:  5,
									OutputTokens: 0,
								},
							},
						},
					},
				},
				{
					Event: "message_delta",
					Data: EventData{
						Data: MessageDelta{
							Type: "message_delta",
							Delta: map[string]interface{}{
								"output_tokens": float64(10),
							},
							Usage: struct {
								OutputTokens int "json:\"output_tokens\""
							}{
								OutputTokens: 10,
							},
						},
					},
				},
				{
					Event: "message_stop",
					Data: EventData{
						Data: MessageStopData{
							Type: "message_stop",
						},
					},
				},
			},
		},
		{
			name: "ping event",
			input: `event: message_start
		data: {"type": "message_start", "message": {"id": "1", "type": "text_completion", "role": "assistant", "content": [], "model": "claude", "stop_reason": "stop_sequence", "stop_sequence": null, "usage": {"input_tokens": 5, "output_tokens": 0}}}
		
		event: ping
		data: {"type": "ping"}
		
		event: message_stop
		data: {"type": "message_stop"}
		`,
			expectedEvents: []*MessageEventPayload{
				{
					Event: "message_start",
					Data: EventData{
						Data: MessageStart{
							Type: "message_start",
							Message: struct {
								ID           string   "json:\"id\""
								Type         string   "json:\"type\""
								Role         string   "json:\"role\""
								Content      []string "json:\"content\""
								Model        string   "json:\"model\""
								StopReason   string   "json:\"stop_reason\""
								StopSequence string   "json:\"stop_sequence\""
								Usage        struct {
									InputTokens  int "json:\"input_tokens\""
									OutputTokens int "json:\"output_tokens\""
								} "json:\"usage\""
							}{
								ID:           "1",
								Type:         "text_completion",
								Role:         "assistant",
								Content:      []string{},
								Model:        "claude",
								StopReason:   "stop_sequence",
								StopSequence: "",
								Usage: struct {
									InputTokens  int "json:\"input_tokens\""
									OutputTokens int "json:\"output_tokens\""
								}{
									InputTokens:  5,
									OutputTokens: 0,
								},
							},
						},
					},
				},
				{
					Event: "ping",
					Data: EventData{
						Data: PingData{
							Type: "ping",
						},
					},
				},
				{
					Event: "message_stop",
					Data: EventData{
						Data: MessageStopData{
							Type: "message_stop",
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			decoder := NewMessageSSEDecoder(strings.NewReader(tc.input))

			var events []*MessageEventPayload
			for {
				event, err := decoder.Decode(tc.options)
				if err != nil {
					if tc.expectedError != nil {
						assert.EqualError(t, err, tc.expectedError.Error())
					} else {
						require.NoError(t, err)
					}
					break
				}

				if event == nil {
					break
				}

				events = append(events, event)
			}

			assert.Equal(t, tc.expectedEvents, events)
		})
	}
}

func TestMessageSSEDecoder_DecodeOptions(t *testing.T) {
	decoder := NewMessageSSEDecoder(strings.NewReader(""))

	_, err := decoder.Decode(DecodeOptions{}, DecodeOptions{})
	assert.EqualError(t, err, "too many options provided, expected at most one")
}

func TestMessageSSEDecoder_updateContent(t *testing.T) {
	decoder := NewMessageSSEDecoder(nil)

	decoder.updateContent(0, "Hello")
	assert.Equal(t, []string{"Hello"}, decoder.content)

	decoder.updateContent(2, "!")
	assert.Equal(t, []string{"Hello", "", "!"}, decoder.content)

	decoder.updateContent(1, " world")
	assert.Equal(t, []string{"Hello", " world", "!"}, decoder.content)
}

type errReader struct{}

func (r errReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

func TestMessageSSEDecoder_DecodeError(t *testing.T) {
	decoder := NewMessageSSEDecoder(errReader{})

	_, err := decoder.Decode()
	assert.EqualError(t, err, io.ErrUnexpectedEOF.Error())
}
