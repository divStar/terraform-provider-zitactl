// Copyright (c) Igor Voronin
// SPDX-License-Identifier: MIT

package application_oidc

import (
	"context"
	"fmt"
	"time"

	"github.com/divStar/terraform-provider-zitactl/internal/provider/helper"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/zitadel/zitadel-go/v3/pkg/client"
	appApi "github.com/zitadel/zitadel-go/v3/pkg/client/zitadel/app/v2beta"
	"google.golang.org/protobuf/types/known/durationpb"
)

// Update updates a Zitadel OIDC application resource (`_application_oidc`) in the Zitadel instance.
func (r *ApplicationOIDCResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ApplicationOIDCResourceModel
	var state ApplicationOIDCResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Update called", map[string]any{
		"plan_name":  data.Name.ValueString(),
		"state_name": state.Name.ValueString(),
	})

	// Lazy client initialization
	zitadelClient, errClientCreation := r.clientInfo.GetClient(ctx)
	if errClientCreation != nil {
		resp.Diagnostics.AddError("Client configuration not possible!", errClientCreation.Error())
		return
	}

	appId := data.Id.ValueString()
	projectId := data.ProjectId.ValueString()

	// Check if name changed
	nameChanged := state.Name.ValueString() != data.Name.ValueString()

	// Check if OIDC config changed - simple check on the main fields
	oidcConfigChanged :=
		!state.DevMode.Equal(data.DevMode) ||
			!state.GrantTypes.Equal(data.GrantTypes) ||
			!state.ResponseTypes.Equal(data.ResponseTypes) ||
			!state.RedirectUris.Equal(data.RedirectUris) ||
			!state.AccessTokenRoleAssertion.Equal(data.AccessTokenRoleAssertion) ||
			!state.IdTokenRoleAssertion.Equal(data.IdTokenRoleAssertion) ||
			!state.IdTokenUserinfoAssertion.Equal(data.IdTokenUserinfoAssertion) ||
			!state.SkipNativeAppSuccessPage.Equal(data.SkipNativeAppSuccessPage) ||
			!state.AccessTokenType.Equal(data.AccessTokenType) ||
			!state.AppType.Equal(data.AppType) ||
			!state.AuthMethodType.Equal(data.AuthMethodType) ||
			!state.Version.Equal(data.Version) ||
			!state.PostLogoutRedirectUris.Equal(data.PostLogoutRedirectUris) ||
			!state.AdditionalOrigins.Equal(data.AdditionalOrigins) ||
			!state.ClockSkew.Equal(data.ClockSkew)

	tflog.Debug(ctx, "Change detection", map[string]any{
		"name_changed":        nameChanged,
		"oidc_config_changed": oidcConfigChanged,
	})

	// Perform update if anything changed
	if nameChanged || oidcConfigChanged {
		if err := r.updateApplication(ctx, appId, projectId, &data, nameChanged, oidcConfigChanged, zitadelClient, resp); err != nil {
			resp.Diagnostics.AddError(
				"Error updating OIDC application",
				fmt.Sprintf("Could not update OIDC application %s: %s", appId, err.Error()),
			)
			return
		}
	}

	// Update state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Refresh from remote
	readReq := resource.ReadRequest{State: resp.State}
	readResp := &resource.ReadResponse{State: resp.State, Diagnostics: resp.Diagnostics}
	r.Read(ctx, readReq, readResp)

	resp.Diagnostics = readResp.Diagnostics
	resp.State = readResp.State
}

