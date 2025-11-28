// Copyright (c) Adfinis
// SPDX-License-Identifier: GPL-3.0-or-later

package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/adfinis/terraform-provider-bastion/bastion"
	"github.com/adfinis/terraform-provider-bastion/internal/provider/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &GroupResource{}
var _ resource.ResourceWithImportState = &GroupResource{}
var _ resource.ResourceWithConfigure = &GroupResource{}

// NewGroupResource is a helper function to simplify the provider implementation.
func NewGroupResource() resource.Resource {
	return &GroupResource{}
}

// GroupResource is the resource implementation.
type GroupResource struct {
	client *bastion.Client
}

// GroupResourceModel describes the resource data model.
type GroupResourceModel struct {
	Group           types.String `tfsdk:"group"`
	Owner           types.String `tfsdk:"owner"`
	KeyAlgo         types.String `tfsdk:"key_algo"`
	MFARequired     types.String `tfsdk:"mfa_required"`
	IdleLockTimeout types.Int64  `tfsdk:"idle_lock_timeout"`
	IdleKillTimeout types.Int64  `tfsdk:"idle_kill_timeout"`
	GuestTtlLimit   types.Int64  `tfsdk:"guest_ttl_limit"`
	Owners          types.List   `tfsdk:"owners"`
	Members         types.List   `tfsdk:"members"`
	Gatekeepers     types.List   `tfsdk:"gatekeepers"`
	ACLKeepers      types.List   `tfsdk:"aclkeepers"`
}

