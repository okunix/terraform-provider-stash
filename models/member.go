package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type MemberModel struct {
	StashID types.String `tfsdk:"stash_id"`
	UserID  types.String `tfsdk:"user_id"`
	Since   types.String `tfsdk:"since"`
}

type MemberResourceModel struct {
	MemberModel
	LastUpdated types.String `tfsdk:"last_updated"`
}
