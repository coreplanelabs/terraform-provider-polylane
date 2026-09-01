package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	api "github.com/coreplanelabs/terraform-provider-polylane/internal/client/generated"
)

// The methods in this file are intentionally thin adapters around the
// generated client. The adapter keeps Terraform-facing types stable while the
// generated package owns paths, request bodies, response envelopes, and API
// schema validation.

func (c *Client) GetWorkspace(ctx context.Context, workspaceID string) (*Workspace, error) {
	response, err := c.generated.WorkspacesGetWithResponse(ctx, workspaceID)
	if err != nil {
		return nil, requestError(err)
	}
	if err := generatedResponseError(response.StatusCode(), response.Body); err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, missingGeneratedResult("get workspace")
	}

	return convertResult[Workspace](response.JSON200.Result)
}

func (c *Client) UpdateWorkspace(ctx context.Context, workspaceID string, input UpdateWorkspaceInput) (*Workspace, error) {
	// The OpenAPI patch schema has optional nullable credential fields. Plain
	// Go pointers cannot distinguish omitted from explicit null, so the generated
	// JSON body would send null for credentials Terraform does not manage. Use
	// the generated path/response method with this narrow body to preserve them.
	encoded, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("encode Polylane workspace update: %w", err)
	}
	response, err := c.generated.WorkspacesPatchWithBodyWithResponse(ctx, workspaceID, "application/json", bytes.NewReader(encoded))
	if err != nil {
		return nil, requestError(err)
	}
	if err := generatedResponseError(response.StatusCode(), response.Body); err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, missingGeneratedResult("update workspace")
	}

	return convertResult[Workspace](response.JSON200.Result)
}

func (c *Client) GetWorkspaceMember(ctx context.Context, workspaceID, userID string) (*WorkspaceMember, error) {
	response, err := c.generated.WorkspaceMembersGetWithResponse(ctx, workspaceID, userID)
	if err != nil {
		return nil, requestError(err)
	}
	if err := generatedResponseError(response.StatusCode(), response.Body); err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, missingGeneratedResult("get workspace member")
	}

	return convertResult[WorkspaceMember](response.JSON200.Result)
}

func (c *Client) UpdateWorkspaceMember(ctx context.Context, workspaceID, userID string, input UpdateWorkspaceMemberInput) (*WorkspaceMember, error) {
	body := api.WorkspaceMemberPatchJSONRequestBody{}
	if input.Position != nil {
		position := api.WorkspaceMemberPatchJSONBodyPosition(*input.Position)
		body.Position = &position
	}
	if input.Scopes != nil {
		scopes := make([]api.WorkspaceMemberPatchJSONBodyScopes, len(*input.Scopes))
		for index, scope := range *input.Scopes {
			scopes[index] = api.WorkspaceMemberPatchJSONBodyScopes(scope)
		}
		body.Scopes = &scopes
	}

	response, err := c.generated.WorkspaceMemberPatchWithResponse(ctx, workspaceID, userID, body)
	if err != nil {
		return nil, requestError(err)
	}
	if err := generatedResponseError(response.StatusCode(), response.Body); err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, missingGeneratedResult("update workspace member")
	}

	return convertResult[WorkspaceMember](response.JSON200.Result)
}

func (c *Client) CreateTeam(ctx context.Context, input CreateTeamInput) (*Team, error) {
	body := api.TeamsPostJSONRequestBody{
		WorkspaceId: input.WorkspaceID,
		Name:        input.Name,
		Slug:        input.Slug,
		Description: input.Description,
		Color:       input.Color,
		Icon:        input.Icon,
	}
	response, err := c.generated.TeamsPostWithResponse(ctx, body)
	if err != nil {
		return nil, requestError(err)
	}
	if err := generatedResponseError(response.StatusCode(), response.Body); err != nil {
		return nil, err
	}
	if response.JSON201 == nil {
		return nil, missingGeneratedResult("create team")
	}

	return convertResult[Team](response.JSON201.Result)
}

