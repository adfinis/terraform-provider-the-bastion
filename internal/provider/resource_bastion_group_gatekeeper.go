// Copyright (c) Adfinis
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"slices"

	"github.com/adfinis/terraform-provider-bastion/bastion"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &GroupGatekeeperResource{}
var _ resource.ResourceWithImportState = &GroupGatekeeperResource{}
var _ resource.ResourceWithConfigure = &GroupGatekeeperResource{}

// NewGroupGatekeeperResource is a helper function to simplify the provider implementation.
func NewGroupGatekeeperResource() resource.Resource {
	return &GroupGatekeeperResource{}
}

// GroupGatekeeperResource is the resource implementation.
type GroupGatekeeperResource struct {
	client *bastion.Client
}

// GroupGatekeeperResourceModel describes the resource data model.
type GroupGatekeeperResourceModel struct {
	Group   types.String `tfsdk:"group"`
	Account types.String `tfsdk:"account"`
	ID      types.String `tfsdk:"id"`
}

// Metadata returns the resource type name.
func (r *GroupGatekeeperResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group_gatekeeper"
}

// Schema defines the schema for the resource.
func (r *GroupGatekeeperResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Bastion group gatekeeper membership",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The resource identifier (group:account)",
				Computed:            true,
			},
			"group": schema.StringAttribute{
				MarkdownDescription: "The name of the Bastion group",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"account": schema.StringAttribute{
				MarkdownDescription: "The account name to add as a gatekeeper",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

// Configure adds the bastion client to the resource.
func (r *GroupGatekeeperResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*bastion.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *bastion.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create creates the resource and sets the initial Terraform state.
func (r *GroupGatekeeperResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan GroupGatekeeperResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.GroupAddGatekeeper(plan.Group.ValueString(), plan.Account.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Adding Group Gatekeeper",
			fmt.Sprintf("Could not add gatekeeper %s to group %s: %s", plan.Account.ValueString(), plan.Group.ValueString(), err.Error()),
		)
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s:%s", plan.Group.ValueString(), plan.Account.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *GroupGatekeeperResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state GroupGatekeeperResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	group, err := r.client.GroupInfo(state.Group.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Group Information",
			fmt.Sprintf("Could not read group %s: %s", state.Group.ValueString(), err.Error()),
		)
		return
	}

	if !slices.Contains(group.Gatekeepers, state.Account.ValueString()) {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *GroupGatekeeperResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Since both group and account require replacement, this should never be called
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Group gatekeeper membership cannot be updated. This is a bug in the provider.",
	)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *GroupGatekeeperResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state GroupGatekeeperResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.GroupRemoveGatekeeper(state.Group.ValueString(), state.Account.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Removing Group Gatekeeper",
			fmt.Sprintf("Could not remove gatekeeper %s from group %s: %s", state.Account.ValueString(), state.Group.ValueString(), err.Error()),
		)
		return
	}
}

// ImportState imports the resource state.
func (r *GroupGatekeeperResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importID := req.ID
	parts := []string{}
	for i, part := range []rune(importID) {
		if part == ':' {
			parts = []string{importID[:i], importID[i+1:]}
			break
		}
	}

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in the format 'group:account', got: %s", importID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("group"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("account"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), importID)...)
}
