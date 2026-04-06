package datasources

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/okunix/stash-sdk/stash/v1"
	"github.com/okunix/terraform-provider-stash/models"
)

type stashByIDDataSource struct {
	client *stash.Client
}

var (
	_ datasource.DataSource              = (*stashByIDDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*stashByIDDataSource)(nil)
)

func NewStashByIDDataSource() datasource.DataSource {
	return &stashByIDDataSource{}
}

func (s *stashByIDDataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_stash_by_id"
}

func (s *stashByIDDataSource) Read(
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

	stashResponse, err := s.client.GetStashByID(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to fetch stash", err.Error())
		return
	}

	state.CreatedAt = types.StringValue(stashResponse.CreatedAt.Format(time.RFC3339))
	state.Locked = types.BoolValue(stashResponse.Locked)
	state.MaintainerID = types.StringValue(stashResponse.MaintainerID)
	state.Name = types.StringValue(stashResponse.Name)
	if stashResponse.Description != nil {
		state.Description = types.StringValue(*stashResponse.Description)
	}

	diag = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diag...)
}

func (s *stashByIDDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
			"description": schema.StringAttribute{
				Computed: true,
				Optional: true,
			},
			"maintainer_id": schema.StringAttribute{
				Computed: true,
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

func (s *stashByIDDataSource) Configure(
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