func (c *Client) GetTeam(ctx context.Context, workspaceID, teamID string) (*Team, error) {
	response, err := c.generated.TeamsGetWithResponse(ctx, workspaceID, teamID)
	if err != nil {
		return nil, requestError(err)
	}
	if err := generatedResponseError(response.StatusCode(), response.Body); err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, missingGeneratedResult("get team")
	}

	return convertResult[Team](response.JSON200.Result)
}

func (c *Client) UpdateTeam(ctx context.Context, teamID string, input UpdateTeamInput) (*Team, error) {
	body := api.TeamsPatchJSONRequestBody{
		WorkspaceId: input.WorkspaceID,
		Name:        input.Name,
		Slug:        input.Slug,
		Description: input.Description,
		Color:       input.Color,
		Icon:        input.Icon,
	}
	response, err := c.generated.TeamsPatchWithResponse(ctx, teamID, body)
	if err != nil {
		return nil, requestError(err)
	}
	if err := generatedResponseError(response.StatusCode(), response.Body); err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, missingGeneratedResult("update team")
	}

	return convertResult[Team](response.JSON200.Result)
}

func (c *Client) DeleteTeam(ctx context.Context, workspaceID, teamID string) error {
	response, err := c.generated.TeamsDelWithResponse(ctx, workspaceID, teamID)
	if err != nil {
		return requestError(err)
	}
	return generatedResponseError(response.StatusCode(), response.Body)
}

func (c *Client) AddTeamMember(ctx context.Context, workspaceID, teamID, userID string) (*TeamMember, error) {
	body := api.TeamMembersPostJSONRequestBody{
		WorkspaceId: workspaceID,
		TeamId:      teamID,
		UserId:      userID,
	}
	response, err := c.generated.TeamMembersPostWithResponse(ctx, body)
	if err != nil {
		return nil, requestError(err)
	}
	if err := generatedResponseError(response.StatusCode(), response.Body); err != nil {
		return nil, err
	}
	if response.JSON201 == nil {
		return nil, missingGeneratedResult("add team member")
	}

	return convertResult[TeamMember](response.JSON201.Result)
}

func (c *Client) GetTeamMember(ctx context.Context, workspaceID, teamID, userID string) (*TeamMember, error) {
	response, err := c.generated.TeamMembersListWithResponse(ctx, workspaceID, teamID)
	if err != nil {
		return nil, requestError(err)
	}
	if err := generatedResponseError(response.StatusCode(), response.Body); err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, missingGeneratedResult("list team members")
	}
	for _, member := range response.JSON200.Result.Items {
		if member.UserId == userID {
			return convertResult[TeamMember](member)
		}
	}

	return nil, &APIError{StatusCode: http.StatusNotFound, Message: "team member does not exist"}
}

func (c *Client) RemoveTeamMember(ctx context.Context, workspaceID, teamID, userID string) error {
	body := api.TeamMembersDelJSONRequestBody{
		WorkspaceId: workspaceID,
		TeamId:      teamID,
		UserId:      userID,
	}
	response, err := c.generated.TeamMembersDelWithResponse(ctx, body)
	if err != nil {
		return requestError(err)
	}
	return generatedResponseError(response.StatusCode(), response.Body)
}

