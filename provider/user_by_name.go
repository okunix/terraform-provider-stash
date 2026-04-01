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

var (
	_ datasource.DataSource              = (*userByNameDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*userByNameDataSource)(nil)
)

type userByNameDataSource struct {
	client *stash.Client
}

func NewUserByNameDataSource() datasource.DataSource {
	return &userByNameDataSource{}
}

func (u *userByNameDataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_user_by_name"
}

func (u *userByNameDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var state userDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// this endpoint handles both username and id
	userResp, err := u.client.GetUserByID(ctx, state.Username.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to fetch username by id", err.Error())
		return
	}

	state.CreatedAt = types.StringValue(userResp.CreatedAt.Format(time.RFC3339))
	state.ID = types.StringValue(userResp.ID)
	state.Locked = types.BoolValue(userResp.Locked)
	if userResp.ExpiredAt != nil {
		state.ExpiredAt = types.StringValue((*userResp.ExpiredAt).Format(time.RFC3339))
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (u *userByNameDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"username": schema.StringAttribute{
				Required: true,
			},
			"locked": schema.BoolAttribute{
				Computed: true,
			},
			"created_at": schema.StringAttribute{
				Computed: true,
			},
			"expired_at": schema.StringAttribute{
				Computed: true,
				Optional: true,
			},
		},
	}
}

func (u *userByNameDataSource) Configure(
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
			fmt.Sprintf("expected *stash.Client got %T", req.ProviderData),
		)
		return
	}

	u.client = client
}
