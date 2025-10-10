// Copyright (c) Igor Voronin
// SPDX-License-Identifier: MIT

package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// TestAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var TestAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"zitactl": providerserver.NewProtocol6WithError(New("test")()),
}

// TestAccPreCheck validates that the required environment variables are set
// for acceptance tests. This function should be called in the PreCheck function
// of acceptance tests.
func TestAccPreCheck(t *testing.T) {
	// Check for required environment variables for acceptance tests
	if v := os.Getenv("ZITACTL_DOMAIN"); v == "" {
		t.Skip("ZITACTL_DOMAIN must be set for acceptance tests")
	}
	if v := os.Getenv("ZITACTL_SERVICE_ACCOUNT_KEY"); v == "" {
		t.Skip("ZITACTL_SERVICE_ACCOUNT_KEY must be set for acceptance tests")
	}
}

// convertTypesStringToTFType converts a types.String to tftypes.Value
// for use in test configurations.
func convertTypesStringToTFType(s types.String) tftypes.Value {
	if s.IsNull() {
		return tftypes.NewValue(tftypes.String, nil)
	}
	if s.IsUnknown() {
		return tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
	}
	return tftypes.NewValue(tftypes.String, s.ValueString())
}

// convertTypesBoolToTFType converts a types.Bool to tftypes.Value
// for use in test configurations.
func convertTypesBoolToTFType(b types.Bool) tftypes.Value {
	if b.IsNull() {
		return tftypes.NewValue(tftypes.Bool, nil)
	}
	if b.IsUnknown() {
		return tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue)
	}
	return tftypes.NewValue(tftypes.Bool, b.ValueBool())
}
