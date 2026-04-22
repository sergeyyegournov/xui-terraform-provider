package provider

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/syegournov/xkeen-gen/terraform-provider-xui/internal/xui"
)

var _ datasource.DataSource = (*inboundsDataSource)(nil)

type inboundsDataSource struct {
	client *xui.Client
}

func NewInboundsDataSource() datasource.DataSource {
	return &inboundsDataSource{}
}

func (d *inboundsDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "xui_inbounds"
}

func (d *inboundsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "List inbounds from `/panel/api/inbounds/list`. Use `json` with `jsondecode()` in Terraform for structured access.",
		Attributes: map[string]schema.Attribute{
			"protocol": schema.StringAttribute{
				MarkdownDescription: "If set, filter results to this protocol (e.g. `vless`).",
				Optional:            true,
			},
			"json": schema.StringAttribute{
				MarkdownDescription: "JSON array of inbound objects after optional protocol filter.",
				Computed:            true,
			},
		},
	}
}

func (d *inboundsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	cli, ok := req.ProviderData.(*xui.Client)
	if !ok {
		resp.Diagnostics.AddError("Internal error", "invalid provider data type")
		return
	}
	d.client = cli
}

type inboundsDSModel struct {
	Protocol types.String `tfsdk:"protocol"`
	JSON     types.String `tfsdk:"json"`
}

func (d *inboundsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg inboundsDSModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	raw, err := d.client.ListInbounds()
	if err != nil {
		resp.Diagnostics.AddError("API error", err.Error())
		return
	}
	filter := ""
	if !cfg.Protocol.IsNull() && cfg.Protocol.ValueString() != "" {
		filter = cfg.Protocol.ValueString()
	}
	encoded, err := filterInboundsJSON(raw, filter)
	if err != nil {
		resp.Diagnostics.AddError("Decode error", err.Error())
		return
	}
	cfg.JSON = types.StringValue(string(encoded))
	resp.Diagnostics.Append(resp.State.Set(ctx, cfg)...)
}

func filterInboundsJSON(raw []byte, protocol string) ([]byte, error) {
	var arr []map[string]any
	if err := json.Unmarshal(raw, &arr); err != nil {
		return nil, err
	}
	filter := strings.TrimSpace(protocol)
	var out []map[string]any
	for _, m := range arr {
		if filter != "" && stringFromMap(m, "protocol") != filter {
			continue
		}
		out = append(out, m)
	}
	return json.MarshalIndent(out, "", "  ")
}
