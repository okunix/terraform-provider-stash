package datasources

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/okunix/stash-sdk/stash/v1"
	"github.com/okunix/terraform-provider-stash/models"
)

var (
	_ datasource.DataSource                     = (*userDataSource)(nil)
	_ datasource.DataSourceWithConfigure        = (*userDataSource)(nil)
	_ datasource.DataSourceWithConfigValidators = (*userDataSource)(nil)
)

type userDataSource struct {
	client *stash.Client
}

func NewUserDataSource() datasource.DataSource {
	return &userDataSource{}
}

func (u *userDataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (u *userDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var state models.UserModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var userResponse *stash.UserResponse
	var err error

	if !state.Username.IsNull() {
		userResponse, err = u.client.GetUserByID(ctx, state.Username.ValueString())
	} else if !state.ID.IsNull() {
		userResponse, err = u.client.GetUserByID(ctx, state.ID.ValueString())
	} else {
		userResponse, err = u.client.Whoami(ctx)
	}
	if err != nil {
		resp.Diagnostics.AddError("failed to fetch user", err.Error())
		return
	}

	state.CreatedAt = types.StringValue(userResponse.CreatedAt.Format(time.RFC3339))
	state.ID = types.StringValue(userResponse.ID)
	state.Username = types.StringValue(userResponse.Username)
	state.Locked = types.BoolValue(userResponse.Locked)
	if userResponse.ExpiredAt != nil {
		state.ExpiredAt = types.StringValue((*userResponse.ExpiredAt).Format(time.RFC3339))
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (u *userDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional: true,
			},
			"username": schema.StringAttribute{
				Optional: true,
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

func (u *userDataSource) Configure(
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

func (u *userDataSource) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.Conflicting(
			path.MatchRoot("username"),
			path.MatchRoot("id"),
		),
	}
}
