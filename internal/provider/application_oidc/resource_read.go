// Copyright (c) Igor Voronin
// SPDX-License-Identifier: MIT

package application_oidc

import (
	"context"
	"fmt"

	"github.com/divStar/terraform-provider-zitactl/internal/provider/helper"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	appApi "github.com/zitadel/zitadel-go/v3/pkg/client/zitadel/app/v2beta"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Read reads a Zitadel OIDC application resource (`_application_oidc`) from the Zitadel instance.
func (r *ApplicationOIDCResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ApplicationOIDCResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Lazy client initialization
	zitadelClient, errClientCreation := r.clientInfo.GetClient(ctx)
	if errClientCreation != nil {
		// Check if this is due to unknown provider configuration during plan refresh
		if r.clientInfo.Config != nil {
			hasUnknown := r.clientInfo.Config.Domain.IsUnknown() ||
				r.clientInfo.Config.SkipTlsVerification.IsUnknown() ||
				r.clientInfo.Config.ServiceAccountKey.IsUnknown()

			if hasUnknown {
				// During plan phase with unknown provider config, we cannot refresh -> return WITHOUT an error, keep the existing state
				tflog.Warn(ctx, "Skipping refresh due to unknown provider configuration", map[string]any{
					"id": data.Id.ValueString(),
				})
				return
			}
		}

		resp.Diagnostics.AddError("Client configuration not possible!", errClientCreation.Error())
		return
	}

	projectId := data.ProjectId.ValueString()
	appId := data.Id.ValueString()

	tflog.Debug(ctx, "reading OIDC application", map[string]any{
		"project_id": projectId,
		"app_id":     appId,
	})

	getResp, err := zitadelClient.AppServiceV2Beta().GetApplication(ctx, &appApi.GetApplicationRequest{
		Id: appId,
	})

	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			tflog.Warn(ctx, "OIDC application not found, removing from state", map[string]any{
				"app_id": appId,
			})
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Error reading OIDC application",
				fmt.Sprintf("Could not read OIDC application %s: %s", appId, err.Error()),
			)
		}
		return
	}

	app := getResp.GetApp()
	if app != nil {
		data.Name = types.StringValue(app.GetName())

		// Get OIDC config
		oidcConfig := app.GetOidcConfig()
		if oidcConfig != nil {
			data.GrantTypes = helper.ConvertEnumSliceToList(oidcConfig.GetGrantTypes())
			data.ResponseTypes = helper.ConvertEnumSliceToList(oidcConfig.GetResponseTypes())
			data.RedirectUris = helper.ConvertStringSliceToList(oidcConfig.GetRedirectUris())
			data.PostLogoutRedirectUris = helper.ConvertStringSliceToList(oidcConfig.GetPostLogoutRedirectUris())
			data.AdditionalOrigins = helper.ConvertStringSliceToList(oidcConfig.GetAdditionalOrigins())

			data.AccessTokenRoleAssertion = types.BoolValue(oidcConfig.GetAccessTokenRoleAssertion())
			data.IdTokenRoleAssertion = types.BoolValue(oidcConfig.GetIdTokenRoleAssertion())
			data.IdTokenUserinfoAssertion = types.BoolValue(oidcConfig.GetIdTokenUserinfoAssertion())
			data.DevMode = types.BoolValue(oidcConfig.GetDevMode())
			data.SkipNativeAppSuccessPage = types.BoolValue(oidcConfig.GetSkipNativeAppSuccessPage())
			data.AccessTokenType = types.StringValue(oidcConfig.GetAccessTokenType().String())
			data.AppType = types.StringValue(oidcConfig.GetAppType().String())
			data.AuthMethodType = types.StringValue(oidcConfig.GetAuthMethodType().String())
			data.Version = types.StringValue(oidcConfig.GetVersion().String())

			if !data.ClockSkew.IsNull() && data.ClockSkew.ValueString() != "" {
				data.ClockSkew = types.StringValue(oidcConfig.ClockSkew.AsDuration().String())
			} else {
				data.ClockSkew = types.StringNull()
			}
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
