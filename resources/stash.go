package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/okunix/stash-sdk/stash/v1"
	"github.com/okunix/terraform-provider-stash/models"
)

type stashResource struct {
	client *stash.Client
}

var (
	_ resource.Resource              = (*stashResource)(nil)
	_ resource.ResourceWithConfigure = (*stashResource)(nil)
)

func NewStashResource() resource.Resource {
	return &stashResource{}
}

func (s *stashResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan models.StashResourceModel
	diag := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	iam, err := s.client.Whoami(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to fetch current user information", err.Error())
		return
	}

	request := stash.CreateStashRequest{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueStringPointer(),
		Password:    plan.Password.ValueString(),
	}
	err = s.client.CreateStash(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("failed to create stash", err.Error())
		return
	}

	stashResponse, err := s.client.GetStashByName(ctx, iam.ID, request.Name)
	if err != nil {
		resp.Diagnostics.AddError("failed to fetch created stash", err.Error())
		return
	}

	s.client.Unlock(ctx, stashResponse.ID, plan.Password.ValueString())

	plan.CreatedAt = types.StringValue(stashResponse.CreatedAt.Format(time.RFC3339))
	plan.ID = types.StringValue(stashResponse.ID)
	plan.MaintainerID = types.StringValue(stashResponse.MaintainerID)
	plan.Locked = types.BoolValue(stashResponse.Locked)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC3339))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (s *stashResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state models.StashResourceModel
	diag := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := s.client.DeleteStash(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to delete stash", err.Error())
		return
	}
}

func (s *stashResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state models.StashResourceModel
	diag := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	stashResp, err := s.client.GetStashByID(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to fetch stash by id", err.Error())
		return
	}

	state.Name = types.StringValue(stashResp.Name)
	state.MaintainerID = types.StringValue(stashResp.MaintainerID)
	state.ID = types.StringValue(stashResp.ID)
	state.Locked = types.BoolValue(stashResp.Locked)
	if stashResp.Description != nil {
		state.Description = types.StringValue(*stashResp.Description)
	}
	state.CreatedAt = types.StringValue(stashResp.CreatedAt.Format(time.RFC3339))

	diag = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diag...)
}

func (s *stashResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan models.StashResourceModel
	var state models.StashResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	newName := plan.Name.ValueString()
	request := stash.UpdateStashRequest{
		Name:        &newName,
		Description: plan.Description.ValueStringPointer(),
	}
	err := s.client.UpdateStash(ctx, state.ID.ValueString(), request)
	if err != nil {
		resp.Diagnostics.AddError("failed to update stash", err.Error())
		return
	}

	state.Name = plan.Name
	state.Description = plan.Description
	state.LastUpdated = types.StringValue(time.Now().Format(time.RFC3339))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (s *stashResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
			},
			"password": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"locked": schema.BoolAttribute{
				Computed: true,
			},
			"maintainer_id": schema.StringAttribute{
				Computed: true,
			},
			"created_at": schema.StringAttribute{
				Computed: true,
			},
			"last_updated": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (s *stashResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_stash"
}

func (s *stashResource) Configure(
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

	s.client = client
}
