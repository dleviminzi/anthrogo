package anthrogo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// CompletePayload contains the necessary data for the completion request.
type CompletePayload struct {
	MaxTokensToSample int            `json:"max_tokens_to_sample"`
	Model             AnthropicModel `json:"model"`
	Prompt            string         `json:"prompt"`
	CompleteOptions
}

// CompleteOptions holds optional parameters for the complete request.
type CompleteOptions struct {
	Metadata      any      `json:"metadata,omitempty"`
	StopSequences []string `json:"stop_sequences,omitempty"`
	Stream        bool     `json:"stream,omitempty"`
	Temperature   float64  `json:"temperature,omitempty"`
	TopK          int      `json:"top_k,omitempty"`
	TopP          float64  `json:"top_p,omitempty"`
}

// ErrorResponse holds the error details in the response.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail describes the error type and message.
type ErrorDetail struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// CompleteResponse contains the completion result or error details.
type CompleteResponse struct {
	Completion string `json:"completion"`
	StopReason string `json:"stop_reason"`
	Model      string `json:"model"`
}

// CompleteStreamResponse contains the server sent events decoder, the response body from the request, and a
// cancel function to enforce a timeout.
type CompleteStreamResponse struct {
	decoder *SSEDecoder
	body    io.ReadCloser
	cancel  context.CancelFunc
}

// Decode is a method for CompleteStreamResponse that returns the next event
// from the server-sent events decoder, or an error if one occurred.
func (c CompleteStreamResponse) Decode() (*Event, error) {
	return c.decoder.Decode()
}

// Cancel is a method for CompleteStreamResponse that invokes the associated
// cancel function to stop the request prematurely.
func (c CompleteStreamResponse) Cancel() {
	c.cancel()
}

// Close is a method for CompleteStreamResponse that closes the response body.
// If the response body has been read, Close returns nil. Otherwise, it returns
// an error.
func (c CompleteStreamResponse) Close() error {
	return c.body.Close()
}

// Complete sends a complete request to the server and returns the response or error.
func (c *Client) Complete(payload *CompletePayload) (*CompleteResponse, error) {
	// force stream off if user uses this method
	payload.Stream = false

	req, cancel, err := c.createRequest(payload)
	if err != nil {
		return nil, err
	}
	defer cancel()

	res, err := c.doRequestWithRetries(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var response CompleteResponse
	if res.StatusCode != http.StatusOK {
		var errorResponse ErrorResponse
		err = json.Unmarshal(body, &errorResponse)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("%s: %s", errorResponse.Error.Type, errorResponse.Error.Message)
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// CompleteStream is a method for Client that sends a request to the server with
// streaming enabled. It marshals the payload into a JSON object and sends it
// to the server in a POST request. If the request is successful, it returns a
// pointer to a CompleteStreamResponse object. Otherwise, it returns an error.
func (c *Client) CompleteStream(payload *CompletePayload) (*CompleteStreamResponse, error) {
	// force stream to true if user calls this method
	payload.Stream = true

	req, cancel, err := c.createRequest(payload)
	if err != nil {
		return nil, err
	}

	res, err := c.doRequestWithRetries(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		var errorResponse ErrorResponse
		err = json.Unmarshal(body, &errorResponse)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("%s: %s", errorResponse.Error.Type, errorResponse.Error.Message)
	}

	return &CompleteStreamResponse{NewSSEDecoder(res.Body), res.Body, cancel}, nil
}
