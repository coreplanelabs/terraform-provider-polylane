package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/coreplanelabs/terraform-provider-polylane/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type gcpProviderTransport func(*http.Request) (*http.Response, error)

func (f gcpProviderTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestGCPRequestCreateUsesFreshDefaultIdentity(t *testing.T) {
	t.Parallel()
	keys := []string{}
	c, err := client.New("test", "https://example.invalid/v1", "test", &http.Client{Transport: gcpProviderTransport(func(r *http.Request) (*http.Response, error) {
		var input client.CreateGCPConnectionRequestInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			return nil, err
		}
		keys = append(keys, input.IdempotencyKey)
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(fmt.Sprintf(`{"success":true,"result":{"requestId":"request","workspaceId":%q,"projectId":%q,"subject":"subject","issuerUrl":"https://example.invalid/issuer","pushEndpoint":"https://example.invalid/push","resourcePrefix":"polylane-1234567890abcdef","status":"pending","expiresAt":"2026-10-06T00:00:00Z"}}`, input.WorkspaceID, input.ProjectID)))}, nil
	})})
	if err != nil {
		t.Fatal(err)
	}
	r := &gcpConnectionRequestResource{client: c}
	var schema resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schema)
	for range 2 {
		model := gcpConnectionRequestResourceModel{WorkspaceID: types.StringValue("ws_00000000000000000000000000000000"), ProjectID: types.StringValue("example-project"), IdempotencyKey: types.StringUnknown()}
		plan := tfsdk.Plan{Schema: schema.Schema}
		if d := plan.Set(context.Background(), model); d.HasError() {
			t.Fatal(d)
		}
		response := resource.CreateResponse{State: tfsdk.State{Schema: schema.Schema}}
		r.Create(context.Background(), resource.CreateRequest{Plan: plan}, &response)
		if response.Diagnostics.HasError() {
			t.Fatal(response.Diagnostics)
		}
		var state gcpConnectionRequestResourceModel
		if d := response.State.Get(context.Background(), &state); d.HasError() {
			t.Fatal(d)
		}
		if state.IdempotencyKey.ValueString() == "" || state.IdempotencyKey.ValueString() != keys[len(keys)-1] {
			t.Fatal("creation key was not persisted")
		}
	}
	if keys[0] == keys[1] {
		t.Fatal("recreation reused identity")
	}
}

func TestGCPDisconnectFailurePreservesState(t *testing.T) {
	t.Parallel()
	c, err := client.New("test", "https://example.invalid/v1", "test", &http.Client{Transport: gcpProviderTransport(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 503, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"success":false,"message":"retry later"}`))}, nil
	})})
	if err != nil {
		t.Fatal(err)
	}
	r := &gcpConnectionResource{client: c}
	var schema resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schema)
	state := tfsdk.State{Schema: schema.Schema}
	if d := state.Set(context.Background(), gcpConnectionResourceModel{ID: types.StringValue("account"), WorkspaceID: types.StringValue("workspace"), RequestID: types.StringValue("request"), ProjectNumber: types.StringValue("123456789"), Status: types.StringValue("registered")}); d.HasError() {
		t.Fatal(d)
	}
	response := resource.DeleteResponse{State: state}
	r.Delete(context.Background(), resource.DeleteRequest{State: state}, &response)
	if !response.Diagnostics.HasError() {
		t.Fatal("disconnect failure must block destroy")
	}
	var model gcpConnectionResourceModel
	if d := response.State.Get(context.Background(), &model); d.HasError() {
		t.Fatal(d)
	}
	if model.ID.ValueString() != "account" {
		t.Fatal("disconnect error discarded state")
	}
}
