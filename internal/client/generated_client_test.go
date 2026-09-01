package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGeneratedClientRoutes(t *testing.T) {
	t.Parallel()

	teamResult := `{"id":"team_123","workspaceId":"ws_123","name":"Platform","slug":"platform","description":null,"color":null,"icon":null,"createdBy":"usr_123","created":"2026-01-01T00:00:00Z","updated":"2026-01-01T00:00:00Z","_object":"team"}`
	workspaceMemberResult := `{"workspaceId":"ws_123","userId":"usr_123","email":"dev@example.com","username":"dev","forename":"Dev","surname":null,"invitedBy":null,"position":"admin","scopes":["teams:read"],"lastActive":null,"created":"2026-01-01T00:00:00Z","updated":"2026-01-01T00:00:00Z","_object":"workspace_member"}`
	teamMemberResult := `{"teamId":"team_123","userId":"usr_123","addedBy":"usr_admin","created":"2026-01-01T00:00:00Z","_object":"team_member"}`

	tests := []struct {
		name             string
		wantMethod       string
		wantPath         string
		wantBodyContains string
		status           int
		response         string
		invoke           func(context.Context, *Client) error
	}{
		{
			name:             "update workspace without credentials",
			wantMethod:       http.MethodPatch,
			wantPath:         "/v1/workspaces/ws_123",
			wantBodyContains: `"name":"Platform"`,
			status:           http.StatusOK,
			response:         `{"success":true,"error":null,"message":{"message":"updated"},"result":{"id":"ws_123","name":"Platform","slug":"platform","description":null,"domain":null,"autoJoinEnabled":null,"ownerId":"usr_123","createdBy":"usr_123","lastEditedBy":"usr_123","llmProvider":"openai","llmProviderApiKey":null,"llmProviderAuthHeader":null,"llmProviderBaseUrl":null,"llmProviderModelsUrl":null,"llmModelTiers":null,"created":"2026-01-01T00:00:00Z","updated":"2026-01-01T00:00:00Z","_object":"workspace"}}`,
			invoke: func(ctx context.Context, apiClient *Client) error {
				name := "Platform"
				_, err := apiClient.UpdateWorkspace(ctx, "ws_123", UpdateWorkspaceInput{Name: &name})
				return err
			},
		},
		{
			name:             "create team",
			wantMethod:       http.MethodPost,
			wantPath:         "/v1/teams",
			wantBodyContains: `"workspaceId":"ws_123"`,
			status:           http.StatusCreated,
			response:         fmt.Sprintf(`{"success":true,"error":null,"message":{"message":"created"},"result":%s}`, teamResult),
			invoke: func(ctx context.Context, apiClient *Client) error {
				team, err := apiClient.CreateTeam(ctx, CreateTeamInput{WorkspaceID: "ws_123", Name: "Platform"})
				if err == nil && team.ID != "team_123" {
					return fmt.Errorf("unexpected team ID %q", team.ID)
				}
				return err
			},
		},
		{
			name:       "get team",
			wantMethod: http.MethodGet,
			wantPath:   "/v1/teams/ws_123/team_123",
			status:     http.StatusOK,
			response:   fmt.Sprintf(`{"success":true,"error":null,"message":{"message":"ok"},"result":%s}`, teamResult),
			invoke: func(ctx context.Context, apiClient *Client) error {
				_, err := apiClient.GetTeam(ctx, "ws_123", "team_123")
				return err
			},
		},
		{
			name:             "update team",
			wantMethod:       http.MethodPatch,
			wantPath:         "/v1/teams/team_123",
			wantBodyContains: `"name":"Platform"`,
			status:           http.StatusOK,
			response:         fmt.Sprintf(`{"success":true,"error":null,"message":{"message":"updated"},"result":%s}`, teamResult),
			invoke: func(ctx context.Context, apiClient *Client) error {
				name := "Platform"
				_, err := apiClient.UpdateTeam(ctx, "team_123", UpdateTeamInput{WorkspaceID: "ws_123", Name: &name})
				return err
			},
		},
		{
			name:       "delete team",
			wantMethod: http.MethodDelete,
			wantPath:   "/v1/teams/ws_123/team_123",
			status:     http.StatusOK,
			response:   fmt.Sprintf(`{"success":true,"error":null,"message":{"message":"deleted"},"result":%s}`, teamResult),
			invoke: func(ctx context.Context, apiClient *Client) error {
				return apiClient.DeleteTeam(ctx, "ws_123", "team_123")
			},
		},
		{
			name:       "get workspace member",
			wantMethod: http.MethodGet,
			wantPath:   "/v1/workspace_members/ws_123/usr_123",
			status:     http.StatusOK,
			response:   fmt.Sprintf(`{"success":true,"error":null,"message":{"message":"ok"},"result":%s}`, workspaceMemberResult),
			invoke: func(ctx context.Context, apiClient *Client) error {
				member, err := apiClient.GetWorkspaceMember(ctx, "ws_123", "usr_123")
				if err == nil && member.Position != "admin" {
					return fmt.Errorf("unexpected position %q", member.Position)
				}
				return err
			},
		},
		{
			name:             "update workspace member",
			wantMethod:       http.MethodPatch,
			wantPath:         "/v1/workspace_members/ws_123/usr_123",
			wantBodyContains: `"position":"admin"`,
			status:           http.StatusOK,
			response:         fmt.Sprintf(`{"success":true,"error":null,"message":{"message":"updated"},"result":%s}`, workspaceMemberResult),
			invoke: func(ctx context.Context, apiClient *Client) error {
				position := "admin"
				_, err := apiClient.UpdateWorkspaceMember(ctx, "ws_123", "usr_123", UpdateWorkspaceMemberInput{Position: &position})
				return err
			},
		},
		{
			name:             "add team member",
			wantMethod:       http.MethodPost,
			wantPath:         "/v1/teams/members",
			wantBodyContains: `"teamId":"team_123"`,
			status:           http.StatusCreated,
			response:         fmt.Sprintf(`{"success":true,"error":null,"message":{"message":"created"},"result":%s}`, teamMemberResult),
			invoke: func(ctx context.Context, apiClient *Client) error {
				_, err := apiClient.AddTeamMember(ctx, "ws_123", "team_123", "usr_123")
				return err
			},
		},
		{
			name:       "get team member",
			wantMethod: http.MethodGet,
			wantPath:   "/v1/teams/ws_123/team_123/members",
			status:     http.StatusOK,
			response:   fmt.Sprintf(`{"success":true,"error":null,"message":{"message":"ok"},"result":{"count":1,"items":[%s]}}`, teamMemberResult),
			invoke: func(ctx context.Context, apiClient *Client) error {
				_, err := apiClient.GetTeamMember(ctx, "ws_123", "team_123", "usr_123")
				return err
			},
		},
		{
			name:             "remove team member",
			wantMethod:       http.MethodPost,
			wantPath:         "/v1/teams/members/remove",
			wantBodyContains: `"userId":"usr_123"`,
			status:           http.StatusOK,
			response:         fmt.Sprintf(`{"success":true,"error":null,"message":{"message":"removed"},"result":%s}`, teamMemberResult),
			invoke: func(ctx context.Context, apiClient *Client) error {
				return apiClient.RemoveTeamMember(ctx, "ws_123", "team_123", "usr_123")
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			type capturedRequest struct {
				method string
				path   string
				body   string
			}
			captured := make(chan capturedRequest, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				captured <- capturedRequest{method: r.Method, path: r.URL.Path, body: string(body)}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(test.status)
				_, _ = w.Write([]byte(test.response))
			}))
			t.Cleanup(server.Close)

			apiClient, err := New("test-key", server.URL+"/v1", "test", server.Client())
			if err != nil {
				t.Fatalf("New returned an error: %v", err)
			}
			if err := test.invoke(context.Background(), apiClient); err != nil {
				t.Fatalf("API operation returned an error: %v", err)
			}

			request := <-captured
			if request.method != test.wantMethod {
				t.Errorf("unexpected method: got %q, want %q", request.method, test.wantMethod)
			}
			if request.path != test.wantPath {
				t.Errorf("unexpected path: got %q, want %q", request.path, test.wantPath)
			}
			if test.wantBodyContains != "" && !strings.Contains(request.body, test.wantBodyContains) {
				t.Errorf("request body %q does not contain %q", request.body, test.wantBodyContains)
			}
			if test.name == "update workspace without credentials" && strings.Contains(request.body, "llmProvider") {
				t.Errorf("workspace update unexpectedly included unmanaged LLM fields: %s", request.body)
			}
		})
	}
}
