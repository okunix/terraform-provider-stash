package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

type userDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Username  types.String `tfsdk:"username"`
	Locked    types.Bool   `tfsdk:"locked"`
	CreatedAt types.String `tfsdk:"created_at"`
	ExpiredAt types.String `tfsdk:"expired_at"`
}
