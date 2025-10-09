// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package application_oidc

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// ImportState imports the state of an existing resource.
// Use the format `project_id:app_id`. The project with the given `project_id` must already exist.
func (r *ApplicationOIDCResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ":")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID format: 'project_id:app_id', got: %s", req.ID),
		)
		return
	}

	projectId := parts[0]
	appId := parts[1]

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), projectId)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), appId)...)
}
