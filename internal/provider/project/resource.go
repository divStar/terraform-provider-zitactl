// Copyright (c) Igor Voronin
// SPDX-License-Identifier: MIT

package project

import (
	"context"
	"fmt"

	"github.com/divStar/terraform-provider-zitactl/internal/provider/client"
	"github.com/divStar/terraform-provider-zitactl/internal/provider/helper"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/zitadel/zitadel-go/v3/pkg/client/zitadel/admin"
	projectApi "github.com/zitadel/zitadel-go/v3/pkg/client/zitadel/project/v2beta"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ resource.Resource = &ProjectResource{}
var _ resource.ResourceWithImportState = &ProjectResource{}

func NewProjectResource() resource.Resource {
	return &ProjectResource{}
}

// ProjectResource defines the resource implementation.
type ProjectResource struct {
	clientInfo *client.ClientInfo
}

// ProjectResourceModel describes the resource data model.
type ProjectResourceModel struct {
	Name                   types.String `tfsdk:"name"`
	OrgId                  types.String `tfsdk:"org_id"`
	HasProjectCheck        types.Bool   `tfsdk:"has_project_check"`
	PrivateLabelingSetting types.String `tfsdk:"private_labeling_setting"`
	ProjectRoleAssertion   types.Bool   `tfsdk:"project_role_assertion"`
	ProjectRoleCheck       types.Bool   `tfsdk:"project_role_check"`
	Id                     types.String `tfsdk:"id"`
	State                  types.String `tfsdk:"state"`
}

// Metadata sets the resource type name.
func (r *ProjectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

// Schema defines the resource schema.
func (r *ProjectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a ZITADEL project",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the project",
				Required:            true,
			},
			"org_id": schema.StringAttribute{
				MarkdownDescription: "ID of the organization",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"has_project_check": schema.BoolAttribute{
				MarkdownDescription: "ZITADEL checks if the org of the user has permission to this project",
				Optional:            true,
				Computed:            true,
			},
			"private_labeling_setting": schema.StringAttribute{
				MarkdownDescription: "Defines from where the private labeling should be triggered, supported values: PRIVATE_LABELING_SETTING_UNSPECIFIED, PRIVATE_LABELING_SETTING_ENFORCE_PROJECT_RESOURCE_OWNER_POLICY, PRIVATE_LABELING_SETTING_ALLOW_LOGIN_USER_RESOURCE_OWNER_POLICY",
				Optional:            true,
				Computed:            true,
			},
			"project_role_assertion": schema.BoolAttribute{
				MarkdownDescription: "Describes if roles of user should be added in token",
				Optional:            true,
				Computed:            true,
			},
			"project_role_check": schema.BoolAttribute{
				MarkdownDescription: "ZITADEL checks if the user has at least one role on this project",
				Optional:            true,
				Computed:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of this resource",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"state": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "State of the project",
			},
		},
	}
}

// Configure configures the resource.
func (r *ProjectResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.clientInfo = clientInfo
}

