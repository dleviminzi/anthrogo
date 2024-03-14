package anthrogo

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// MessageEventPayload is the decoded event from anthropic
type MessageEventPayload struct {
	Event string
	Data  EventData
}

// EventData contains content which will be whatever the model output
// and Data which is the full data from the event
type EventData struct {
	Content string
	Data    any
}

// MessageStart is one of the data types for events and it represents the start of a
// a stream of messages. It contains metadata about the request.
type MessageStart struct {
	Type    string `json:"type"`
	Message struct {
		ID           string   `json:"id"`
		Type         string   `json:"type"`
		Role         string   `json:"role"`
		Content      []string `json:"content"`
		Model        string   `json:"model"`
		StopReason   string   `json:"stop_reason"`
		StopSequence string   `json:"stop_sequence"`
		Usage        struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	} `json:"message"`
}

// ContentBlockStart marks the start of a new content block in the message stream.
type ContentBlockStart struct {
	Type         string `json:"type"`
	Index        int    `json:"index"`
	ContentBlock struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content_block"`
}

// PingData is a ping event
type PingData struct {
	Type string `json:"type"`
}

// ContentBlockDelta carries new content for a content block in the message stream.
type ContentBlockDelta struct {
	Type  string `json:"type"`
	Index int    `json:"index"`
	Delta struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"delta"`
}

// ContentBlockStop marks the end of a content block in the message stream.
type ContentBlockStop struct {
	Type  string `json:"type"`
	Index int    `json:"index"`
}

// MessageDelta events indicate top-level changes to the final message.
type MessageDelta struct {
	Type  string      `json:"type"`
	Delta interface{} `json:"delta"`
	Usage struct {
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// MessageStopData is the final event in a message stream.
type MessageStopData struct {
	Type string `json:"type"`
}

// ErrorData is the event type for errors.
type ErrorData struct {
	Type  string `json:"type"`
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// MessageEvent is the event type for messages. It contains the message payload
// and an error if one occurred.
type MessageEvent struct {
	Message *MessageEventPayload
	Err     *error
}

// MessageSSEDecoder is a decoder for the SSE stream from the message endpoint.
type MessageSSEDecoder struct {
	reader  *bufio.Reader
	content []string
}

// DecodeOptions are options for decoding the SSE stream.
type DecodeOptions struct {
	ContentOnly bool
}

// NewMessageSSEDecoder creates a new MessageSSEDecoder.
func NewMessageSSEDecoder(reader io.Reader) *MessageSSEDecoder {
	return &MessageSSEDecoder{
		reader:  bufio.NewReader(reader),
		content: make([]string, 0),
	}
}

// Decode reads the next event from the SSE stream.
func (d *MessageSSEDecoder) Decode(opts ...DecodeOptions) (*MessageEventPayload, error) {
	var options DecodeOptions
	if len(opts) > 1 {
		return nil, fmt.Errorf("too many options provided, expected at most one")
	} else if len(opts) == 1 {
		options = opts[0]
	}

	line, err := d.reader.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			return nil, nil
		}
		return nil, err
	}

	line = strings.TrimSpace(line)
	if line == "" {
		// Recursively call Decode to read the next event
		return d.Decode(opts...)
	}

	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid SSE format")
	}

	field := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	if field == "event" {
		data, err := d.decodeData(value)
		if err != nil {
			return nil, err
		}

		if data.Content != "" || !options.ContentOnly || value == "message_stop" {
			return &MessageEventPayload{
				Event: value,
				Data:  data,
			}, nil
		}
	}
	// Recursively call Decode to read the next event if we didn't have one here
	return d.Decode(opts...)
}

func (d *MessageSSEDecoder) decodeData(event string) (EventData, error) {
	var eventData EventData

	for {
		line, err := d.reader.ReadString('\n')
		if err != nil {
			break
		}

		line = strings.TrimSpace(line)
		if line == "" {
			break
		}

		if strings.HasPrefix(line, "data:") {
			jsonData := strings.TrimSpace(line[5:])

			switch event {
			case "message_start":
				var messageStartData MessageStart
				err := json.Unmarshal([]byte(jsonData), &messageStartData)
				if err != nil {
					return eventData, err
				}
				eventData.Data = messageStartData
			case "content_block_start":
				var contentBlockStartData ContentBlockStart
				err := json.Unmarshal([]byte(jsonData), &contentBlockStartData)
				if err != nil {
					return eventData, err
				}
				eventData.Data = contentBlockStartData
				eventData.Content = contentBlockStartData.ContentBlock.Text
				d.updateContent(contentBlockStartData.Index, contentBlockStartData.ContentBlock.Text)
			case "ping":
				var pingData PingData
				err := json.Unmarshal([]byte(jsonData), &pingData)
				if err != nil {
					return eventData, err
				}
				eventData.Data = pingData
			case "content_block_delta":
				var contentBlockDeltaData ContentBlockDelta
				err := json.Unmarshal([]byte(jsonData), &contentBlockDeltaData)
				if err != nil {
					return eventData, err
				}
				eventData.Data = contentBlockDeltaData
				eventData.Content = contentBlockDeltaData.Delta.Text
				d.updateContent(contentBlockDeltaData.Index, contentBlockDeltaData.Delta.Text)
			case "content_block_stop":
				var contentBlockStopData ContentBlockStop
				err := json.Unmarshal([]byte(jsonData), &contentBlockStopData)
				if err != nil {
					return eventData, err
				}
				eventData.Data = contentBlockStopData
			case "message_delta":
				var messageDeltaData MessageDelta
				err := json.Unmarshal([]byte(jsonData), &messageDeltaData)
				if err != nil {
					return eventData, err
				}
				eventData.Data = messageDeltaData
			case "message_stop":
				var messageStopData MessageStopData
				err := json.Unmarshal([]byte(jsonData), &messageStopData)
				if err != nil {
					return eventData, err
				}
				eventData.Data = messageStopData
			case "error":
				var errorData ErrorData
				err := json.Unmarshal([]byte(jsonData), &errorData)
				if err != nil {
					return eventData, err
				}
				return eventData, fmt.Errorf("error(%s) -  %s", errorData.Error.Type, errorData.Error.Message)
			}
		}
	}

	return eventData, nil
}

func (d *MessageSSEDecoder) updateContent(index int, content string) {
	if index >= len(d.content) {
		d.content = append(d.content, make([]string, index-len(d.content)+1)...)
	}
	d.content[index] += content
}
