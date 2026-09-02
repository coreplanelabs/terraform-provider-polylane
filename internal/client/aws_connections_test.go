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

const (
	testWorkspaceID  = "ws_00000000000000000000000000000000"
	testRequestID    = "cloudformation_000000000000000000000000"
	testConnectionID = "cloud_account_000000000000000000000000"
)

func TestAWSConnectionRequestRoutes(t *testing.T) {
	t.Parallel()

	requestResult := fmt.Sprintf(`{"requestId":%q,"externalId":%q,"awsAccountId":"123456789012","region":"us-east-1","regions":["us-east-1"],"trustedPrincipal":{"accountId":"251714435813","arn":"arn:aws:iam::251714435813:root"},"subscriptionEndpoint":"https://aws.polylane.com/sns","status":"pending","cloudAccountId":null}`, testRequestID, testRequestID)
	tests := []struct {
		name        string
		method      string
		path        string
		body        string
		contentType string
		result      string
		invoke      func(context.Context, *Client) error
	}{
		{
			name:        "create",
			method:      http.MethodPost,
			path:        "/v1/aws_connection_requests",
			body:        `"awsAccountId":"123456789012"`,
			contentType: "application/json",
			result:      requestResult,
			invoke: func(ctx context.Context, apiClient *Client) error {
				result, err := apiClient.CreateAWSConnectionRequest(ctx, CreateAWSConnectionRequestInput{
					WorkspaceID: testWorkspaceID, AccountID: "123456789012", Regions: []string{"us-east-1"}, IdempotencyKey: "terraform_test",
				})
				if err == nil && (result.ID != testRequestID || result.SubscriptionEndpoint != "https://aws.polylane.com/sns") {
					return fmt.Errorf("unexpected connection request: %#v", result)
				}
				return err
			},
		},
		{
			name:   "get",
			method: http.MethodGet,
			path:   "/v1/aws_connection_requests/" + testWorkspaceID + "/" + testRequestID,
			result: requestResult,
			invoke: func(ctx context.Context, apiClient *Client) error {
				result, err := apiClient.GetAWSConnectionRequest(ctx, testWorkspaceID, testRequestID)
				if err == nil && result.PrincipalARN != "arn:aws:iam::251714435813:root" {
					return fmt.Errorf("unexpected principal ARN %q", result.PrincipalARN)
				}
				return err
			},
		},
		{
			name:        "delete",
			method:      http.MethodDelete,
			path:        "/v1/aws_connection_requests/" + testWorkspaceID + "/" + testRequestID,
			contentType: "application/json",
			result:      fmt.Sprintf(`{"requestId":%q}`, testRequestID),
			invoke: func(ctx context.Context, apiClient *Client) error {
				return apiClient.DeleteAWSConnectionRequest(ctx, testWorkspaceID, testRequestID)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			var requestBody string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
				body, _ := io.ReadAll(request.Body)
				requestBody = string(body)
				if request.Method != test.method || request.URL.Path != test.path {
					t.Errorf("unexpected request: %s %s", request.Method, request.URL.Path)
				}
				if got := request.Header.Get("Content-Type"); got != test.contentType {
					t.Errorf("unexpected content type: got %q, want %q", got, test.contentType)
				}
				writeTestEnvelope(w, test.result)
			}))
			t.Cleanup(server.Close)

			apiClient := newTestClient(t, server)
			if err := test.invoke(context.Background(), apiClient); err != nil {
				t.Fatalf("operation returned an error: %v", err)
			}
			if test.body != "" && !strings.Contains(requestBody, test.body) {
				t.Errorf("request body %q does not contain %q", requestBody, test.body)
			}
			for _, forbidden := range []string{"accessKey", "secretAccessKey", "sessionToken"} {
				if strings.Contains(requestBody, forbidden) {
					t.Errorf("request body contains forbidden AWS credential field %q", forbidden)
				}
			}
		})
	}
}

