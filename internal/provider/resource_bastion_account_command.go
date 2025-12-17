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

var _ resource.Resource = &AccountCommandResource{}
var _ resource.ResourceWithImportState = &AccountCommandResource{}
var _ resource.ResourceWithConfigure = &AccountCommandResource{}

// NewAccountCommandResource is a helper function to simplify the provider implementation.
func NewAccountCommandResource() resource.Resource {
	return &AccountCommandResource{}
}

// AccountCommandResource is the resource implementation.
type AccountCommandResource struct {
	client *bastion.Client
}

// AccountCommandResourceModel describes the resource data model.
type AccountCommandResourceModel struct {
	Account types.String `tfsdk:"account"`
	Command types.String `tfsdk:"command"`
	ID      types.String `tfsdk:"id"`
}

// Metadata returns the resource type name.
func (r *AccountCommandResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_account_command"
}

// Schema defines the schema for the resource.
func (r *AccountCommandResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Bastion account command grant",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The resource identifier (account:command)",
				Computed:            true,
			},
			"account": schema.StringAttribute{
				MarkdownDescription: "The name of the Bastion account",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"command": schema.StringAttribute{
				MarkdownDescription: "The command to grant to the account",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

// Configure adds the bastion client to the resource.
func (r *AccountCommandResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *AccountCommandResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AccountCommandResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.AccountGrantCommand(plan.Account.ValueString(), plan.Command.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Granting Account Command",
			fmt.Sprintf("Could not grant command %s to account %s: %s", plan.Command.ValueString(), plan.Account.ValueString(), err.Error()),
		)
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s:%s", plan.Account.ValueString(), plan.Command.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *AccountCommandResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AccountCommandResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	account, err := r.client.AccountInfo(state.Account.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Account Information",
			fmt.Sprintf("Could not read account %s: %s", state.Account.ValueString(), err.Error()),
		)
		return
	}

	// auditor is special
	if state.Command.ValueString() == "auditor" {
		if !account.IsAuditor.Bool() {
			resp.State.RemoveResource(ctx)
			return
		}
	} else {
		if !slices.Contains(account.AllowedCommands, state.Command.ValueString()) {
			resp.State.RemoveResource(ctx)
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *AccountCommandResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Since both account and command require replacement, this should never be called
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Account command grants cannot be updated. This is a bug in the provider.",
	)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *AccountCommandResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AccountCommandResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.AccountRevokeCommand(state.Account.ValueString(), state.Command.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Revoking Account Command",
			fmt.Sprintf("Could not revoke command %s from account %s: %s", state.Command.ValueString(), state.Account.ValueString(), err.Error()),
		)
		return
	}
}

// ImportState imports the resource state.
func (r *AccountCommandResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
			fmt.Sprintf("Expected import ID in the format 'account:command', got: %s", importID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("account"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("command"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), importID)...)
}
