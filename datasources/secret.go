package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	stash "github.com/okunix/stash-sdk/stash/v1"
	"github.com/okunix/terraform-provider-stash/models"
)

type secretDataSource struct {
	client *stash.Client
}

var (
	_ datasource.DataSource              = (*secretDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*secretDataSource)(nil)
)

func NewSecretDataSource() datasource.DataSource {
	return &secretDataSource{}
}

func (s *secretDataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_secret"
}

func (s *secretDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var state models.SecretModel
	diag := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	value, err := s.client.GetSecretsEntry(
		ctx,
		state.StashID.ValueString(),
		state.Name.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("failed to fetch secret", err.Error())
		return
	}

	state.Value = types.StringValue(value)

	diag = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diag...)
}

func (s *secretDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"stash_id": schema.StringAttribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"value": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func (s *secretDataSource) Configure(
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
