// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package application_oidc

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	appApi "github.com/zitadel/zitadel-go/v3/pkg/client/zitadel/app/v2beta"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Delete deletes a Zitadel OIDC application resource (`_application_oidc`).
func (r *ApplicationOIDCResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ApplicationOIDCResourceModel
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

	appId := data.Id.ValueString()
	projectId := data.ProjectId.ValueString()

	tflog.Debug(ctx, "deleting OIDC application", map[string]any{
		"app_id":     appId,
		"project_id": projectId,
	})

	_, err := zitadelClient.AppServiceV2Beta().DeleteApplication(ctx, &appApi.DeleteApplicationRequest{
		Id:        appId,
		ProjectId: projectId,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			tflog.Warn(ctx, "OIDC application already deleted or does not exist", map[string]any{
				"app_id": appId,
			})
			return
		}

		resp.Diagnostics.AddError(
			"Error deleting OIDC application",
			fmt.Sprintf("Could not delete OIDC application %s: %s", appId, err.Error()),
		)
		return
	}

	tflog.Trace(ctx, "deleted OIDC application", map[string]any{
		"app_id":     appId,
		"project_id": projectId,
	})
}
