package client

import (
	"context"
	"net/http"

	api "github.com/coreplanelabs/terraform-provider-polylane/internal/client/generated"
)

// CreateAWSConnectionRequest creates the Polylane half of a customer-managed
// AWS connection handshake. It never sends AWS credentials to Polylane.
func (c *Client) CreateAWSConnectionRequest(ctx context.Context, input CreateAWSConnectionRequestInput) (*AWSConnectionRequest, error) {
	regions := append([]string(nil), input.Regions...)
	body := api.AwsConnectionRequestsCreateJSONRequestBody{
		WorkspaceId:    input.WorkspaceID,
		AwsAccountId:   input.AccountID,
		IdempotencyKey: input.IdempotencyKey,
		Regions:        &regions,
	}
	response, err := c.generated.AwsConnectionRequestsCreateWithResponse(ctx, body)
	if err != nil {
		return nil, requestError(err)
	}
	if err := generatedResponseError(response.StatusCode(), response.Body); err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, missingGeneratedResult("create AWS connection request")
	}

	return convertAWSConnectionRequest(input.WorkspaceID, response.JSON200.Result), nil
}

func (c *Client) GetAWSConnectionRequest(ctx context.Context, workspaceID, requestID string) (*AWSConnectionRequest, error) {
	response, err := c.generated.AwsConnectionRequestsGetWithResponse(ctx, workspaceID, requestID)
	if err != nil {
		return nil, requestError(err)
	}
	if err := generatedResponseError(response.StatusCode(), response.Body); err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, missingGeneratedResult("get AWS connection request")
	}

	return convertAWSConnectionRequest(workspaceID, response.JSON200.Result), nil
}

func (c *Client) DeleteAWSConnectionRequest(ctx context.Context, workspaceID, requestID string) error {
	response, err := c.generated.AwsConnectionRequestsDeleteWithResponse(ctx, workspaceID, requestID)
	if err != nil {
		return requestError(err)
	}
	return generatedResponseError(response.StatusCode(), response.Body)
}

func convertAWSConnectionRequest(workspaceID string, result api.AwsConnectionRequest) *AWSConnectionRequest {
	var regions []string
	if result.Regions != nil {
		regions = append(regions, (*result.Regions)...)
	}

	return &AWSConnectionRequest{
		ID:                   result.RequestId,
		WorkspaceID:          workspaceID,
		AccountID:            result.AwsAccountId,
		Regions:              regions,
		ExternalID:           result.ExternalId,
		PrincipalARN:         result.TrustedPrincipal.Arn,
		SubscriptionEndpoint: result.SubscriptionEndpoint,
		CloudAccountID:       valueOrEmpty(result.CloudAccountId),
		Region:               result.Region,
		Status:               string(result.Status),
	}
}

func (c *Client) ActivateAWSConnection(ctx context.Context, input ActivateAWSConnectionInput) (*AWSConnection, error) {
	body := api.AwsConnectionsActivateJSONRequestBody{
		WorkspaceId:          input.WorkspaceID,
		RequestId:            input.RequestID,
		Region:               input.Region,
		RoleArn:              input.RoleARN,
		BucketName:           input.BucketName,
		TopicArn:             input.TopicARN,
		TopicSubscriptionArn: input.TopicSubscriptionARN,
		CloudTrailName:       input.CloudTrailName,
	}
	response, err := c.generated.AwsConnectionsActivateWithResponse(ctx, body)
	if err != nil {
		return nil, requestError(err)
	}
	if err := generatedResponseError(response.StatusCode(), response.Body); err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, missingGeneratedResult("activate AWS connection")
	}

	connectionRequest, err := c.GetAWSConnectionRequest(ctx, input.WorkspaceID, input.RequestID)
	if err != nil {
		return nil, err
	}
	if connectionRequest.CloudAccountID != response.JSON200.Result.CloudAccountId {
		return nil, missingGeneratedResult("activate AWS connection registration")
	}
	return &AWSConnection{
		ID:                   response.JSON200.Result.CloudAccountId,
		WorkspaceID:          input.WorkspaceID,
		RequestID:            input.RequestID,
		AccountID:            connectionRequest.AccountID,
		Regions:              connectionRequest.Regions,
		Region:               input.Region,
		RoleARN:              input.RoleARN,
		BucketName:           input.BucketName,
		TopicARN:             input.TopicARN,
		TopicSubscriptionARN: input.TopicSubscriptionARN,
		CloudTrailName:       input.CloudTrailName,
		Status:               string(response.JSON200.Result.ProvisioningStatus),
	}, nil
}

func (c *Client) GetAWSConnection(ctx context.Context, workspaceID, connectionID, requestID string) (*AWSConnection, error) {
	connectionRequest, err := c.GetAWSConnectionRequest(ctx, workspaceID, requestID)
	if err != nil {
		return nil, err
	}
	if connectionRequest.CloudAccountID == "" || connectionRequest.CloudAccountID != connectionID {
		return nil, &APIError{StatusCode: http.StatusNotFound, Message: "customer-managed AWS connection does not exist"}
	}

	return &AWSConnection{
		ID:          connectionRequest.CloudAccountID,
		WorkspaceID: workspaceID,
		RequestID:   requestID,
		AccountID:   connectionRequest.AccountID,
		Regions:     connectionRequest.Regions,
		Region:      connectionRequest.Region,
		Status:      connectionRequest.Status,
	}, nil
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

// DeleteAWSConnection removes only Polylane's registration. Customer-managed
// AWS resources are deliberately left to the customer's AWS provider.
func (c *Client) DeleteAWSConnection(ctx context.Context, workspaceID, connectionID string) error {
	response, err := c.generated.CloudAccountsDisconnectWithResponse(ctx, workspaceID, connectionID)
	if err != nil {
		return requestError(err)
	}
	return generatedResponseError(response.StatusCode(), response.Body)
}
