// Copyright (c) Adfinis
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/adfinis/terraform-provider-bastion/bastion"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &GroupGuestAccessResource{}
var _ resource.ResourceWithConfigure = &GroupGuestAccessResource{}
var _ resource.ResourceWithImportState = &GroupGuestAccessResource{}

// NewGroupGuestAccessResource is a helper function to simplify the provider implementation.
func NewGroupGuestAccessResource() resource.Resource {
	return &GroupGuestAccessResource{}
}

// GroupGuestAccessResource is the resource implementation.
type GroupGuestAccessResource struct {
	client *bastion.Client
}

// GroupGuestAccessResourceModel describes the resource data model.
type GroupGuestAccessResourceModel struct {
	ID         types.String `tfsdk:"id"`
	Group      types.String `tfsdk:"group"`
	Account    types.String `tfsdk:"account"`
	IP         types.String `tfsdk:"ip"`
	Port       types.String `tfsdk:"port"`
	User       types.String `tfsdk:"user"`
	Protocol   types.String `tfsdk:"protocol"`
	ProxyIP    types.String `tfsdk:"proxy_ip"`
	ProxyPort  types.String `tfsdk:"proxy_port"`
	ProxyUser  types.String `tfsdk:"proxy_user"`
	Comment    types.String `tfsdk:"comment"`
	TTL        types.Int64  `tfsdk:"ttl"`
	RemotePort types.Int64  `tfsdk:"remote_port"`
}

// Metadata returns the resource type name.
func (r *GroupGuestAccessResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group_guest_access"
}

