package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

type gcpTransport func(*http.Request) (*http.Response, error)

func (f gcpTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestGCPConnectionContract(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, method, path, body, result string
		invoke                           func(context.Context, *Client) error
	}{
		{"create", "POST", "/v1/gcp_connection_requests", `"projectId":"example-project"`, `{"requestId":"gcp_request","workspaceId":"workspace","projectId":"example-project","status":"pending"}`, func(ctx context.Context, c *Client) error {
			_, err := c.CreateGCPConnectionRequest(ctx, CreateGCPConnectionRequestInput{WorkspaceID: "workspace", ProjectID: "example-project", IdempotencyKey: "unique"})
			return err
		}},
		{"read", "GET", "/v1/gcp_connection_requests/workspace/gcp_request", "", `{"requestId":"gcp_request","workspaceId":"workspace","projectId":"example-project","status":"active","cloudAccountId":"account"}`, func(ctx context.Context, c *Client) error {
			v, err := c.GetGCPConnection(ctx, "workspace", "account", "gcp_request")
			if err == nil && v.Status != "registered" {
				return fmt.Errorf("unstable registration status: %s", v.Status)
			}
			return err
		}},
		{"revoke", "DELETE", "/v1/gcp_connection_requests/workspace/gcp_request", "{}", `{}`, func(ctx context.Context, c *Client) error {
			return c.DeleteGCPConnectionRequest(ctx, "workspace", "gcp_request")
		}},
		{"activate", "POST", "/v1/gcp_connections", `"projectNumber":"123456789"`, `{"requestId":"gcp_request","cloudAccountId":"account","provisioningStatus":"registered"}`, func(ctx context.Context, c *Client) error {
			_, err := c.ActivateGCPConnection(ctx, ActivateGCPConnectionInput{WorkspaceID: "workspace", RequestID: "gcp_request", ProjectNumber: "123456789"})
			return err
		}},
		{"disconnect", "DELETE", "/v1/cloud_accounts/workspace/account", "", `{}`, func(ctx context.Context, c *Client) error { return c.DeleteGCPConnection(ctx, "workspace", "account") }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c, err := New("test-key", "https://example.invalid/v1", "test", &http.Client{Transport: gcpTransport(func(r *http.Request) (*http.Response, error) {
				if r.Method != tt.method || r.URL.Path != tt.path {
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
				}
				if r.Header.Get("X-API-Key") != "test-key" {
					t.Error("missing authentication")
				}
				if tt.method != "GET" && r.Header.Get("Content-Type") != "application/json" {
					t.Error("missing JSON content type")
				}
				body := ""
				if r.Body != nil {
					b, _ := io.ReadAll(r.Body)
					body = string(b)
				}
				if !strings.Contains(body, tt.body) {
					t.Errorf("body %s lacks %s", body, tt.body)
				}
				return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"success":true,"result":` + tt.result + `}`))}, nil
			})})
			if err != nil {
				t.Fatal(err)
			}
			if err = tt.invoke(context.Background(), c); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestGCPReadRejectsMismatchedOrRevokedConnection(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ name, result string }{
		{"wrong request", `{"requestId":"other","workspaceId":"workspace","status":"active","cloudAccountId":"account"}`},
		{"wrong workspace", `{"requestId":"request","workspaceId":"other","status":"active","cloudAccountId":"account"}`},
		{"revoked", `{"requestId":"request","workspaceId":"workspace","status":"revoked","cloudAccountId":"account"}`},
		{"different account", `{"requestId":"request","workspaceId":"workspace","status":"active","cloudAccountId":"other"}`},
		{"missing result", `null`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c, err := New("key", "https://example.invalid/v1", "test", &http.Client{Transport: gcpTransport(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"success":true,"result":` + tt.result + `}`))}, nil
			})})
			if err != nil {
				t.Fatal(err)
			}
			if _, err = c.GetGCPConnection(context.Background(), "workspace", "account", "request"); err == nil {
				t.Fatal("accepted mismatched or inactive connection")
			}
		})
	}
}
