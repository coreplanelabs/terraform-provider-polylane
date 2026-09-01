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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var workspaceIDPattern = regexp.MustCompile(`^ws_[0-9a-z]{32}$`)

type workspaceResource struct {
	client *client.Client
}

type workspaceResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Slug        types.String `tfsdk:"slug"`
	Description types.String `tfsdk:"description"`
	Domain      types.String `tfsdk:"domain"`
	AutoJoin    types.Bool   `tfsdk:"auto_join"`
	OwnerID     types.String `tfsdk:"owner_id"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

var (
	_ resource.Resource                = &workspaceResource{}
	_ resource.ResourceWithConfigure   = &workspaceResource{}
	_ resource.ResourceWithImportState = &workspaceResource{}
)

func NewWorkspaceResource() resource.Resource {
	return &workspaceResource{}
}

func (r *workspaceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace"
}

func (r *workspaceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Adopts an existing Polylane workspace so its safe, non-secret properties can be managed with Terraform. Workspaces must be imported; this resource does not create or delete them.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique Polylane workspace ID.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(workspaceIDPattern, "must be a valid Polylane workspace ID"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Human-readable workspace name.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(4, 128),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"slug": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "URL-friendly workspace slug.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(4, 1024),
					stringvalidator.RegexMatches(regexp.MustCompile(`^[a-z0-9_]+(?:-[a-z0-9_]+)*$`), "must contain lowercase letters, numbers, underscores, and single hyphen separators only"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Workspace description.",
				Validators: []validator.String{
					stringvalidator.LengthAtMost(2000),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"domain": schema.StringAttribute{
				Computed:    true,
				Description: "Email domain claimed by the workspace, if any.",
			},
			"auto_join": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether verified users with the workspace email domain can join automatically.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"owner_id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the workspace owner.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "RFC 3339 timestamp when the workspace was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "RFC 3339 timestamp when the workspace was last updated.",
			},
		},
	}
}

func (r *workspaceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	apiClient, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data Type",
			fmt.Sprintf("Expected *client.Client, got %T. Please report this issue to the Polylane provider developers.", req.ProviderData),
		)
		return
	}

	r.client = apiClient
}

func (r *workspaceResource) Create(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.AddError(
		"Workspace Must Be Imported",
		"Polylane API keys are workspace-bound and cannot create a workspace. Create the workspace in Polylane, then import it with `terraform import polylane_workspace.example <workspace-id>` or an import block.",
	)
}

func (r *workspaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Polylane Client", "The provider client was not configured. Please report this issue to the Polylane provider developers.")
		return
	}

	var state workspaceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	workspace, err := r.client.GetWorkspace(ctx, state.ID.ValueString())
	if client.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Polylane Workspace", err.Error())
		return
	}

	setWorkspaceState(&state, workspace)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *workspaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Polylane Client", "The provider client was not configured. Please report this issue to the Polylane provider developers.")
		return
	}

	var plan workspaceResourceModel
	var state workspaceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := client.UpdateWorkspaceInput{}
	changed := false
	if !plan.Name.Equal(state.Name) {
		value := plan.Name.ValueString()
		input.Name = &value
		changed = true
	}
	if !plan.Slug.Equal(state.Slug) {
		value := plan.Slug.ValueString()
		input.Slug = &value
		changed = true
	}
	if !plan.Description.Equal(state.Description) && !plan.Description.IsNull() {
		value := plan.Description.ValueString()
		input.Description = &value
		changed = true
	}
	if !plan.AutoJoin.Equal(state.AutoJoin) {
		value := plan.AutoJoin.ValueBool()
		input.AutoJoin = &value
		changed = true
	}

	var workspace *client.Workspace
	var err error
	if changed {
		workspace, err = r.client.UpdateWorkspace(ctx, state.ID.ValueString(), input)
	} else {
		workspace, err = r.client.GetWorkspace(ctx, state.ID.ValueString())
	}
	if client.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update Polylane Workspace", err.Error())
		return
	}

	setWorkspaceState(&plan, workspace)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *workspaceResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// This adoption-oriented resource intentionally leaves the workspace in
	// Polylane. Removing it from configuration only stops Terraform management.
}

func (r *workspaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func setWorkspaceState(state *workspaceResourceModel, workspace *client.Workspace) {
	state.ID = types.StringValue(workspace.ID)
	state.Name = types.StringValue(workspace.Name)
	state.Slug = types.StringValue(workspace.Slug)
	state.Description = nullableString(workspace.Description)
	state.Domain = nullableString(workspace.Domain)
	state.AutoJoin = types.BoolValue(workspace.AutoJoinEnabled != nil)
	state.OwnerID = types.StringValue(workspace.OwnerID)
	state.CreatedAt = nullableString(workspace.Created)
	state.UpdatedAt = nullableString(workspace.Updated)
}

func nullableString(value *string) types.String {
	if value == nil {
		return types.StringNull()
	}
	return types.StringValue(*value)
}
