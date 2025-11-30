// Copyright (c) Adfinis
// SPDX-License-Identifier: GPL-3.0-or-later

package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/adfinis/terraform-provider-bastion/bastion"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &GroupServerResource{}
var _ resource.ResourceWithConfigure = &GroupServerResource{}
var _ resource.ResourceWithImportState = &GroupServerResource{}

// NewGroupServerResource is a helper function to simplify the provider implementation.
func NewGroupServerResource() resource.Resource {
	return &GroupServerResource{}
}

// GroupServerResource is the resource implementation.
type GroupServerResource struct {
	client *bastion.Client
}

// GroupServerResourceModel describes the resource data model.
type GroupServerResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Group         types.String `tfsdk:"group"`
	IP            types.String `tfsdk:"ip"`
	Port          types.String `tfsdk:"port"`
	User          types.String `tfsdk:"user"`
	Protocol      types.String `tfsdk:"protocol"`
	ProxyIP       types.String `tfsdk:"proxy_ip"`
	ProxyPort     types.String `tfsdk:"proxy_port"`
	ProxyUser     types.String `tfsdk:"proxy_user"`
	Comment       types.String `tfsdk:"comment"`
	ForceKey      types.String `tfsdk:"force_key"`
	ForcePassword types.String `tfsdk:"force_password"`
	TTL           types.Int64  `tfsdk:"ttl"`
	Force         types.Bool   `tfsdk:"force"`
}

// Metadata returns the resource type name.
func (r *GroupServerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group_server"
}

// Schema defines the schema for the resource.
func (r *GroupServerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Bastion group access",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The resource identifier (group:ip:port:user[:proxy_ip:proxy_port:proxy_user])",
				Computed:            true,
			},
			"group": schema.StringAttribute{
				MarkdownDescription: "The Bastion group name to add the access to",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ip": schema.StringAttribute{
				MarkdownDescription: "IP, subnet of the access target. (hostname does not work)",
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
				MarkdownDescription: "Username for the access, use '*' to allow ssh access for all users.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"protocol": schema.StringAttribute{
				MarkdownDescription: "Protocol to grant access for. Valid values are 'sftp', 'scpup', 'scpdown', 'rsync'. When set, 'user' must be empty. A base access must already exist for the server.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"proxy_ip": schema.StringAttribute{
				MarkdownDescription: "IP of the proxy server",
				Optional:            true,
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
				MarkdownDescription: "Comment for the access",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"force_key": schema.StringAttribute{
				MarkdownDescription: "Force a specific SSH key for the access",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"force_password": schema.StringAttribute{
				MarkdownDescription: "Force a specific password for the access",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ttl": schema.Int64Attribute{
				MarkdownDescription: "Time to live in seconds for the access",
				Optional:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"force": schema.BoolAttribute{
				MarkdownDescription: "Force adding the access even if it cannot be verified",
				Optional:            true,
				WriteOnly:           true,
			},
		},
	}
}