func (r *ApplicationOIDCResource) updateApplication(
	ctx context.Context,
	appId, projectId string,
	data *ApplicationOIDCResourceModel,
	nameChanged, oidcConfigChanged bool,
	zitadelClient *client.Client,
	respDiags *resource.UpdateResponse,
) error {
	tflog.Debug(ctx, "updating OIDC application", map[string]any{
		"app_id":              appId,
		"project_id":          projectId,
		"name_changed":        nameChanged,
		"oidc_config_changed": oidcConfigChanged,
	})

	// Create the base update request
	updateReq := &appApi.UpdateApplicationRequest{
		Id:        appId,
		ProjectId: projectId,
	}

	// Set name if changed
	if nameChanged {
		updateReq.Name = data.Name.ValueString()
	}

	// Set OIDC config if changed
	if oidcConfigChanged {
		// Convert grant types
		grantTypesRaw, ok := helper.ExtractStringList(ctx, data.GrantTypes, &respDiags.Diagnostics)
		if !ok {
			return fmt.Errorf("failed to convert grant types")
		}
		grantTypes := helper.ConvertEnumList[appApi.OIDCGrantType](grantTypesRaw, appApi.OIDCGrantType_value)

		// Convert response types
		responseTypesRaw, ok := helper.ExtractStringList(ctx, data.ResponseTypes, &respDiags.Diagnostics)
		if !ok {
			return fmt.Errorf("failed to convert response types")
		}
		responseTypes := helper.ConvertEnumList[appApi.OIDCResponseType](responseTypesRaw, appApi.OIDCResponseType_value)

		// Convert redirect URIs
		redirectUris, ok := helper.ExtractStringList(ctx, data.RedirectUris, &respDiags.Diagnostics)
		if !ok {
			return fmt.Errorf("failed to convert redirect URIs")
		}

		// Build the OIDC config update
		oidcConfig := &appApi.UpdateOIDCApplicationConfigurationRequest{
			RedirectUris:             redirectUris,
			ResponseTypes:            responseTypes,
			GrantTypes:               grantTypes,
			AccessTokenRoleAssertion: helper.Ptr(data.AccessTokenRoleAssertion.ValueBool()),
			IdTokenRoleAssertion:     helper.Ptr(data.IdTokenRoleAssertion.ValueBool()),
			IdTokenUserinfoAssertion: helper.Ptr(data.IdTokenUserinfoAssertion.ValueBool()),
			DevMode:                  helper.Ptr(data.DevMode.ValueBool()),
			SkipNativeAppSuccessPage: helper.Ptr(data.SkipNativeAppSuccessPage.ValueBool()),
		}

		// Set optional enum fields
		if !data.AccessTokenType.IsNull() {
			if atValue, ok := appApi.OIDCTokenType_value[data.AccessTokenType.ValueString()]; ok {
				tokenType := appApi.OIDCTokenType(atValue)
				oidcConfig.AccessTokenType = &tokenType
			}
		}

		if !data.AppType.IsNull() {
			if appTypeValue, ok := appApi.OIDCAppType_value[data.AppType.ValueString()]; ok {
				appType := appApi.OIDCAppType(appTypeValue)
				oidcConfig.AppType = &appType
			}
		}

		if !data.AuthMethodType.IsNull() {
			if authValue, ok := appApi.OIDCAuthMethodType_value[data.AuthMethodType.ValueString()]; ok {
				authMethod := appApi.OIDCAuthMethodType(authValue)
				oidcConfig.AuthMethodType = &authMethod
			}
		}

		if !data.Version.IsNull() {
			if versionValue, ok := appApi.OIDCVersion_value[data.Version.ValueString()]; ok {
				version := appApi.OIDCVersion(versionValue)
				oidcConfig.Version = &version
			}
		}

		// Set optional list fields
		if !data.PostLogoutRedirectUris.IsNull() {
			postLogoutUris, ok := helper.ExtractStringList(ctx, data.PostLogoutRedirectUris, &respDiags.Diagnostics)
			if ok {
				oidcConfig.PostLogoutRedirectUris = postLogoutUris
			}
		}

		if !data.AdditionalOrigins.IsNull() {
			additionalOrigins, ok := helper.ExtractStringList(ctx, data.AdditionalOrigins, &respDiags.Diagnostics)
			if ok {
				oidcConfig.AdditionalOrigins = additionalOrigins
			}
		}

		if !data.ClockSkew.IsNull() && data.ClockSkew.ValueString() != "" {
			duration, err := time.ParseDuration(data.ClockSkew.ValueString())
			if err != nil {
				respDiags.Diagnostics.AddError(
					"Invalid ClockSkew",
					fmt.Sprintf("Could not parse clock_skew duration: %s", err.Error()),
				)
				return err
			}
			oidcConfig.ClockSkew = durationpb.New(duration)
		}

		updateReq.UpdateRequestType = &appApi.UpdateApplicationRequest_OidcConfigurationRequest{
			OidcConfigurationRequest: oidcConfig,
		}
	}

	_, err := zitadelClient.AppServiceV2Beta().UpdateApplication(ctx, updateReq)
	return err
}
