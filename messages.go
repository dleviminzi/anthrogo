package anthrogo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ContentType is the type of content in a message.
type ContentType string

// RoleType is the role of the message.
type RoleType string

const (
	ContentTypeText  ContentType = "text"
	ContentTypeImage ContentType = "image"

	RoleTypeUser      RoleType = "user"
	RoleTypeAssistant RoleType = "assistant"
)

// MessagePayload is the request payload for the /messages endpoint.
type MessagePayload struct {
	// The model to use for the request.
	Model AnthropicModel `json:"model"`
	// The messages to send to the model.
	Messages []Message `json:"messages"`
	// The maximum number of tokens to generate.
	MaxTokens int `json:"max_tokens"`
	// Sequences that will cause the model to stop generating.
	StopSequences []string `json:"stop_sequences,omitempty"`
	// Amount of randomness injected into the response.
	Temperature *float64 `json:"temperature,omitempty"`
	// Nucleus sampling.
	TopP *float64 `json:"top_p,omitempty"`
	// Only sample from the top K options for each subsequent token.
	TopK *int `json:"top_k,omitempty"`
	// An object describing metadata about the request.
	Metadata *Metadata `json:"metadata,omitempty"`
	// System prompt to provide to the model.
	System *string `json:"system,omitempty"`
	// Stream the response using server-sent events.
	Stream *bool `json:"stream,omitempty"`
}

// Message is composed of a role and content. The role is either "user" or "assistant"
// and the content is a list of message content, which may contain text or images.
type Message struct {
	Role    RoleType         `json:"role"`
	Content []MessageContent `json:"content"`
}

// MessageContent is the content of a message. It can be either text or an image.
type MessageContent struct {
	Type  ContentType  `json:"type,omitempty"`
	Text  *string      `json:"text,omitempty"`
	Image *ImageSource `json:"image,omitempty"`
}

// ImageSource describes an image that is sent to the model in base64 (type).
// The following media types are accepted: image/jpeg, image/png, image/gif, image/webp.
type ImageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

// Metadata is an object describing metadata about the request.
// At the moment this only supports a user ID.
type Metadata struct {
	UserID string `json:"user_id,omitempty"`
}

// MessageResponse is the response payload for the /messages endpoint.
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

// ContentBlock is a block of content in a message response.
// Currently the model will only return text content.
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// Usage contains information about the number of input and output tokens.
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// MessageRequest sends a message to the model and returns the response.
func (c *Client) MessageRequest(ctx context.Context, payload MessagePayload) (MessageResponse, error) {
	var resp MessageResponse
	stream := false
	payload.Stream = &stream

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

// MessageStreamRequest sends a message to the model and returns the body for the user to consume
func (c *Client) MessageStreamRequest(ctx context.Context, payload MessagePayload) (io.ReadCloser, context.CancelFunc, error) {
	stream := true
	payload.Stream = &stream

	req, cancel, err := c.createRequest(ctx, payload, RequestTypeMessages)
	if err != nil {
		return nil, nil, err
	}

	res, err := c.doRequestWithRetries(req)
	if err != nil {
		return nil, nil, err
	}

	if res.StatusCode != http.StatusOK {
		var errorResponse ErrorResponse

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, nil, err
		}

		err = json.Unmarshal(body, &errorResponse)
		if err != nil {
			return nil, nil, err
		}

		return nil, nil, fmt.Errorf("%s: %s", errorResponse.Error.Type, errorResponse.Error.Message)
	}

	return res.Body, cancel, nil
}
