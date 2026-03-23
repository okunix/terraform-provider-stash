package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/okunix/terraform-provider-stash/version"
	"github.com/okunix/terraform-provider-stash/webutil"
)

func New() func() provider.Provider {
	return func() provider.Provider {
		return &stashProvider{}
	}
}

type stashProvider struct{}

type StashProviderUserModel struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

type StashProviderModel struct {
	ApiVersion types.String            `tfsdk:"api_version"`
	ApiServer  types.String            `tfsdk:"api_server"`
	ApiToken   types.String            `tfsdk:"api_token"`
	User       *StashProviderUserModel `tfsdk:"user"`
}

func (s *stashProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewSecretDataSource(),
	}
}

func (s *stashProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}

func (s *stashProvider) Schema(
	ctx context.Context,
	req provider.SchemaRequest,
	resp *provider.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_version": schema.StringAttribute{
				Required:    true,
				Description: "stash-server api version",
			},
			"api_server": schema.StringAttribute{
				Required:    true,
				Description: "stash-server api server address",
			},
			"api_token": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "stash-server api token",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("user")),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"user": schema.SingleNestedBlock{
				Description: "stash-server user credentials",
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(path.MatchRoot("token")),
				},
				Attributes: map[string]schema.Attribute{
					"username": schema.StringAttribute{
						Required:    true,
						Description: "account username",
					},
					"password": schema.StringAttribute{
						Required:    true,
						Sensitive:   true,
						Description: "account password",
					},
				},
			},
		},
	}
}

func (s *stashProvider) Metadata(
	ctx context.Context,
	req provider.MetadataRequest,
	resp *provider.MetadataResponse,
) {
	resp.Version = version.Version()
	resp.TypeName = "stash"
}

func (s *stashProvider) Configure(
	ctx context.Context,
	req provider.ConfigureRequest,
	resp *provider.ConfigureResponse,
) {
	var data StashProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client := webutil.NewClientWithAuth(data.ApiToken.ValueString(), 60*time.Second)

	resp.ResourceData = client
	resp.DataSourceData = client
}
