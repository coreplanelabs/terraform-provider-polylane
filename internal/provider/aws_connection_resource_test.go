package provider

import (
	"context"
	"reflect"
	"testing"

	"github.com/coreplanelabs/terraform-provider-polylane/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestAWSConnectionResourceSchemas(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		newResource func() resource.Resource
		required    []string
		computed    []string
		forbidden   []string
	}{
		{
			name:        "connection request exposes handshake values",
			newResource: NewAWSConnectionRequestResource,
			required:    []string{"workspace_id", "account_id", "regions"},
			computed:    []string{"id", "external_id", "principal_arn", "subscription_endpoint", "region", "status"},
			forbidden:   []string{"api_key", "aws_access_key", "aws_secret_key", "role_arn"},
		},
		{
			name:        "connection accepts only customer owned identifiers",
			newResource: NewAWSConnectionResource,
			required: []string{
				"workspace_id",
				"request_id",
				"region",
				"role_arn",
				"bucket_name",
				"topic_arn",
				"topic_subscription_arn",
				"cloudtrail_name",
			},
			computed:  []string{"id", "account_id", "regions", "status"},
			forbidden: []string{"api_key", "aws_access_key", "aws_secret_key", "stack_name", "template_body"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var response resource.SchemaResponse
			test.newResource().Schema(context.Background(), resource.SchemaRequest{}, &response)
			if response.Diagnostics.HasError() {
				t.Fatalf("unexpected schema diagnostics: %v", response.Diagnostics)
			}

			for _, name := range test.required {
				attribute, ok := response.Schema.Attributes[name]
				if !ok {
					t.Errorf("schema is missing required attribute %q", name)
					continue
				}
				if !attribute.IsRequired() {
					t.Errorf("attribute %q must be required", name)
				}
			}
			for _, name := range test.computed {
				attribute, ok := response.Schema.Attributes[name]
				if !ok {
					t.Errorf("schema is missing computed attribute %q", name)
					continue
				}
				if !attribute.IsComputed() {
					t.Errorf("attribute %q must be computed", name)
				}
			}
			for _, name := range test.forbidden {
				if _, ok := response.Schema.Attributes[name]; ok {
					t.Errorf("schema must not expose %q", name)
				}
			}
		})
	}
}

func TestSetAWSConnectionRequestState(t *testing.T) {
	t.Parallel()

	remote := &client.AWSConnectionRequest{
		ID:                   "cloudformation_000000000000000000000000",
		WorkspaceID:          "ws_00000000000000000000000000000000",
		AccountID:            "123456789012",
		Regions:              []string{"us-west-2", "us-east-1"},
		ExternalID:           "cloudformation_000000000000000000000000",
		PrincipalARN:         "arn:aws:iam::251714435813:root",
		SubscriptionEndpoint: "https://aws.polylane.com/sns",
		Region:               "us-east-1",
		Status:               "pending",
	}
	var state awsConnectionRequestResourceModel
	var diagnostics diag.Diagnostics
	setAWSConnectionRequestState(context.Background(), &state, remote, &diagnostics)

	if diagnostics.HasError() {
		t.Fatalf("unexpected state diagnostics: %v", diagnostics)
	}
	if state.ID.ValueString() != remote.ID || state.ExternalID.ValueString() != remote.ExternalID {
		t.Fatalf("unexpected request identity in state: %#v", state)
	}
	var regions []string
	diagnostics.Append(state.Regions.ElementsAs(context.Background(), &regions, false)...)
	if diagnostics.HasError() {
		t.Fatalf("decode regions: %v", diagnostics)
	}
	if !reflect.DeepEqual(regions, remote.Regions) {
		t.Fatalf("unexpected regions: got %#v, want %#v", regions, remote.Regions)
	}
}

func TestAWSConnectionRequestIdempotencyKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		workspace string
		account   string
		regions   []string
	}{
		{name: "single region", workspace: "ws_one", account: "123456789012", regions: []string{"us-east-1"}},
		{name: "multiple regions", workspace: "ws_one", account: "123456789012", regions: []string{"us-west-2", "us-east-1"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			first := awsConnectionRequestIdempotencyKey(test.workspace, test.account, test.regions)
			reversed := append([]string(nil), test.regions...)
			for left, right := 0, len(reversed)-1; left < right; left, right = left+1, right-1 {
				reversed[left], reversed[right] = reversed[right], reversed[left]
			}
			second := awsConnectionRequestIdempotencyKey(test.workspace, test.account, reversed)

			if first != second {
				t.Fatalf("key depends on region order: %q != %q", first, second)
			}
			if len(first) != len("terraform_")+32 {
				t.Fatalf("unexpected key length %d for %q", len(first), first)
			}
		})
	}

	base := awsConnectionRequestIdempotencyKey("ws_one", "123456789012", []string{"us-east-1"})
	changedAccount := awsConnectionRequestIdempotencyKey("ws_one", "999999999999", []string{"us-east-1"})
	changedRegions := awsConnectionRequestIdempotencyKey("ws_one", "123456789012", []string{"us-west-2"})
	if base == changedAccount || base == changedRegions {
		t.Fatal("different connection identities produced the same idempotency key")
	}
}

func TestSetAWSConnectionState(t *testing.T) {
	t.Parallel()

	remote := &client.AWSConnection{
		ID:                   "acc_000000000000000000000000",
		WorkspaceID:          "ws_00000000000000000000000000000000",
		RequestID:            "cloudformation_000000000000000000000000",
		AccountID:            "123456789012",
		Regions:              []string{"us-east-1"},
		Region:               "us-east-1",
		RoleARN:              "arn:aws:iam::123456789012:role/polylane",
		BucketName:           "erebor-polylane-cloudtrail",
		TopicARN:             "arn:aws:sns:us-east-1:123456789012:polylane",
		TopicSubscriptionARN: "arn:aws:sns:us-east-1:123456789012:polylane:00000000-0000-0000-0000-000000000000",
		CloudTrailName:       "polylane",
		Status:               "new",
	}
	var state awsConnectionResourceModel
	var diagnostics diag.Diagnostics
	setAWSConnectionState(context.Background(), &state, remote, &diagnostics)

	if diagnostics.HasError() {
		t.Fatalf("unexpected state diagnostics: %v", diagnostics)
	}
	if state.ID.ValueString() != remote.ID || state.RoleARN.ValueString() != remote.RoleARN || state.TopicSubscriptionARN.ValueString() != remote.TopicSubscriptionARN {
		t.Fatalf("unexpected connection state: %#v", state)
	}
}
