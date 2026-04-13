package provider

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/syegournov/xkeen-gen/terraform-provider-xui/internal/xui"
)

var _ resource.Resource = (*vlessClientResource)(nil)
var _ resource.ResourceWithImportState = (*vlessClientResource)(nil)

type vlessClientResource struct {
	client *xui.Client
}

func NewVLESSClientResource() resource.Resource {
	return &vlessClientResource{}
}

func (r *vlessClientResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "xui_vless_client"
}

func (r *vlessClientResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "VLESS user (client) on an existing 3x-ui inbound. Uses `/panel/api/inbounds/addClient` and related routes.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Client UUID (`id` in Xray VLESS settings).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"inbound_id": schema.Int64Attribute{
				MarkdownDescription: "Panel inbound id (number from URL / API).",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "Unique client email / label in the panel.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"uuid": schema.StringAttribute{
				MarkdownDescription: "Static VLESS UUID. If empty, one is generated on create.",
				Optional:            true,
				Computed:            true,
			},
			"flow": schema.StringAttribute{
				MarkdownDescription: "e.g. `xtls-rprx-vision` for XTLS Vision.",
				Optional:            true,
			},
			"enable": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"limit_ip": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
			},
			"total_gb": schema.Int64Attribute{
				MarkdownDescription: "Traffic limit in **bytes** (panel field `totalGB`; 0 = unlimited).",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
			},
			"expiry_time": schema.Int64Attribute{
				MarkdownDescription: "Expiry in milliseconds since Unix epoch (0 = never).",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
			},
			"tg_id": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
			},
			"sub_id": schema.StringAttribute{
				Optional: true,
			},
			"comment": schema.StringAttribute{
				Optional: true,
			},
			"reset": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
			},
		},
	}
}

func (r *vlessClientResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type vlessClientModel struct {
	ID         types.String `tfsdk:"id"`
	InboundID  types.Int64  `tfsdk:"inbound_id"`
	Email      types.String `tfsdk:"email"`
	UUID       types.String `tfsdk:"uuid"`
	Flow       types.String `tfsdk:"flow"`
	Enable     types.Bool   `tfsdk:"enable"`
	LimitIP    types.Int64  `tfsdk:"limit_ip"`
	TotalGB    types.Int64  `tfsdk:"total_gb"`
	ExpiryTime types.Int64  `tfsdk:"expiry_time"`
	TgID       types.Int64  `tfsdk:"tg_id"`
	SubID      types.String `tfsdk:"sub_id"`
	Comment    types.String `tfsdk:"comment"`
	Reset      types.Int64  `tfsdk:"reset"`
}

func (r *vlessClientResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan vlessClientModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	uid := strings.TrimSpace(plan.UUID.ValueString())
	if uid == "" {
		uid = uuid.New().String()
	}
	if _, err := uuid.Parse(uid); err != nil {
		resp.Diagnostics.AddError("Invalid uuid", err.Error())
		return
	}
	clientObj := r.clientMapFromPlan(plan, uid)
	settings := map[string]any{"clients": []any{clientObj}}
	raw, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		resp.Diagnostics.AddError("Internal error", err.Error())
		return
	}
	if err := r.client.AddInboundClient(int(plan.InboundID.ValueInt64()), string(raw)); err != nil {
		resp.Diagnostics.AddError("API error", err.Error())
		return
	}
	plan.ID = types.StringValue(uid)
	plan.UUID = types.StringValue(uid)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *vlessClientResource) clientMapFromPlan(plan vlessClientModel, uid string) map[string]any {
	m := map[string]any{
		"id":         uid,
		"email":      plan.Email.ValueString(),
		"enable":     plan.Enable.ValueBool(),
		"limitIp":    plan.LimitIP.ValueInt64(),
		"totalGB":    plan.TotalGB.ValueInt64(),
		"expiryTime": plan.ExpiryTime.ValueInt64(),
		"tgId":       plan.TgID.ValueInt64(),
		"reset":      plan.Reset.ValueInt64(),
	}
	if !plan.Flow.IsNull() && plan.Flow.ValueString() != "" {
		m["flow"] = plan.Flow.ValueString()
	} else {
		m["flow"] = ""
	}
	if !plan.SubID.IsNull() {
		m["subId"] = plan.SubID.ValueString()
	} else {
		m["subId"] = ""
	}
	if !plan.Comment.IsNull() {
		m["comment"] = plan.Comment.ValueString()
	} else {
		m["comment"] = ""
	}
	return m
}

