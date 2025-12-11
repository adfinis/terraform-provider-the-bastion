// Copyright (c) Adfinis
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/adfinis/terraform-provider-bastion/bastion"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &AccountResource{}
var _ resource.ResourceWithImportState = &AccountResource{}
var _ resource.ResourceWithConfigure = &AccountResource{}

// NewAccountResource is a helper function to simplify the provider implementation.
func NewAccountResource() resource.Resource {
	return &AccountResource{}
}

// AccountResource is the resource implementation.
type AccountResource struct {
	client *bastion.Client
}

// AccountResourceModel describes the resource data model.
type AccountResourceModel struct {
	Account                     types.String `tfsdk:"account"`
	UID                         types.Int64  `tfsdk:"uid"`
	UIDAuto                     types.Bool   `tfsdk:"uid_auto"`
	PublicKey                   types.String `tfsdk:"public_key"`
	NoKey                       types.Bool   `tfsdk:"no_key"`
	ImmutableKey                types.Bool   `tfsdk:"immutable_key"`
	Comment                     types.String `tfsdk:"comment"`
	TTL                         types.Int64  `tfsdk:"ttl"`
	AlwaysActive                types.Bool   `tfsdk:"always_active"`
	OshOnly                     types.Bool   `tfsdk:"osh_only"`
	MaxInactiveDays             types.Int64  `tfsdk:"max_inactive_days"`
	PamAuthBypass               types.Bool   `tfsdk:"pam_auth_bypass"`
	MFAPasswordRequired         types.String `tfsdk:"mfa_password_required"`
	MFATOTPRequired             types.String `tfsdk:"mfa_totp_required"`
	EgressStrictHostKeyChecking types.String `tfsdk:"egress_strict_host_key_checking"`
	EgressSessionMultiplexing   types.String `tfsdk:"egress_session_multiplexing"`
	PersonalEgressMFARequired   types.String `tfsdk:"personal_egress_mfa_required"`
	IdleIgnore                  types.Bool   `tfsdk:"idle_ignore"`
	PubkeyAuthOptional          types.Bool   `tfsdk:"pubkey_auth_optional"`
}

func (r *AccountResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_account"
}

// Schema defines the schema for the resource.
func (r *AccountResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Bastion account",
		Attributes: map[string]schema.Attribute{
			"account": schema.StringAttribute{
				MarkdownDescription: "The name of the Bastion account",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"uid": schema.Int64Attribute{
				MarkdownDescription: "The UID of the Bastion account. Mutually exclusive with uid_auto.",
				Optional:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				WriteOnly: true,
			},
			"uid_auto": schema.BoolAttribute{
				MarkdownDescription: "Whether to automatically assign a UID. Mutually exclusive with uid.",
				Optional:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
				WriteOnly: true,
			},
			"public_key": schema.StringAttribute{
				MarkdownDescription: "The public key to assign to the account upon creation.",
				Optional:            true,
				WriteOnly:           true,
			},
			"no_key": schema.BoolAttribute{
				MarkdownDescription: "Whether to create the account without an initial public key.",
				Optional:            true,
				WriteOnly:           true,
			},
			"immutable_key": schema.BoolAttribute{
				MarkdownDescription: "Whether the account's public key is immutable.",
				Optional:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "A comment for the account.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ttl": schema.Int64Attribute{
				MarkdownDescription: "Time to live for the account in seconds.",
				Optional:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"always_active": schema.BoolAttribute{
				MarkdownDescription: "Whether the account is always active.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"osh_only": schema.BoolAttribute{
				MarkdownDescription: "Whether the account can only use osh (bastion) commands.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"max_inactive_days": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of days of inactivity before the account is considered inactive.",
				Optional:            true,
			},
			"pam_auth_bypass": schema.BoolAttribute{
				MarkdownDescription: "Whether PAM authentication is bypassed for this account.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"mfa_password_required": schema.StringAttribute{
				MarkdownDescription: "MFA password policy. Valid values: yes, no, bypass.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("no"),
				Validators: []validator.String{
					stringvalidator.OneOf("yes", "no", "bypass"),
				},
			},
			"mfa_totp_required": schema.StringAttribute{
				MarkdownDescription: "MFA TOTP policy. Valid values: yes, no, bypass.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("no"),
				Validators: []validator.String{
					stringvalidator.OneOf("yes", "no", "bypass"),
				},
			},
			"egress_strict_host_key_checking": schema.StringAttribute{
				MarkdownDescription: "Egress strict host key checking policy. Valid values: yes, accept-new, no, ask, default, bypass.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("yes", "accept-new", "no", "ask", "default", "bypass"),
				},
			},
			"egress_session_multiplexing": schema.StringAttribute{
				MarkdownDescription: "Egress session multiplexing policy. Valid values: yes, no, default.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("yes", "no", "default"),
				},
			},
			"personal_egress_mfa_required": schema.StringAttribute{
				MarkdownDescription: "Personal egress MFA policy. Valid values: password, totp, any, none.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("none"),
				Validators: []validator.String{
					stringvalidator.OneOf("password", "totp", "any", "none"),
				},
			},
			"idle_ignore": schema.BoolAttribute{
				MarkdownDescription: "Whether to ignore idle timeouts for this account.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"pubkey_auth_optional": schema.BoolAttribute{
				MarkdownDescription: "Whether public key authentication is optional for this account.",
				Optional:            true,
			},
		},
	}
}

