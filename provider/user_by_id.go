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

type userByIDDataSource struct {
	client *stash.Client
}

var (
	_ datasource.DataSource              = (*userByIDDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*userByIDDataSource)(nil)
)

func NewUserByIDDataSource() datasource.DataSource {
	return &userByIDDataSource{}
}

func (u *userByIDDataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_user_by_id"
}

func (u *userByIDDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var state userDataSourceModel
	diag := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	userResp, err := u.client.GetUserByID(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to fetch user information", err.Error())
		return
	}

	state.CreatedAt = types.StringValue(userResp.CreatedAt.Format(time.RFC3339))
	state.Username = types.StringValue(userResp.Username)
	state.Locked = types.BoolValue(userResp.Locked)
	if userResp.ExpiredAt != nil {
		state.ExpiredAt = types.StringValue((*userResp.ExpiredAt).Format(time.RFC3339))
	}

	diag = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diag...)
}

func (u *userByIDDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "user id",
			},
			"username": schema.StringAttribute{
				Computed:    true,
				Description: "user name",
			},
			"locked": schema.BoolAttribute{
				Computed:    true,
				Description: "is user account locked",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "when user was created",
			},
			"expired_at": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Description: "when user expires",
			},
		},
	}
}

func (u *userByIDDataSource) Configure(
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
