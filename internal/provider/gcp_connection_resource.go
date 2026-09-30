package provider

import (
	"context"
	"fmt"

	"github.com/coreplanelabs/terraform-provider-polylane/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type gcpConnectionResource struct{ client *client.Client }
type gcpConnectionResourceModel struct {
	WorkspaceID   types.String `tfsdk:"workspace_id"`
	RequestID     types.String `tfsdk:"request_id"`
	ProjectNumber types.String `tfsdk:"project_number"`
	ID            types.String `tfsdk:"id"`
	Status        types.String `tfsdk:"status"`
}

var _ resource.ResourceWithConfigure = &gcpConnectionResource{}

func NewGCPConnectionResource() resource.Resource { return &gcpConnectionResource{} }
func (r *gcpConnectionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gcp_connection"
}
func (r *gcpConnectionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Description: "Activates a beta Google Cloud connection after customer-managed Google resources and IAM grants are ready. Destroy disconnects Polylane before Google teardown.", Attributes: map[string]schema.Attribute{
		"workspace_id":   schema.StringAttribute{Required: true, Description: "Polylane workspace ID.", Validators: []validator.String{stringvalidator.RegexMatches(workspaceIDPattern, "must be a valid workspace_id")}, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"request_id":     schema.StringAttribute{Required: true, Description: "ID returned by polylane_gcp_connection_request.", Validators: []validator.String{stringvalidator.LengthBetween(1, 128)}, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"project_number": schema.StringAttribute{Required: true, Description: "Numeric Google Cloud project number.", Validators: []validator.String{stringvalidator.RegexMatches(gcpProjectNumberPattern, "must be a valid project_number")}, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"id":             schema.StringAttribute{Computed: true, Description: "Polylane cloud-account ID.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"status":         schema.StringAttribute{Computed: true, Description: "Registration status. Initial synchronization continues asynchronously."},
	}}
}
func (r *gcpConnectionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	apiClient, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data Type", fmt.Sprintf("Expected *client.Client, got %T", req.ProviderData))
		return
	}
	r.client = apiClient
}
func (r *gcpConnectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Polylane Client", "Configure the Polylane provider before managing connections.")
		return
	}
	var plan gcpConnectionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := activateGCPWithRetry(ctx, r.client, client.ActivateGCPConnectionInput{WorkspaceID: plan.WorkspaceID.ValueString(), RequestID: plan.RequestID.ValueString(), ProjectNumber: plan.ProjectNumber.ValueString()}, waitForGCPRetry)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create GCP Connection", err.Error())
		return
	}
	setGCPConnectionState(&plan, remote)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *gcpConnectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Polylane Client", "Configure the Polylane provider before managing connections.")
		return
	}
	var state gcpConnectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := r.client.GetGCPConnection(ctx, state.WorkspaceID.ValueString(), state.ID.ValueString(), state.RequestID.ValueString())
	if client.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read GCP Connection", err.Error())
		return
	}
	setGCPConnectionState(&state, remote)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *gcpConnectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state gcpConnectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *gcpConnectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Polylane Client", "Configure the Polylane provider before managing connections.")
		return
	}
	var state gcpConnectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteGCPConnection(ctx, state.WorkspaceID.ValueString(), state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to Delete GCP Connection", err.Error())
	}
}
func setGCPConnectionState(state *gcpConnectionResourceModel, remote *client.GCPConnection) {
	state.ID = types.StringValue(remote.ID)
	state.Status = types.StringValue(remote.Status)
}
