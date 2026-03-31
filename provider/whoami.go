package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/okunix/stash-sdk/stash/v1"
)

type whoamiDataSource struct {
	client *stash.Client
}

type whoamiDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Username  types.String `tfsdk:"username"`
	CreatedAt types.String `tfsdk:"created_at"`
	Locked    types.Bool   `tfsdk:"locked"`
}

var (
	_ datasource.DataSource              = (*whoamiDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*whoamiDataSource)(nil)
)

func NewWhoamiDataSource() datasource.DataSource {
	return &whoamiDataSource{}
}

func (w *whoamiDataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_whoami"
}

func (w *whoamiDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	whoamiResp, err := w.client.Whoami(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to perform whoami function", err.Error())
		return
	}

	var state whoamiDataSourceModel
	state.ID = types.StringValue(whoamiResp.ID)
	state.Username = types.StringValue(whoamiResp.Username)
	state.CreatedAt = types.StringValue(whoamiResp.CreatedAt.Format(time.RFC3339))
	state.Locked = types.BoolValue(whoamiResp.Locked)
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (w *whoamiDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"created_at": schema.StringAttribute{
				Computed: true,
			},
			"username": schema.StringAttribute{
				Computed: true,
			},
			"id": schema.StringAttribute{
				Computed: true,
			},
			"locked": schema.BoolAttribute{
				Computed: true,
			},
		},
	}
}

func (w *whoamiDataSource) Configure(
	ctx context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*stash.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"unexpected data source configure type",
			fmt.Sprintf("expected *stash.Client, got %T", req.ProviderData),
		)
	}
	w.client = client
}
