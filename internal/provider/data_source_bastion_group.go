// Copyright (c) Adfinis
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/adfinis/terraform-provider-bastion/bastion"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &GroupDataSource{}
var _ datasource.DataSourceWithConfigure = &GroupDataSource{}

// NewGroupDataSource is a helper function to simplify the provider implementation.
func NewGroupDataSource() datasource.DataSource {
	return &GroupDataSource{}
}

// GroupDataSource is the data source implementation.
type GroupDataSource struct {
	client *bastion.Client
}

// groupDataSourceModel describes the data source data model.
type groupDataSourceModel struct {
	Group       types.String `tfsdk:"group"`
	Owners      types.List   `tfsdk:"owners"`
	Members     types.List   `tfsdk:"members"`
	Gatekeepers types.List   `tfsdk:"gatekeepers"`
	ACLKeepers  types.List   `tfsdk:"aclkeepers"`
}

// Metadata returns the data source type name.
func (d *GroupDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

// Schema defines the schema for the data source.
func (d *GroupDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"group": schema.StringAttribute{
				MarkdownDescription: "The name of the Bastion group",
				Required:            true,
			},
			"owners": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "The owners of the Bastion group",
				Computed:            true,
			},
			"members": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "The members of the Bastion group",
				Computed:            true,
			},
			"gatekeepers": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "The gatekeepers of the Bastion group",
				Computed:            true,
			},
			"aclkeepers": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "The ACL keepers of the Bastion group",
				Computed:            true,
			},
		},
	}
}

// Configure adds the bastion client to the data source.
func (d *GroupDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*bastion.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			"Expected *bastion.Client type for data source configuration.",
		)
		return
	}

	d.client = client
}

// Read refreshes the Terraform state with the latest data.
func (d *GroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data groupDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	group, err := d.client.GroupInfo(data.Group.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Bastion Group",
			err.Error(),
		)
		return
	}

	owners, diags := types.ListValueFrom(ctx, types.StringType, group.Owners)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Owners = owners

	members, diags := types.ListValueFrom(ctx, types.StringType, group.Members)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Members = members

	gatekeepers, diags := types.ListValueFrom(ctx, types.StringType, group.Gatekeepers)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Gatekeepers = gatekeepers

	aclkeepers, diags := types.ListValueFrom(ctx, types.StringType, group.ACLKeepers)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ACLKeepers = aclkeepers

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
