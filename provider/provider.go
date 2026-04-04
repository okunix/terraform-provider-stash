package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	stash "github.com/okunix/stash-sdk/stash/v1"
	"github.com/okunix/terraform-provider-stash/version"
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
	ApiServer  types.String            `tfsdk:"api_server"`
	ApiToken   types.String            `tfsdk:"api_token"`
	ApiVersion types.String            `tfsdk:"api_version"`
	User       *StashProviderUserModel `tfsdk:"user"`
}

func (s *stashProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewWhoamiDataSource,
		NewUserByIDDataSource,
		NewUserByNameDataSource,
		NewStashByNameDataSource,
		NewStashByIDDataSource,
		NewSecretDataSource,
	}
}

func (s *stashProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewMemberResource,
	}
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
				Description: "Credentials for authenticating with the API.",
				Attributes: map[string]schema.Attribute{
					"username": schema.StringAttribute{
						Required:    true,
						Description: "Username for authentication.",
					},
					"password": schema.StringAttribute{
						Required:    true,
						Sensitive:   true,
						Description: "Password for authentication.",
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

	if data.ApiVersion.ValueString() != "v1" {
		resp.Diagnostics.AddError("unsupported version", "unsupported version secified in provider")
		return
	}

	client, err := stash.NewClient(
		stash.WithAddr(data.ApiServer.ValueString()),
		stash.WithUser(
			data.User.Username.ValueString(),
			data.User.Password.ValueString(),
		),
	)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("failed to initiate client", err.Error()))
		return
	}

	if err := client.Ping(ctx); err != nil {
		resp.Diagnostics.AddError("stash server ping failed", err.Error())
		return
	}

	resp.ResourceData = client
	resp.DataSourceData = client
}
