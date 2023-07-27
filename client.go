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
	DefaultMaxRetries = 3
	DefaultTimeout    = time.Minute
	DefaultVersion    = "2023-06-01"
	jitterFactor      = 0.5
)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client is a structure holding all necessary fields for making requests to the API.
type Client struct {
	BaseURL                  string
	Version                  string
	StrictResponseValidation bool
	MaxRetries               int
	Timeout                  time.Duration
	CustomHeaders            map[string]string
	HttpClient               HttpClient
	ApiKey                   string
}

// NewClient creates and returns a new Client. It applies the provided options to the client.
// If no API key is provided as an option, it looks for the API key in the environment variable ANTHROPIC_API_KEY.
func NewClient(options ...func(*Client)) (*Client, error) {
	client := &Client{
		Version:    DefaultVersion,
		MaxRetries: DefaultMaxRetries,
		Timeout:    DefaultTimeout,
		HttpClient: &http.Client{},
		ApiKey:     "",
		BaseURL:    "https://api.anthropic.com/v1/",
	}

	for _, option := range options {
		option(client)
	}

	if client.ApiKey == "" {
		apiKey, exists := os.LookupEnv("ANTHROPIC_API_KEY")
		if !exists {
			return nil, errors.New("ANTHROPIC_API_KEY not found in environment and not provided as option")
		}
		client.ApiKey = apiKey
	}

	return client, nil
}

// Optional settings

// WithApiKey is an option to provide an API key for the Client.
func WithApiKey(apiKey string) func(*Client) {
	return func(c *Client) {
		c.ApiKey = apiKey
	}
}

// WithMaxRetries is an option to set the maximum number of retries for the Client.
func WithMaxRetries(maxRetries int) func(*Client) {
	return func(c *Client) {
		c.MaxRetries = maxRetries
	}
}

// WithTimeout is an option to set the timeout for the Client.
func WithTimeout(timeout time.Duration) func(*Client) {
	return func(c *Client) {
		c.Timeout = timeout
	}
}

// WithCustomHeaders is an option to set custom headers for the Client.
func WithCustomHeaders(headers map[string]string) func(*Client) {
	return func(c *Client) {
		c.CustomHeaders = headers
	}
}

// WithVersion is an option to set the API version for the Client.
func WithVersion(version string) func(*Client) {
	return func(c *Client) {
		c.Version = version
	}
}

// setRequestHeaders sets the necessary headers for the HTTP request.
func (c *Client) setRequestHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Anthropic-Version", c.Version)
	req.Header.Set("x-api-key", c.ApiKey)

	for key, value := range c.CustomHeaders {
		req.Header.Set(key, value)
	}
}

// createRequest creates and returns a new HTTP request with necessary headers.
func (c *Client) createRequest(payload *CompletePayload) (*http.Request, context.CancelFunc, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, err
	}

	req, err := http.NewRequest("POST", c.BaseURL+"complete", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, nil, err
	}

	c.setRequestHeaders(req)

	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	req = req.WithContext(ctx)

	return req, cancel, nil
}

// doRequest sends an HTTP request and returns the response.
func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	res, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// doRequestWithRetries sends the HTTP request and retries upon failure up to the maximum retry limit.
func (c *Client) doRequestWithRetries(req *http.Request) (*http.Response, error) {
	for i := 0; i < c.MaxRetries; i++ {
		response, err := c.doRequest(req)
		if err != nil {
			if i == c.MaxRetries-1 {
				return nil, err
			}

			time.Sleep(c.getSleepDuration(i))
			continue
		}

		return response, nil
	}

	return nil, fmt.Errorf("failed to complete request after %d retries", c.MaxRetries)
}

// getSleepDuration calculates and returns the sleep duration based on the retry attempt with added jitter.
func (c *Client) getSleepDuration(retry int) time.Duration {
	sleepDuration := time.Duration(retry) * time.Second

	jitter := time.Duration(rand.Float64() * jitterFactor * float64(sleepDuration))

	return sleepDuration + jitter
}
