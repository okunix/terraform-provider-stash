package provider

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
)

type secretResource struct {
	client *stash.Client
}

type secretResourceModel struct {
	StashID     types.String `tfsdk:"stash_id"`
	Name        types.String `tfsdk:"name"`
	Value       types.String `tfsdk:"value"`
	LastUpdated types.String `tfsdk:"last_updated"`
}

var (
	_ resource.Resource              = (*secretResource)(nil)
	_ resource.ResourceWithConfigure = (*secretResource)(nil)
)

func NewSecretResource() resource.Resource {
	return &secretResource{}
}

func (s *secretResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan secretResourceModel
	diag := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	stashID := plan.StashID.ValueString()
	name := plan.Name.ValueString()
	value := plan.Value.ValueString()
	err := s.client.AddSecretsEntry(ctx, stashID, stash.AddSecretRequest{Name: name, Value: value})
	if err != nil {
		resp.Diagnostics.AddError("failed to add a new secret", err.Error())
		return
	}

	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC3339))
	diag = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diag...)
}

func (s *secretResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state secretResourceModel
	diag := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	stashID := state.StashID.ValueString()
	name := state.Name.ValueString()
	if err := s.client.RemoveSecretsEntry(ctx, stashID, name); err != nil {
		resp.Diagnostics.AddError("failed to remove a secret", err.Error())
		return
	}
}

func (s *secretResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan secretResourceModel
	diag := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state secretResourceModel
	diag = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	oldStashID := state.StashID.ValueString()
	oldName := state.Name.ValueString()

	newStashID := plan.StashID.ValueString()
	newName := plan.Name.ValueString()
	newValue := plan.Value.ValueString()

	stashResponse, err := s.client.GetStashByID(ctx, newStashID)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("failed to fetch stash id=%s", newStashID),
			err.Error(),
		)
		return
	}
	if stashResponse.Locked {
		resp.Diagnostics.AddError(
			fmt.Sprintf("failed to add entry to stash id=%s", newStashID),
			"stash is locked",
		)
		return
	}

	if err := s.client.RemoveSecretsEntry(ctx, oldStashID, oldName); err != nil {
		resp.Diagnostics.AddError("failed to remove old entry", err.Error())
		return
	}

	addReq := stash.AddSecretRequest{Name: newName, Value: newValue}
	if err := s.client.AddSecretsEntry(ctx, newStashID, addReq); err != nil {
		resp.Diagnostics.AddError("failed to add new entry", err.Error())
		return
	}

	state.StashID = types.StringValue(newStashID)
	state.Name = plan.Name
	state.Value = plan.Value
	state.LastUpdated = types.StringValue(time.Now().Format(time.RFC3339))

	diag = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diag...)
}

func (s *secretResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state secretResourceModel
	diag := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()
	stashID := state.StashID.ValueString()
	value, err := s.client.GetSecretsEntry(ctx, stashID, name)
	if err != nil {
		resp.Diagnostics.AddError("failed to fetch secret value", err.Error())
		return
	}
	if value != state.Value.ValueString() {
		resp.Diagnostics.AddError("invalid state", "value in stash doesn't match terraform state")
		return
	}

	diag = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diag...)
}

func (s *secretResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_secret"
}

func (s *secretResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"stash_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"value": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
			"last_updated": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (s *secretResource) Configure(
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
