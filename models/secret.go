package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type SecretModel struct {
	StashID types.String `tfsdk:"stash_id"`
	Name    types.String `tfsdk:"name"`
	Value   types.String `tfsdk:"value"`
}

type SecretResourceModel struct {
	SecretModel
	LastUpdated types.String `tfsdk:"last_updated"`
}
