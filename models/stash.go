package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type StashModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	MaintainerID types.String `tfsdk:"maintainer_id"`
	CreatedAt    types.String `tfsdk:"created_at"`
	Locked       types.Bool   `tfsdk:"locked"`
}

type StashResourceModel struct {
	StashModel
	Password    types.String `tfsdk:"password"`
	LastUpdated types.String `tfsdk:"last_updated"`
}
