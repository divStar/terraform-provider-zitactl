// Copyright (c) Igor Voronin
// SPDX-License-Identifier: MIT

package application_oidc

import (
	"context"
	"fmt"

	"github.com/divStar/terraform-provider-zitactl/internal/provider/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ApplicationOIDCResource{}
var _ resource.ResourceWithImportState = &ApplicationOIDCResource{}

// NewApplicationOIDCResource returns a new resource.Resource.
func NewApplicationOIDCResource() resource.Resource {
	return &ApplicationOIDCResource{}
}

// ApplicationOIDCResource defines the resource implementation.
type ApplicationOIDCResource struct {
	clientInfo *client.ClientInfo
}

// ApplicationOIDCResourceModel describes the resource data model.
type ApplicationOIDCResourceModel struct {
	// Required fields
	Name          types.String `tfsdk:"name"`
	ProjectId     types.String `tfsdk:"project_id"`
	GrantTypes    types.List   `tfsdk:"grant_types"`
	RedirectUris  types.List   `tfsdk:"redirect_uris"`
	ResponseTypes types.List   `tfsdk:"response_types"`
	// Optional + Computed fields
	AccessTokenRoleAssertion types.Bool   `tfsdk:"access_token_role_assertion"`
	AccessTokenType          types.String `tfsdk:"access_token_type"`
	AppType                  types.String `tfsdk:"app_type"`
	AuthMethodType           types.String `tfsdk:"auth_method_type"`
	ClockSkew                types.String `tfsdk:"clock_skew"`
	IdTokenRoleAssertion     types.Bool   `tfsdk:"id_token_role_assertion"`
	IdTokenUserinfoAssertion types.Bool   `tfsdk:"id_token_userinfo_assertion"`
	SkipNativeAppSuccessPage types.Bool   `tfsdk:"skip_native_app_success_page"`
	Version                  types.String `tfsdk:"version"`
	// Optional fields
	AdditionalOrigins      types.List `tfsdk:"additional_origins"`
	DevMode                types.Bool `tfsdk:"dev_mode"`
	PostLogoutRedirectUris types.List `tfsdk:"post_logout_redirect_uris"`
	// Computed fields (outputs)
	Id           types.String `tfsdk:"id"`
	ClientId     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
}

// Metadata sets the resource type name.
func (r *ApplicationOIDCResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application_oidc"
}

// Schema defines the resource schema.
func (r *ApplicationOIDCResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a ZITADEL OIDC Application",

		Attributes: map[string]schema.Attribute{
			// Required fields
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the application",
			},
			"project_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the project",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"grant_types": schema.ListAttribute{
				Required:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Grant types, supported values: OIDC_GRANT_TYPE_AUTHORIZATION_CODE, OIDC_GRANT_TYPE_IMPLICIT, OIDC_GRANT_TYPE_REFRESH_TOKEN, OIDC_GRANT_TYPE_DEVICE_CODE, OIDC_GRANT_TYPE_TOKEN_EXCHANGE",
			},
			"redirect_uris": schema.ListAttribute{
				Required:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Redirect URIs",
			},
			"response_types": schema.ListAttribute{
				Required:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Response types, supported values: OIDC_RESPONSE_TYPE_CODE, OIDC_RESPONSE_TYPE_ID_TOKEN, OIDC_RESPONSE_TYPE_ID_TOKEN_TOKEN",
			},

			// Optional + Computed fields (alphabetically sorted)
			"access_token_role_assertion": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Access token role assertion",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"access_token_type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Access token type, supported values: OIDC_TOKEN_TYPE_BEARER, OIDC_TOKEN_TYPE_JWT",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"app_type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "App type, supported values: OIDC_APP_TYPE_WEB, OIDC_APP_TYPE_USER_AGENT, OIDC_APP_TYPE_NATIVE",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"auth_method_type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Auth method type, supported values: OIDC_AUTH_METHOD_TYPE_BASIC, OIDC_AUTH_METHOD_TYPE_POST, OIDC_AUTH_METHOD_TYPE_NONE, OIDC_AUTH_METHOD_TYPE_PRIVATE_KEY_JWT",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"clock_skew": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Clock skew",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"id_token_role_assertion": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "ID token role assertion",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"id_token_userinfo_assertion": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "ID token userinfo assertion",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"skip_native_app_success_page": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Skip the successful login page on native apps and directly redirect the user to the callback",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"version": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Version, supported values: OIDC_VERSION_1_0",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			// Optional fields (alphabetically sorted)
			"additional_origins": schema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Additional origins",
			},
			"post_logout_redirect_uris": schema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Post logout redirect URIs",
			},
			// Computed fields (outputs)
			"dev_mode": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Dev mode (can be set by the user, will set to `false` otherwise",
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of this resource",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"client_id": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "Generated client ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"client_secret": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "Generated client secret",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure configures the resource.
func (r *ApplicationOIDCResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
