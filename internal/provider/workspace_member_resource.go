package provider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/coreplanelabs/terraform-provider-polylane/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var subjectIDPattern = regexp.MustCompile(`^(system|usr_[0-9a-z]{32}|team_[0-9a-z]{24})$`)

type workspaceMemberResource struct {
	client *client.Client
}

type workspaceMemberResourceModel struct {
	WorkspaceID types.String `tfsdk:"workspace_id"`
	UserID      types.String `tfsdk:"user_id"`
	Email       types.String `tfsdk:"email"`
	Username    types.String `tfsdk:"username"`
	Forename    types.String `tfsdk:"forename"`
	Surname     types.String `tfsdk:"surname"`
	InvitedBy   types.String `tfsdk:"invited_by"`
	Position    types.String `tfsdk:"position"`
	Scopes      types.Set    `tfsdk:"scopes"`
	LastActive  types.String `tfsdk:"last_active"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

var (
	_ resource.Resource                = &workspaceMemberResource{}
	_ resource.ResourceWithConfigure   = &workspaceMemberResource{}
	_ resource.ResourceWithImportState = &workspaceMemberResource{}
)

func NewWorkspaceMemberResource() resource.Resource {
	return &workspaceMemberResource{}
}

func (r *workspaceMemberResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace_member"
}

func (r *workspaceMemberResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Adopts an existing Polylane workspace member so their role and permission scopes can be managed. Membership must be established in Polylane and imported; destroy only stops Terraform management.",
		Attributes: map[string]schema.Attribute{
			"workspace_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the workspace containing the member.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(workspaceIDPattern, "must be a valid Polylane workspace ID"),
				},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"user_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the workspace member.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(subjectIDPattern, "must be a valid Polylane user or team ID"),
				},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"position": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Workspace role. Terraform may assign admin, member, or read-only; ownership transfers remain a console operation.",
				Validators: []validator.String{
					stringvalidator.OneOf("admin", "member", "read-only"),
				},
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"scopes": schema.SetAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Explicit Polylane permission scopes assigned to the member.",
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"email":       schema.StringAttribute{Computed: true, Description: "Member email address."},
			"username":    schema.StringAttribute{Computed: true, Description: "Member username."},
			"forename":    schema.StringAttribute{Computed: true, Description: "Member given name."},
			"surname":     schema.StringAttribute{Computed: true, Description: "Member surname, if set."},
			"invited_by":  schema.StringAttribute{Computed: true, Description: "ID of the user who invited this member, if applicable."},
			"last_active": schema.StringAttribute{Computed: true, Description: "RFC 3339 timestamp of the member's latest activity, if known."},
			"created_at":  schema.StringAttribute{Computed: true, Description: "RFC 3339 timestamp when the membership was created."},
			"updated_at":  schema.StringAttribute{Computed: true, Description: "RFC 3339 timestamp when the membership was last updated."},
		},
	}
}

func (r *workspaceMemberResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *workspaceMemberResource) Create(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.AddError(
		"Workspace Member Must Be Imported",
		"The Polylane API does not create workspace membership directly. Invite or add the member in Polylane, then import it with `terraform import polylane_workspace_member.example <workspace-id>/<user-id>` or an import block.",
	)
}

func (r *workspaceMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Polylane Client", "The provider client was not configured. Please report this issue to the Polylane provider developers.")
		return
	}

	var state workspaceMemberResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	member, err := r.client.GetWorkspaceMember(ctx, state.WorkspaceID.ValueString(), state.UserID.ValueString())
	if client.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Polylane Workspace Member", err.Error())
		return
	}

	resp.Diagnostics.Append(setWorkspaceMemberState(ctx, &state, member)...)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	}
}

func (r *workspaceMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Polylane Client", "The provider client was not configured. Please report this issue to the Polylane provider developers.")
		return
	}

	var plan workspaceMemberResourceModel
	var state workspaceMemberResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := client.UpdateWorkspaceMemberInput{}
	if !plan.Position.Equal(state.Position) {
		position := plan.Position.ValueString()
		input.Position = &position
	}
	if !plan.Scopes.Equal(state.Scopes) {
		var scopes []string
		resp.Diagnostics.Append(plan.Scopes.ElementsAs(ctx, &scopes, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		input.Scopes = &scopes
	}

	member, err := r.client.UpdateWorkspaceMember(ctx, state.WorkspaceID.ValueString(), state.UserID.ValueString(), input)
	if client.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update Polylane Workspace Member", err.Error())
		return
	}

	resp.Diagnostics.Append(setWorkspaceMemberState(ctx, &plan, member)...)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}

func (r *workspaceMemberResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// There is no corresponding create operation, so this adoption resource
	// deliberately does not remove the member from the workspace.
}

func (r *workspaceMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := parseCompositeImportID(req.ID, 2)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Workspace Member Import ID", "Expected `<workspace-id>/<user-id>`. "+err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("workspace_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user_id"), parts[1])...)
}

func setWorkspaceMemberState(ctx context.Context, state *workspaceMemberResourceModel, member *client.WorkspaceMember) diag.Diagnostics {
	var diagnostics diag.Diagnostics
	state.WorkspaceID = types.StringValue(member.WorkspaceID)
	state.UserID = types.StringValue(member.UserID)
	state.Email = types.StringValue(member.Email)
	state.Username = types.StringValue(member.Username)
	state.Forename = types.StringValue(member.Forename)
	state.Surname = nullableString(member.Surname)
	state.InvitedBy = nullableString(member.InvitedBy)
	state.Position = types.StringValue(member.Position)
	state.LastActive = nullableString(member.LastActive)
	state.CreatedAt = nullableString(member.Created)
	state.UpdatedAt = nullableString(member.Updated)
	state.Scopes, diagnostics = types.SetValueFrom(ctx, types.StringType, member.Scopes)
	return diagnostics
}
