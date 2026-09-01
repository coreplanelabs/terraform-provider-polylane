package provider

import (
	"testing"

	"github.com/coreplanelabs/terraform-provider-polylane/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSettingsResourceDescriptors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		constructor func() resource.Resource
		path        string
	}{
		{name: "autofix", constructor: NewWorkspaceAutofixSettingsResource, path: "autofix_settings"},
		{name: "digest", constructor: NewWorkspaceDigestSettingsResource, path: "digest_settings"},
		{name: "investigations", constructor: NewWorkspaceInvestigationsSettingsResource, path: "investigations_settings"},
		{name: "investigation limits", constructor: NewWorkspaceInvestigationLimitsSettingsResource, path: "investigation_limits_settings"},
		{name: "model training", constructor: NewWorkspaceModelTrainingSettingsResource, path: "model_training_settings"},
		{name: "observability", constructor: NewWorkspaceObservabilitySettingsResource, path: "observability_settings"},
		{name: "pull request reviews", constructor: NewWorkspacePRReviewSettingsResource, path: "pr_review_settings"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			settingsResource, ok := test.constructor().(*workspaceSettingsResource)
			if !ok {
				t.Fatalf("constructor returned %T", test.constructor())
			}
			if settingsResource.descriptor.settingsPath != test.path {
				t.Errorf("unexpected settings path: got %q, want %q", settingsResource.descriptor.settingsPath, test.path)
			}
			if _, ok := settingsResource.descriptor.attributes["workspace_id"]; !ok {
				t.Error("schema is missing workspace_id")
			}
		})
	}
}

func TestSettingsResponsesUseIntentValues(t *testing.T) {
	t.Parallel()

	autofix, ok := NewWorkspaceAutofixSettingsResource().(*workspaceSettingsResource)
	if !ok {
		t.Fatal("autofix settings constructor returned an unexpected resource type")
	}
	autofixModel := &autofixSettingsModel{WorkspaceID: types.StringValue("ws_123")}
	autofixResponse := &client.AutofixSettings{}
	autofixResponse.Autofix.TriggerMode = "approval"
	disabledAt := "2026-01-01T00:00:00Z"
	autofixResponse.Autofix.DisabledAt = &disabledAt

	if err := autofix.descriptor.setFromResponse(autofixModel, autofixResponse); err != nil {
		t.Fatalf("setFromResponse returned an error: %v", err)
	}
	if autofixModel.Enabled.ValueBool() {
		t.Error("expected disabledAt to map to enabled=false")
	}
	if got := autofixModel.TriggerMode.ValueString(); got != "approval" {
		t.Errorf("unexpected trigger mode: %q", got)
	}

	digest, ok := NewWorkspaceDigestSettingsResource().(*workspaceSettingsResource)
	if !ok {
		t.Fatal("digest settings constructor returned an unexpected resource type")
	}
	digestModel := &digestSettingsModel{WorkspaceID: types.StringValue("ws_123")}
	digestResponse := &client.DigestSettings{}
	disabledMillis := float32(1)
	digestResponse.DailyTrendsDisabledAt = &disabledMillis

	if err := digest.descriptor.setFromResponse(digestModel, digestResponse); err != nil {
		t.Fatalf("setFromResponse returned an error: %v", err)
	}
	if !digestModel.Enabled.ValueBool() || digestModel.DailyTrendsEnabled.ValueBool() {
		t.Errorf("unexpected digest intent values: %#v", digestModel)
	}
}