// Configure adds the bastion client to the resource.
func (r *GroupServerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *GroupServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan GroupServerResourceModel
	var config GroupServerResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// either user or protocol must exist but not both
	if (plan.User.IsNull() && plan.Protocol.IsNull()) || (!plan.User.IsNull() && !plan.Protocol.IsNull()) {
		resp.Diagnostics.AddError(
			"Error Creating Group Server Access",
			"Either 'user' or 'protocol' must be set, but not both.",
		)
		return
	}

	// Build options
	options := &bastion.GroupAddServerOptions{
		Force: config.Force.ValueBool(),
	}

	if !plan.ForceKey.IsNull() {
		options.ForceKey = plan.ForceKey.ValueString()
	}

	if !plan.ForcePassword.IsNull() {
		options.ForcePassword = plan.ForcePassword.ValueString()
	}

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

	// Add the server access
	server, err := r.client.GroupAddServer(
		plan.Group.ValueString(),
		plan.IP.ValueString(),
		plan.Port.ValueString(),
		plan.User.ValueString(),
		options,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Adding Group Server Access",
			fmt.Sprintf("Could not add server access to group %s: %s", plan.Group.ValueString(), err.Error()),
		)
		return
	}

	plan.IP = types.StringValue(server.IP)
	if server.Port != nil {
		plan.Port = types.StringValue(server.Port.ValueString())
	} else {
		plan.Port = types.StringValue("*")
	}

	// Handle protocols
	if server.User != nil && strings.HasPrefix(*server.User, "!") {
		plan.Protocol = types.StringValue(strings.TrimPrefix(*server.User, "!"))
		plan.User = types.StringNull()
	} else if server.User != nil {
		plan.User = types.StringValue(*server.User)
		plan.Protocol = types.StringNull()
	} else {
		plan.User = types.StringValue("*")
		plan.Protocol = types.StringNull()
	}

	if server.ProxyIP != nil {
		plan.ProxyIP = types.StringValue(*server.ProxyIP)
	} else {
		plan.ProxyIP = types.StringNull()
	}
	if server.ProxyPort != nil {
		plan.ProxyPort = types.StringValue(server.ProxyPort.ValueString())
	} else if !plan.ProxyPort.IsNull() && plan.ProxyPort.ValueString() == "*" {
		plan.ProxyPort = types.StringValue("*")
	} else {
		plan.ProxyPort = types.StringNull()
	}
	if server.ProxyUser != nil {
		plan.ProxyUser = types.StringValue(*server.ProxyUser)
	} else {
		plan.ProxyUser = types.StringNull()
	}

	if server.Comment != nil {
		plan.Comment = types.StringValue(*server.Comment)
	} else {
		plan.Comment = types.StringNull()
	}

	// Generate ID
	plan.ID = types.StringValue(generateServerAccessID(&plan))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *GroupServerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state GroupServerResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// List all servers for the group
	servers, err := r.client.GroupListServers(state.Group.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Group Server Accesses",
			fmt.Sprintf("Could not read server accesses for group %s: %s", state.Group.ValueString(), err.Error()),
		)
		return
	}

	// Find matching server access
	// Note: API returns null for port, user, proxyPort when set to "*"
	var found *bastion.GroupServer
	for _, server := range servers {
		if matchesServerAccess(&state, server) {
			found = server
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

	// for some reason comment becomes userComment
	if found.UserComment != nil {
		state.Comment = types.StringValue(*found.UserComment)
	} else {
		state.Comment = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *GroupServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Group server accesses cannot be updated. This is a bug in the provider.",
	)
}

func (r *GroupServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state GroupServerResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build proxy options if needed
	var proxyOpts *bastion.ProxyOptions
	if !state.ProxyIP.IsNull() {
		proxyOpts = &bastion.ProxyOptions{
			ProxyHost: state.ProxyIP.ValueString(),
			ProxyPort: state.ProxyPort.ValueString(),
			ProxyUser: state.ProxyUser.ValueString(),
		}
	}

	err := r.client.GroupDelServer(
		state.Group.ValueString(),
		state.IP.ValueString(),
		state.Port.ValueString(),
		state.User.ValueString(),
		state.Protocol.ValueString(),
		proxyOpts,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Group Server Access",
			fmt.Sprintf("Could not delete server access from group %s: %s", state.Group.ValueString(), err.Error()),
		)
		return
	}
}

// ImportState imports the resource state.
func (r *GroupServerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Expected formats:
	// - group:ip:port:user
	// - group:ip:port:user:protocol
	// - group:ip:port:user:proxy_ip:proxy_port:proxy_user
	// - group:ip:port:user:protocol:proxy_ip:proxy_port:proxy_user
	importID := req.ID
	parts := parseImportID(importID)

	if len(parts) < 4 || len(parts) == 6 || len(parts) > 8 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in the format 'group:ip:port:user', 'group:ip:port:user:protocol', 'group:ip:port:user:proxy_ip:proxy_port:proxy_user', or 'group:ip:port:user:protocol:proxy_ip:proxy_port:proxy_user', got: %s", importID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("group"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ip"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("port"), parts[2])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user"), parts[3])...)

	if len(parts) == 5 {
		// group:ip:port:user:protocol
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("protocol"), parts[4])...)
	} else if len(parts) == 7 {
		// group:ip:port:user:proxy_ip:proxy_port:proxy_user
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("proxy_ip"), parts[4])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("proxy_port"), parts[5])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("proxy_user"), parts[6])...)
	} else if len(parts) == 8 {
		// group:ip:port:user:protocol:proxy_ip:proxy_port:proxy_user
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("protocol"), parts[4])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("proxy_ip"), parts[5])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("proxy_port"), parts[6])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("proxy_user"), parts[7])...)
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), importID)...)
}

