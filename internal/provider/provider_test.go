package provider

import (
	"context"
	"testing"

	fwprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func TestProviderMetadata(t *testing.T) {
	t.Parallel()

	p := New("1.2.3")()

	resp := fwprovider.MetadataResponse{}
	p.Metadata(context.Background(), fwprovider.MetadataRequest{}, &resp)

	if resp.TypeName != "polylane" {
		t.Errorf("unexpected provider type name: got %q, want %q", resp.TypeName, "polylane")
	}
	if resp.Version != "1.2.3" {
		t.Errorf("unexpected provider version: got %q, want %q", resp.Version, "1.2.3")
	}
}

func TestProviderSchema(t *testing.T) {
	t.Parallel()

	p := New("test")()

	resp := fwprovider.SchemaResponse{}
	p.Schema(context.Background(), fwprovider.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected schema diagnostics: %v", resp.Diagnostics)
	}

	attrs := resp.Schema.Attributes
	apiKey, ok := attrs["api_key"]
	if !ok {
		t.Fatal("schema is missing the api_key attribute")
	}
	if !apiKey.IsSensitive() {
		t.Error("api_key attribute must be sensitive")
	}
	if !apiKey.IsOptional() {
		t.Error("api_key attribute must be optional")
	}

	endpoint, ok := attrs["endpoint"]
	if !ok {
		t.Fatal("schema is missing the endpoint attribute")
	}
	if !endpoint.IsOptional() {
		t.Error("endpoint attribute must be optional")
	}
}

// TestProviderServerSchema round-trips the provider through the protocol v6
// server, which runs the framework's own schema implementation validation.
func TestProviderServerSchema(t *testing.T) {
	t.Parallel()

	server, err := providerserver.NewProtocol6WithError(New("test")())()
	if err != nil {
		t.Fatalf("failed to create protocol v6 server: %v", err)
	}

	resp, err := server.GetProviderSchema(context.Background(), &tfprotov6.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatalf("GetProviderSchema failed: %v", err)
	}

	for _, diag := range resp.Diagnostics {
		if diag.Severity == tfprotov6.DiagnosticSeverityError {
			t.Errorf("unexpected schema diagnostic: %s: %s", diag.Summary, diag.Detail)
		}
	}

	resourceNames := []string{
		"polylane_aws_connection",
		"polylane_aws_connection_request",
		"polylane_team",
		"polylane_team_member",
		"polylane_workspace",
		"polylane_workspace_member",
		"polylane_workspace_autofix_settings",
		"polylane_workspace_digest_settings",
		"polylane_workspace_investigation_limits_settings",
		"polylane_workspace_investigations_settings",
		"polylane_workspace_model_training_settings",
		"polylane_workspace_observability_settings",
		"polylane_workspace_pr_review_settings",
	}
	if len(resp.ResourceSchemas) != len(resourceNames) {
		t.Fatalf("unexpected resource schema count: got %d, want %d", len(resp.ResourceSchemas), len(resourceNames))
	}
	for _, name := range resourceNames {
		if _, ok := resp.ResourceSchemas[name]; !ok {
			t.Errorf("provider schema is missing resource %q", name)
		}
	}
	if len(resp.DataSourceSchemas) != 0 {
		t.Errorf("expected no data sources in the skeleton provider, got %d", len(resp.DataSourceSchemas))
	}
}
