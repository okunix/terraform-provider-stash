package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type stashResponseModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	MaintainerID types.String `tfsdk:"maintainer_id"`
	CreatedAt    types.String `tfsdk:"created_at"`
	Locked       types.Bool   `tfsdk:"locked"`
}