// Configure adds the bastion client to the resource.
func (r *AccountResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *AccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AccountResourceModel
	var config AccountResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.UID.IsNull() && !config.UIDAuto.IsNull() {
		resp.Diagnostics.AddError(
			"Conflicting Configuration",
			"Cannot specify both 'uid' and 'uid_auto'. Please specify only one.",
		)
		return
	}

	if config.UID.IsNull() && config.UIDAuto.IsNull() {
		resp.Diagnostics.AddError(
			"Missing Configuration",
			"Either 'uid' or 'uid_auto' must be specified.",
		)
		return
	}

	var uidO bastion.UIDOpt
	if !config.UIDAuto.IsNull() && config.UIDAuto.ValueBool() {
		uidO = bastion.WithAutoUID()
	} else if !config.UID.IsNull() {
		uidO = bastion.WithSpecificUID(uint(config.UID.ValueInt64()))
	}

	if config.NoKey.IsNull() && config.PublicKey.IsNull() {
		resp.Diagnostics.AddError(
			"Missing Public Key",
			"Either 'public_key' must be set or 'no_key' must be true to create the account without an initial public key.",
		)
		return
	}

	if !config.NoKey.IsNull() && !config.PublicKey.IsNull() {
		resp.Diagnostics.AddError(
			"Conflicting Configuration",
			"Cannot specify both 'public_key' and 'no_key'. Please specify only one.",
		)
		return
	}

	createOpts := &bastion.CreateAccountOptions{}
	if !config.NoKey.IsNull() {
		createOpts.NoKey = config.NoKey.ValueBool()
	}

	if !config.PublicKey.IsNull() {
		createOpts.PublicKey = config.PublicKey.ValueString()
	}

	if !plan.AlwaysActive.IsNull() {
		createOpts.AlwaysActive = plan.AlwaysActive.ValueBool()
	}

	if !plan.OshOnly.IsNull() {
		createOpts.OshOnly = plan.OshOnly.ValueBool()
	}

	if !plan.ImmutableKey.IsNull() {
		createOpts.ImmutableKey = plan.ImmutableKey.ValueBool()
	}

	if !plan.MaxInactiveDays.IsNull() {
		createOpts.MaxInactiveDays = uint(plan.MaxInactiveDays.ValueInt64())
	}

	if !plan.Comment.IsNull() {
		createOpts.Comment = plan.Comment.ValueString()
	}

	if !plan.TTL.IsNull() {
		createOpts.TTL = int(plan.TTL.ValueInt64())
	}

	if err := r.client.CreateAccount(plan.Account.ValueString(), uidO, createOpts); err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Account",
			fmt.Sprintf("Could not create account %s: %s", plan.Account.ValueString(), err.Error()),
		)
		return
	}

	// Apply modify-only options if specified
	modifyOpts := &bastion.ModifyAccountOptions{}
	needsModify := false

	if !plan.PamAuthBypass.IsNull() {
		val := plan.PamAuthBypass.ValueBool()
		modifyOpts.PamAuthBypass = &val
		needsModify = true
	}

	if !plan.MFAPasswordRequired.IsNull() {
		val := bastion.YesNoBypass(plan.MFAPasswordRequired.ValueString())
		modifyOpts.MFAPasswordRequired = &val
		needsModify = true
	}

	if !plan.MFATOTPRequired.IsNull() {
		val := bastion.YesNoBypass(plan.MFATOTPRequired.ValueString())
		modifyOpts.MFATOTPRequired = &val
		needsModify = true
	}

	if !plan.EgressStrictHostKeyChecking.IsNull() {
		val := bastion.EgressStrictHostKeyCheckingPolicy(plan.EgressStrictHostKeyChecking.ValueString())
		modifyOpts.EgressStrictHostKeyChecking = &val
		needsModify = true
	}

	if !plan.EgressSessionMultiplexing.IsNull() {
		val := bastion.YesNoDefault(plan.EgressSessionMultiplexing.ValueString())
		modifyOpts.EgressSessionMultiplexing = &val
		needsModify = true
	}

	if !plan.PersonalEgressMFARequired.IsNull() {
		val := bastion.MFARequiredPolicy(plan.PersonalEgressMFARequired.ValueString())
		modifyOpts.PersonalEgressMFARequired = &val
		needsModify = true
	}

	if !plan.IdleIgnore.IsNull() {
		val := plan.IdleIgnore.ValueBool()
		modifyOpts.IdleIgnore = &val
		needsModify = true
	}

	if !plan.PubkeyAuthOptional.IsNull() {
		val := plan.PubkeyAuthOptional.ValueBool()
		modifyOpts.PubkeyAuthOptional = &val
		needsModify = true
	}

	if !plan.MaxInactiveDays.IsNull() {
		val := int(plan.MaxInactiveDays.ValueInt64())
		modifyOpts.MaxInactiveDays = &val
		needsModify = true
	}

	if needsModify {
		if err := r.client.ModifyAccount(plan.Account.ValueString(), modifyOpts); err != nil {
			resp.Diagnostics.AddError(
				"Error Modifying Account After Creation",
				fmt.Sprintf("Could not modify account %s: %s", plan.Account.ValueString(), err.Error()),
			)

			// delete the account to avoid orphaned resources
			delErr := r.client.DeleteAccount(plan.Account.ValueString())
			if delErr != nil {
				resp.Diagnostics.AddError(
					"Error Cleaning Up After Failed Account Modify",
					fmt.Sprintf("Could not delete account %s after failed modify: %s", plan.Account.ValueString(), delErr.Error()),
				)
			}
			return
		}
	}

	account, err := r.client.AccountInfo(plan.Account.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Retrieving Account Info",
			fmt.Sprintf("Could not retrieve info for account %s: %s", plan.Account.ValueString(), err.Error()),
		)
		return
	}

	plan.Account = types.StringValue(account.Account)
	plan.AlwaysActive = types.BoolValue(account.AlwaysActive.Bool())
	plan.OshOnly = types.BoolValue(account.OshOnly.Bool())
	plan.PamAuthBypass = types.BoolValue(account.PamAuthBypass.Bool())
	plan.IdleIgnore = types.BoolValue(account.IdleIgnore.Bool())
	plan.PersonalEgressMFARequired = types.StringValue(string(account.PersonalEgressMFARequired))

	// Map MFA settings
	if account.MFAPasswordBypass.Bool() {
		plan.MFAPasswordRequired = types.StringValue("bypass")
	} else if account.MFAPasswordRequired.Bool() {
		plan.MFAPasswordRequired = types.StringValue("yes")
	} else {
		plan.MFAPasswordRequired = types.StringValue("no")
	}

	if account.MFATOTPBypass.Bool() {
		plan.MFATOTPRequired = types.StringValue("bypass")
	} else if account.MFATOTPRequired.Bool() {
		plan.MFATOTPRequired = types.StringValue("yes")
	} else {
		plan.MFATOTPRequired = types.StringValue("no")
	}

	if account.MaxInactiveDays != "" {
		maxInactiveDays, err := strconv.Atoi(account.MaxInactiveDays)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Converting Max Inactive Days",
				fmt.Sprintf("Could not convert max_inactive_days '%s' to integer: %s", account.MaxInactiveDays, err.Error()),
			)
			return
		}
		plan.MaxInactiveDays = types.Int64Value(int64(maxInactiveDays))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *AccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AccountResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	account, err := r.client.AccountInfo(state.Account.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Read Account",
			"Could not read account: "+err.Error(),
		)
		return
	}

	// Map response body to model
	state.Account = types.StringValue(account.Account)
	state.AlwaysActive = types.BoolValue(account.AlwaysActive.Bool())
	state.OshOnly = types.BoolValue(account.OshOnly.Bool())
	state.PamAuthBypass = types.BoolValue(account.PamAuthBypass.Bool())
	state.IdleIgnore = types.BoolValue(account.IdleIgnore.Bool())

	// Map MFA settings
	if account.MFAPasswordBypass.Bool() {
		state.MFAPasswordRequired = types.StringValue("bypass")
	} else if account.MFAPasswordRequired.Bool() {
		state.MFAPasswordRequired = types.StringValue("yes")
	} else {
		state.MFAPasswordRequired = types.StringValue("no")
	}

	if account.MFATOTPBypass.Bool() {
		state.MFATOTPRequired = types.StringValue("bypass")
	} else if account.MFATOTPRequired.Bool() {
		state.MFATOTPRequired = types.StringValue("yes")
	} else {
		state.MFATOTPRequired = types.StringValue("no")
	}

	if account.PersonalEgressMFARequired != "" {
		state.PersonalEgressMFARequired = types.StringValue(string(account.PersonalEgressMFARequired))
	}

	if account.MaxInactiveDays != "" {
		maxInactiveDays, err := strconv.Atoi(account.MaxInactiveDays)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Converting Max Inactive Days",
				fmt.Sprintf("Could not convert max_inactive_days '%s' to integer: %s", account.MaxInactiveDays, err.Error()),
			)
			return
		}
		state.MaxInactiveDays = types.Int64Value(int64(maxInactiveDays))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *AccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state AccountResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if any modifiable attributes have changed
	if !plan.AlwaysActive.Equal(state.AlwaysActive) ||
		!plan.OshOnly.Equal(state.OshOnly) ||
		!plan.MaxInactiveDays.Equal(state.MaxInactiveDays) ||
		!plan.PamAuthBypass.Equal(state.PamAuthBypass) ||
		!plan.MFAPasswordRequired.Equal(state.MFAPasswordRequired) ||
		!plan.MFATOTPRequired.Equal(state.MFATOTPRequired) ||
		!plan.EgressStrictHostKeyChecking.Equal(state.EgressStrictHostKeyChecking) ||
		!plan.EgressSessionMultiplexing.Equal(state.EgressSessionMultiplexing) ||
		!plan.PersonalEgressMFARequired.Equal(state.PersonalEgressMFARequired) ||
		!plan.IdleIgnore.Equal(state.IdleIgnore) ||
		!plan.PubkeyAuthOptional.Equal(state.PubkeyAuthOptional) {

		modifyOpts := &bastion.ModifyAccountOptions{}

		if !plan.AlwaysActive.IsNull() {
			val := plan.AlwaysActive.ValueBool()
			modifyOpts.AlwaysActive = &val
		}

		if !plan.OshOnly.IsNull() {
			val := plan.OshOnly.ValueBool()
			modifyOpts.OshOnly = &val
		}

		if !plan.MaxInactiveDays.IsNull() {
			val := int(plan.MaxInactiveDays.ValueInt64())
			modifyOpts.MaxInactiveDays = &val
		}

		if !plan.PamAuthBypass.IsNull() {
			val := plan.PamAuthBypass.ValueBool()
			modifyOpts.PamAuthBypass = &val
		}

		if !plan.MFAPasswordRequired.IsNull() {
			val := bastion.YesNoBypass(plan.MFAPasswordRequired.ValueString())
			modifyOpts.MFAPasswordRequired = &val
		}

		if !plan.MFATOTPRequired.IsNull() {
			val := bastion.YesNoBypass(plan.MFATOTPRequired.ValueString())
			modifyOpts.MFATOTPRequired = &val
		}

		if !plan.EgressStrictHostKeyChecking.IsNull() {
			val := bastion.EgressStrictHostKeyCheckingPolicy(plan.EgressStrictHostKeyChecking.ValueString())
			modifyOpts.EgressStrictHostKeyChecking = &val
		}

		if !plan.EgressSessionMultiplexing.IsNull() {
			val := bastion.YesNoDefault(plan.EgressSessionMultiplexing.ValueString())
			modifyOpts.EgressSessionMultiplexing = &val
		}

		if !plan.PersonalEgressMFARequired.IsNull() {
			val := bastion.MFARequiredPolicy(plan.PersonalEgressMFARequired.ValueString())
			modifyOpts.PersonalEgressMFARequired = &val
		}

		if !plan.IdleIgnore.IsNull() {
			val := plan.IdleIgnore.ValueBool()
			modifyOpts.IdleIgnore = &val
		}

		if !plan.PubkeyAuthOptional.IsNull() {
			val := plan.PubkeyAuthOptional.ValueBool()
			modifyOpts.PubkeyAuthOptional = &val
		}

		err := r.client.ModifyAccount(plan.Account.ValueString(), modifyOpts)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Modifying Account",
				fmt.Sprintf("Could not modify account %s: %s", plan.Account.ValueString(), err.Error()),
			)
			return
		}
	}

	// Read back the account to ensure state is consistent
	account, err := r.client.AccountInfo(plan.Account.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Account After Update",
			fmt.Sprintf("Could not read account %s after update: %s", plan.Account.ValueString(), err.Error()),
		)
		return
	}

	plan.Account = types.StringValue(account.Account)
	plan.AlwaysActive = types.BoolValue(account.AlwaysActive.Bool())
	plan.OshOnly = types.BoolValue(account.OshOnly.Bool())
	plan.PamAuthBypass = types.BoolValue(account.PamAuthBypass.Bool())
	plan.IdleIgnore = types.BoolValue(account.IdleIgnore.Bool())

	// Map MFA settings
	if account.MFAPasswordBypass.Bool() {
		plan.MFAPasswordRequired = types.StringValue("bypass")
	} else if account.MFAPasswordRequired.Bool() {
		plan.MFAPasswordRequired = types.StringValue("yes")
	} else {
		plan.MFAPasswordRequired = types.StringValue("no")
	}

	if account.MFATOTPBypass.Bool() {
		plan.MFATOTPRequired = types.StringValue("bypass")
	} else if account.MFATOTPRequired.Bool() {
		plan.MFATOTPRequired = types.StringValue("yes")
	} else {
		plan.MFATOTPRequired = types.StringValue("no")
	}

	if account.PersonalEgressMFARequired != "" {
		plan.PersonalEgressMFARequired = types.StringValue(string(account.PersonalEgressMFARequired))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *AccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AccountResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteAccount(state.Account.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Delete Account",
			"Could not delete account: "+err.Error(),
		)
		return
	}
}

// ImportState imports the resource state.
func (r *AccountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("account"), req, resp)
}
