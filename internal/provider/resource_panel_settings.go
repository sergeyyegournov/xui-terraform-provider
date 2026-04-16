package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/syegournov/xkeen-gen/terraform-provider-xui/internal/xui"
)

var _ resource.Resource = (*panelSettingsResource)(nil)
var _ resource.ResourceWithImportState = (*panelSettingsResource)(nil)

type panelSettingsResource struct {
	client *xui.Client
}

func NewPanelSettingsResource() resource.Resource {
	return &panelSettingsResource{}
}

func (r *panelSettingsResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "xui_panel_settings"
}

func (r *panelSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages 3x-ui panel settings (`/panel/setting/update`). This is a singleton resource — only one instance should exist per panel. All attributes are optional and default to the panel's built-in defaults. Set `restart_panel` to true if you want to restart the panel after applying changes (required for web listen/port/cert changes to take effect).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Static resource id (`panel-settings`).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			// Web server
			"web_listen": schema.StringAttribute{
				MarkdownDescription: "Web panel listen IP address (empty = all interfaces).",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"web_domain": schema.StringAttribute{
				MarkdownDescription: "Web panel domain for validation.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"web_port": schema.Int64Attribute{
				MarkdownDescription: "Web panel port.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(2053),
			},
			"web_cert_file": schema.StringAttribute{
				MarkdownDescription: "Path to SSL certificate file for the web panel.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"web_key_file": schema.StringAttribute{
				MarkdownDescription: "Path to SSL private key file for the web panel.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"web_base_path": schema.StringAttribute{
				MarkdownDescription: "Base path for panel URLs (e.g. `/<uuid>/`).",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("/"),
			},
			"session_max_age": schema.Int64Attribute{
				MarkdownDescription: "Session maximum age in minutes.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(60),
			},

			// UI
			"page_size": schema.Int64Attribute{
				MarkdownDescription: "Number of items per page in lists.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(50),
			},
			"expire_diff": schema.Int64Attribute{
				MarkdownDescription: "Expiration warning threshold in days.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
			},
			"traffic_diff": schema.Int64Attribute{
				MarkdownDescription: "Traffic warning threshold percentage.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
			},
			"remark_model": schema.StringAttribute{
				MarkdownDescription: "Remark model pattern for inbounds.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"datepicker": schema.StringAttribute{
				MarkdownDescription: "Date picker format.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("gregorian"),
			},

			// Telegram bot
			"tg_bot_enable": schema.BoolAttribute{
				MarkdownDescription: "Enable Telegram bot notifications.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"tg_bot_token": schema.StringAttribute{
				MarkdownDescription: "Telegram bot token.",
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
				Default:             stringdefault.StaticString(""),
			},
			"tg_bot_proxy": schema.StringAttribute{
				MarkdownDescription: "Proxy URL for Telegram bot.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"tg_bot_api_server": schema.StringAttribute{
				MarkdownDescription: "Custom API server for Telegram bot.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"tg_bot_chat_id": schema.StringAttribute{
				MarkdownDescription: "Telegram chat ID for notifications.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"tg_run_time": schema.StringAttribute{
				MarkdownDescription: "Cron schedule for Telegram notifications (e.g. `@daily`).",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("@daily"),
			},
			"tg_bot_backup": schema.BoolAttribute{
				MarkdownDescription: "Enable database backup via Telegram.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"tg_bot_login_notify": schema.BoolAttribute{
				MarkdownDescription: "Send login notifications via Telegram.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"tg_cpu": schema.Int64Attribute{
				MarkdownDescription: "CPU usage threshold percentage for Telegram alerts.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(80),
			},
			"tg_lang": schema.StringAttribute{
				MarkdownDescription: "Telegram bot language.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("en-US"),
			},

			// Security
			"time_location": schema.StringAttribute{
				MarkdownDescription: "Time zone location (e.g. `UTC`, `Asia/Tehran`).",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("UTC"),
			},

			// Subscription server
			"sub_enable": schema.BoolAttribute{
				MarkdownDescription: "Enable subscription server.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"sub_json_enable": schema.BoolAttribute{
				MarkdownDescription: "Enable JSON subscription endpoint.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"sub_title": schema.StringAttribute{
				MarkdownDescription: "Subscription title.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"sub_support_url": schema.StringAttribute{
				MarkdownDescription: "Subscription support URL.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"sub_profile_url": schema.StringAttribute{
				MarkdownDescription: "Subscription profile URL.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"sub_announce": schema.StringAttribute{
				MarkdownDescription: "Subscription announcement.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"sub_listen": schema.StringAttribute{
				MarkdownDescription: "Subscription server listen IP.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"sub_port": schema.Int64Attribute{
				MarkdownDescription: "Subscription server port.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(2096),
			},
			"sub_path": schema.StringAttribute{
				MarkdownDescription: "Base path for subscription URLs.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("/sub/"),
			},
			"sub_domain": schema.StringAttribute{
				MarkdownDescription: "Domain for subscription server validation.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"sub_cert_file": schema.StringAttribute{
				MarkdownDescription: "SSL certificate file for subscription server.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"sub_key_file": schema.StringAttribute{
				MarkdownDescription: "SSL private key file for subscription server.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"sub_updates": schema.Int64Attribute{
				MarkdownDescription: "Subscription update interval in minutes.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(12),
			},
			"sub_encrypt": schema.BoolAttribute{
				MarkdownDescription: "Encrypt subscription responses.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"sub_show_info": schema.BoolAttribute{
				MarkdownDescription: "Show client information in subscriptions.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"sub_uri": schema.StringAttribute{
				MarkdownDescription: "Subscription server URI.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"sub_json_path": schema.StringAttribute{
				MarkdownDescription: "Path for JSON subscription endpoint.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("/json/"),
			},
			"sub_json_uri": schema.StringAttribute{
				MarkdownDescription: "JSON subscription server URI.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"sub_json_fragment": schema.StringAttribute{
				MarkdownDescription: "JSON subscription fragment configuration.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"sub_json_noises": schema.StringAttribute{
				MarkdownDescription: "JSON subscription noise configuration.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"sub_json_mux": schema.StringAttribute{
				MarkdownDescription: "JSON subscription mux configuration.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"sub_json_rules": schema.StringAttribute{
				MarkdownDescription: "JSON subscription routing rules.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"sub_enable_routing": schema.BoolAttribute{
				MarkdownDescription: "Enable routing for subscription.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"sub_routing_rules": schema.StringAttribute{
				MarkdownDescription: "Subscription global routing rules.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"external_traffic_inform_enable": schema.BoolAttribute{
				MarkdownDescription: "Enable external traffic reporting.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"external_traffic_inform_uri": schema.StringAttribute{
				MarkdownDescription: "URI for external traffic reporting.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},

			// Restart
			"restart_panel": schema.BoolAttribute{
				MarkdownDescription: "If true, restart the panel after applying changes. Required for web listen/port/cert changes to take effect.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		},
	}
}

func (r *panelSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type panelSettingsModel struct {
	ID types.String `tfsdk:"id"`

	// Web
	WebListen     types.String `tfsdk:"web_listen"`
	WebDomain     types.String `tfsdk:"web_domain"`
	WebPort       types.Int64  `tfsdk:"web_port"`
	WebCertFile   types.String `tfsdk:"web_cert_file"`
	WebKeyFile    types.String `tfsdk:"web_key_file"`
	WebBasePath   types.String `tfsdk:"web_base_path"`
	SessionMaxAge types.Int64  `tfsdk:"session_max_age"`

	// UI
	PageSize    types.Int64  `tfsdk:"page_size"`
	ExpireDiff  types.Int64  `tfsdk:"expire_diff"`
	TrafficDiff types.Int64  `tfsdk:"traffic_diff"`
	RemarkModel types.String `tfsdk:"remark_model"`
	Datepicker  types.String `tfsdk:"datepicker"`

	// Telegram
	TgBotEnable      types.Bool   `tfsdk:"tg_bot_enable"`
	TgBotToken       types.String `tfsdk:"tg_bot_token"`
	TgBotProxy       types.String `tfsdk:"tg_bot_proxy"`
	TgBotAPIServer   types.String `tfsdk:"tg_bot_api_server"`
	TgBotChatID      types.String `tfsdk:"tg_bot_chat_id"`
	TgRunTime        types.String `tfsdk:"tg_run_time"`
	TgBotBackup      types.Bool   `tfsdk:"tg_bot_backup"`
	TgBotLoginNotify types.Bool   `tfsdk:"tg_bot_login_notify"`
	TgCPU            types.Int64  `tfsdk:"tg_cpu"`
	TgLang           types.String `tfsdk:"tg_lang"`

	// Security
	TimeLocation types.String `tfsdk:"time_location"`

	// Subscription
	SubEnable                   types.Bool   `tfsdk:"sub_enable"`
	SubJSONEnable               types.Bool   `tfsdk:"sub_json_enable"`
	SubTitle                    types.String `tfsdk:"sub_title"`
	SubSupportURL               types.String `tfsdk:"sub_support_url"`
	SubProfileURL               types.String `tfsdk:"sub_profile_url"`
	SubAnnounce                 types.String `tfsdk:"sub_announce"`
	SubListen                   types.String `tfsdk:"sub_listen"`
	SubPort                     types.Int64  `tfsdk:"sub_port"`
	SubPath                     types.String `tfsdk:"sub_path"`
	SubDomain                   types.String `tfsdk:"sub_domain"`
	SubCertFile                 types.String `tfsdk:"sub_cert_file"`
	SubKeyFile                  types.String `tfsdk:"sub_key_file"`
	SubUpdates                  types.Int64  `tfsdk:"sub_updates"`
	SubEncrypt                  types.Bool   `tfsdk:"sub_encrypt"`
	SubShowInfo                 types.Bool   `tfsdk:"sub_show_info"`
	SubURI                      types.String `tfsdk:"sub_uri"`
	SubJSONPath                 types.String `tfsdk:"sub_json_path"`
	SubJSONURI                  types.String `tfsdk:"sub_json_uri"`
	SubJSONFragment             types.String `tfsdk:"sub_json_fragment"`
	SubJSONNoises               types.String `tfsdk:"sub_json_noises"`
	SubJSONMux                  types.String `tfsdk:"sub_json_mux"`
	SubJSONRules                types.String `tfsdk:"sub_json_rules"`
	SubEnableRouting            types.Bool   `tfsdk:"sub_enable_routing"`
	SubRoutingRules             types.String `tfsdk:"sub_routing_rules"`
	ExternalTrafficInformEnable types.Bool   `tfsdk:"external_traffic_inform_enable"`
	ExternalTrafficInformURI    types.String `tfsdk:"external_traffic_inform_uri"`

	// Restart
	RestartPanel types.Bool `tfsdk:"restart_panel"`
}

func (r *panelSettingsResource) modelToPayload(m *panelSettingsModel) map[string]any {
	p := map[string]any{
		"webListen":                   m.WebListen.ValueString(),
		"webDomain":                   m.WebDomain.ValueString(),
		"webPort":                     m.WebPort.ValueInt64(),
		"webCertFile":                 m.WebCertFile.ValueString(),
		"webKeyFile":                  m.WebKeyFile.ValueString(),
		"webBasePath":                 m.WebBasePath.ValueString(),
		"sessionMaxAge":               m.SessionMaxAge.ValueInt64(),
		"pageSize":                    m.PageSize.ValueInt64(),
		"expireDiff":                  m.ExpireDiff.ValueInt64(),
		"trafficDiff":                 m.TrafficDiff.ValueInt64(),
		"remarkModel":                 m.RemarkModel.ValueString(),
		"datepicker":                  m.Datepicker.ValueString(),
		"tgBotEnable":                 m.TgBotEnable.ValueBool(),
		"tgBotToken":                  m.TgBotToken.ValueString(),
		"tgBotProxy":                  m.TgBotProxy.ValueString(),
		"tgBotAPIServer":              m.TgBotAPIServer.ValueString(),
		"tgBotChatId":                 m.TgBotChatID.ValueString(),
		"tgRunTime":                   m.TgRunTime.ValueString(),
		"tgBotBackup":                 m.TgBotBackup.ValueBool(),
		"tgBotLoginNotify":            m.TgBotLoginNotify.ValueBool(),
		"tgCpu":                       m.TgCPU.ValueInt64(),
		"tgLang":                      m.TgLang.ValueString(),
		"timeLocation":                m.TimeLocation.ValueString(),
		"subEnable":                   m.SubEnable.ValueBool(),
		"subJsonEnable":               m.SubJSONEnable.ValueBool(),
		"subTitle":                    m.SubTitle.ValueString(),
		"subSupportUrl":               m.SubSupportURL.ValueString(),
		"subProfileUrl":               m.SubProfileURL.ValueString(),
		"subAnnounce":                 m.SubAnnounce.ValueString(),
		"subListen":                   m.SubListen.ValueString(),
		"subPort":                     m.SubPort.ValueInt64(),
		"subPath":                     m.SubPath.ValueString(),
		"subDomain":                   m.SubDomain.ValueString(),
		"subCertFile":                 m.SubCertFile.ValueString(),
		"subKeyFile":                  m.SubKeyFile.ValueString(),
		"subUpdates":                  m.SubUpdates.ValueInt64(),
		"subEncrypt":                  m.SubEncrypt.ValueBool(),
		"subShowInfo":                 m.SubShowInfo.ValueBool(),
		"subURI":                      m.SubURI.ValueString(),
		"subJsonPath":                 m.SubJSONPath.ValueString(),
		"subJsonURI":                  m.SubJSONURI.ValueString(),
		"subJsonFragment":             m.SubJSONFragment.ValueString(),
		"subJsonNoises":               m.SubJSONNoises.ValueString(),
		"subJsonMux":                  m.SubJSONMux.ValueString(),
		"subJsonRules":                m.SubJSONRules.ValueString(),
		"subEnableRouting":            m.SubEnableRouting.ValueBool(),
		"subRoutingRules":             m.SubRoutingRules.ValueString(),
		"externalTrafficInformEnable": m.ExternalTrafficInformEnable.ValueBool(),
		"externalTrafficInformURI":    m.ExternalTrafficInformURI.ValueString(),
	}
	return p
}

func (r *panelSettingsResource) apiToModel(m map[string]any, state *panelSettingsModel) {
	state.WebListen = types.StringValue(stringFromMap(m, "webListen"))
	state.WebDomain = types.StringValue(stringFromMap(m, "webDomain"))
	state.WebPort = types.Int64Value(int64FromMap(m, "webPort"))
	state.WebCertFile = types.StringValue(stringFromMap(m, "webCertFile"))
	state.WebKeyFile = types.StringValue(stringFromMap(m, "webKeyFile"))
	state.WebBasePath = types.StringValue(stringFromMap(m, "webBasePath"))
	state.SessionMaxAge = types.Int64Value(int64FromMap(m, "sessionMaxAge"))
	state.PageSize = types.Int64Value(int64FromMap(m, "pageSize"))
	state.ExpireDiff = types.Int64Value(int64FromMap(m, "expireDiff"))
	state.TrafficDiff = types.Int64Value(int64FromMap(m, "trafficDiff"))
	state.RemarkModel = types.StringValue(stringFromMap(m, "remarkModel"))
	state.Datepicker = types.StringValue(stringFromMap(m, "datepicker"))
	state.TgBotEnable = types.BoolValue(boolFromMap(m, "tgBotEnable"))
	state.TgBotToken = types.StringValue(stringFromMap(m, "tgBotToken"))
	state.TgBotProxy = types.StringValue(stringFromMap(m, "tgBotProxy"))
	state.TgBotAPIServer = types.StringValue(stringFromMap(m, "tgBotAPIServer"))
	state.TgBotChatID = types.StringValue(stringFromMap(m, "tgBotChatId"))
	state.TgRunTime = types.StringValue(stringFromMap(m, "tgRunTime"))
	state.TgBotBackup = types.BoolValue(boolFromMap(m, "tgBotBackup"))
	state.TgBotLoginNotify = types.BoolValue(boolFromMap(m, "tgBotLoginNotify"))
	state.TgCPU = types.Int64Value(int64FromMap(m, "tgCpu"))
	state.TgLang = types.StringValue(stringFromMap(m, "tgLang"))
	state.TimeLocation = types.StringValue(stringFromMap(m, "timeLocation"))
	state.SubEnable = types.BoolValue(boolFromMap(m, "subEnable"))
	state.SubJSONEnable = types.BoolValue(boolFromMap(m, "subJsonEnable"))
	state.SubTitle = types.StringValue(stringFromMap(m, "subTitle"))
	state.SubSupportURL = types.StringValue(stringFromMap(m, "subSupportUrl"))
	state.SubProfileURL = types.StringValue(stringFromMap(m, "subProfileUrl"))
	state.SubAnnounce = types.StringValue(stringFromMap(m, "subAnnounce"))
	state.SubListen = types.StringValue(stringFromMap(m, "subListen"))
	state.SubPort = types.Int64Value(int64FromMap(m, "subPort"))
	state.SubPath = types.StringValue(stringFromMap(m, "subPath"))
	state.SubDomain = types.StringValue(stringFromMap(m, "subDomain"))
	state.SubCertFile = types.StringValue(stringFromMap(m, "subCertFile"))
	state.SubKeyFile = types.StringValue(stringFromMap(m, "subKeyFile"))
	state.SubUpdates = types.Int64Value(int64FromMap(m, "subUpdates"))
	state.SubEncrypt = types.BoolValue(boolFromMap(m, "subEncrypt"))
	state.SubShowInfo = types.BoolValue(boolFromMap(m, "subShowInfo"))
	state.SubURI = types.StringValue(stringFromMap(m, "subURI"))
	state.SubJSONPath = types.StringValue(stringFromMap(m, "subJsonPath"))
	state.SubJSONURI = types.StringValue(stringFromMap(m, "subJsonURI"))
	state.SubJSONFragment = types.StringValue(stringFromMap(m, "subJsonFragment"))
	state.SubJSONNoises = types.StringValue(stringFromMap(m, "subJsonNoises"))
	state.SubJSONMux = types.StringValue(stringFromMap(m, "subJsonMux"))
	state.SubJSONRules = types.StringValue(stringFromMap(m, "subJsonRules"))
	state.SubEnableRouting = types.BoolValue(boolFromMap(m, "subEnableRouting"))
	state.SubRoutingRules = types.StringValue(stringFromMap(m, "subRoutingRules"))
	state.ExternalTrafficInformEnable = types.BoolValue(boolFromMap(m, "externalTrafficInformEnable"))
	state.ExternalTrafficInformURI = types.StringValue(stringFromMap(m, "externalTrafficInformURI"))
}

func (r *panelSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan panelSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	payload := r.modelToPayload(&plan)
	if err := r.client.UpdatePanelSettings(payload); err != nil {
		resp.Diagnostics.AddError("API error", err.Error())
		return
	}
	if plan.RestartPanel.ValueBool() {
		if err := r.client.RestartPanel(); err != nil {
			resp.Diagnostics.AddWarning("Panel restart failed", fmt.Sprintf("Settings were saved but panel restart failed: %s", err.Error()))
		}
	}
	plan.ID = types.StringValue("panel-settings")
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *panelSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state panelSettingsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	m, err := r.client.GetPanelSettings()
	if err != nil {
		resp.Diagnostics.AddError("API error", err.Error())
		return
	}
	r.apiToModel(m, &state)
	if state.ID.IsNull() || state.ID.ValueString() == "" {
		state.ID = types.StringValue("panel-settings")
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *panelSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan panelSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	payload := r.modelToPayload(&plan)
	if err := r.client.UpdatePanelSettings(payload); err != nil {
		resp.Diagnostics.AddError("API error", err.Error())
		return
	}
	if plan.RestartPanel.ValueBool() {
		if err := r.client.RestartPanel(); err != nil {
			resp.Diagnostics.AddWarning("Panel restart failed", fmt.Sprintf("Settings were saved but panel restart failed: %s", err.Error()))
		}
	}
	plan.ID = types.StringValue("panel-settings")
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *panelSettingsResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Panel settings cannot be deleted; removing from state is sufficient.
}

func (r *panelSettingsResource) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	m, err := r.client.GetPanelSettings()
	if err != nil {
		resp.Diagnostics.AddError("API error", err.Error())
		return
	}
	var state panelSettingsModel
	state.ID = types.StringValue("panel-settings")
	state.RestartPanel = types.BoolValue(false)
	r.apiToModel(m, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