func (r *vlessClientResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state vlessClientModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	raw, err := r.client.GetInbound(int(state.InboundID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("API error", err.Error())
		return
	}
	m, err := inboundMapFromJSON(raw)
	if err != nil {
		resp.Diagnostics.AddError("Decode error", err.Error())
		return
	}
	settingsJSON := stringFromMap(m, "settings")
	cm, err := findVLESSClientByEmail(settingsJSON, state.Email.ValueString())
	if err != nil {
		resp.State.RemoveResource(ctx)
		return
	}
	uid := clientUUID(cm)
	state.ID = types.StringValue(uid)
	state.UUID = types.StringValue(uid)
	if v, ok := cm["flow"].(string); ok {
		state.Flow = types.StringValue(v)
	} else {
		state.Flow = types.StringNull()
	}
	if v, ok := cm["enable"].(bool); ok {
		state.Enable = types.BoolValue(v)
	}
	state.LimitIP = types.Int64Value(int64FromAny(cm["limitIp"]))
	state.TotalGB = types.Int64Value(int64FromAny(cm["totalGB"]))
	state.ExpiryTime = types.Int64Value(int64FromAny(cm["expiryTime"]))
	state.TgID = types.Int64Value(int64FromAny(cm["tgId"]))
	if v, ok := cm["subId"].(string); ok {
		state.SubID = types.StringValue(v)
	} else {
		state.SubID = types.StringNull()
	}
	if v, ok := cm["comment"].(string); ok {
		state.Comment = types.StringValue(v)
	} else {
		state.Comment = types.StringNull()
	}
	state.Reset = types.Int64Value(int64FromAny(cm["reset"]))
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func int64FromAny(v any) int64 {
	switch n := v.(type) {
	case float64:
		return int64(n)
	case int64:
		return n
	case int:
		return int64(n)
	default:
		return 0
	}
}

func (r *vlessClientResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state vlessClientModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	uid := state.ID.ValueString()
	clientObj := r.clientMapFromPlan(plan, uid)
	settings := map[string]any{"clients": []any{clientObj}}
	raw, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		resp.Diagnostics.AddError("Internal error", err.Error())
		return
	}
	payload := map[string]any{
		"id":       int(plan.InboundID.ValueInt64()),
		"settings": string(raw),
	}
	if err := r.client.UpdateInboundClient(uid, payload); err != nil {
		resp.Diagnostics.AddError("API error", err.Error())
		return
	}
	state.Flow = plan.Flow
	state.Enable = plan.Enable
	state.LimitIP = plan.LimitIP
	state.TotalGB = plan.TotalGB
	state.ExpiryTime = plan.ExpiryTime
	state.TgID = plan.TgID
	state.SubID = plan.SubID
	state.Comment = plan.Comment
	state.Reset = plan.Reset
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *vlessClientResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state vlessClientModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteInboundClient(int(state.InboundID.ValueInt64()), state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("API error", err.Error())
	}
}

func (r *vlessClientResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid id", "Expected `inbound_id:email` (e.g. `3:user@example.com`).")
		return
	}
	inboundID, err := parseInt64Trim(parts[0])
	if err != nil {
		resp.Diagnostics.AddError("Invalid inbound_id", err.Error())
		return
	}
	email := strings.TrimSpace(parts[1])
	if email == "" {
		resp.Diagnostics.AddError("Invalid email", "Empty email in import id")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("inbound_id"), types.Int64Value(inboundID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("email"), types.StringValue(email))...)
}

func parseInt64Trim(s string) (int64, error) {
	return strconv.ParseInt(strings.TrimSpace(s), 10, 64)
}
