package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetWorkspace(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: got %q, want %q", r.Method, http.MethodGet)
		}
		if r.URL.Path != "/v1/workspaces/ws_123" {
			t.Errorf("unexpected path: got %q", r.URL.Path)
		}
		if got := r.Header.Get("X-API-Key"); got != "test-key" {
			t.Errorf("unexpected API key: got %q", got)
		}
		if got := r.Header.Get("User-Agent"); got != "terraform-provider-polylane/1.2.3" {
			t.Errorf("unexpected user agent: got %q", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"error":null,"message":{"message":"Successful request"},"result":{"id":"ws_123","name":"Acme","slug":"acme","description":null,"domain":"acme.example","autoJoinEnabled":null,"ownerId":"usr_123","created":"2026-01-01T00:00:00Z","updated":"2026-01-02T00:00:00Z"}}`))
	}))
	t.Cleanup(server.Close)

	apiClient, err := New("test-key", server.URL+"/v1", "1.2.3", server.Client())
	if err != nil {
		t.Fatalf("New returned an error: %v", err)
	}

	workspace, err := apiClient.GetWorkspace(context.Background(), "ws_123")
	if err != nil {
		t.Fatalf("GetWorkspace returned an error: %v", err)
	}
	if workspace.ID != "ws_123" || workspace.Name != "Acme" {
		t.Fatalf("unexpected workspace: %#v", workspace)
	}
	if workspace.Description != nil {
		t.Errorf("expected a null description, got %q", *workspace.Description)
	}
	if workspace.Domain == nil || *workspace.Domain != "acme.example" {
		t.Errorf("unexpected domain: %#v", workspace.Domain)
	}
}

func TestUpdateWorkspaceSettings(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("unexpected method: got %q, want %q", r.Method, http.MethodPatch)
		}
		if r.URL.Path != "/v1/workspaces/ws_123/observability_settings" {
			t.Errorf("unexpected path: got %q", r.URL.Path)
		}

		var body map[string]map[string]bool
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if !body["nativeObservability"]["enabled"] {
			t.Errorf("unexpected request body: %#v", body)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"result":{"nativeObservability":{"enabled":true}}}`))
	}))
	t.Cleanup(server.Close)

	apiClient, err := New("test-key", server.URL+"/v1", "test", server.Client())
	if err != nil {
		t.Fatalf("New returned an error: %v", err)
	}

	var result ObservabilitySettings
	err = apiClient.UpdateWorkspaceSettings(
		context.Background(),
		"ws_123",
		"observability_settings",
		map[string]any{"nativeObservability": map[string]bool{"enabled": true}},
		&result,
	)
	if err != nil {
		t.Fatalf("UpdateWorkspaceSettings returned an error: %v", err)
	}
	if !result.NativeObservability.Enabled {
		t.Error("expected native observability to be enabled")
	}
}

func TestAPIError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"success":false,"error":{"detail":"workspace does not exist"},"result":null}`))
	}))
	t.Cleanup(server.Close)

	apiClient, err := New("test-key", server.URL, "test", server.Client())
	if err != nil {
		t.Fatalf("New returned an error: %v", err)
	}

	_, err = apiClient.GetWorkspace(context.Background(), "ws_missing")
	if !IsNotFound(err) {
		t.Fatalf("expected a not-found error, got %v", err)
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.Message != "workspace does not exist" {
		t.Errorf("unexpected API error message: %q", apiErr.Message)
	}
}

func TestNewRejectsInvalidEndpoints(t *testing.T) {
	t.Parallel()

	tests := []string{
		"api.polylane.com/v1",
		"ftp://api.polylane.com/v1",
		"https:///v1",
		"https://user:password@api.polylane.com/v1",
		"https://api.polylane.com/v1?debug=true",
		"https://api.polylane.com/v1#fragment",
	}

	for _, endpoint := range tests {
		t.Run(endpoint, func(t *testing.T) {
			t.Parallel()
			if _, err := New("test-key", endpoint, "test", nil); err == nil {
				t.Fatalf("expected endpoint %q to be rejected", endpoint)
			}
		})
	}
}

func TestClientDoesNotFollowRedirects(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"generated", "handwritten"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			destination := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Error("redirect destination must not receive a request")
				w.WriteHeader(http.StatusNoContent)
			}))
			t.Cleanup(destination.Close)
			origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("X-API-Key") != "test-key" {
					t.Error("original request must be authenticated")
				}
				http.Redirect(w, r, destination.URL, http.StatusTemporaryRedirect)
			}))
			t.Cleanup(origin.Close)
			transport := origin.Client()
			apiClient, err := New("test-key", origin.URL, "test", transport)
			if err != nil {
				t.Fatal(err)
			}
			if name == "generated" {
				_, err = apiClient.GetWorkspace(context.Background(), "ws_123")
			} else {
				var settings map[string]any
				err = apiClient.GetWorkspaceSettings(context.Background(), "ws_123", "model_training_settings", &settings)
			}
			var apiErr *APIError
			if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusTemporaryRedirect {
				t.Fatalf("expected redirect API error, got %v", err)
			}
			if transport.CheckRedirect != nil {
				t.Error("caller-owned HTTP client was modified")
			}
		})
	}
}
