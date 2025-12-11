// Copyright (c) Adfinis
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/adfinis/terraform-provider-bastion/bastion"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &AccountPIVPolicyResource{}
var _ resource.ResourceWithImportState = &AccountPIVPolicyResource{}
var _ resource.ResourceWithConfigure = &AccountPIVPolicyResource{}

// NewAccountPIVPolicyResource is a helper function to simplify the provider implementation.
func NewAccountPIVPolicyResource() resource.Resource {
	return &AccountPIVPolicyResource{}
}

// AccountPIVPolicyResource is the resource implementation.
type AccountPIVPolicyResource struct {
	client *bastion.Client
}

// AccountPIVPolicyResourceModel describes the resource data model.
type AccountPIVPolicyResourceModel struct {
	Account types.String `tfsdk:"account"`
	Policy  types.String `tfsdk:"policy"`
}

// Metadata returns the resource type name.
func (r *AccountPIVPolicyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_account_piv_policy"
}

// Schema defines the schema for the resource.
func (r *AccountPIVPolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the PIV (Personal Identity Verification) policy for a Bastion account's ingress keys. " +
			"This resource manages stable PIV policies. For temporary grace periods, use the `ephemeral:bastion_account_piv_grace` ephemeral resource.",
		Attributes: map[string]schema.Attribute{
			"account": schema.StringAttribute{
				MarkdownDescription: "The name of the Bastion account",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"policy": schema.StringAttribute{
				MarkdownDescription: "The PIV policy for the account. Valid values: `default` (use bastion global policy), " +
					"`enforce` (only PIV-verified keys allowed, disables non-PIV keys immediately), " +
					"`never` (never require PIV keys regardless of global setting). " +
					"Note: The `grace` policy must be managed using the ephemeral resource.",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("default"),
				Validators: []validator.String{
					stringvalidator.OneOf("default", "enforce", "never"),
				},
			},
		},
	}
}

// Configure adds the bastion client to the resource.
func (r *AccountPIVPolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *AccountPIVPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AccountPIVPolicyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy := bastion.PIVPolicy(plan.Policy.ValueString())
	err := r.client.AccountSetPIVPolicy(plan.Account.ValueString(), policy)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Setting Account PIV Policy",
			fmt.Sprintf("Could not set PIV policy for account %s: %s", plan.Account.ValueString(), err.Error()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *AccountPIVPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AccountPIVPolicyResourceModel

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

	if account.IngressPIVGrace.Enabled.Bool() {
		resp.Diagnostics.AddWarning(
			"Account has PIV grace policy active",
			fmt.Sprintf("Account %s currently has a grace policy active. This should be managed via the ephemeral resource, not this stable policy resource. The state will reflect the underlying stable policy once grace expires.", state.Account.ValueString()),
		)
	}

	if account.IngressPIVPolicy != "" {
		state.Policy = types.StringValue(string(account.IngressPIVPolicy))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *AccountPIVPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan AccountPIVPolicyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy := bastion.PIVPolicy(plan.Policy.ValueString())
	err := r.client.AccountSetPIVPolicy(plan.Account.ValueString(), policy)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Account PIV Policy",
			fmt.Sprintf("Could not update PIV policy for account %s: %s", plan.Account.ValueString(), err.Error()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *AccountPIVPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AccountPIVPolicyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Reset to default policy on delete
	err := r.client.AccountSetPIVPolicy(state.Account.ValueString(), bastion.PIVPolicyDefault)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Resetting Account PIV Policy",
			fmt.Sprintf("Could not reset PIV policy for account %s: %s", state.Account.ValueString(), err.Error()),
		)
		return
	}
}

// ImportState imports the resource state.
func (r *AccountPIVPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("account"), req, resp)
}