// Create creates a new Zitadel project resource (`_project`) and reads it back.
func (r *ProjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ProjectResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Lazy client initialization
	zitadelClient, errClientCreation := r.clientInfo.GetClient(ctx)
	if errClientCreation != nil {
		resp.Diagnostics.AddError("Client configuration not possible!", errClientCreation.Error())
		return
	}

	orgId := data.OrgId.ValueString()

	_, err := zitadelClient.AdminService().GetOrgByID(ctx, &admin.GetOrgByIDRequest{
		Id: orgId,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			resp.Diagnostics.AddError(
				"Invalid Organization ID",
				fmt.Sprintf("Organization with ID %s does not exist. Please provide a valid organization ID.", orgId),
			)
		} else {
			resp.Diagnostics.AddError(
				"Error Validating Organization",
				fmt.Sprintf("Could not validate organization %s: %s", orgId, err.Error()),
			)
		}
		return
	}

	// Validate that the organization exists before creating the project
	tflog.Debug(ctx, "validating organization exists", map[string]any{
		"org_id": orgId,
	})

	createReq := &projectApi.CreateProjectRequest{
		Name:                  data.Name.ValueString(),
		OrganizationId:        data.OrgId.ValueString(),
		ProjectRoleAssertion:  data.ProjectRoleAssertion.ValueBool(),
		AuthorizationRequired: data.ProjectRoleCheck.ValueBool(),
		ProjectAccessRequired: data.HasProjectCheck.ValueBool(),
	}

	if !data.PrivateLabelingSetting.IsNull() {
		settingStr := data.PrivateLabelingSetting.ValueString()
		if settingValue, ok := projectApi.PrivateLabelingSetting_value[settingStr]; ok {
			createReq.PrivateLabelingSetting = projectApi.PrivateLabelingSetting(settingValue)
		}
	}

	tflog.Debug(ctx, "creating project", map[string]any{
		"name":   data.Name.ValueString(),
		"org_id": data.OrgId.ValueString(),
	})

	createResp, err := zitadelClient.ProjectServiceV2Beta().CreateProject(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating project",
			fmt.Sprintf("Could not create project: %s", err.Error()),
		)
		return
	}

	data.Id = types.StringValue(createResp.GetId())

	tflog.Trace(ctx, "created project", map[string]any{
		"project_id": data.Id.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call Read to populate all computed fields; provide the current partial state to Read
	readReq := resource.ReadRequest{State: resp.State}
	readResp := &resource.ReadResponse{State: resp.State, Diagnostics: resp.Diagnostics}
	r.Read(ctx, readReq, readResp)

	// Copy diagnostics and state back to Create
	resp.Diagnostics = readResp.Diagnostics
	resp.State = readResp.State
}

// Read reads a Zitadel project resource (`_project`) from the Zitadel instance.
func (r *ProjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProjectResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Lazy client initialization
	zitadelClient, errClientCreation := r.clientInfo.GetClient(ctx)
	if errClientCreation != nil {
		resp.Diagnostics.AddError("Client configuration not possible!", errClientCreation.Error())
		return
	}

	projectId := data.Id.ValueString()
	orgId := data.OrgId.ValueString()

	tflog.Debug(ctx, "reading project", map[string]any{
		"project_id": projectId,
		"org_id":     orgId,
	})

	queryResponse, err := zitadelClient.ProjectServiceV2Beta().GetProject(ctx, &projectApi.GetProjectRequest{
		Id: projectId,
	})

	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			tflog.Warn(ctx, "project not found, removing from state", map[string]any{
				"project_id": projectId,
			})
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Error reading project",
				fmt.Sprintf("Could not read project %s: %s", projectId, err.Error()),
			)
		}
		return
	}

	retrievedProject := queryResponse.GetProject()
	if retrievedProject != nil {
		data.Name = types.StringValue(retrievedProject.GetName())
		data.ProjectRoleAssertion = types.BoolValue(retrievedProject.GetProjectRoleAssertion())
		data.State = types.StringValue(retrievedProject.GetState().String())
		data.OrgId = types.StringValue(retrievedProject.GetOrganizationId())
		data.ProjectRoleCheck = types.BoolValue(retrievedProject.GetAuthorizationRequired())
		data.HasProjectCheck = types.BoolValue(retrievedProject.GetProjectAccessRequired())
		data.PrivateLabelingSetting = types.StringValue(retrievedProject.GetPrivateLabelingSetting().String())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates a Zitadel project resource (`_project`) in the Zitadel instance.
func (r *ProjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ProjectResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Lazy client initialization
	zitadelClient, errClientCreation := r.clientInfo.GetClient(ctx)
	if errClientCreation != nil {
		resp.Diagnostics.AddError("Client configuration not possible!", errClientCreation.Error())
		return
	}

	projectId := data.Id.ValueString()

	updateReq := &projectApi.UpdateProjectRequest{
		Id:                   projectId,
		Name:                 helper.Ptr(data.Name.ValueString()),
		ProjectRoleAssertion: helper.Ptr(data.ProjectRoleAssertion.ValueBool()),
		ProjectRoleCheck:     helper.Ptr(data.ProjectRoleCheck.ValueBool()),
		HasProjectCheck:      helper.Ptr(data.HasProjectCheck.ValueBool()),
	}

	if !data.PrivateLabelingSetting.IsNull() {
		settingStr := data.PrivateLabelingSetting.ValueString()
		if settingValue, ok := projectApi.PrivateLabelingSetting_value[settingStr]; ok {
			setting := projectApi.PrivateLabelingSetting(settingValue)
			updateReq.PrivateLabelingSetting = &setting
		}
	}

	tflog.Debug(ctx, "updating project", map[string]any{
		"project_id": projectId,
	})

	_, err := zitadelClient.ProjectServiceV2Beta().UpdateProject(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating project",
			fmt.Sprintf("Could not update project %s: %s", projectId, err.Error()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	readReq := resource.ReadRequest{State: resp.State}
	readResp := &resource.ReadResponse{State: resp.State, Diagnostics: resp.Diagnostics}
	r.Read(ctx, readReq, readResp)

	resp.Diagnostics = readResp.Diagnostics
	resp.State = readResp.State
}

// Delete deletes a Zitadel project resource (`_project`).
func (r *ProjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ProjectResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Lazy client initialization
	zitadelClient, errClientCreation := r.clientInfo.GetClient(ctx)
	if errClientCreation != nil {
		resp.Diagnostics.AddError("Client configuration not possible!", errClientCreation.Error())
		return
	}

	projectId := data.Id.ValueString()

	tflog.Debug(ctx, "deleting project", map[string]any{
		"project_id": projectId,
	})

	_, err := zitadelClient.ProjectServiceV2Beta().DeleteProject(ctx, &projectApi.DeleteProjectRequest{
		Id: projectId,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			tflog.Warn(ctx, "project already deleted or does not exist", map[string]any{
				"project_id": projectId,
			})
			return
		}

		resp.Diagnostics.AddError(
			"Error deleting project",
			fmt.Sprintf("Could not delete project %s: %s", projectId, err.Error()),
		)
		return
	}

	tflog.Trace(ctx, "deleted project", map[string]any{
		"project_id": projectId,
	})
}

// ImportState imports the state of an existing resource.
// Use the format `id`. The project with the given `id` must already exist.
func (r *ProjectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
