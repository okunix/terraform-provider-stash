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
	stash "github.com/okunix/stash-sdk/stash/v1"
	"github.com/okunix/terraform-provider-stash/models"
)

type stashDataSource struct {
	client *stash.Client
}

var (
	_ datasource.DataSource                     = (*stashDataSource)(nil)
	_ datasource.DataSourceWithConfigure        = (*stashDataSource)(nil)
	_ datasource.DataSourceWithConfigValidators = (*stashDataSource)(nil)
)

func NewStashDataSource() datasource.DataSource {
	return &stashDataSource{}
}

func (s *stashDataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_stash"
}

func (s *stashDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var state models.StashModel
	diag := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}
	var stashResponse *stash.StashResponse
	var err error

	if !state.Name.IsNull() {
		stashResponse, err = s.client.GetStashByName(ctx,
			state.MaintainerID.ValueString(),
			state.Name.ValueString(),
		)
	} else if !state.ID.IsNull() {
		stashResponse, err = s.client.GetStashByID(ctx, state.ID.ValueString())
	} else {
		resp.Diagnostics.AddError("failed to fetch stash", "no request data provided")
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("failed to fetch stash", err.Error())
		return
	}

	state.Name = types.StringValue(stashResponse.Name)
	state.MaintainerID = types.StringValue(stashResponse.MaintainerID)
	state.CreatedAt = types.StringValue(stashResponse.CreatedAt.Format(time.RFC3339))
	if stashResponse.Description != nil {
		state.Description = types.StringValue(*stashResponse.Description)
	}
	state.Locked = types.BoolValue(stashResponse.Locked)
	state.ID = types.StringValue(stashResponse.ID)

	diag = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diag...)
}

func (s *stashDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional: true,
			},
			"name": schema.StringAttribute{
				Optional: true,
			},
			"description": schema.StringAttribute{
				Computed: true,
				Optional: true,
			},
			"maintainer_id": schema.StringAttribute{
				Optional: true,
			},
			"created_at": schema.StringAttribute{
				Computed: true,
			},
			"locked": schema.BoolAttribute{
				Computed: true,
			},
		},
	}
}

func (s *stashDataSource) Configure(
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

	s.client = client
}

func (s *stashDataSource) ConfigValidators(
	ctx context.Context,
) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.ExactlyOneOf(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
		),
		datasourcevalidator.Conflicting(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
		),
		datasourcevalidator.Conflicting(
			path.MatchRoot("id"),
			path.MatchRoot("maintainer_id"),
		),
		datasourcevalidator.RequiredTogether(
			path.MatchRoot("name"),
			path.MatchRoot("maintainer_id"),
		),
	}
}
