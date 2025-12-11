// Copyright (c) Adfinis
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/adfinis/terraform-provider-bastion/bastion"
)

// Ensure BastionProvider satisfies various provider interfaces.
var _ provider.Provider = &BastionProvider{}

// BastionProvider defines the provider implementation.
type BastionProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// BastionProviderModel describes the provider data model.
type BastionProviderModel struct {
	Host                  types.String `tfsdk:"host"`
	Port                  types.Int64  `tfsdk:"port"`
	Username              types.String `tfsdk:"username"`
	PrivateKey            types.String `tfsdk:"private_key"`
	PrivateKeyFile        types.String `tfsdk:"private_key_file"`
	UseAgent              types.Bool   `tfsdk:"use_agent"`
	Timeout               types.Int64  `tfsdk:"timeout"`
	StrictHostKeyChecking types.Bool   `tfsdk:"strict_host_key_checking"`
}

func (p *BastionProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "bastion"
	resp.Version = p.version
}

func (p *BastionProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				MarkdownDescription: "The Bastion host to connect to",
				Required:            true,
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "The SSH port to connect to (default: 22)",
				Optional:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "SSH username for The Bastion",
				Required:            true,
			},
			"private_key": schema.StringAttribute{
				MarkdownDescription: "SSH private key content",
				Optional:            true,
				Sensitive:           true,
			},
			"private_key_file": schema.StringAttribute{
				MarkdownDescription: "Path to SSH private key file",
				Optional:            true,
			},
			"use_agent": schema.BoolAttribute{
				MarkdownDescription: "Use SSH agent for authentication (default: false)",
				Optional:            true,
			},
			"timeout": schema.Int64Attribute{
				MarkdownDescription: "SSH connection timeout in seconds (default: 30)",
				Optional:            true,
			},
			"strict_host_key_checking": schema.BoolAttribute{
				MarkdownDescription: "Enable strict host key checking (default: true)",
				Optional:            true,
			},
		},
	}
}

func (p *BastionProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data BastionProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Port.IsNull() {
		data.Port = types.Int64Value(22)
	}

	if data.Timeout.IsNull() {
		data.Timeout = types.Int64Value(30)
	}

	if data.StrictHostKeyChecking.IsNull() {
		data.StrictHostKeyChecking = types.BoolValue(true)
	}

	// Environment variable support
	host := os.Getenv("BASTION_HOST")
	if host != "" {
		data.Host = types.StringValue(host)
	}

	port := os.Getenv("BASTION_PORT")
	if port != "" {
		portInt, err := strconv.ParseInt(port, 10, 64)
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("port"),
				"Invalid Bastion Port",
				"The BASTION_PORT environment variable must be a valid integer. "+
					"Error: "+err.Error(),
			)
		} else {
			data.Port = types.Int64Value(portInt)
		}
	}

	username := os.Getenv("BASTION_USERNAME")
	if username != "" {
		data.Username = types.StringValue(username)
	}

	keyFile := os.Getenv("BASTION_PRIVATE_KEY_FILE")
	if keyFile != "" {
		data.PrivateKeyFile = types.StringValue(keyFile)
	}

	// Validation
	if data.Host.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing Bastion Host",
			"The provider cannot create the Bastion client as there is a missing or empty value for the Bastion host. "+
				"Set the host value in the configuration or use the BASTION_HOST environment variable.",
		)
	}

	if data.Username.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Missing Bastion Username",
			"The provider cannot create the Bastion client as there is a missing or empty value for the Bastion username. "+
				"Set the username value in the configuration or use the BASTION_USERNAME environment variable.",
		)
	}

	config := &bastion.Config{
		Host:                  data.Host.ValueString(),
		Port:                  int(data.Port.ValueInt64()),
		Username:              data.Username.ValueString(),
		Timeout:               int(data.Timeout.ValueInt64()),
		StrictHostKeyChecking: data.StrictHostKeyChecking.ValueBool(),
	}

	authMethods := make([]bastion.SSHAuthMethod, 0)
	if !data.PrivateKey.IsNull() {
		authMethods = append(
			authMethods,
			bastion.WithPrivateKeyAuth(data.PrivateKey.ValueString()),
		)
	}

	if !data.PrivateKeyFile.IsNull() {
		authMethods = append(
			authMethods,
			bastion.WithPrivateKeyFileAuth(data.PrivateKeyFile.ValueString()),
		)
	}

	if data.UseAgent.ValueBool() {
		authMethods = append(
			authMethods,
			bastion.WithSSHAgentAuth(),
		)
	}

	if len(authMethods) == 0 {
		resp.Diagnostics.AddError(
			"Missing Bastion Authentication Method",
			"The provider cannot create the Bastion client as there is no SSH authentication method configured. "+
				"Set at least one of private_key, private_key_file, password, or enable use_agent.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := bastion.New(config, authMethods...)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Bastion Client",
			"An unexpected error occurred when creating the Bastion client: "+err.Error(),
		)
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
	resp.EphemeralResourceData = client

	tflog.Info(ctx, "Configured Bastion client", map[string]any{"success": true})
}

func (p *BastionProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAccountResource,
		NewAccountCommandResource,
		NewAccountPIVPolicyResource,
		NewGroupResource,
		NewGroupOwnerResource,
		NewGroupGatekeeperResource,
		NewGroupACLKeeperResource,
		NewGroupMemberResource,
		NewGroupServerResource,
		NewGroupGuestAccessResource,
	}
}

func (p *BastionProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewGroupDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &BastionProvider{
			version: version,
		}
	}
}
