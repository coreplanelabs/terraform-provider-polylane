package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	api "github.com/coreplanelabs/terraform-provider-polylane/internal/client/generated"
)

const maxResponseSize = 4 << 20

// Client is a small client for the subset of the Polylane API exposed by the
// Terraform provider. Keeping this client hand-written makes the provider's
// supported surface explicit instead of coupling it to every API operation.
type Client struct {
	apiKey     string
	endpoint   *url.URL
	httpClient *http.Client
	userAgent  string
	generated  *api.ClientWithResponses
}

type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("Polylane API returned HTTP %d", e.StatusCode)
	}

	return fmt.Sprintf("Polylane API returned HTTP %d: %s", e.StatusCode, e.Message)
}

func IsNotFound(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound
}

func New(apiKey, endpoint, version string, httpClient *http.Client) (*Client, error) {
	parsedEndpoint, err := url.Parse(strings.TrimRight(endpoint, "/"))
	if err != nil {
		return nil, fmt.Errorf("parse endpoint: %w", err)
	}
	if parsedEndpoint.Scheme != "http" && parsedEndpoint.Scheme != "https" {
		return nil, fmt.Errorf("endpoint must use http or https")
	}
	if parsedEndpoint.Host == "" {
		return nil, fmt.Errorf("endpoint must include a host")
	}
	if parsedEndpoint.RawQuery != "" || parsedEndpoint.Fragment != "" {
		return nil, fmt.Errorf("endpoint must not include a query string or fragment")
	}

	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	userAgent := "terraform-provider-polylane/" + version
	generatedClient, err := api.NewClientWithResponses(
		parsedEndpoint.String(),
		api.WithHTTPClient(httpClient),
		api.WithRequestEditorFn(func(_ context.Context, req *http.Request) error {
			req.Header.Set("Accept", "application/json")
			req.Header.Set("X-API-Key", apiKey)
			req.Header.Set("User-Agent", userAgent)
			// Hono's CSRF middleware treats an unsafe request without a content
			// type as form-shaped text/plain. Mark bodyless DELETE requests as
			// API requests so they are not rejected before API-key auth runs.
			if req.Method == http.MethodPost || req.Method == http.MethodPut || req.Method == http.MethodPatch || req.Method == http.MethodDelete {
				req.Header.Set("Content-Type", "application/json")
			}
			return nil
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("create generated Polylane API client: %w", err)
	}

	return &Client{
		apiKey:     apiKey,
		endpoint:   parsedEndpoint,
		httpClient: httpClient,
		userAgent:  userAgent,
		generated:  generatedClient,
	}, nil
}

type envelope struct {
	Success bool            `json:"success"`
	Error   json.RawMessage `json:"error"`
	Message json.RawMessage `json:"message"`
	Result  json.RawMessage `json:"result"`
}

func (c *Client) do(ctx context.Context, method, requestPath string, input, result any) error {
	var requestBody io.Reader
	if input != nil {
		encoded, err := json.Marshal(input)
		if err != nil {
			return fmt.Errorf("encode Polylane API request: %w", err)
		}
		requestBody = bytes.NewReader(encoded)
	}

	requestURL := *c.endpoint
	requestURL.Path = strings.TrimRight(c.endpoint.Path, "/") + "/" + strings.TrimLeft(requestPath, "/")

	req, err := http.NewRequestWithContext(ctx, method, requestURL.String(), requestBody)
	if err != nil {
		return fmt.Errorf("create Polylane API request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("User-Agent", c.userAgent)
	if input != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	response, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send Polylane API request: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(io.LimitReader(response.Body, maxResponseSize+1))
	if err != nil {
		return fmt.Errorf("read Polylane API response: %w", err)
	}
	if len(body) > maxResponseSize {
		return fmt.Errorf("API response exceeded %d bytes", maxResponseSize)
	}

	var responseEnvelope envelope
	if len(body) != 0 {
		if err := json.Unmarshal(body, &responseEnvelope); err != nil {
			if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
				return &APIError{StatusCode: response.StatusCode, Message: strings.TrimSpace(string(body))}
			}
			return fmt.Errorf("decode Polylane API response: %w", err)
		}
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices || !responseEnvelope.Success {
		return &APIError{StatusCode: response.StatusCode, Message: envelopeMessage(responseEnvelope)}
	}
	if result == nil || len(responseEnvelope.Result) == 0 || string(responseEnvelope.Result) == "null" {
		return nil
	}
	if err := json.Unmarshal(responseEnvelope.Result, result); err != nil {
		return fmt.Errorf("decode Polylane API result: %w", err)
	}

	return nil
}

func envelopeMessage(response envelope) string {
	for _, raw := range []json.RawMessage{response.Error, response.Message} {
		if len(raw) == 0 || string(raw) == "null" {
			continue
		}

		var text string
		if json.Unmarshal(raw, &text) == nil && text != "" {
			return text
		}

		var object struct {
			Message string `json:"message"`
			Detail  string `json:"detail"`
		}
		if json.Unmarshal(raw, &object) == nil {
			if object.Detail != "" {
				return object.Detail
			}
			if object.Message != "" {
				return object.Message
			}
		}
	}

	return "request failed"
}
