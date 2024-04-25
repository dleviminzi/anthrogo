package anthrogo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_MessageRequest(t *testing.T) {
	testCases := []struct {
		name          string
		responseCode  int
		responseBody  string
		expectedError string
	}{
		{
			name:         "successful request",
			responseCode: http.StatusOK,
			responseBody: `{
				"id": "1",
				"type": "text_completion",
				"role": "assistant",
				"content": [{"type": "text", "text": "Hello! How can I assist you today?"}],
				"model": "claude-3-sonnet-20240229",
				"usage": {"input_tokens": 5, "output_tokens": 10}
			}`,
		},
		{
			name:          "error response",
			responseCode:  http.StatusBadRequest,
			responseBody:  `{"error": {"type": "invalid_request_error", "message": "Invalid model"}}`,
			expectedError: "invalid_request_error: Invalid model",
		},
		{
			name:         "request failure",
			responseCode: http.StatusInternalServerError,
			responseBody: `{
				"error": {
					"type": "server_error",
					"message": "Internal server error"
				}
			}`,
			expectedError: "server_error: Internal server error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/messages", r.URL.Path)
				assert.Equal(t, "POST", r.Method)

				var payload MessagePayload
				err := json.NewDecoder(r.Body).Decode(&payload)
				require.NoError(t, err)

				assert.Equal(t, string(ModelClaude3Sonnet), string(payload.Model))
				assert.Len(t, payload.Messages, 1)
				assert.Equal(t, RoleTypeUser, payload.Messages[0].Role)
				assert.Equal(t, "Hello!", *payload.Messages[0].Content[0].Text)
				assert.Equal(t, 100, payload.MaxTokens)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.responseCode)
				fmt.Fprint(w, tc.responseBody)
			}))
			defer ts.Close()

			client, err := NewClient(WithApiKey("fake-key"))
			require.NoError(t, err)
			client.baseURL = ts.URL + "/"

			ctx := context.Background()
			var s = "Hello!"
			payload := MessagePayload{
				Model: ModelClaude3Sonnet,
				Messages: []Message{
					{
						Role: RoleTypeUser,
						Content: []MessageContent{
							{
								Type: ContentTypeText,
								Text: &s,
							},
						},
					},
				},
				MaxTokens: 100,
			}

			resp, err := client.MessageRequest(ctx, payload)

			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, "1", resp.ID)
				assert.Equal(t, "text_completion", resp.Type)
				assert.Equal(t, RoleTypeAssistant, resp.Role)
				assert.Len(t, resp.Content, 1)
				assert.Equal(t, "text", resp.Content[0].Type)
				assert.Equal(t, "Hello! How can I assist you today?", resp.Content[0].Text)
				assert.Equal(t, string(ModelClaude3Sonnet), resp.Model)
				assert.Equal(t, 5, resp.Usage.InputTokens)
				assert.Equal(t, 10, resp.Usage.OutputTokens)
			}
		})
	}
}

func TestClient_MessageStreamRequest(t *testing.T) {
	testCases := []struct {
		name           string
		responseCode   int
		responseBody   string
		expectedError  string
		expectedStream string
	}{
		{
			name:         "successful request",
			responseCode: http.StatusOK,
			responseBody: `{"type": "stream_response", "content": [{"type": "text", "text": "Hello"}]}
{"type": "stream_response", "content": [{"type": "text", "text": " world!"}]}
[DONE]`,
			expectedStream: `{"type": "stream_response", "content": [{"type": "text", "text": "Hello"}]}
{"type": "stream_response", "content": [{"type": "text", "text": " world!"}]}
[DONE]`,
		},
		{
			name:          "error response",
			responseCode:  http.StatusBadRequest,
			responseBody:  `{"error": {"type": "invalid_request_error", "message": "Invalid model"}}`,
			expectedError: "invalid_request_error: Invalid model",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/messages", r.URL.Path)
				assert.Equal(t, "POST", r.Method)

				var payload MessagePayload
				err := json.NewDecoder(r.Body).Decode(&payload)
				require.NoError(t, err)

				assert.Equal(t, string(ModelClaude3Sonnet), string(payload.Model))
				assert.Len(t, payload.Messages, 1)
				assert.Equal(t, RoleTypeUser, payload.Messages[0].Role)
				assert.Equal(t, "Hello!", *payload.Messages[0].Content[0].Text)
				assert.Equal(t, 100, payload.MaxTokens)
				assert.True(t, *payload.Stream)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.responseCode)
				fmt.Fprint(w, tc.responseBody)
			}))
			defer ts.Close()

			client, err := NewClient(WithApiKey("fake-key"))
			require.NoError(t, err)
			client.baseURL = ts.URL + "/"

			ctx := context.Background()
			var s = "Hello!"
			payload := MessagePayload{
				Model: ModelClaude3Sonnet,
				Messages: []Message{
					{
						Role: RoleTypeUser,
						Content: []MessageContent{
							{
								Type: ContentTypeText,
								Text: &s,
							},
						},
					},
				},
				MaxTokens: 100,
			}

			body, cancel, err := client.MessageStreamRequest(ctx, payload)
			defer func() {
				if cancel != nil {
					cancel()
				}
				if body != nil {
					body.Close()
				}
			}()

			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				require.NoError(t, err)

				var sb strings.Builder
				_, err = io.Copy(&sb, body)
				require.NoError(t, err)

				assert.Equal(t, tc.expectedStream, sb.String())
			}
		})
	}
}
