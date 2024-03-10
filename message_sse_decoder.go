package anthrogo

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type MessageSSEPayload struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

type MessageStartData struct {
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

type ContentBlockStartData struct {
	Type         string `json:"type"`
	Index        int    `json:"index"`
	ContentBlock struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content_block"`
}

type PingData struct {
	Type string `json:"type"`
}

type ContentBlockDeltaData struct {
	Type  string    `json:"type"`
	Index int       `json:"index"`
	Delta TextDelta `json:"delta"`
}

type TextDelta struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ContentBlockStopData struct {
	Type  string `json:"type"`
	Index int    `json:"index"`
}

type MessageDeltaData struct {
	Type  string      `json:"type"`
	Delta interface{} `json:"delta"`
	Usage struct {
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

type MessageStopData struct {
	Type string `json:"type"`
}

type ErrorData struct {
	Type  string `json:"type"`
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

type MessageSSEDecoder struct {
	reader  *bufio.Reader
	content []string
}

func NewSSEDecoder(reader io.Reader) *MessageSSEDecoder {
	return &MessageSSEDecoder{
		reader:  bufio.NewReader(reader),
		content: make([]string, 0),
	}
}

func (d *MessageSSEDecoder) Decode() (*MessageSSEPayload, error) {
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
		return d.Decode()
	}

	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid SSE format")
	}

	field := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	if field == "event" {
		data := d.decodeData(value)

		// Update the content array based on the event type
		switch value {
		case "content_block_start":
			contentBlockStartData := data.(ContentBlockStartData)
			index := contentBlockStartData.Index
			if index >= len(d.content) {
				d.content = append(d.content, make([]string, index-len(d.content)+1)...)
			}
			d.content[index] = contentBlockStartData.ContentBlock.Text
		case "content_block_delta":
			contentBlockDeltaData := data.(ContentBlockDeltaData)
			index := contentBlockDeltaData.Index
			if index >= len(d.content) {
				d.content = append(d.content, make([]string, index-len(d.content)+1)...)
			}
			d.content[index] += contentBlockDeltaData.Delta.Text
		}

		return &MessageSSEPayload{
			Event: value,
			Data:  data,
		}, nil
	}
	// Recursively call Decode to read the next event if we didn't have one here
	return d.Decode()
}

// TODO: check for errors and return them here.
func (d *MessageSSEDecoder) decodeData(event string) any {
	var data any

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
				var messageStartData MessageStartData
				json.Unmarshal([]byte(jsonData), &messageStartData)
				data = messageStartData
			case "content_block_start":
				var contentBlockStartData ContentBlockStartData
				json.Unmarshal([]byte(jsonData), &contentBlockStartData)
				data = contentBlockStartData
			case "ping":
				var pingData PingData
				json.Unmarshal([]byte(jsonData), &pingData)
				data = pingData
			case "content_block_delta":
				var contentBlockDeltaData ContentBlockDeltaData
				json.Unmarshal([]byte(jsonData), &contentBlockDeltaData)
				data = contentBlockDeltaData
			case "content_block_stop":
				var contentBlockStopData ContentBlockStopData
				json.Unmarshal([]byte(jsonData), &contentBlockStopData)
				data = contentBlockStopData
			case "message_delta":
				var messageDeltaData MessageDeltaData
				json.Unmarshal([]byte(jsonData), &messageDeltaData)
				data = messageDeltaData
			case "message_stop":
				var messageStopData MessageStopData
				json.Unmarshal([]byte(jsonData), &messageStopData)
				data = messageStopData
			case "error":
				var errorData ErrorData
				json.Unmarshal([]byte(jsonData), &errorData)
				data = errorData
			}
		}
	}

	return data
}