func (c *Client) GetWorkspaceSettings(ctx context.Context, workspaceID, settingsPath string, result any) error {
	switch settingsPath {
	case "autofix_settings":
		response, err := c.generated.WorkspacesAutofixSettingsGetWithResponse(ctx, workspaceID)
		if err != nil {
			return requestError(err)
		}
		if err := generatedResponseError(response.StatusCode(), response.Body); err != nil {
			return err
		}
		if response.JSON200 == nil {
			return missingGeneratedResult("get autofix settings")
		}
		return copyResult(response.JSON200.Result, result)
	case "digest_settings":
		response, err := c.generated.WorkspacesDigestSettingsGetWithResponse(ctx, workspaceID)
		if err != nil {
			return requestError(err)
		}
		if err := generatedResponseError(response.StatusCode(), response.Body); err != nil {
			return err
		}
		if response.JSON200 == nil {
			return missingGeneratedResult("get digest settings")
		}
		return copyResult(response.JSON200.Result, result)
	case "investigation_limits_settings":
		response, err := c.generated.WorkspacesInvestigationLimitsSettingsGetWithResponse(ctx, workspaceID)
		if err != nil {
			return requestError(err)
		}
		if err := generatedResponseError(response.StatusCode(), response.Body); err != nil {
			return err
		}
		if response.JSON200 == nil {
			return missingGeneratedResult("get investigation limits settings")
		}
		return copyResult(response.JSON200.Result, result)
	case "investigations_settings":
		response, err := c.generated.WorkspacesInvestigationsSettingsGetWithResponse(ctx, workspaceID)
		if err != nil {
			return requestError(err)
		}
		if err := generatedResponseError(response.StatusCode(), response.Body); err != nil {
			return err
		}
		if response.JSON200 == nil {
			return missingGeneratedResult("get investigations settings")
		}
		return copyResult(response.JSON200.Result, result)
	case "model_training_settings":
		return c.do(ctx, http.MethodGet, workspaceSettingsPath(workspaceID, settingsPath), nil, result)
	case "observability_settings":
		response, err := c.generated.WorkspacesObservabilitySettingsGetWithResponse(ctx, workspaceID)
		if err != nil {
			return requestError(err)
		}
		if err := generatedResponseError(response.StatusCode(), response.Body); err != nil {
			return err
		}
		if response.JSON200 == nil {
			return missingGeneratedResult("get observability settings")
		}
		return copyResult(response.JSON200.Result, result)
	case "pr_review_settings":
		response, err := c.generated.WorkspacesPrReviewSettingsGetWithResponse(ctx, workspaceID)
		if err != nil {
			return requestError(err)
		}
		if err := generatedResponseError(response.StatusCode(), response.Body); err != nil {
			return err
		}
		if response.JSON200 == nil {
			return missingGeneratedResult("get pull request review settings")
		}
		return copyResult(response.JSON200.Result, result)
	default:
		return fmt.Errorf("unsupported workspace settings path %q", settingsPath)
	}
}

