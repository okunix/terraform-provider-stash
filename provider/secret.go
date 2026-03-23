package provider

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

type secretDataSource struct {
	client *http.Client
}

func NewSecretDataSource() func() datasource.DataSource {
	return func() datasource.DataSource {
		return &secretDataSource{}
	}
}

func (s *secretDataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_secret"
}

func (s *secretDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "name of a secret",
				Required:    true,
			},
			"stash_id": schema.StringAttribute{
				Description: "id of a stash that stores the secret",
				Required:    true,
			},
			"value": schema.StringAttribute{
				Description: "value of a secret",
				Computed:    true,
			},
		},
	}
}

type secretDataSourceData struct {
	Name    string `tfsdk:"name"`
	StashID string `tfsdk:"stash_id"`
	Value   string `tfsdk:"value"`
}

func (s *secretDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data secretDataSourceData
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (s *secretDataSource) Configure(
	ctx context.Context,
	req datasource.ConfigureRequest,
	resp datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}
}
