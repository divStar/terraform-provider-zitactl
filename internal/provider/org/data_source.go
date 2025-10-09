// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package org

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/divStar/terraform-provider-zitactl/internal/provider/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	objectV2 "github.com/zitadel/zitadel-go/v3/pkg/client/zitadel/object/v2"
	"github.com/zitadel/zitadel-go/v3/pkg/client/zitadel/org/v2"
)

var _ datasource.DataSource = &OrgsDataSource{}

func NewOrgsDataSource() datasource.DataSource {
	return &OrgsDataSource{}
}

// OrgsDataSource defines the orgs data source implementation.
type OrgsDataSource struct {
	clientInfo *client.ClientInfo
}

// OrgsDataSourceModel describes the orgs data source data model.
type OrgsDataSourceModel struct {
	Ids        []types.String `tfsdk:"ids"`
	Name       types.String   `tfsdk:"name"`
	NameMethod types.String   `tfsdk:"name_method"`
}

func (d *OrgsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_orgs"
}

func (d *OrgsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Datasource representing organizations in ZITADEL.

Organizations are the highest level after the instance and contain several 
other resources including policies if the configuration differs from the 
default policies on the instance.`,
		Attributes: map[string]schema.Attribute{
			"ids": schema.ListAttribute{
				MarkdownDescription: "List of organization IDs matching the given name",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the organization to search for",
				Required:            true,
			},
			"name_method": schema.StringAttribute{
				MarkdownDescription: "Method for querying orgs by name",
				Optional:            true,
			},
		},
	}
}

func (d *OrgsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	clientInfo, ok := req.ProviderData.(*client.ClientInfo)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected resource configure type",
			fmt.Sprintf("Expected *ProviderData, got: %T.", req.ProviderData),
		)
		return
	}

	d.clientInfo = clientInfo
}

// Read reads the `_orgs` data source, returning a list of organization IDs.
func (d *OrgsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OrgsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Lazy client initialization
	zitadelClient, errClientCreation := d.clientInfo.GetClient(ctx)
	if errClientCreation != nil {
		resp.Diagnostics.AddError("Client configuration not possible!", errClientCreation.Error())
		return
	}

	orgName := data.Name.ValueString()

	queryMethod := objectV2.TextQueryMethod_TEXT_QUERY_METHOD_EQUALS_IGNORE_CASE
	if !data.NameMethod.IsNull() && !data.NameMethod.IsUnknown() {
		methodStr := data.NameMethod.ValueString()

		if enumValue, ok := objectV2.TextQueryMethod_value[methodStr]; ok {
			queryMethod = objectV2.TextQueryMethod(enumValue)
		} else {
			validNames := slices.Collect(maps.Keys(objectV2.TextQueryMethod_value))

			resp.Diagnostics.AddError(
				"Invalid name_method",
				fmt.Sprintf("The provided name_method '%s' is not valid. Valid values are: %v", methodStr, validNames),
			)
			return
		}
	}

	tflog.Debug(ctx, "Searching for organizations", map[string]any{
		"name":        orgName,
		"name_method": queryMethod,
	})

	queryResponse, err := zitadelClient.OrganizationServiceV2().ListOrganizations(ctx, &org.ListOrganizationsRequest{
		Queries: []*org.SearchQuery{
			{
				Query: &org.SearchQuery_NameQuery{
					NameQuery: &org.OrganizationNameQuery{
						Name:   orgName,
						Method: queryMethod,
					},
				},
			},
		},
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to list organizations",
			fmt.Sprintf("Unable to search for organizations with name '%s': %s", orgName, err),
		)
		return
	}

	var ids []types.String
	for _, currentOrganization := range queryResponse.Result {
		ids = append(ids, types.StringValue(currentOrganization.Id))
	}
	data.Ids = ids

	tflog.Trace(ctx, "Successfully read organization data")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
