package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// GCP uses the beta API contract until these operations reach the published OpenAPI document.
type GCPConnectionRequest struct {
	ID             string  `json:"requestId"`
	WorkspaceID    string  `json:"workspaceId"`
	ProjectID      string  `json:"projectId"`
	Subject        string  `json:"subject"`
	IssuerURL      string  `json:"issuerUrl"`
	PushEndpoint   string  `json:"pushEndpoint"`
	ResourcePrefix string  `json:"resourcePrefix"`
	Status         string  `json:"status"`
	CloudAccountID *string `json:"cloudAccountId"`
	ExpiresAt      string  `json:"expiresAt"`
	CreatedAt      string  `json:"createdAt"`
}

type CreateGCPConnectionRequestInput struct {
	WorkspaceID    string `json:"workspaceId"`
	ProjectID      string `json:"projectId"`
	IdempotencyKey string `json:"idempotencyKey"`
}

type ActivateGCPConnectionInput struct {
	WorkspaceID   string `json:"workspaceId"`
	RequestID     string `json:"requestId"`
	ProjectNumber string `json:"projectNumber"`
}

type GCPConnection struct {
	ID        string `json:"cloudAccountId"`
	RequestID string `json:"requestId"`
	Status    string `json:"provisioningStatus"`
}

func (c *Client) CreateGCPConnectionRequest(ctx context.Context, input CreateGCPConnectionRequestInput) (*GCPConnectionRequest, error) {
	var result GCPConnectionRequest
	if err := c.do(ctx, http.MethodPost, "/gcp_connection_requests", input, &result); err != nil {
		return nil, err
	}
	if result.ID == "" || result.WorkspaceID != input.WorkspaceID || result.ProjectID != input.ProjectID {
		return nil, fmt.Errorf("GCP connection request response did not match the requested workspace and project")
	}
	return &result, nil
}

func (c *Client) GetGCPConnectionRequest(ctx context.Context, workspaceID, requestID string) (*GCPConnectionRequest, error) {
	var result GCPConnectionRequest
	if err := c.do(ctx, http.MethodGet, gcpRequestPath(workspaceID, requestID), nil, &result); err != nil {
		return nil, err
	}
	if result.ID != requestID || result.WorkspaceID != workspaceID {
		return nil, fmt.Errorf("GCP connection request response did not match the requested identity")
	}
	return &result, nil
}

func (c *Client) DeleteGCPConnectionRequest(ctx context.Context, workspaceID, requestID string) error {
	return c.do(ctx, http.MethodDelete, gcpRequestPath(workspaceID, requestID), struct{}{}, nil)
}

func gcpRequestPath(workspaceID, requestID string) string {
	return "/gcp_connection_requests/" + url.PathEscape(workspaceID) + "/" + url.PathEscape(requestID)
}

func (c *Client) ActivateGCPConnection(ctx context.Context, input ActivateGCPConnectionInput) (*GCPConnection, error) {
	var result GCPConnection
	if err := c.do(ctx, http.MethodPost, "/gcp_connections", input, &result); err != nil {
		return nil, err
	}
	if result.ID == "" || result.RequestID != input.RequestID || result.Status != "registered" {
		return nil, fmt.Errorf("GCP activation response did not confirm registration")
	}
	return &result, nil
}

func (c *Client) GetGCPConnection(ctx context.Context, workspaceID, connectionID, requestID string) (*GCPConnection, error) {
	result, err := c.GetGCPConnectionRequest(ctx, workspaceID, requestID)
	if err != nil {
		return nil, err
	}
	if result.Status != "active" || result.CloudAccountID == nil || *result.CloudAccountID != connectionID {
		return nil, &APIError{StatusCode: http.StatusNotFound, Message: "GCP connection is no longer active"}
	}
	return &GCPConnection{ID: connectionID, RequestID: requestID, Status: "registered"}, nil
}

func (c *Client) DeleteGCPConnection(ctx context.Context, workspaceID, connectionID string) error {
	response, err := c.generated.CloudAccountsDisconnectWithResponse(ctx, workspaceID, connectionID)
	if err != nil {
		return requestError(err)
	}
	return generatedResponseError(response.StatusCode(), response.Body)
}