func (c *Client) UpdateWorkspaceSettings(ctx context.Context, workspaceID, settingsPath string, input, result any) error {
	switch settingsPath {
	case "autofix_settings":
		var body api.WorkspacesAutofixSettingsPatchJSONRequestBody
		if err := copyResult(input, &body); err != nil {
			return err
		}
		response, err := c.generated.WorkspacesAutofixSettingsPatchWithResponse(ctx, workspaceID, body)
		if err != nil {
			return requestError(err)
		}
		if err := generatedResponseError(response.StatusCode(), response.Body); err != nil {
			return err
		}
		if response.JSON200 == nil {
			return missingGeneratedResult("update autofix settings")
		}
		return copyResult(response.JSON200.Result, result)
	case "digest_settings":
		var body api.WorkspacesDigestSettingsPatchJSONRequestBody
		if err := copyResult(input, &body); err != nil {
			return err
		}
		response, err := c.generated.WorkspacesDigestSettingsPatchWithResponse(ctx, workspaceID, body)
		if err != nil {
			return requestError(err)
		}
		if err := generatedResponseError(response.StatusCode(), response.Body); err != nil {
			return err
		}
		if response.JSON200 == nil {
			return missingGeneratedResult("update digest settings")
		}
		return copyResult(response.JSON200.Result, result)
	case "investigation_limits_settings":
		var body api.WorkspacesInvestigationLimitsSettingsPatchJSONRequestBody
		if err := copyResult(input, &body); err != nil {
			return err
		}
		response, err := c.generated.WorkspacesInvestigationLimitsSettingsPatchWithResponse(ctx, workspaceID, body)
		if err != nil {
			return requestError(err)
		}
		if err := generatedResponseError(response.StatusCode(), response.Body); err != nil {
			return err
		}
		if response.JSON200 == nil {
			return missingGeneratedResult("update investigation limits settings")
		}
		return copyResult(response.JSON200.Result, result)
	case "investigations_settings":
		var body api.WorkspacesInvestigationsSettingsPatchJSONRequestBody
		if err := copyResult(input, &body); err != nil {
			return err
		}
		response, err := c.generated.WorkspacesInvestigationsSettingsPatchWithResponse(ctx, workspaceID, body)
		if err != nil {
			return requestError(err)
		}
		if err := generatedResponseError(response.StatusCode(), response.Body); err != nil {
			return err
		}
		if response.JSON200 == nil {
			return missingGeneratedResult("update investigations settings")
		}
		return copyResult(response.JSON200.Result, result)
	case "model_training_settings":
		return c.do(ctx, http.MethodPatch, workspaceSettingsPath(workspaceID, settingsPath), input, result)
	case "observability_settings":
		var body api.WorkspacesObservabilitySettingsPatchJSONRequestBody
		if err := copyResult(input, &body); err != nil {
			return err
		}
		response, err := c.generated.WorkspacesObservabilitySettingsPatchWithResponse(ctx, workspaceID, body)
		if err != nil {
			return requestError(err)
		}
		if err := generatedResponseError(response.StatusCode(), response.Body); err != nil {
			return err
		}
		if response.JSON200 == nil {
			return missingGeneratedResult("update observability settings")
		}
		return copyResult(response.JSON200.Result, result)
	case "pr_review_settings":
		var body api.WorkspacesPrReviewSettingsPatchJSONRequestBody
		if err := copyResult(input, &body); err != nil {
			return err
		}
		response, err := c.generated.WorkspacesPrReviewSettingsPatchWithResponse(ctx, workspaceID, body)
		if err != nil {
			return requestError(err)
		}
		if err := generatedResponseError(response.StatusCode(), response.Body); err != nil {
			return err
		}
		if response.JSON200 == nil {
			return missingGeneratedResult("update pull request review settings")
		}
		return copyResult(response.JSON200.Result, result)
	default:
		return fmt.Errorf("unsupported workspace settings path %q", settingsPath)
	}
}

func workspaceSettingsPath(workspaceID, settingsPath string) string {
	return "/workspaces/" + url.PathEscape(workspaceID) + "/" + settingsPath
}

func requestError(err error) error {
	return fmt.Errorf("send Polylane API request: %w", err)
}

func generatedResponseError(statusCode int, body []byte) error {
	var responseEnvelope envelope
	if len(body) != 0 {
		if err := json.Unmarshal(body, &responseEnvelope); err != nil {
			if statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices {
				return &APIError{StatusCode: statusCode, Message: string(body)}
			}
			return fmt.Errorf("decode Polylane API response: %w", err)
		}
	}

	if statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices || !responseEnvelope.Success {
		return &APIError{StatusCode: statusCode, Message: envelopeMessage(responseEnvelope)}
	}
	return nil
}

func copyResult(source, target any) error {
	encoded, err := json.Marshal(source)
	if err != nil {
		return fmt.Errorf("encode generated Polylane API value: %w", err)
	}
	if err := json.Unmarshal(encoded, target); err != nil {
		return fmt.Errorf("decode generated Polylane API value: %w", err)
	}
	return nil
}

func convertResult[T any](source any) (*T, error) {
	var target T
	if err := copyResult(source, &target); err != nil {
		return nil, err
	}
	return &target, nil
}

func missingGeneratedResult(operation string) error {
	return fmt.Errorf("polylane API returned a successful response without a typed result for %s", operation)
}
