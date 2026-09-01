package provider

import (
	"context"
	"fmt"

	"github.com/coreplanelabs/terraform-provider-polylane/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type teamMemberResource struct {
	client *client.Client
}

type teamMemberResourceModel struct {
	WorkspaceID types.String `tfsdk:"workspace_id"`
	TeamID      types.String `tfsdk:"team_id"`
	UserID      types.String `tfsdk:"user_id"`
	AddedBy     types.String `tfsdk:"added_by"`
	CreatedAt   types.String `tfsdk:"created_at"`
}

var (
	_ resource.Resource                = &teamMemberResource{}
	_ resource.ResourceWithConfigure   = &teamMemberResource{}
	_ resource.ResourceWithImportState = &teamMemberResource{}
)

func NewTeamMemberResource() resource.Resource {
	return &teamMemberResource{}
}

func (r *teamMemberResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team_member"
}

func (r *teamMemberResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	requiresReplace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		Description: "Manages a user's membership in a Polylane team.",
		Attributes: map[string]schema.Attribute{
			"workspace_id": schema.StringAttribute{
				Required:      true,
				Description:   "ID of the workspace containing the team.",
				Validators:    []validator.String{stringvalidator.RegexMatches(workspaceIDPattern, "must be a valid Polylane workspace ID")},
				PlanModifiers: requiresReplace,
			},
			"team_id": schema.StringAttribute{
				Required:      true,
				Description:   "ID of the team.",
				Validators:    []validator.String{stringvalidator.RegexMatches(teamIDPattern, "must be a valid Polylane team ID")},
				PlanModifiers: requiresReplace,
			},
			"user_id": schema.StringAttribute{
				Required:      true,
				Description:   "ID of the user to include in the team.",
				Validators:    []validator.String{stringvalidator.RegexMatches(subjectIDPattern, "must be a valid Polylane user or team ID")},
				PlanModifiers: requiresReplace,
			},
			"added_by":   schema.StringAttribute{Computed: true, Description: "ID of the user who added this member."},
			"created_at": schema.StringAttribute{Computed: true, Description: "RFC 3339 timestamp when the team membership was created."},
		},
	}
}

func (r *teamMemberResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	apiClient, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data Type", fmt.Sprintf("Expected *client.Client, got %T. Please report this issue to the Polylane provider developers.", req.ProviderData))
		return
	}
	r.client = apiClient
}

func (r *teamMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Polylane Client", "The provider client was not configured. Please report this issue to the Polylane provider developers.")
		return
	}

	var plan teamMemberResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	member, err := r.client.AddTeamMember(ctx, plan.WorkspaceID.ValueString(), plan.TeamID.ValueString(), plan.UserID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Add Polylane Team Member", err.Error())
		return
	}
	setTeamMemberState(&plan, member)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *teamMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Polylane Client", "The provider client was not configured. Please report this issue to the Polylane provider developers.")
		return
	}

	var state teamMemberResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	member, err := r.client.GetTeamMember(ctx, state.WorkspaceID.ValueString(), state.TeamID.ValueString(), state.UserID.ValueString())
	if client.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Polylane Team Member", err.Error())
		return
	}

	setTeamMemberState(&state, member)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *teamMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All configurable fields are membership identity and require replacement.
	// Refresh here to keep protocol behavior complete if the framework calls it.
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Polylane Client", "The provider client was not configured. Please report this issue to the Polylane provider developers.")
		return
	}
	var state teamMemberResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	member, err := r.client.GetTeamMember(ctx, state.WorkspaceID.ValueString(), state.TeamID.ValueString(), state.UserID.ValueString())
	if client.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Polylane Team Member", err.Error())
		return
	}
	setTeamMemberState(&state, member)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *teamMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Polylane Client", "The provider client was not configured. Please report this issue to the Polylane provider developers.")
		return
	}

	var state teamMemberResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.RemoveTeamMember(ctx, state.WorkspaceID.ValueString(), state.TeamID.ValueString(), state.UserID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to Remove Polylane Team Member", err.Error())
	}
}

func (r *teamMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := parseCompositeImportID(req.ID, 3)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Team Member Import ID", "Expected `<workspace-id>/<team-id>/<user-id>`. "+err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("workspace_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("team_id"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user_id"), parts[2])...)
}

func setTeamMemberState(state *teamMemberResourceModel, member *client.TeamMember) {
	state.TeamID = types.StringValue(member.TeamID)
	state.UserID = types.StringValue(member.UserID)
	state.AddedBy = types.StringValue(member.AddedBy)
	state.CreatedAt = nullableString(member.Created)
}
