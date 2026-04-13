package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/syegournov/xkeen-gen/terraform-provider-xui/internal/xui"
)

var _ resource.Resource = (*xrayTemplateResource)(nil)

type xrayTemplateResource struct {
	client *xui.Client
}

func NewXrayTemplateResource() resource.Resource {
	return &xrayTemplateResource{}
}

func (r *xrayTemplateResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "xui_xray_template"
}

func (r *xrayTemplateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages full Xray template config JSON (`/panel/xray/update`). This is intentionally unopinionated: provide the full template JSON body and optionally restart Xray after apply.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Static resource id (`xray-template`).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"json": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Full Xray template JSON body sent to 3x-ui `POST /panel/xray/update` as `xraySetting`.",
			},
			"restart_xray": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "If true, call `POST /panel/api/server/restartXrayService` after updating template.",
			},
		},
	}
}

func (r *xrayTemplateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	cli, ok := req.ProviderData.(*xui.Client)
	if !ok {
		resp.Diagnostics.AddError("Internal error", "invalid provider data type")
		return
	}
	r.client = cli
}

type xrayTemplateModel struct {
	ID          types.String `tfsdk:"id"`
	JSON        types.String `tfsdk:"json"`
	RestartXray types.Bool   `tfsdk:"restart_xray"`
}

func (r *xrayTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan xrayTemplateModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := validateJSONString(plan.JSON.ValueString(), "json"); err != nil {
		resp.Diagnostics.AddError("Invalid json", err.Error())
		return
	}
	if err := r.client.UpdateXrayTemplate(plan.JSON.ValueString()); err != nil {
		resp.Diagnostics.AddError("API error", err.Error())
		return
	}
	if plan.RestartXray.ValueBool() {
		if err := r.client.RestartXrayService(); err != nil {
			resp.Diagnostics.AddError("API error", err.Error())
			return
		}
	}
	plan.ID = types.StringValue("xray-template")
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *xrayTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state xrayTemplateModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	raw, err := r.client.GetXrayTemplate()
	if err != nil {
		resp.Diagnostics.AddError("API error", err.Error())
		return
	}
	state.JSON = types.StringValue(raw)
	if state.ID.IsNull() || state.ID.ValueString() == "" {
		state.ID = types.StringValue("xray-template")
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *xrayTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan xrayTemplateModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := validateJSONString(plan.JSON.ValueString(), "json"); err != nil {
		resp.Diagnostics.AddError("Invalid json", err.Error())
		return
	}
	if err := r.client.UpdateXrayTemplate(plan.JSON.ValueString()); err != nil {
		resp.Diagnostics.AddError("API error", err.Error())
		return
	}
	if plan.RestartXray.ValueBool() {
		if err := r.client.RestartXrayService(); err != nil {
			resp.Diagnostics.AddError("API error", err.Error())
			return
		}
	}
	plan.ID = types.StringValue("xray-template")
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *xrayTemplateResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// No delete endpoint on 3x-ui side; Terraform state removal is sufficient.
}