// Metadata returns the resource type name.
func (r *GroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

// Schema defines the schema for the resource.
func (r *GroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Bastion group",
		Attributes: map[string]schema.Attribute{
			"group": schema.StringAttribute{
				MarkdownDescription: "The name of the Bastion group",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"owner": schema.StringAttribute{
				MarkdownDescription: "The initial owner of the Bastion group (only used during creation)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"key_algo": schema.StringAttribute{
				MarkdownDescription: "The SSH key algorithm for the group's initial key. Valid values: ed25519, rsa2048, rsa4096, rsa8192, ecdsa256, ecdsa384, ecdsa521. Defaults to ed25519. This value is only used during creation and cannot be changed afterward.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("ed25519"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"mfa_required": schema.StringAttribute{
				MarkdownDescription: "MFA policy for the group. Valid values: password, totp, any, none. If not specified, the group's current setting is preserved.",
				Optional:            true,
			},
			"idle_lock_timeout": schema.Int64Attribute{
				MarkdownDescription: "Idle lock timeout in seconds. After this duration of inactivity, the session will be locked.",
				Optional:            true,
			},
			"idle_kill_timeout": schema.Int64Attribute{
				MarkdownDescription: "Idle kill timeout in seconds. After this duration of inactivity, the session will be terminated.",
				Optional:            true,
			},
			"guest_ttl_limit": schema.Int64Attribute{
				MarkdownDescription: "Maximum TTL (time to live) for guest accesses in seconds.",
				Optional:            true,
			},
			"owners": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "The owners of the Bastion group",
				Computed:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"members": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "The members of the Bastion group",
				Computed:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"gatekeepers": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "The gatekeepers of the Bastion group",
				Computed:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"aclkeepers": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "The ACL keepers of the Bastion group",
				Computed:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the bastion client to the resource.
func (r *GroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *GroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan GroupResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	keyAlgo := bastion.KeyAlgo(plan.KeyAlgo.ValueString())
	_, err := r.client.CreateGroup(plan.Group.ValueString(), plan.Owner.ValueString(), keyAlgo)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Bastion Group",
			fmt.Sprintf("Could not create group %s: %s", plan.Group.ValueString(), err.Error()),
		)
		return
	}

	// Apply modify-only options if specified
	modifyOpts := &bastion.GroupModifyOptions{}
	needsModify := false

	if !plan.MFARequired.IsNull() {
		mfaPolicy := bastion.MFARequiredPolicy(plan.MFARequired.ValueString())
		modifyOpts.MFARequired = &mfaPolicy
		needsModify = true
	}

	if !plan.IdleLockTimeout.IsNull() {
		modifyOpts.IdleLockTimeout = utils.ToPtr(fmt.Sprintf("%d", plan.IdleLockTimeout.ValueInt64()))
		needsModify = true
	}

	if !plan.IdleKillTimeout.IsNull() {
		modifyOpts.IdleKillTimeout = utils.ToPtr(fmt.Sprintf("%d", plan.IdleKillTimeout.ValueInt64()))
		needsModify = true
	}

	if !plan.GuestTtlLimit.IsNull() {
		modifyOpts.GuestTtlLimit = utils.ToPtr(fmt.Sprintf("%d", plan.GuestTtlLimit.ValueInt64()))
		needsModify = true
	}

	if needsModify {
		err := r.client.ModifyGroup(plan.Group.ValueString(), modifyOpts)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Modifying Bastion Group After Creation",
				fmt.Sprintf("Could not modify group %s: %s", plan.Group.ValueString(), err.Error()),
			)
			return
		}
	}

	// because the createGroup call doesn't return the same data structure as groupInfo, we need to call groupInfo to get the full data
	group, err := r.client.GroupInfo(plan.Group.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Retrieving Bastion Group After Creation",
			fmt.Sprintf("Could not read group %s after creation: %s", plan.Group.ValueString(), err.Error()),
		)
		return
	}

	plan.Group = types.StringValue(group.Group)

	owners, diags := types.ListValueFrom(ctx, types.StringType, group.Owners)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Owners = owners

	members, diags := types.ListValueFrom(ctx, types.StringType, group.Members)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Members = members

	gatekeepers, diags := types.ListValueFrom(ctx, types.StringType, group.Gatekeepers)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Gatekeepers = gatekeepers

	aclkeepers, diags := types.ListValueFrom(ctx, types.StringType, group.ACLKeepers)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ACLKeepers = aclkeepers

	if group.MFARequired != nil {
		plan.MFARequired = types.StringValue(string(*group.MFARequired))
	}

	if group.IdleLockTimeout != nil {
		idleLockTimeout, err := strconv.ParseInt(*group.IdleLockTimeout, 10, 64)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Parsing Idle Lock Timeout",
				fmt.Sprintf("Could not parse idle lock timeout for group %s: %s", plan.Group.ValueString(), err.Error()),
			)
			return
		}
		plan.IdleLockTimeout = types.Int64Value(idleLockTimeout)
	}

	if group.IdleKillTimeout != nil {
		idleKillTimeout, err := strconv.ParseInt(*group.IdleKillTimeout, 10, 64)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Parsing Idle Kill Timeout",
				fmt.Sprintf("Could not parse idle kill timeout for group %s: %s", plan.Group.ValueString(), err.Error()),
			)
			return
		}
		plan.IdleKillTimeout = types.Int64Value(idleKillTimeout)
	}

	if group.GuestTtlLimit != nil {
		guestTtlLimit, err := strconv.ParseInt(*group.GuestTtlLimit, 10, 64)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Parsing Guest TTL Limit",
				fmt.Sprintf("Could not parse guest TTL limit for group %s: %s", plan.Group.ValueString(), err.Error()),
			)
			return
		}
		plan.GuestTtlLimit = types.Int64Value(guestTtlLimit)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *GroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state GroupResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	group, err := r.client.GroupInfo(state.Group.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Bastion Group",
			fmt.Sprintf("Could not read group %s: %s", state.Group.ValueString(), err.Error()),
		)
		return
	}

	state.Group = types.StringValue(group.Group)

	owners, diags := types.ListValueFrom(ctx, types.StringType, group.Owners)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Owners = owners

	members, diags := types.ListValueFrom(ctx, types.StringType, group.Members)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Members = members

	gatekeepers, diags := types.ListValueFrom(ctx, types.StringType, group.Gatekeepers)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Gatekeepers = gatekeepers

	aclkeepers, diags := types.ListValueFrom(ctx, types.StringType, group.ACLKeepers)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.ACLKeepers = aclkeepers

	if group.MFARequired != nil {
		state.MFARequired = types.StringValue(string(*group.MFARequired))
	}

	if group.IdleLockTimeout != nil {
		idleLockTimeout, err := strconv.ParseInt(*group.IdleLockTimeout, 10, 64)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Parsing Idle Lock Timeout",
				fmt.Sprintf("Could not parse idle lock timeout for group %s: %s", state.Group.ValueString(), err.Error()),
			)
			return
		}
		state.IdleLockTimeout = types.Int64Value(idleLockTimeout)
	}

	if group.IdleKillTimeout != nil {
		idleKillTimeout, err := strconv.ParseInt(*group.IdleKillTimeout, 10, 64)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Parsing Idle Kill Timeout",
				fmt.Sprintf("Could not parse idle kill timeout for group %s: %s", state.Group.ValueString(), err.Error()),
			)
			return
		}
		state.IdleKillTimeout = types.Int64Value(idleKillTimeout)
	}

	if group.GuestTtlLimit != nil {
		guestTtlLimit, err := strconv.ParseInt(*group.GuestTtlLimit, 10, 64)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Parsing Guest TTL Limit",
				fmt.Sprintf("Could not parse guest TTL limit for group %s: %s", state.Group.ValueString(), err.Error()),
			)
			return
		}
		state.GuestTtlLimit = types.Int64Value(guestTtlLimit)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *GroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state GroupResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if any modifiable attributes have changed
	if !plan.MFARequired.Equal(state.MFARequired) ||
		!plan.IdleLockTimeout.Equal(state.IdleLockTimeout) ||
		!plan.IdleKillTimeout.Equal(state.IdleKillTimeout) ||
		!plan.GuestTtlLimit.Equal(state.GuestTtlLimit) {

		modifyOpts := &bastion.GroupModifyOptions{}

		if !plan.MFARequired.IsNull() {
			mfaPolicy := bastion.MFARequiredPolicy(plan.MFARequired.ValueString())
			modifyOpts.MFARequired = &mfaPolicy
		}

		if !plan.IdleLockTimeout.IsNull() {
			modifyOpts.IdleLockTimeout = utils.ToPtr(fmt.Sprintf("%d", plan.IdleLockTimeout.ValueInt64()))
		}

		if !plan.IdleKillTimeout.IsNull() {
			modifyOpts.IdleKillTimeout = utils.ToPtr(fmt.Sprintf("%d", plan.IdleKillTimeout.ValueInt64()))
		}

		if !plan.GuestTtlLimit.IsNull() {
			modifyOpts.GuestTtlLimit = utils.ToPtr(fmt.Sprintf("%d", plan.GuestTtlLimit.ValueInt64()))
		}

		err := r.client.ModifyGroup(plan.Group.ValueString(), modifyOpts)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Modifying Bastion Group",
				fmt.Sprintf("Could not modify group %s: %s", plan.Group.ValueString(), err.Error()),
			)
			return
		}
	}

	group, err := r.client.GroupInfo(plan.Group.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Bastion Group",
			fmt.Sprintf("Could not read group %s: %s", plan.Group.ValueString(), err.Error()),
		)
		return
	}

	plan.Group = types.StringValue(group.Group)

	owners, diags := types.ListValueFrom(ctx, types.StringType, group.Owners)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Owners = owners

	members, diags := types.ListValueFrom(ctx, types.StringType, group.Members)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Members = members

	gatekeepers, diags := types.ListValueFrom(ctx, types.StringType, group.Gatekeepers)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Gatekeepers = gatekeepers

	aclkeepers, diags := types.ListValueFrom(ctx, types.StringType, group.ACLKeepers)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ACLKeepers = aclkeepers

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *GroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state GroupResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DestroyGroup(state.Group.ValueString())
	if err != nil {
		// If DestroyGroup fails, try DeleteGroup
		err = r.client.DeleteGroup(state.Group.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Deleting Bastion Group",
				fmt.Sprintf("Could not delete group %s: %s", state.Group.ValueString(), err.Error()),
			)
			return
		}
	}
}

// ImportState imports the resource state.
func (r *GroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("group"), req, resp)
}
