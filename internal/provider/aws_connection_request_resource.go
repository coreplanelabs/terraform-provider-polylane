package provider

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/coreplanelabs/terraform-provider-polylane/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	awsAccountIDPattern = regexp.MustCompile(`^\d{12}$`)
	awsRegionPattern    = regexp.MustCompile(`^[a-z]{2}(?:-gov)?-[a-z]+-\d$`)
)

type awsConnectionRequestResource struct {
	client *client.Client
}

type awsConnectionRequestResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	WorkspaceID          types.String `tfsdk:"workspace_id"`
	AccountID            types.String `tfsdk:"account_id"`
	Regions              types.Set    `tfsdk:"regions"`
	ExternalID           types.String `tfsdk:"external_id"`
	PrincipalARN         types.String `tfsdk:"principal_arn"`
	SubscriptionEndpoint types.String `tfsdk:"subscription_endpoint"`
	Region               types.String `tfsdk:"region"`
	Status               types.String `tfsdk:"status"`
}

var (
	_ resource.Resource              = &awsConnectionRequestResource{}
	_ resource.ResourceWithConfigure = &awsConnectionRequestResource{}
)

func NewAWSConnectionRequestResource() resource.Resource {
	return &awsConnectionRequestResource{}
}

func (r *awsConnectionRequestResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_aws_connection_request"
}

func (r *awsConnectionRequestResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	requiresReplace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		Description: "Requests the Polylane-issued identity needed to connect customer-managed AWS resources. This resource does not create or modify anything in AWS.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "Stable ID of the pending Polylane AWS connection request.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"workspace_id": schema.StringAttribute{
				Required:      true,
				Description:   "ID of the Polylane workspace that will own the connection.",
				Validators:    []validator.String{stringvalidator.RegexMatches(workspaceIDPattern, "must be a valid Polylane workspace ID")},
				PlanModifiers: requiresReplace,
			},
			"account_id": schema.StringAttribute{
				Required:      true,
				Description:   "Twelve-digit AWS account ID to connect.",
				Validators:    []validator.String{stringvalidator.RegexMatches(awsAccountIDPattern, "must be a twelve-digit AWS account ID")},
				PlanModifiers: requiresReplace,
			},
			"regions": schema.SetAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "AWS regions Polylane may read and synchronize.",
				Validators: []validator.Set{
					setvalidator.SizeBetween(1, 30),
					setvalidator.ValueStringsAre(stringvalidator.RegexMatches(awsRegionPattern, "must be a valid AWS region name")),
				},
				PlanModifiers: []planmodifier.Set{setplanmodifier.RequiresReplace()},
			},
			"external_id": schema.StringAttribute{
				Computed:      true,
				Description:   "Polylane-issued STS external ID. Reference this value in the customer-managed IAM role trust policy.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"principal_arn": schema.StringAttribute{
				Computed:      true,
				Description:   "AWS principal ARN to trust in the customer-managed IAM role.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"subscription_endpoint": schema.StringAttribute{
				Computed:      true,
				Description:   "Polylane HTTPS endpoint for the customer-managed SNS topic subscription.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"region": schema.StringAttribute{
				Computed:      true,
				Description:   "AWS region Polylane selected for the connection handshake.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Current lifecycle status of the connection request.",
			},
		},
	}
}

func (r *awsConnectionRequestResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *awsConnectionRequestResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Polylane Client", "The provider client was not configured. Please report this issue to the Polylane provider developers.")
		return
	}

	var plan awsConnectionRequestResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var regions []string
	resp.Diagnostics.Append(plan.Regions.ElementsAs(ctx, &regions, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	connectionRequest, err := r.client.CreateAWSConnectionRequest(ctx, client.CreateAWSConnectionRequestInput{
		WorkspaceID:    plan.WorkspaceID.ValueString(),
		AccountID:      plan.AccountID.ValueString(),
		Regions:        regions,
		IdempotencyKey: awsConnectionRequestIdempotencyKey(plan.WorkspaceID.ValueString(), plan.AccountID.ValueString(), regions),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create Polylane AWS Connection Request", err.Error())
		return
	}

	setAWSConnectionRequestState(ctx, &plan, connectionRequest, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func awsConnectionRequestIdempotencyKey(workspaceID, accountID string, regions []string) string {
	orderedRegions := slices.Clone(regions)
	slices.Sort(orderedRegions)
	digest := sha256.Sum256([]byte(strings.Join(append([]string{workspaceID, accountID}, orderedRegions...), "\x00")))
	return "terraform_" + hex.EncodeToString(digest[:16])
}

func (r *awsConnectionRequestResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Polylane Client", "The provider client was not configured. Please report this issue to the Polylane provider developers.")
		return
	}

	var state awsConnectionRequestResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	connectionRequest, err := r.client.GetAWSConnectionRequest(ctx, state.WorkspaceID.ValueString(), state.ID.ValueString())
	if client.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Polylane AWS Connection Request", err.Error())
		return
	}

	setAWSConnectionRequestState(ctx, &state, connectionRequest, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *awsConnectionRequestResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All configurable attributes are immutable handshake identity and require replacement.
	var state awsConnectionRequestResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *awsConnectionRequestResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Polylane Client", "The provider client was not configured. Please report this issue to the Polylane provider developers.")
		return
	}

	var state awsConnectionRequestResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteAWSConnectionRequest(ctx, state.WorkspaceID.ValueString(), state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to Delete Polylane AWS Connection Request", err.Error())
	}
}

func setAWSConnectionRequestState(ctx context.Context, state *awsConnectionRequestResourceModel, connectionRequest *client.AWSConnectionRequest, diagnostics *diag.Diagnostics) {
	state.ID = types.StringValue(connectionRequest.ID)
	state.WorkspaceID = types.StringValue(connectionRequest.WorkspaceID)
	state.AccountID = types.StringValue(connectionRequest.AccountID)
	state.ExternalID = types.StringValue(connectionRequest.ExternalID)
	state.PrincipalARN = types.StringValue(connectionRequest.PrincipalARN)
	state.SubscriptionEndpoint = types.StringValue(connectionRequest.SubscriptionEndpoint)
	state.Region = types.StringValue(connectionRequest.Region)
	state.Status = types.StringValue(connectionRequest.Status)

	regions, regionsDiagnostics := types.SetValueFrom(ctx, types.StringType, connectionRequest.Regions)
	diagnostics.Append(regionsDiagnostics...)
	if !regionsDiagnostics.HasError() {
		state.Regions = regions
	}
}
