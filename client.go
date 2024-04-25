package anthrogo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"
)

const (
	DefaultMaxRetries   = 3
	DefaultTimeout      = time.Minute
	DefaultVersion      = "2023-06-01"
	jitterFactor        = 0.5
	RequestTypeComplete = "complete"
	RequestTypeMessages = "messages"
)

// ErrorResponse holds the error details in the response.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail describes the error type and message.
type ErrorDetail struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client is a structure holding all necessary fields for making requests to the API.
type Client struct {
	baseURL       string
	version       string
	maxRetries    int
	timeout       time.Duration
	customHeaders map[string]string
	httpClient    HttpClient
	apiKey        string
}

// NewClient creates and returns a new Client. It applies the provided options to the client.
// If no API key is provided as an option, it looks for the API key in the environment variable ANTHROPIC_API_KEY.
func NewClient(options ...func(*Client)) (*Client, error) {
	client := &Client{
		version:    DefaultVersion,
		maxRetries: DefaultMaxRetries,
		timeout:    DefaultTimeout,
		httpClient: &http.Client{},
		apiKey:     "",
		baseURL:    "https://api.anthropic.com/v1/",
	}

	for _, option := range options {
		option(client)
	}

	if client.apiKey == "" {
		apiKey, exists := os.LookupEnv("ANTHROPIC_API_KEY")
		if !exists {
			return nil, errors.New("ANTHROPIC_API_KEY not found in environment and not provided as option")
		}
		client.apiKey = apiKey
	}

	return client, nil
}

// Optional settings

// WithApiKey is an option to provide an API key for the Client.
func WithApiKey(apiKey string) func(*Client) {
	return func(c *Client) {
		c.apiKey = apiKey
	}
}

// WithMaxRetries is an option to set the maximum number of retries for the Client.
func WithMaxRetries(maxRetries int) func(*Client) {
	return func(c *Client) {
		c.maxRetries = maxRetries
	}
}

// WithTimeout is an option to set the timeout for the Client.
func WithTimeout(timeout time.Duration) func(*Client) {
	return func(c *Client) {
		c.timeout = timeout
	}
}

// WithCustomHeaders is an option to set custom headers for the Client.
func WithCustomHeaders(headers map[string]string) func(*Client) {
	return func(c *Client) {
		c.customHeaders = headers
	}
}

// WithVersion is an option to set the API version for the Client.
func WithVersion(version string) func(*Client) {
	return func(c *Client) {
		c.version = version
	}
}

// setRequestHeaders sets the necessary headers for the HTTP request.
func (c *Client) setRequestHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Anthropic-Version", c.version)
	req.Header.Set("x-api-key", c.apiKey)

	for key, value := range c.customHeaders {
		req.Header.Set(key, value)
	}
}

// createRequest creates and returns a new HTTP request with necessary headers.
func (c *Client) createRequest(ctx context.Context, payload any, requestType string) (*http.Request, context.CancelFunc, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, err
	}

	req, err := http.NewRequest("POST", c.baseURL+requestType, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, nil, err
	}

	c.setRequestHeaders(req)

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	req = req.WithContext(ctx)

	return req, cancel, nil
}

// doRequest sends an HTTP request and returns the response.
func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// doRequestWithRetries sends the HTTP request and retries upon failure up to the maximum retry limit.
func (c *Client) doRequestWithRetries(req *http.Request) (*http.Response, error) {
	for i := 0; i < c.maxRetries; i++ {
		response, err := c.doRequest(req)
		if err != nil {
			if i == c.maxRetries-1 {
				return nil, err
			}

			time.Sleep(c.getSleepDuration(i))
			continue
		}

		return response, nil
	}

	return nil, fmt.Errorf("failed to complete request after %d retries", c.maxRetries)
}

// getSleepDuration calculates and returns the sleep duration based on the retry attempt with added jitter.
func (c *Client) getSleepDuration(retry int) time.Duration {
	sleepDuration := time.Duration(retry) * time.Second

	jitter := time.Duration(rand.Float64() * jitterFactor * float64(sleepDuration))

	return sleepDuration + jitter
}
