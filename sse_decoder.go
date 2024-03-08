package anthrogo

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// CompletionEventData represents the data payload in a Server-Sent Events (SSE) message.
type CompletionEventData struct {
	Completion string `json:"completion"`
	StopReason string `json:"stop_reason"`
	Model      string `json:"model"`
	Stop       string `json:"stop"`
	LogID      string `json:"log_id"`
}

// CompletionEvent represents a single Server-Sent CompletionEvent. It includes the event type, data, ID, and retry fields.
type CompletionEvent struct {
	Event string
	Data  *CompletionEventData
	ID    string
	Retry int
}

// CompletionSSEDecoder is a decoder for Server-Sent Events. It maintains a buffer reader and the current event being processed.
type CompletionSSEDecoder struct {
	currentEvent CompletionEvent
	Reader       *bufio.Reader
}

// NewCompletionSSEDecoder initializes a new SSEDecoder with the provided reader.
func NewCompletionSSEDecoder(r io.Reader) *CompletionSSEDecoder {
	return &CompletionSSEDecoder{
		Reader: bufio.NewReader(r),
	}
}

// Decode reads from the buffered reader line by line, parses Server-Sent Events and sets fields on the current event.
// It returns the complete event when encountering an empty line, and nil otherwise. It will return EOF when nothing is left.
func (d *CompletionSSEDecoder) Decode() (*CompletionEvent, error) {
	line, err := d.Reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	line = strings.TrimSuffix(line, "\n")

	if line == "\r" {
		if d.currentEvent.Event == "" && d.currentEvent.Data == nil && d.currentEvent.ID == "" && d.currentEvent.Retry == 0 {
			return nil, nil
		}

		ev := d.currentEvent
		d.currentEvent = CompletionEvent{ID: ev.ID} // preserve LastEventID for the next event
		return &ev, nil
	}

	if strings.HasPrefix(line, ":") {
		return nil, nil
	}

	fields := strings.SplitN(line, ":", 2)
	if len(fields) < 2 {
		return nil, nil
	}

	fieldName := strings.TrimSpace(fields[0])
	fieldValue := strings.TrimSpace(fields[1])

	switch fieldName {
	case "id":
		if !strings.Contains(fieldValue, "\000") {
			d.currentEvent.ID = fieldValue
		}
	case "event":
		d.currentEvent.Event = fieldValue
	case "data":
		var data CompletionEventData
		err := json.Unmarshal([]byte(fieldValue), &data)
		if err != nil {
			return nil, fmt.Errorf("error decoding data field: %w", err)
		}
		d.currentEvent.Data = &data
	case "retry":
		retry, err := strconv.Atoi(fieldValue)
		if err == nil {
			d.currentEvent.Retry = retry
		}
	}

	return nil, nil
}
