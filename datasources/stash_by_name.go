package datasources

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	stash "github.com/okunix/stash-sdk/stash/v1"
	"github.com/okunix/terraform-provider-stash/models"
)

type stashByNameDataSource struct {
	client *stash.Client
}

var (
	_ datasource.DataSource              = (*stashByNameDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*stashByNameDataSource)(nil)
)

func NewStashByNameDataSource() datasource.DataSource {
	return &stashByNameDataSource{}
}

func (s *stashByNameDataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_stash_by_name"
}

func (s *stashByNameDataSource) Read(
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
	stashResponse, err := s.client.GetStashByName(ctx,
		state.MaintainerID.ValueString(),
		state.Name.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("failed to fetch stash by name", err.Error())
		return
	}

	state.CreatedAt = types.StringValue(stashResponse.CreatedAt.Format(time.RFC3339))
	if stashResponse.Description != nil {
		state.Description = types.StringValue(*stashResponse.Description)
	}
	state.Locked = types.BoolValue(stashResponse.Locked)
	state.ID = types.StringValue(stashResponse.ID)

	diag = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diag...)
}

func (s *stashByNameDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"description": schema.StringAttribute{
				Computed: true,
				Optional: true,
			},
			"maintainer_id": schema.StringAttribute{
				Required: true,
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

func (s *stashByNameDataSource) Configure(
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
