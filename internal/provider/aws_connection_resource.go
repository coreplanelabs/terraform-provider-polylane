package provider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/coreplanelabs/terraform-provider-polylane/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	awsRoleARNPattern         = regexp.MustCompile(`^arn:(?:aws|aws-us-gov|aws-cn):iam::\d{12}:role/.+$`)
	awsTopicARNPattern        = regexp.MustCompile(`^arn:(?:aws|aws-us-gov|aws-cn):sns:[a-z0-9-]+:\d{12}:[A-Za-z0-9_-]+$`)
	awsSubscriptionARNPattern = regexp.MustCompile(`^arn:(?:aws|aws-us-gov|aws-cn):sns:[a-z0-9-]+:\d{12}:[A-Za-z0-9_-]+:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	awsBucketNamePattern      = regexp.MustCompile(`^[a-z0-9][a-z0-9.-]*[a-z0-9]$`)
	awsCloudTrailNamePattern  = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)
)

type awsConnectionResource struct {
	client *client.Client
}

type awsConnectionResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	WorkspaceID          types.String `tfsdk:"workspace_id"`
	RequestID            types.String `tfsdk:"request_id"`
	AccountID            types.String `tfsdk:"account_id"`
	Regions              types.Set    `tfsdk:"regions"`
	Region               types.String `tfsdk:"region"`
	RoleARN              types.String `tfsdk:"role_arn"`
	BucketName           types.String `tfsdk:"bucket_name"`
	TopicARN             types.String `tfsdk:"topic_arn"`
	TopicSubscriptionARN types.String `tfsdk:"topic_subscription_arn"`
	CloudTrailName       types.String `tfsdk:"cloudtrail_name"`
	Status               types.String `tfsdk:"status"`
}

var (
	_ resource.Resource              = &awsConnectionResource{}
	_ resource.ResourceWithConfigure = &awsConnectionResource{}
)

func NewAWSConnectionResource() resource.Resource {
	return &awsConnectionResource{}
}

func (r *awsConnectionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_aws_connection"
}

func (r *awsConnectionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	requiresReplace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		Description: "Activates a Polylane connection using AWS resources managed by the customer. Polylane validates and reads these resources but does not manage their AWS lifecycle.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "Unique Polylane cloud-account ID.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"workspace_id": schema.StringAttribute{
				Required:      true,
				Description:   "ID of the Polylane workspace that owns the connection.",
				Validators:    []validator.String{stringvalidator.RegexMatches(workspaceIDPattern, "must be a valid Polylane workspace ID")},
				PlanModifiers: requiresReplace,
			},
			"request_id": schema.StringAttribute{
				Required:      true,
				Description:   "ID returned by polylane_aws_connection_request.",
				Validators:    []validator.String{stringvalidator.LengthAtLeast(1)},
				PlanModifiers: requiresReplace,
			},
			"region": schema.StringAttribute{
				Required:      true,
				Description:   "AWS handshake region returned by polylane_aws_connection_request.",
				Validators:    []validator.String{stringvalidator.RegexMatches(awsRegionPattern, "must be a valid AWS region name")},
				PlanModifiers: requiresReplace,
			},
			"role_arn": schema.StringAttribute{
				Required:      true,
				Description:   "ARN of the customer-managed IAM role Polylane may assume.",
				Validators:    []validator.String{stringvalidator.RegexMatches(awsRoleARNPattern, "must be an IAM role ARN")},
				PlanModifiers: requiresReplace,
			},
			"bucket_name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the customer-managed S3 bucket receiving CloudTrail objects.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 63),
					stringvalidator.RegexMatches(awsBucketNamePattern, "must be a valid S3 bucket name"),
				},
				PlanModifiers: requiresReplace,
			},
			"topic_arn": schema.StringAttribute{
				Required:      true,
				Description:   "ARN of the customer-managed SNS topic receiving CloudTrail notifications.",
				Validators:    []validator.String{stringvalidator.RegexMatches(awsTopicARNPattern, "must be an SNS topic ARN")},
				PlanModifiers: requiresReplace,
			},
			"topic_subscription_arn": schema.StringAttribute{
				Required:      true,
				Description:   "ARN of the customer-managed, confirmed SNS HTTPS subscription that delivers notifications to Polylane.",
				Validators:    []validator.String{stringvalidator.RegexMatches(awsSubscriptionARNPattern, "must be a confirmed SNS subscription ARN")},
				PlanModifiers: requiresReplace,
			},
			"cloudtrail_name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the customer-managed CloudTrail trail.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 128),
					stringvalidator.RegexMatches(awsCloudTrailNamePattern, "must contain only letters, numbers, periods, underscores, and hyphens"),
				},
				PlanModifiers: requiresReplace,
			},
			"account_id": schema.StringAttribute{
				Computed:    true,
				Description: "Twelve-digit AWS account ID registered with Polylane.",
			},
			"regions": schema.SetAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "AWS regions Polylane may read and synchronize.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Polylane registration status. Registered means the connection is durable; initial synchronization may continue asynchronously.",
			},
		},
	}
}

func (r *awsConnectionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *awsConnectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Polylane Client", "The provider client was not configured. Please report this issue to the Polylane provider developers.")
		return
	}

	var plan awsConnectionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	connection, err := r.client.ActivateAWSConnection(ctx, client.ActivateAWSConnectionInput{
		WorkspaceID:          plan.WorkspaceID.ValueString(),
		RequestID:            plan.RequestID.ValueString(),
		Region:               plan.Region.ValueString(),
		RoleARN:              plan.RoleARN.ValueString(),
		BucketName:           plan.BucketName.ValueString(),
		TopicARN:             plan.TopicARN.ValueString(),
		TopicSubscriptionARN: plan.TopicSubscriptionARN.ValueString(),
		CloudTrailName:       plan.CloudTrailName.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to Activate Polylane AWS Connection", err.Error())
		return
	}

	setAWSConnectionState(ctx, &plan, connection, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *awsConnectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Polylane Client", "The provider client was not configured. Please report this issue to the Polylane provider developers.")
		return
	}

	var state awsConnectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	connection, err := r.client.GetAWSConnection(ctx, state.WorkspaceID.ValueString(), state.ID.ValueString(), state.RequestID.ValueString())
	if client.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Polylane AWS Connection", err.Error())
		return
	}

	setAWSConnectionState(ctx, &state, connection, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *awsConnectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Customer-owned AWS identifiers are immutable connection identity and require replacement.
	var state awsConnectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *awsConnectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Polylane Client", "The provider client was not configured. Please report this issue to the Polylane provider developers.")
		return
	}

	var state awsConnectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteAWSConnection(ctx, state.WorkspaceID.ValueString(), state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to Disconnect Polylane AWS Connection", err.Error())
	}
}

func setAWSConnectionState(ctx context.Context, state *awsConnectionResourceModel, connection *client.AWSConnection, diagnostics *diag.Diagnostics) {
	state.ID = types.StringValue(connection.ID)
	state.WorkspaceID = types.StringValue(connection.WorkspaceID)
	if connection.RequestID != "" {
		state.RequestID = types.StringValue(connection.RequestID)
	}
	state.AccountID = types.StringValue(connection.AccountID)
	state.Region = types.StringValue(connection.Region)
	if connection.RoleARN != "" {
		state.RoleARN = types.StringValue(connection.RoleARN)
	}
	if connection.BucketName != "" {
		state.BucketName = types.StringValue(connection.BucketName)
	}
	if connection.TopicARN != "" {
		state.TopicARN = types.StringValue(connection.TopicARN)
	}
	if connection.TopicSubscriptionARN != "" {
		state.TopicSubscriptionARN = types.StringValue(connection.TopicSubscriptionARN)
	}
	if connection.CloudTrailName != "" {
		state.CloudTrailName = types.StringValue(connection.CloudTrailName)
	}
	state.Status = types.StringValue(connection.Status)

	regions, regionsDiagnostics := types.SetValueFrom(ctx, types.StringType, connection.Regions)
	diagnostics.Append(regionsDiagnostics...)
	if !regionsDiagnostics.HasError() {
		state.Regions = regions
	}
}
