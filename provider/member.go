package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/okunix/stash-sdk/stash/v1"
)

type memberResource struct {
	client *stash.Client
}

var (
	_ resource.Resource              = (*memberResource)(nil)
	_ resource.ResourceWithConfigure = (*memberResource)(nil)
)

type memberModel struct {
	StashID     types.String `tfsdk:"stash_id"`
	UserID      types.String `tfsdk:"user_id"`
	Since       types.String `tfsdk:"since"`
	LastUpdated types.String `tfsdk:"last_updated"`
}

func NewMemberResource() resource.Resource {
	return &memberResource{}
}

func (m *memberResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_member"
}

func (m *memberResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan memberModel
	diag := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	stashID := plan.StashID.ValueString()
	userID := plan.UserID.ValueString()
	err := m.client.AddStashMember(ctx, stashID, userID)
	if err != nil {
		resp.Diagnostics.AddError("failed to add a new member",
			fmt.Sprintf("creation failed: %s", err.Error()),
		)
		return
	}

	newMember, err := m.client.GetStashMember(ctx, stashID, userID)
	if err != nil {
		resp.Diagnostics.AddError("failed to add a new member",
			fmt.Sprintf("lookup failed: %s", err.Error()),
		)
		return
	}

	plan.Since = types.StringValue(newMember.Since.Format(time.RFC3339))
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC3339))

	diag = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diag...)
}

func (m *memberResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state memberModel
	diag := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	stashMember, err := m.client.GetStashMember(
		ctx,
		state.StashID.ValueString(),
		state.UserID.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("failed to fetch stash member", err.Error())
		return
	}

	state.Since = types.StringValue(stashMember.Since.Format(time.RFC3339))

	diag = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diag...)
}

func (m *memberResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan memberModel
	diag := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state memberModel
	diag = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	oldStashID := state.StashID.ValueString()
	oldUserID := state.UserID.ValueString()
	newStashID := plan.StashID.ValueString()
	newUserID := plan.UserID.ValueString()

	if _, err := m.client.GetStashByID(ctx, newStashID); err != nil {
		resp.Diagnostics.AddError("failed to update stash member", err.Error())
	}

	if _, err := m.client.GetUserByID(ctx, newUserID); err != nil {
		resp.Diagnostics.AddError("failed to update stash member", err.Error())
	}

	err := m.client.RemoveStashMember(ctx, oldStashID, oldUserID)
	if err != nil {
		resp.Diagnostics.AddError("failed to update stash member", err.Error())
		return
	}

	err = m.client.AddStashMember(ctx, newStashID, newUserID)
	if err != nil {
		resp.Diagnostics.AddError("failed to update stash member", err.Error())
		return
	}

	memberResp, err := m.client.GetStashMember(ctx, newStashID, newUserID)
	if err != nil {
		resp.Diagnostics.AddError("failed to update stash member", err.Error())
	}

	state.UserID = plan.UserID
	state.StashID = plan.StashID
	state.Since = types.StringValue(memberResp.Since.Format(time.RFC3339))
	state.LastUpdated = types.StringValue(time.Now().Format(time.RFC3339))

	diag = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diag...)
}

func (m *memberResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state memberModel
	diag := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := m.client.RemoveStashMember(ctx, state.StashID.ValueString(), state.UserID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to remove stash member", err.Error())
		return
	}

}

func (m *memberResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"since": schema.StringAttribute{
				Computed:    true,
				Description: "when member joined",
			},
			"stash_id": schema.StringAttribute{
				Required: true,
			},
			"user_id": schema.StringAttribute{
				Required: true,
			},
			"last_updated": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (m *memberResource) Configure(
	ctx context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
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

	m.client = client
}
