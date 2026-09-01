package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/coreplanelabs/terraform-provider-polylane/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type workspaceSettingsModel interface {
	getWorkspaceID() types.String
}

type workspaceSettingsDescriptor struct {
	typeSuffix      string
	settingsPath    string
	description     string
	attributes      map[string]schema.Attribute
	newModel        func() workspaceSettingsModel
	newResponse     func() any
	patchBody       func(workspaceSettingsModel) (any, error)
	setFromResponse func(workspaceSettingsModel, any) error
}

type workspaceSettingsResource struct {
	client     *client.Client
	descriptor workspaceSettingsDescriptor
}

var (
	_ resource.Resource                = &workspaceSettingsResource{}
	_ resource.ResourceWithConfigure   = &workspaceSettingsResource{}
	_ resource.ResourceWithImportState = &workspaceSettingsResource{}
)

func (r *workspaceSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace_" + r.descriptor.typeSuffix
}

func (r *workspaceSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: r.descriptor.description + " This is an existing workspace singleton: create and update use PATCH, while destroy only stops Terraform management and does not reset the remote settings.",
		Attributes:  r.descriptor.attributes,
	}
}

func (r *workspaceSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *workspaceSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	model := r.descriptor.newModel()
	resp.Diagnostics.Append(req.Plan.Get(ctx, model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.apply(ctx, model); err != nil {
		resp.Diagnostics.AddError("Unable to Update Polylane Workspace Settings", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (r *workspaceSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	model := r.descriptor.newModel()
	resp.Diagnostics.Append(req.State.Get(ctx, model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.read(ctx, model); client.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError("Unable to Read Polylane Workspace Settings", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (r *workspaceSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	model := r.descriptor.newModel()
	resp.Diagnostics.Append(req.Plan.Get(ctx, model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.apply(ctx, model); client.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError("Unable to Update Polylane Workspace Settings", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (r *workspaceSettingsResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Polylane settings endpoints have no delete/reset operation. Removing this
	// resource from configuration intentionally leaves the settings unchanged.
}

func (r *workspaceSettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("workspace_id"), req, resp)
}

func (r *workspaceSettingsResource) apply(ctx context.Context, model workspaceSettingsModel) error {
	if r.client == nil {
		return fmt.Errorf("provider client was not configured; please report this issue to the Polylane provider developers")
	}

	response := r.descriptor.newResponse()
	requestBody, err := r.descriptor.patchBody(model)
	if err != nil {
		return err
	}
	if err := r.client.UpdateWorkspaceSettings(
		ctx,
		model.getWorkspaceID().ValueString(),
		r.descriptor.settingsPath,
		requestBody,
		response,
	); err != nil {
		return err
	}

	return r.descriptor.setFromResponse(model, response)
}

func (r *workspaceSettingsResource) read(ctx context.Context, model workspaceSettingsModel) error {
	if r.client == nil {
		return fmt.Errorf("provider client was not configured; please report this issue to the Polylane provider developers")
	}

	response := r.descriptor.newResponse()
	if err := r.client.GetWorkspaceSettings(ctx, model.getWorkspaceID().ValueString(), r.descriptor.settingsPath, response); err != nil {
		return err
	}

	return r.descriptor.setFromResponse(model, response)
}

func workspaceIDSettingsAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		Required:    true,
		Description: "ID of the workspace whose settings are managed.",
		Validators: []validator.String{
			stringvalidator.RegexMatches(workspaceIDPattern, "must be a valid Polylane workspace ID"),
		},
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
}

func disabledAt(enabled bool) any {
	if enabled {
		return nil
	}
	return time.Now().UTC().Format(time.RFC3339Nano)
}

func terraformString(value types.String) any {
	if value.IsNull() {
		return nil
	}
	return value.ValueString()
}

func terraformInt64(value types.Int64) any {
	if value.IsNull() {
		return nil
	}
	return value.ValueInt64()
}

func terraformBool(value types.Bool) any {
	if value.IsNull() {
		return nil
	}
	return value.ValueBool()
}

func nullableInt64(value *int64) types.Int64 {
	if value == nil {
		return types.Int64Null()
	}
	return types.Int64Value(*value)
}

func nullableBool(value *bool) types.Bool {
	if value == nil {
		return types.BoolNull()
	}
	return types.BoolValue(*value)
}

func settingsPointer[T any](value any) (*T, error) {
	pointer, ok := value.(*T)
	if !ok {
		return nil, fmt.Errorf("unexpected settings value type %T", value)
	}
	return pointer, nil
}

type autofixSettingsModel struct {
	WorkspaceID    types.String `tfsdk:"workspace_id"`
	Enabled        types.Bool   `tfsdk:"enabled"`
	TriggerMode    types.String `tfsdk:"trigger_mode"`
	ReviewBotLogin types.String `tfsdk:"review_bot_login"`
}

func (m *autofixSettingsModel) getWorkspaceID() types.String { return m.WorkspaceID }

func NewWorkspaceAutofixSettingsResource() resource.Resource {
	return &workspaceSettingsResource{descriptor: workspaceSettingsDescriptor{
		typeSuffix:   "autofix_settings",
		settingsPath: "autofix_settings",
		description:  "Manages workspace-wide autofix behavior.",
		attributes: map[string]schema.Attribute{
			"workspace_id": workspaceIDSettingsAttribute(),
			"enabled": schema.BoolAttribute{
				Required:    true,
				Description: "Whether Polylane may open new autofix pull requests for repositories in this workspace.",
			},
			"trigger_mode": schema.StringAttribute{
				Required:    true,
				Description: "How machine-initiated autofixes start. `auto` starts immediately; `approval` waits for approval in the console.",
				Validators: []validator.String{
					stringvalidator.OneOf("auto", "approval"),
				},
			},
			"review_bot_login": schema.StringAttribute{
				Optional:    true,
				Description: "GitHub login whose review is requested on each autofix pull request. Omit to request no bot review.",
			},
		},
		newModel:    func() workspaceSettingsModel { return &autofixSettingsModel{} },
		newResponse: func() any { return &client.AutofixSettings{} },
		patchBody: func(raw workspaceSettingsModel) (any, error) {
			model, err := settingsPointer[autofixSettingsModel](raw)
			if err != nil {
				return nil, err
			}
			return map[string]any{
				"disabledAt":     disabledAt(model.Enabled.ValueBool()),
				"triggerMode":    model.TriggerMode.ValueString(),
				"reviewBotLogin": terraformString(model.ReviewBotLogin),
			}, nil
		},
		setFromResponse: func(rawModel workspaceSettingsModel, rawResponse any) error {
			model, err := settingsPointer[autofixSettingsModel](rawModel)
			if err != nil {
				return err
			}
			response, err := settingsPointer[client.AutofixSettings](rawResponse)
			if err != nil {
				return err
			}
			model.Enabled = types.BoolValue(response.Autofix.DisabledAt == nil)
			model.TriggerMode = types.StringValue(response.Autofix.TriggerMode)
			model.ReviewBotLogin = nullableString(response.Autofix.ReviewBotLogin)
			return nil
		},
	}}
}

type digestSettingsModel struct {
	WorkspaceID        types.String `tfsdk:"workspace_id"`
	Enabled            types.Bool   `tfsdk:"enabled"`
	DailyTrendsEnabled types.Bool   `tfsdk:"daily_trends_enabled"`
}

func (m *digestSettingsModel) getWorkspaceID() types.String { return m.WorkspaceID }

func NewWorkspaceDigestSettingsResource() resource.Resource {
	return &workspaceSettingsResource{descriptor: workspaceSettingsDescriptor{
		typeSuffix:   "digest_settings",
		settingsPath: "digest_settings",
		description:  "Manages workspace digest email settings.",
		attributes: map[string]schema.Attribute{
			"workspace_id": workspaceIDSettingsAttribute(),
			"enabled": schema.BoolAttribute{
				Required:    true,
				Description: "Whether the weekly workspace digest is sent.",
			},
			"daily_trends_enabled": schema.BoolAttribute{
				Required:    true,
				Description: "Whether the daily log trends email is sent.",
			},
		},
		newModel:    func() workspaceSettingsModel { return &digestSettingsModel{} },
		newResponse: func() any { return &client.DigestSettings{} },
		patchBody: func(raw workspaceSettingsModel) (any, error) {
			model, err := settingsPointer[digestSettingsModel](raw)
			if err != nil {
				return nil, err
			}
			return map[string]any{
				"enabled":            model.Enabled.ValueBool(),
				"dailyTrendsEnabled": model.DailyTrendsEnabled.ValueBool(),
			}, nil
		},
		setFromResponse: func(rawModel workspaceSettingsModel, rawResponse any) error {
			model, err := settingsPointer[digestSettingsModel](rawModel)
			if err != nil {
				return err
			}
			response, err := settingsPointer[client.DigestSettings](rawResponse)
			if err != nil {
				return err
			}
			model.Enabled = types.BoolValue(response.WeeklyDigestDisabledAt == nil)
			model.DailyTrendsEnabled = types.BoolValue(response.DailyTrendsDisabledAt == nil)
			return nil
		},
	}}
}

type investigationsSettingsModel struct {
	WorkspaceID         types.String `tfsdk:"workspace_id"`
	PassesPerHypothesis types.Int64  `tfsdk:"passes_per_hypothesis"`
}

func (m *investigationsSettingsModel) getWorkspaceID() types.String { return m.WorkspaceID }

func NewWorkspaceInvestigationsSettingsResource() resource.Resource {
	return &workspaceSettingsResource{descriptor: workspaceSettingsDescriptor{
		typeSuffix:   "investigations_settings",
		settingsPath: "investigations_settings",
		description:  "Manages workspace investigation execution settings.",
		attributes: map[string]schema.Attribute{
			"workspace_id": workspaceIDSettingsAttribute(),
			"passes_per_hypothesis": schema.Int64Attribute{
				Required:    true,
				Description: "Number of parallel analysis passes run for each hypothesis.",
				Validators: []validator.Int64{
					int64validator.Between(1, 12),
				},
			},
		},
		newModel:    func() workspaceSettingsModel { return &investigationsSettingsModel{} },
		newResponse: func() any { return &client.InvestigationsSettings{} },
		patchBody: func(raw workspaceSettingsModel) (any, error) {
			model, err := settingsPointer[investigationsSettingsModel](raw)
			if err != nil {
				return nil, err
			}
			return map[string]any{
				"investigations": map[string]any{
					"passesPerHypothesis": model.PassesPerHypothesis.ValueInt64(),
				},
			}, nil
		},
		setFromResponse: func(rawModel workspaceSettingsModel, rawResponse any) error {
			model, err := settingsPointer[investigationsSettingsModel](rawModel)
			if err != nil {
				return err
			}
			response, err := settingsPointer[client.InvestigationsSettings](rawResponse)
			if err != nil {
				return err
			}
			model.PassesPerHypothesis = types.Int64Value(response.Investigations.PassesPerHypothesis)
			return nil
		},
	}}
}

type investigationLimitsSettingsModel struct {
	WorkspaceID                types.String `tfsdk:"workspace_id"`
	Per24h                     types.Int64  `tfsdk:"per_24h"`
	MinAutoInvestigateSeverity types.String `tfsdk:"minimum_auto_investigate_severity"`
}

func (m *investigationLimitsSettingsModel) getWorkspaceID() types.String { return m.WorkspaceID }

func NewWorkspaceInvestigationLimitsSettingsResource() resource.Resource {
	return &workspaceSettingsResource{descriptor: workspaceSettingsDescriptor{
		typeSuffix:   "investigation_limits_settings",
		settingsPath: "investigation_limits_settings",
		description:  "Manages workspace-wide investigation limits and automatic-investigation thresholds.",
		attributes: map[string]schema.Attribute{
			"workspace_id": workspaceIDSettingsAttribute(),
			"per_24h": schema.Int64Attribute{
				Optional:    true,
				Description: "Maximum investigations started in a rolling 24-hour window. Omit for unlimited.",
				Validators: []validator.Int64{
					int64validator.Between(1, 1000),
				},
			},
			"minimum_auto_investigate_severity": schema.StringAttribute{
				Required:    true,
				Description: "Lowest issue severity that starts an investigation automatically.",
				Validators: []validator.String{
					stringvalidator.OneOf("critical", "high", "medium", "low", "info"),
				},
			},
		},
		newModel:    func() workspaceSettingsModel { return &investigationLimitsSettingsModel{} },
		newResponse: func() any { return &client.InvestigationLimitsSettings{} },
		patchBody: func(raw workspaceSettingsModel) (any, error) {
			model, err := settingsPointer[investigationLimitsSettingsModel](raw)
			if err != nil {
				return nil, err
			}
			return map[string]any{
				"per24h":                     terraformInt64(model.Per24h),
				"minAutoInvestigateSeverity": model.MinAutoInvestigateSeverity.ValueString(),
			}, nil
		},
		setFromResponse: func(rawModel workspaceSettingsModel, rawResponse any) error {
			model, err := settingsPointer[investigationLimitsSettingsModel](rawModel)
			if err != nil {
				return err
			}
			response, err := settingsPointer[client.InvestigationLimitsSettings](rawResponse)
			if err != nil {
				return err
			}
			model.Per24h = nullableInt64(response.InvestigationLimits.Per24h)
			model.MinAutoInvestigateSeverity = types.StringValue(response.InvestigationLimits.MinAutoInvestigateSeverity)
			return nil
		},
	}}
}

type modelTrainingSettingsModel struct {
	WorkspaceID        types.String `tfsdk:"workspace_id"`
	AllowModelTraining types.Bool   `tfsdk:"allow_model_training"`
	Effective          types.Bool   `tfsdk:"effective"`
}

func (m *modelTrainingSettingsModel) getWorkspaceID() types.String { return m.WorkspaceID }

func NewWorkspaceModelTrainingSettingsResource() resource.Resource {
	return &workspaceSettingsResource{descriptor: workspaceSettingsDescriptor{
		typeSuffix:   "model_training_settings",
		settingsPath: "model_training_settings",
		description:  "Manages whether workspace data may be used for model training.",
		attributes: map[string]schema.Attribute{
			"workspace_id": workspaceIDSettingsAttribute(),
			"allow_model_training": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether workspace data may be used to improve Polylane models. Omit to use the workspace plan default.",
			},
			"effective": schema.BoolAttribute{
				Computed:    true,
				Description: "Resolved setting after applying the explicit value or workspace plan default.",
			},
		},
		newModel:    func() workspaceSettingsModel { return &modelTrainingSettingsModel{} },
		newResponse: func() any { return &client.ModelTrainingSettings{} },
		patchBody: func(raw workspaceSettingsModel) (any, error) {
			model, err := settingsPointer[modelTrainingSettingsModel](raw)
			if err != nil {
				return nil, err
			}
			return map[string]any{"allowModelTraining": terraformBool(model.AllowModelTraining)}, nil
		},
		setFromResponse: func(rawModel workspaceSettingsModel, rawResponse any) error {
			model, err := settingsPointer[modelTrainingSettingsModel](rawModel)
			if err != nil {
				return err
			}
			response, err := settingsPointer[client.ModelTrainingSettings](rawResponse)
			if err != nil {
				return err
			}
			model.AllowModelTraining = nullableBool(response.ModelTraining.AllowModelTraining)
			model.Effective = types.BoolValue(response.ModelTraining.Effective)
			return nil
		},
	}}
}

type observabilitySettingsModel struct {
	WorkspaceID types.String `tfsdk:"workspace_id"`
	Enabled     types.Bool   `tfsdk:"enabled"`
}

func (m *observabilitySettingsModel) getWorkspaceID() types.String { return m.WorkspaceID }

func NewWorkspaceObservabilitySettingsResource() resource.Resource {
	return &workspaceSettingsResource{descriptor: workspaceSettingsDescriptor{
		typeSuffix:   "observability_settings",
		settingsPath: "observability_settings",
		description:  "Manages Polylane native observability for a workspace.",
		attributes: map[string]schema.Attribute{
			"workspace_id": workspaceIDSettingsAttribute(),
			"enabled": schema.BoolAttribute{
				Required:    true,
				Description: "Whether Polylane native observability is enabled.",
			},
		},
		newModel:    func() workspaceSettingsModel { return &observabilitySettingsModel{} },
		newResponse: func() any { return &client.ObservabilitySettings{} },
		patchBody: func(raw workspaceSettingsModel) (any, error) {
			model, err := settingsPointer[observabilitySettingsModel](raw)
			if err != nil {
				return nil, err
			}
			return map[string]any{
				"nativeObservability": map[string]any{"enabled": model.Enabled.ValueBool()},
			}, nil
		},
		setFromResponse: func(rawModel workspaceSettingsModel, rawResponse any) error {
			model, err := settingsPointer[observabilitySettingsModel](rawModel)
			if err != nil {
				return err
			}
			response, err := settingsPointer[client.ObservabilitySettings](rawResponse)
			if err != nil {
				return err
			}
			model.Enabled = types.BoolValue(response.NativeObservability.Enabled)
			return nil
		},
	}}
}

type prReviewSettingsModel struct {
	WorkspaceID       types.String `tfsdk:"workspace_id"`
	Enabled           types.Bool   `tfsdk:"enabled"`
	GuestLinksEnabled types.Bool   `tfsdk:"guest_links_enabled"`
}

func (m *prReviewSettingsModel) getWorkspaceID() types.String { return m.WorkspaceID }

func NewWorkspacePRReviewSettingsResource() resource.Resource {
	return &workspaceSettingsResource{descriptor: workspaceSettingsDescriptor{
		typeSuffix:   "pr_review_settings",
		settingsPath: "pr_review_settings",
		description:  "Manages workspace-wide pull request review behavior.",
		attributes: map[string]schema.Attribute{
			"workspace_id": workspaceIDSettingsAttribute(),
			"enabled": schema.BoolAttribute{
				Required:    true,
				Description: "Whether pull requests are reviewed for production impact.",
			},
			"guest_links_enabled": schema.BoolAttribute{
				Required:    true,
				Description: "Whether new review threads may use signed guest links in pull request comments.",
			},
		},
		newModel:    func() workspaceSettingsModel { return &prReviewSettingsModel{} },
		newResponse: func() any { return &client.PRReviewSettings{} },
		patchBody: func(raw workspaceSettingsModel) (any, error) {
			model, err := settingsPointer[prReviewSettingsModel](raw)
			if err != nil {
				return nil, err
			}
			return map[string]any{
				"disabledAt":           disabledAt(model.Enabled.ValueBool()),
				"guestLinksDisabledAt": disabledAt(model.GuestLinksEnabled.ValueBool()),
			}, nil
		},
		setFromResponse: func(rawModel workspaceSettingsModel, rawResponse any) error {
			model, err := settingsPointer[prReviewSettingsModel](rawModel)
			if err != nil {
				return err
			}
			response, err := settingsPointer[client.PRReviewSettings](rawResponse)
			if err != nil {
				return err
			}
			model.Enabled = types.BoolValue(response.PRReviews.DisabledAt == nil)
			model.GuestLinksEnabled = types.BoolValue(response.PRReviews.GuestLinksDisabledAt == nil)
			return nil
		},
	}}
}
