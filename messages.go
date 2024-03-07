package anthrogo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type ContentType string
type RoleType string

const (
	ContentTypeText   ContentType = "text"
	ContentTypeImage  ContentType = "image"
	RoleTypeUser      RoleType    = "user"
	RoleTypeAssistant RoleType    = "assistant"
)

type MessagePayload struct {
	Model         AnthropicModel `json:"model"`
	Messages      []Message      `json:"messages"`
	MaxTokens     int            `json:"max_tokens"`
	StopSequences []string       `json:"stop_sequences,omitempty"`
	System        *string        `json:"system,omitempty"`
	Metadata      *Metadata      `json:"metadata,omitempty"`
	Stream        *bool          `json:"stream,omitempty"`
	Temperature   *float64       `json:"temperature,omitempty"`
	TopP          *float64       `json:"top_p,omitempty"`
	TopK          *int           `json:"top_k,omitempty"`
}

type Message struct {
	Role    RoleType         `json:"role"`
	Content []MessageContent `json:"content"`
}

type MessageContent struct {
	Type  ContentType  `json:"type,omitempty"`
	Text  *string      `json:"text,omitempty"`
	Image *ImageSource `json:"image,omitempty"`
}

type ImageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

type Metadata struct {
	UserID string `json:"user_id,omitempty"`
}

type MessageResponse struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	Role         RoleType       `json:"role"`
	Content      []ContentBlock `json:"content"`
	Model        string         `json:"model"`
	StopReason   string         `json:"stop_reason"`
	StopSequence string         `json:"stop_sequence,omitempty"`
	Usage        Usage          `json:"usage"`
}

type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

func (c *Client) MessageRequest(ctx context.Context, payload MessagePayload) (MessageResponse, error) {
	var resp MessageResponse

	req, cancel, err := c.createRequest(ctx, payload, RequestTypeMessages)
	if err != nil {
		return resp, err
	}
	defer cancel()

	res, err := c.doRequestWithRetries(req)
	if err != nil {
		return resp, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return resp, err
	}

	if res.StatusCode != http.StatusOK {
		var errorResponse ErrorResponse
		err = json.Unmarshal(body, &errorResponse)
		if err != nil {
			return resp, err
		}
		return resp, fmt.Errorf("%s: %s", errorResponse.Error.Type, errorResponse.Error.Message)
	}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		return resp, err
	}

	return resp, nil
}