func TestAWSConnectionActivationAndLifecycleRoutes(t *testing.T) {
	t.Parallel()

	registeredRequestResult := fmt.Sprintf(`{"requestId":%q,"externalId":%q,"awsAccountId":"123456789012","region":"us-east-1","regions":["us-east-1"],"trustedPrincipal":{"accountId":"251714435813","arn":"arn:aws:iam::251714435813:root"},"subscriptionEndpoint":"https://aws.polylane.com/sns","status":"registered","cloudAccountId":%q}`, testRequestID, testRequestID, testConnectionID)
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		body, _ := io.ReadAll(request.Body)
		requests = append(requests, request.Method+" "+request.URL.Path+" "+string(body))
		switch {
		case request.Method == http.MethodPost && request.URL.Path == "/v1/aws_connections":
			writeTestEnvelope(w, fmt.Sprintf(`{"requestId":%q,"cloudAccountId":%q,"provisioningStatus":"registered"}`, testRequestID, testConnectionID))
		case request.Method == http.MethodGet && request.URL.Path == "/v1/aws_connection_requests/"+testWorkspaceID+"/"+testRequestID:
			writeTestEnvelope(w, registeredRequestResult)
		case request.Method == http.MethodDelete && request.URL.Path == "/v1/cloud_accounts/"+testWorkspaceID+"/"+testConnectionID:
			writeTestEnvelope(w, fmt.Sprintf(`{"id":%q,"_object":"cloud_account"}`, testConnectionID))
		default:
			http.NotFound(w, request)
		}
	}))
	t.Cleanup(server.Close)

	apiClient := newTestClient(t, server)
	connection, err := apiClient.ActivateAWSConnection(context.Background(), ActivateAWSConnectionInput{
		WorkspaceID:          testWorkspaceID,
		RequestID:            testRequestID,
		Region:               "us-east-1",
		RoleARN:              "arn:aws:iam::123456789012:role/polylane-read",
		BucketName:           "example-cloudtrail",
		TopicARN:             "arn:aws:sns:us-east-1:123456789012:polylane-cloudtrail",
		TopicSubscriptionARN: "arn:aws:sns:us-east-1:123456789012:polylane-cloudtrail:12345678-1234-1234-1234-123456789012",
		CloudTrailName:       "polylane-cloudtrail",
	})
	if err != nil {
		t.Fatalf("ActivateAWSConnection returned an error: %v", err)
	}
	if connection.ID != testConnectionID || connection.RequestID != testRequestID || connection.Status != "registered" {
		t.Fatalf("unexpected connection: %#v", connection)
	}
	if len(requests) != 2 || !strings.HasPrefix(requests[0], "POST /v1/aws_connections ") || !strings.HasPrefix(requests[1], "GET /v1/aws_connection_requests/") {
		t.Fatalf("unexpected activation request sequence: %#v", requests)
	}
	if strings.Contains(requests[0], "accessKey") || strings.Contains(requests[0], "secretAccessKey") {
		t.Fatalf("activation sent AWS credentials: %s", requests[0])
	}

	read, err := apiClient.GetAWSConnection(context.Background(), testWorkspaceID, testConnectionID, testRequestID)
	if err != nil || read.ID != testConnectionID || read.Status != "registered" {
		t.Fatalf("unexpected read result %#v or error %v", read, err)
	}
	if err := apiClient.DeleteAWSConnection(context.Background(), testWorkspaceID, testConnectionID); err != nil {
		t.Fatalf("DeleteAWSConnection returned an error: %v", err)
	}
}

func TestGetAWSConnectionRejectsUnregisteredRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeTestEnvelope(w, fmt.Sprintf(`{"requestId":%q,"externalId":%q,"awsAccountId":"123456789012","region":"us-east-1","regions":["us-east-1"],"trustedPrincipal":{"accountId":"251714435813","arn":"arn:aws:iam::251714435813:root"},"subscriptionEndpoint":"https://aws.polylane.com/sns","status":"pending","cloudAccountId":null}`, testRequestID, testRequestID))
	}))
	t.Cleanup(server.Close)

	_, err := newTestClient(t, server).GetAWSConnection(context.Background(), testWorkspaceID, testConnectionID, testRequestID)
	if !IsNotFound(err) {
		t.Fatalf("expected an unregistered connection to return not found, got %v", err)
	}
}

func newTestClient(t *testing.T, server *httptest.Server) *Client {
	t.Helper()
	apiClient, err := New("test-key", server.URL+"/v1", "test", server.Client())
	if err != nil {
		t.Fatalf("New returned an error: %v", err)
	}
	return apiClient
}

func writeTestEnvelope(writer http.ResponseWriter, result string) {
	writer.Header().Set("Content-Type", "application/json")
	_, _ = fmt.Fprintf(writer, `{"success":true,"error":null,"message":{"message":"ok"},"result":%s}`, result)
}
