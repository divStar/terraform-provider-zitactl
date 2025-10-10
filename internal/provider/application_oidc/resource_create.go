// Copyright (c) Igor Voronin
// SPDX-License-Identifier: MIT

package application_oidc

import (
	"context"
	"fmt"
	"time"

	"github.com/divStar/terraform-provider-zitactl/internal/provider/helper"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	appApi "github.com/zitadel/zitadel-go/v3/pkg/client/zitadel/app/v2beta"
	"google.golang.org/protobuf/types/known/durationpb"
)

// Create creates a new Zitadel OIDC application resource (`_application_oidc`) and reads it back.
func (r *ApplicationOIDCResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ApplicationOIDCResourceModel
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

	projectId := data.ProjectId.ValueString()

	// Convert grant types
	grantTypesRaw, ok := helper.ExtractStringList(ctx, data.GrantTypes, &resp.Diagnostics)
	if !ok {
		return
	}
	grantTypes := helper.ConvertEnumList[appApi.OIDCGrantType](grantTypesRaw, appApi.OIDCGrantType_value)

	// Convert response types
	responseTypesRaw, ok := helper.ExtractStringList(ctx, data.ResponseTypes, &resp.Diagnostics)
	if !ok {
		return
	}
	responseTypes := helper.ConvertEnumList[appApi.OIDCResponseType](responseTypesRaw, appApi.OIDCResponseType_value)

	// Convert redirect URIs
	redirectUris, ok := helper.ExtractStringList(ctx, data.RedirectUris, &resp.Diagnostics)
	if !ok {
		return
	}

	// Convert AppType
	var appType appApi.OIDCAppType
	if !data.AppType.IsNull() {
		if appTypeValue, ok := appApi.OIDCAppType_value[data.AppType.ValueString()]; ok {
			appType = appApi.OIDCAppType(appTypeValue)
		}
	}

	// Convert AuthMethodType
	var authMethodType appApi.OIDCAuthMethodType
	if !data.AuthMethodType.IsNull() {
		if authValue, ok := appApi.OIDCAuthMethodType_value[data.AuthMethodType.ValueString()]; ok {
			authMethodType = appApi.OIDCAuthMethodType(authValue)
		}
	}

	// Convert PostLogoutRedirectUris
	var postLogoutUris []string
	if !data.PostLogoutRedirectUris.IsNull() {
		postLogoutUris, ok = helper.ExtractStringList(ctx, data.PostLogoutRedirectUris, &resp.Diagnostics)
		if !ok {
			return
		}
	}

	// Convert Version
	var version appApi.OIDCVersion
	if !data.Version.IsNull() {
		if versionValue, ok := appApi.OIDCVersion_value[data.Version.ValueString()]; ok {
			version = appApi.OIDCVersion(versionValue)
		}
	}

	// Convert AccessTokenType
	var accessTokenType appApi.OIDCTokenType
	if !data.AccessTokenType.IsNull() {
		if atValue, ok := appApi.OIDCTokenType_value[data.AccessTokenType.ValueString()]; ok {
			accessTokenType = appApi.OIDCTokenType(atValue)
		}
	}

	// Convert AdditionalOrigins
	var additionalOrigins []string
	if !data.AdditionalOrigins.IsNull() {
		additionalOrigins, ok = helper.ExtractStringList(ctx, data.AdditionalOrigins, &resp.Diagnostics)
		if !ok {
			return
		}
	}

	// Convert ClockSkew
	var clockSkew *durationpb.Duration
	if !data.ClockSkew.IsNull() && data.ClockSkew.ValueString() != "" {
		duration, err := time.ParseDuration(data.ClockSkew.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid ClockSkew",
				fmt.Sprintf("Could not parse clock_skew duration: %s", err.Error()),
			)
			return
		}
		clockSkew = durationpb.New(duration)
	}

	// Create the application request
	createReq := &appApi.CreateApplicationRequest{
		ProjectId: projectId,
		Name:      data.Name.ValueString(),
		CreationRequestType: &appApi.CreateApplicationRequest_OidcRequest{
			OidcRequest: &appApi.CreateOIDCApplicationRequest{
				RedirectUris:             redirectUris,
				ResponseTypes:            responseTypes,
				GrantTypes:               grantTypes,
				AppType:                  appType,
				AuthMethodType:           authMethodType,
				PostLogoutRedirectUris:   postLogoutUris,
				Version:                  version,
				DevMode:                  data.DevMode.ValueBool(),
				AccessTokenType:          accessTokenType,
				AccessTokenRoleAssertion: data.AccessTokenRoleAssertion.ValueBool(),
				IdTokenRoleAssertion:     data.IdTokenRoleAssertion.ValueBool(),
				IdTokenUserinfoAssertion: data.IdTokenUserinfoAssertion.ValueBool(),
				ClockSkew:                clockSkew,
				AdditionalOrigins:        additionalOrigins,
				SkipNativeAppSuccessPage: data.SkipNativeAppSuccessPage.ValueBool(),
			},
		},
	}

	tflog.Debug(ctx, "creating OIDC application", map[string]any{
		"name":       data.Name.ValueString(),
		"project_id": projectId,
	})

	createResp, err := zitadelClient.AppServiceV2Beta().CreateApplication(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating OIDC application",
			fmt.Sprintf("Could not create OIDC application: %s", err.Error()),
		)
		return
	}

	data.Id = types.StringValue(createResp.GetAppId())

	// Extract client credentials from the response
	if oidcDetails := createResp.GetOidcResponse(); oidcDetails != nil {
		data.ClientId = types.StringValue(oidcDetails.GetClientId())
		data.ClientSecret = types.StringValue(oidcDetails.GetClientSecret())
	}

	tflog.Trace(ctx, "created OIDC application", map[string]any{
		"app_id": data.Id.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call Read to populate all computed fields
	readReq := resource.ReadRequest{State: resp.State}
	readResp := &resource.ReadResponse{State: resp.State, Diagnostics: resp.Diagnostics}
	r.Read(ctx, readReq, readResp)

	resp.Diagnostics = readResp.Diagnostics
	resp.State = readResp.State
}
