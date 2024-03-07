package anthrogo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/dleviminzi/anthrogo/mocks"
)

func TestComplete(t *testing.T) {
	ctx := context.Background()
	mockRes := &CompleteResponse{
		Completion: "test",
		StopReason: "test",
		Model:      "test",
	}
	resBodyBytes, _ := json.Marshal(mockRes)
	resBodyReader := io.NopCloser(bytes.NewReader(resBodyBytes))

	mockHTTPClient := new(mocks.MockHttpClient)
	mockHTTPClient.On("Do", mock.Anything).Return(&http.Response{
		StatusCode: 200,
		Body:       resBodyReader,
	}, nil)

	client, err := NewClient(WithApiKey("blah"))
	require.NoError(t, err)

	client.HttpClient = mockHTTPClient

	payload := CompletePayload{
		MaxTokensToSample: 10,
		Model:             "test",
		Prompt:            "test",
	}

	response, err := client.Complete(ctx, payload)

	assert.NoError(t, err)
	assert.NotNil(t, response)

	assert.Equal(t, mockRes.Completion, response.Completion)
	assert.Equal(t, mockRes.StopReason, response.StopReason)
	assert.Equal(t, mockRes.Model, response.Model)

	mockHTTPClient.AssertExpectations(t)
}

func TestComplete_HttpClientError(t *testing.T) {
	ctx := context.Background()
	mockHTTPClient := new(mocks.MockHttpClient)
	mockHTTPClient.On("Do", mock.Anything).Return(&http.Response{}, errors.New("some http error"))

	client, err := NewClient(WithApiKey("blah"))
	require.NoError(t, err)

	client.HttpClient = mockHTTPClient

	payload := CompletePayload{}

	response, err := client.Complete(ctx, payload)

	assert.Empty(t, response)
	assert.Error(t, err)

	mockHTTPClient.AssertExpectations(t)
}

func TestComplete_Non200StatusCode(t *testing.T) {
	ctx := context.Background()
	mockHTTPClient := new(mocks.MockHttpClient)
	mockHTTPClient.On("Do", mock.Anything).Return(&http.Response{
		StatusCode: 400,
		Body:       io.NopCloser(bytes.NewReader([]byte(`{"error": {"type": "BadRequest", "message": "Bad request"}}`))),
	}, nil)

	client, err := NewClient(WithApiKey("blah"))
	require.NoError(t, err)

	client.HttpClient = mockHTTPClient

	payload := CompletePayload{}

	_, err = client.Complete(ctx, payload)

	assert.NotNil(t, err)

	mockHTTPClient.AssertExpectations(t)
}

func TestComplete_UnmarshalError(t *testing.T) {
	ctx := context.Background()
	mockHTTPClient := new(mocks.MockHttpClient)
	mockHTTPClient.On("Do", mock.Anything).Return(&http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader([]byte(`{"completion": "test", "invalid": {}`))),
	}, nil)

	client, err := NewClient(WithApiKey("blah"))
	require.NoError(t, err)

	client.HttpClient = mockHTTPClient

	payload := CompletePayload{}

	response, err := client.Complete(ctx, payload)

	assert.Empty(t, response)
	assert.Error(t, err)

	mockHTTPClient.AssertExpectations(t)
}

func TestCompleteStream(t *testing.T) {
	ctx := context.Background()
	expectedBody := "data: {\"completion\":\"testCompletion\",\"stop_reason\":\"testReason\",\"model\":\"testModel\",\"stop\":\"testStop\",\"log_id\":\"testLogId\"}\n\r\n"
	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(expectedBody)),
	}

	mockHTTPClient := new(mocks.MockHttpClient)
	mockHTTPClient.On("Do", mock.Anything).Return(mockResponse, nil)

	client, err := NewClient(WithApiKey("blah"))
	require.NoError(t, err)

	client.HttpClient = mockHTTPClient

	payload := CompletePayload{}

	response, err := client.CompleteStream(ctx, payload)
	defer response.Close()

	assert.NoError(t, err)

	_, _ = response.Decode()
	event, err := response.Decode()
	assert.NoError(t, err)
	assert.Equal(t, "testCompletion", event.Data.Completion)

	mockHTTPClient.AssertExpectations(t)
}

func TestCompleteStream_HttpClientError(t *testing.T) {
	mockHTTPClient := new(mocks.MockHttpClient)
	mockHTTPClient.On("Do", mock.Anything).Return(&http.Response{}, errors.New("some http error"))

	client, err := NewClient(WithApiKey("blah"))
	require.NoError(t, err)

	client.HttpClient = mockHTTPClient

	payload := CompletePayload{}

	response, err := client.CompleteStream(context.Background(), payload)

	assert.Nil(t, response)
	assert.Error(t, err)

	mockHTTPClient.AssertExpectations(t)
}
