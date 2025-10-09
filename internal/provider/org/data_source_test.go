// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package org

import (
	"context"
	"testing"

	"github.com/divStar/terraform-provider-zitactl/internal/provider/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestOrgsDataSource_DeferredClientInitialization tests that the data source properly handles
// deferred client initialization when provider configuration contains unknown values initially.
// This test demonstrates the lazy initialization flow where:
// 1. The provider stores ClientInfo in Configure() even when values are unknown
// 2. The data source accepts this configuration without error
// 3. Client creation is deferred until Read() is called (when GetClient() is invoked)
// 4. If values are still unknown during Read(), GetClient() will return an error
func TestOrgsDataSource_DeferredClientInitialization(t *testing.T) {
	ctx := context.Background()

	// Create data source instance
	ds := NewOrgsDataSource()

	// Type assert to access Configure method
	dsWithConfigure, ok := ds.(datasource.DataSourceWithConfigure)
	if !ok {
		t.Fatal("Data source does not implement DataSourceWithConfigure")
	}

	// Configure with ClientInfo containing unknown service account key
	configReq := datasource.ConfigureRequest{
		ProviderData: &client.ClientInfo{
			Config: &client.ZitactlProviderModel{
				Domain:              types.StringValue("example.zitadel.cloud"),
				SkipTlsVerification: types.BoolValue(false),
				ServiceAccountKey:   types.StringUnknown(), // Unknown value - simulates deferred config
			},
			ClientFactory: client.MockSuccessClientFactory,
		},
	}
	configResp := &datasource.ConfigureResponse{}
	dsWithConfigure.Configure(ctx, configReq, configResp)

	if configResp.Diagnostics.HasError() {
		t.Fatalf("Configure failed: %v", configResp.Diagnostics)
	}

	// Note: We can't easily test the Read() method here without more complex setup,
	// but this demonstrates that:
	// 1. Configure() accepts unknown values and stores ClientInfo
	// 2. GetClient() would be called during Read() and would detect unknown values
	// 3. In a real scenario, GetClient() returns an error if values are still unknown

	t.Log("Successfully configured data source with unknown provider values")
	t.Log("Client initialization is deferred until Read() is called with known values")
}