package provider

import (
	"context"
	"fmt"
	"regexp"

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

var (
	teamIDPattern   = regexp.MustCompile(`^team_[0-9a-z]{24}$`)
	teamSlugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
)

type teamResource struct {
	client *client.Client
}

type teamResourceModel struct {
	ID          types.String `tfsdk:"id"`
	WorkspaceID types.String `tfsdk:"workspace_id"`
	Name        types.String `tfsdk:"name"`
	Slug        types.String `tfsdk:"slug"`
	Description types.String `tfsdk:"description"`
	Color       types.String `tfsdk:"color"`
	Icon        types.String `tfsdk:"icon"`
	CreatedBy   types.String `tfsdk:"created_by"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

var (
	_ resource.Resource                = &teamResource{}
	_ resource.ResourceWithConfigure   = &teamResource{}
	_ resource.ResourceWithImportState = &teamResource{}
)

func NewTeamResource() resource.Resource {
	return &teamResource{}
}

func (r *teamResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team"
}

func (r *teamResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a team within an existing Polylane workspace.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique Polylane team ID.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(teamIDPattern, "must be a valid Polylane team ID"),
				},
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"workspace_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the workspace containing the team.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(workspaceIDPattern, "must be a valid Polylane workspace ID"),
				},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Human-readable team name.",
				Validators:  []validator.String{stringvalidator.LengthBetween(1, 128)},
			},
			"slug": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "URL-friendly team identifier. Polylane generates it from name when omitted.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 128),
					stringvalidator.RegexMatches(teamSlugPattern, "must contain lowercase letters and numbers separated by single hyphens"),
				},
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"description": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "Team description.",
				Validators:    []validator.String{stringvalidator.LengthAtMost(2000)},
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"color": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "Color used for team badges in Polylane.",
				Validators:    []validator.String{stringvalidator.LengthAtMost(32)},
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"icon": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "Polylane icon identifier for the team.",
				Validators:    []validator.String{stringvalidator.LengthAtMost(64)},
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"created_by": schema.StringAttribute{Computed: true, Description: "ID of the user who created the team."},
			"created_at": schema.StringAttribute{Computed: true, Description: "RFC 3339 timestamp when the team was created."},
			"updated_at": schema.StringAttribute{Computed: true, Description: "RFC 3339 timestamp when the team was last updated."},
		},
	}
}

func (r *teamResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *teamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Polylane Client", "The provider client was not configured. Please report this issue to the Polylane provider developers.")
		return
	}

	var plan teamResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	team, err := r.client.CreateTeam(ctx, client.CreateTeamInput{
		WorkspaceID: plan.WorkspaceID.ValueString(),
		Name:        plan.Name.ValueString(),
		Slug:        optionalStringPointer(plan.Slug),
		Description: optionalStringPointer(plan.Description),
		Color:       optionalStringPointer(plan.Color),
		Icon:        optionalStringPointer(plan.Icon),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create Polylane Team", err.Error())
		return
	}

	setTeamState(&plan, team)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *teamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Polylane Client", "The provider client was not configured. Please report this issue to the Polylane provider developers.")
		return
	}

	var state teamResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	team, err := r.client.GetTeam(ctx, state.WorkspaceID.ValueString(), state.ID.ValueString())
	if client.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Polylane Team", err.Error())
		return
	}

	setTeamState(&state, team)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *teamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Polylane Client", "The provider client was not configured. Please report this issue to the Polylane provider developers.")
		return
	}

	var plan teamResourceModel
	var state teamResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	team, err := r.client.UpdateTeam(ctx, state.ID.ValueString(), client.UpdateTeamInput{
		WorkspaceID: state.WorkspaceID.ValueString(),
		Name:        &name,
		Slug:        optionalStringPointer(plan.Slug),
		Description: optionalStringPointer(plan.Description),
		Color:       optionalStringPointer(plan.Color),
		Icon:        optionalStringPointer(plan.Icon),
	})
	if client.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update Polylane Team", err.Error())
		return
	}

	setTeamState(&plan, team)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *teamResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Polylane Client", "The provider client was not configured. Please report this issue to the Polylane provider developers.")
		return
	}

	var state teamResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteTeam(ctx, state.WorkspaceID.ValueString(), state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to Delete Polylane Team", err.Error())
	}
}

func (r *teamResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := parseCompositeImportID(req.ID, 2)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Team Import ID", "Expected `<workspace-id>/<team-id>`. "+err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("workspace_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

func optionalStringPointer(value types.String) *string {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	result := value.ValueString()
	return &result
}

func setTeamState(state *teamResourceModel, team *client.Team) {
	state.ID = types.StringValue(team.ID)
	state.WorkspaceID = types.StringValue(team.WorkspaceID)
	state.Name = types.StringValue(team.Name)
	state.Slug = types.StringValue(team.Slug)
	state.Description = nullableString(team.Description)
	state.Color = nullableString(team.Color)
	state.Icon = nullableString(team.Icon)
	state.CreatedBy = types.StringValue(team.CreatedBy)
	state.CreatedAt = nullableString(team.Created)
	state.UpdatedAt = nullableString(team.Updated)
}