// Schema defines the schema for the resource.
func (r *GroupGuestAccessResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages a Bastion group guest access.

A guest access grants a specific account access to a subset of a group's server accesses.
Note that the server access must exist in the group before granting guest access to it.

Some features like proxyjump accesses and port forwardings are only supported when running [The Bastion fork](https://github.com/adfinis-forks/the-bastion) from Adfinis.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The resource identifier",
				Computed:            true,
			},
			"group": schema.StringAttribute{
				MarkdownDescription: "The Bastion group name owning the server access",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"account": schema.StringAttribute{
				MarkdownDescription: "The account to grant guest access to",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ip": schema.StringAttribute{
				MarkdownDescription: "IP or subnet of the server access target (hostname does not work)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"port": schema.StringAttribute{
				MarkdownDescription: "Port of the access target, use '*' to allow ssh access to all ports",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user": schema.StringAttribute{
				MarkdownDescription: "Username for the server access. Cannot be used together with `protocol`",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("protocol")),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"protocol": schema.StringAttribute{
				MarkdownDescription: "Protocol to grant access for. Valid values are 'sftp', 'scpupload', 'scpdownload', 'rsync', 'portforward'. When set, 'user' must be empty.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("user")),
					stringvalidator.OneOf("sftp", "scpupload", "scpdownload", "rsync", "portforward"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"proxy_ip": schema.StringAttribute{
				MarkdownDescription: "IP address of the proxy server",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.MatchRoot("proxy_port")),
					stringvalidator.AlsoRequires(path.MatchRoot("proxy_user")),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"proxy_port": schema.StringAttribute{
				MarkdownDescription: "Port of the proxy server",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"proxy_user": schema.StringAttribute{
				MarkdownDescription: "Username for the proxy server, use '*' to allow all users",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Comment for the guest access",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ttl": schema.Int64Attribute{
				MarkdownDescription: "Time to live for the guest access in seconds",
				Optional:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"remote_port": schema.Int64Attribute{
				MarkdownDescription: "Remote port forwarded from the target server to The Bastion",
				Optional:            true,
				Validators:          []validator.Int64{},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *GroupGuestAccessResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *GroupGuestAccessResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan GroupGuestAccessResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// either user or protocol must exist but not both
	if (plan.User.IsNull() && plan.Protocol.IsNull()) || (!plan.User.IsNull() && !plan.Protocol.IsNull()) {
		resp.Diagnostics.AddError(
			"Error Creating Group Guest Access",
			"Either 'user' or 'protocol' must be set, but not both.",
		)
		return
	}

	// Build options
	options := &bastion.GroupAddGuestAccessOptions{}

	if !plan.Comment.IsNull() {
		options.Comment = plan.Comment.ValueString()
	}

	if !plan.TTL.IsNull() {
		options.TTL = strconv.FormatInt(plan.TTL.ValueInt64(), 10)
	}

	if !plan.Protocol.IsNull() {
		options.Protocol = plan.Protocol.ValueString()
	}

	// Handle proxy options
	if !plan.ProxyIP.IsNull() || !plan.ProxyPort.IsNull() || !plan.ProxyUser.IsNull() {
		options.ProxyOptions = &bastion.ProxyOptions{
			ProxyHost: plan.ProxyIP.ValueString(),
			ProxyPort: plan.ProxyPort.ValueString(),
			ProxyUser: plan.ProxyUser.ValueString(),
		}
	}

	if !plan.RemotePort.IsNull() {
		remotePort := int(plan.RemotePort.ValueInt64())
		options.RemotePort = &remotePort
	}

	// Add the guest access
	err := r.client.GroupAddGuestAccess(
		plan.Group.ValueString(),
		plan.Account.ValueString(),
		plan.IP.ValueString(),
		plan.Port.ValueString(),
		plan.User.ValueString(),
		options,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Adding Group Guest Access",
			fmt.Sprintf("Could not add guest access for account %s to group %s: %s", plan.Account.ValueString(), plan.Group.ValueString(), err.Error()),
		)
		return
	}

	// Generate ID
	plan.ID = types.StringValue(generateGuestAccessID(&plan))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *GroupGuestAccessResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state GroupGuestAccessResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// List all guest accesses for the group and account
	accesses, err := r.client.GroupListGuestAccesses(state.Group.ValueString(), state.Account.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Group Guest Accesses",
			fmt.Sprintf("Could not read guest accesses for account %s in group %s: %s", state.Account.ValueString(), state.Group.ValueString(), err.Error()),
		)
		return
	}

	// Find matching guest access
	var found *bastion.GroupGuestAccess
	for _, access := range accesses {
		if matchesGuestAccess(&state, access) {
			found = access
			break
		}
	}

	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Update state from API response
	state.IP = types.StringValue(found.IP)
	if found.Port != nil {
		state.Port = types.StringValue(found.Port.ValueString())
	} else {
		state.Port = types.StringValue("*")
	}

	// Handle protocol
	if found.User != nil && strings.HasPrefix(*found.User, "!") {
		// This is a protocol access
		state.Protocol = types.StringValue(strings.TrimPrefix(*found.User, "!"))
		state.User = types.StringNull()
	} else if found.User != nil {
		state.User = types.StringValue(*found.User)
		state.Protocol = types.StringNull()
	} else {
		state.User = types.StringValue("*")
		state.Protocol = types.StringNull()
	}

	if found.ProxyIP != nil {
		state.ProxyIP = types.StringValue(*found.ProxyIP)
	} else {
		state.ProxyIP = types.StringNull()
	}
	if found.ProxyPort != nil {
		state.ProxyPort = types.StringValue(found.ProxyPort.ValueString())
	} else if !state.ProxyPort.IsNull() && state.ProxyPort.ValueString() == "*" {
		state.ProxyPort = types.StringValue("*")
	} else {
		state.ProxyPort = types.StringNull()
	}
	if found.ProxyUser != nil {
		state.ProxyUser = types.StringValue(*found.ProxyUser)
	} else {
		state.ProxyUser = types.StringNull()
	}

	if found.RemotePort != nil {
		state.RemotePort = types.Int64Value(int64(found.RemotePort.ValueInt()))
	} else {
		state.RemotePort = types.Int64Null()
	}

	// for some reason comment becomes userComment
	if found.UserComment != nil {
		state.Comment = types.StringValue(*found.UserComment)
	} else {
		state.Comment = types.StringNull()
	}

	// Update ID
	state.ID = types.StringValue(generateGuestAccessID(&state))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *GroupGuestAccessResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Group guest access updates are not supported. All changes require resource replacement.",
	)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *GroupGuestAccessResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state GroupGuestAccessResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	proxyOpts := buildProxyOptionsFromState(&state)
	var remotePort *int64
	if !state.RemotePort.IsNull() {
		rp := state.RemotePort.ValueInt64()
		remotePort = &rp
	}

	err := r.client.GroupDelGuestAccess(
		state.Group.ValueString(),
		state.Account.ValueString(),
		state.IP.ValueString(),
		state.Port.ValueString(),
		state.User.ValueString(),
		state.Protocol.ValueString(),
		proxyOpts,
		remotePort,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Group Guest Access",
			fmt.Sprintf("Could not delete guest access: %s", err.Error()),
		)
		return
	}
}

// ImportState imports an existing resource by ID.
func (r *GroupGuestAccessResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importID := req.ID
	parts := parseImportID(importID)

	// Minimum is group:account:ip:port:user (5 parts)
	if len(parts) < 5 || len(parts) > 10 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in the format 'group:account:ip:port:user', 'group:account:ip:port:user:protocol', 'group:account:ip:port:user:protocol:remote_port', 'group:account:ip:port:user:proxy_ip:proxy_port:proxy_user', 'group:account:ip:port:user:protocol:proxy_ip:proxy_port:proxy_user', or 'group:account:ip:port:user:protocol:remote_port:proxy_ip:proxy_port:proxy_user', got: %s", importID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("group"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("account"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ip"), parts[2])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("port"), parts[3])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user"), parts[4])...)

	if len(parts) == 6 {
		// group:account:ip:port:user:protocol
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("protocol"), parts[5])...)
	} else if len(parts) == 7 {
		// group:account:ip:port:user:protocol:remote_port (when protocol is "portforward")
		protocol := parts[5]
		if protocol != "portforward" {
			resp.Diagnostics.AddError(
				"Invalid Import ID",
				fmt.Sprintf("When specifying remote_port, protocol must be 'portforward', got: %s", protocol),
			)
			return
		}
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("protocol"), protocol)...)
		remotePort, err := strconv.ParseInt(parts[6], 10, 64)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid Import ID",
				fmt.Sprintf("Invalid remote_port value '%s': %s", parts[6], err.Error()),
			)
			return
		}
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("remote_port"), remotePort)...)
	} else if len(parts) == 8 {
		// group:account:ip:port:user:proxy_ip:proxy_port:proxy_user
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("proxy_ip"), parts[5])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("proxy_port"), parts[6])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("proxy_user"), parts[7])...)
	} else if len(parts) == 9 {
		// group:account:ip:port:user:protocol:proxy_ip:proxy_port:proxy_user
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("protocol"), parts[5])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("proxy_ip"), parts[6])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("proxy_port"), parts[7])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("proxy_user"), parts[8])...)
	} else if len(parts) == 10 {
		// group:account:ip:port:user:protocol:remote_port:proxy_ip:proxy_port:proxy_user
		protocol := parts[5]
		if protocol != "portforward" {
			resp.Diagnostics.AddError(
				"Invalid Import ID",
				fmt.Sprintf("When specifying remote_port with proxy, protocol must be 'portforward', got: %s", protocol),
			)
			return
		}
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("protocol"), protocol)...)
		remotePort, err := strconv.ParseInt(parts[6], 10, 64)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid Import ID",
				fmt.Sprintf("Invalid remote_port value '%s': %s", parts[6], err.Error()),
			)
			return
		}
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("remote_port"), remotePort)...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("proxy_ip"), parts[7])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("proxy_port"), parts[8])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("proxy_user"), parts[9])...)
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), importID)...)
}