// generateServerAccessID generates a unique ID for a server access.
// IPv6 addresses are wrapped in brackets to distinguish colons in the address from delimiter colons.
func generateServerAccessID(model *GroupServerResourceModel) string {
	ip := formatIPForID(model.IP.ValueString())
	proxyIP := ""
	if !model.ProxyIP.IsNull() {
		proxyIP = formatIPForID(model.ProxyIP.ValueString())
	}

	id := fmt.Sprintf("%s:%s:%s:%s",
		model.Group.ValueString(),
		ip,
		model.Port.ValueString(),
		model.User.ValueString(),
	)

	// Add protocol if present
	if !model.Protocol.IsNull() {
		id = fmt.Sprintf("%s:%s", id, model.Protocol.ValueString())
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

// parseImportID parses an import ID into its components.
// Handles IPv6 addresses enclosed in square brackets (e.g., [::1]).
func parseImportID(importID string) []string {
	parts := []string{}
	start := 0
	isIPv6 := false

	for i, char := range importID {
		if char == '[' {
			isIPv6 = true
		} else if char == ']' {
			isIPv6 = false
		} else if char == ':' && !isIPv6 {
			parts = append(parts, importID[start:i])
			start = i + 1
		}
	}
	if start < len(importID) {
		parts = append(parts, importID[start:])
	}

	// Strip brackets from IPv6 addresses
	for i, part := range parts {
		parts[i] = strings.Trim(part, "[]")
	}

	return parts
}

func formatIPForID(ip string) string {
	if strings.HasPrefix(ip, "[") && strings.HasSuffix(ip, "]") {
		return ip
	}
	if strings.Contains(ip, ":") {
		return "[" + ip + "]"
	}
	return ip
}

// matchesServerAccess checks if a server access matches the state.
func matchesServerAccess(state *GroupServerResourceModel, server *bastion.GroupServer) bool {
	if state.IP.ValueString() != server.IP {
		return false
	}

	// Check port (API returns null for "*")
	statePort := state.Port.ValueString()
	serverPort := "*"
	if server.Port != nil {
		serverPort = server.Port.ValueString()
	}
	if statePort != serverPort {
		return false
	}

	// Check user (API returns null for "*")
	// For protocol accesses, API returns username as "!protocol"
	stateUser := state.User.ValueString()
	stateProtocol := ""
	if !state.Protocol.IsNull() {
		stateProtocol = state.Protocol.ValueString()
	}

	serverUser := "*"
	if server.User != nil {
		serverUser = *server.User
	}

	// If state has a protocol, match against "!protocol" in server user
	if stateProtocol != "" {
		expectedUser := "!" + stateProtocol
		if serverUser != expectedUser {
			return false
		}
	} else {
		if stateUser != serverUser {
			return false
		}
	}

	// Check proxy settings
	stateHasProxy := !state.ProxyIP.IsNull()
	serverHasProxy := server.ProxyIP != nil

	if stateHasProxy != serverHasProxy {
		return false
	}

	if stateHasProxy {
		if state.ProxyIP.ValueString() != *server.ProxyIP {
			return false
		}

		// Check proxy port (API returns null for "*")
		stateProxyPort := state.ProxyPort.ValueString()
		serverProxyPort := "*"
		if server.ProxyPort != nil {
			serverProxyPort = server.ProxyPort.ValueString()
		}
		if stateProxyPort != serverProxyPort {
			return false
		}

		if state.ProxyUser.ValueString() != *server.ProxyUser {
			return false
		}
	}

	return true
}