// generateGuestAccessID generates a unique ID for a guest access.
func generateGuestAccessID(model *GroupGuestAccessResourceModel) string {
	ip := formatIPForID(model.IP.ValueString())
	proxyIP := ""
	if !model.ProxyIP.IsNull() {
		proxyIP = formatIPForID(model.ProxyIP.ValueString())
	}

	id := fmt.Sprintf("%s:%s:%s:%s:%s",
		model.Group.ValueString(),
		model.Account.ValueString(),
		ip,
		model.Port.ValueString(),
		model.User.ValueString(),
	)

	// Add protocol if present
	if !model.Protocol.IsNull() {
		id = fmt.Sprintf("%s:%s", id, model.Protocol.ValueString())
	}

	// Add remote_port if present (only valid with portforward protocol)
	if !model.RemotePort.IsNull() {
		id = fmt.Sprintf("%s:%d", id, model.RemotePort.ValueInt64())
	}

	if !model.ProxyIP.IsNull() {
		id = fmt.Sprintf("%s:%s:%s:%s",
			id,
			proxyIP,
			model.ProxyPort.ValueString(),
			model.ProxyUser.ValueString(),
		)
	}

	return id
}

// matchesGuestAccess checks if a guest access matches the state.
func matchesGuestAccess(state *GroupGuestAccessResourceModel, access *bastion.GroupGuestAccess) bool {
	if state.IP.ValueString() != access.IP {
		return false
	}

	// Check port (API returns null for "*")
	statePort := state.Port.ValueString()
	accessPort := "*"
	if access.Port != nil {
		accessPort = access.Port.ValueString()
	}
	if statePort != accessPort {
		return false
	}

	// Check user (API returns null for "*")
	// For protocol accesses, API returns username as "!protocol"
	stateUser := state.User.ValueString()
	stateProtocol := ""
	if !state.Protocol.IsNull() {
		stateProtocol = state.Protocol.ValueString()
	}

	accessUser := "*"
	if access.User != nil {
		accessUser = *access.User
	}

	// If state has a protocol, match against "!protocol" in access user
	if stateProtocol != "" {
		expectedUser := "!" + stateProtocol
		if accessUser != expectedUser {
			return false
		}
	} else {
		if stateUser != accessUser {
			return false
		}
	}

	// Check remote_port (for portforward protocol)
	stateHasRemotePort := !state.RemotePort.IsNull()
	accessHasRemotePort := access.RemotePort != nil

	if stateHasRemotePort != accessHasRemotePort {
		return false
	}

	if stateHasRemotePort {
		if state.RemotePort.ValueInt64() != int64(access.RemotePort.ValueInt()) {
			return false
		}
	}

	// Check proxy settings
	stateHasProxy := !state.ProxyIP.IsNull()
	accessHasProxy := access.ProxyIP != nil

	if stateHasProxy != accessHasProxy {
		return false
	}

	if stateHasProxy {
		if state.ProxyIP.ValueString() != *access.ProxyIP {
			return false
		}

		// Check proxy port (API returns null for "*")
		stateProxyPort := state.ProxyPort.ValueString()
		accessProxyPort := "*"
		if access.ProxyPort != nil {
			accessProxyPort = access.ProxyPort.ValueString()
		}
		if stateProxyPort != accessProxyPort {
			return false
		}

		if state.ProxyUser.ValueString() != *access.ProxyUser {
			return false
		}
	}

	return true
}

// buildProxyOptionsFromState builds ProxyOptions from state values.
func buildProxyOptionsFromState(state *GroupGuestAccessResourceModel) *bastion.ProxyOptions {
	if state.ProxyIP.IsNull() {
		return nil
	}
	return &bastion.ProxyOptions{
		ProxyHost: state.ProxyIP.ValueString(),
		ProxyPort: state.ProxyPort.ValueString(),
		ProxyUser: state.ProxyUser.ValueString(),
	}
}
