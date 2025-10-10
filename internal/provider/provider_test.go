// Copyright (c) Igor Voronin
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/divStar/terraform-provider-zitactl/internal/provider/client"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

const (
	// testServiceAccountKey is a dummy service account key file used in tests.
	// This is not a real token and cannot be used for authentication.
	testServiceAccountKey = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."

	// testDomain is a test Zitadel domain.
	testDomain = "example.zitadel.cloud"

	// testDomainLocalhost is a test localhost domain.
	testDomainLocalhost = "localhost:8080"
)

func TestZitactlProvider_Configure(t *testing.T) {
	tests := []struct {
		name          string
		config        client.ZitactlProviderModel
		envVars       map[string]string
		clientFactory client.ClientFactory
		expectError   bool
		errorContains string
	}{
		{
			name: "successful configuration with explicit values",
			config: client.ZitactlProviderModel{
				Domain:              types.StringValue(testDomain),
				SkipTlsVerification: types.BoolValue(false),
				ServiceAccountKey:   types.StringValue(testServiceAccountKey),
			},
			clientFactory: client.MockSuccessClientFactory,
			expectError:   false,
		},
		{
			name: "successful configuration with insecure connection",
			config: client.ZitactlProviderModel{
				Domain:              types.StringValue(testDomainLocalhost),
				SkipTlsVerification: types.BoolValue(true),
				ServiceAccountKey:   types.StringValue(testServiceAccountKey),
			},
			clientFactory: client.MockSuccessClientFactory,
			expectError:   false,
		},
		{
			name: "successful configuration with environment variables",
			config: client.ZitactlProviderModel{
				Domain:              types.StringNull(),
				SkipTlsVerification: types.BoolNull(),
				ServiceAccountKey:   types.StringNull(),
			},
			envVars: map[string]string{
				"ZITACTL_DOMAIN":                testDomain,
				"ZITACTL_SKIP_TLS_VERIFICATION": "false",
				"ZITACTL_SERVICE_ACCOUNT_KEY":   testServiceAccountKey,
			},
			clientFactory: client.MockSuccessClientFactory,
			expectError:   false,
		},
		{
			name: "configuration with unknown domain value - stores config",
			config: client.ZitactlProviderModel{
				Domain:              types.StringUnknown(),
				SkipTlsVerification: types.BoolValue(false),
				ServiceAccountKey:   types.StringValue(testServiceAccountKey),
			},
			clientFactory: client.MockSuccessClientFactory,
			expectError:   false,
		},
		{
			name: "configuration with unknown skip_tls_verification value - stores config",
			config: client.ZitactlProviderModel{
				Domain:              types.StringValue(testDomain),
				SkipTlsVerification: types.BoolUnknown(),
				ServiceAccountKey:   types.StringValue(testServiceAccountKey),
			},
			clientFactory: client.MockSuccessClientFactory,
			expectError:   false,
		},
		{
			name: "configuration with unknown service_account_key value - stores config",
			config: client.ZitactlProviderModel{
				Domain:              types.StringValue(testDomain),
				SkipTlsVerification: types.BoolValue(false),
				ServiceAccountKey:   types.StringUnknown(),
			},
			clientFactory: client.MockSuccessClientFactory,
			expectError:   false,
		},
		{
			name: "configuration with all values unknown - stores config",
			config: client.ZitactlProviderModel{
				Domain:              types.StringUnknown(),
				SkipTlsVerification: types.BoolUnknown(),
				ServiceAccountKey:   types.StringUnknown(),
			},
			clientFactory: client.MockSuccessClientFactory,
			expectError:   false,
		},
		{
			name: "configuration with null values - stores config",
			config: client.ZitactlProviderModel{
				Domain:              types.StringNull(),
				SkipTlsVerification: types.BoolNull(),
				ServiceAccountKey:   types.StringNull(),
			},
			clientFactory: client.MockSuccessClientFactory,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment variables before each test
			clearTestEnvVars()

			// Set test-specific environment variables
			for key, value := range tt.envVars {
				if err := os.Setenv(key, value); err != nil {
					t.Fatalf("Failed to set environment variable %s: %v", key, err)
				}
			}
			defer clearTestEnvVars()

			// Create provider instance with mock client factory
			p := &ZitactlProvider{
				version:       "test",
				clientFactory: tt.clientFactory,
			}

			// Create schema for conversion
			ctx := context.Background()
			schemaReq := provider.SchemaRequest{}
			schemaResp := &provider.SchemaResponse{}
			p.Schema(ctx, schemaReq, schemaResp)

			if schemaResp.Diagnostics.HasError() {
				t.Fatalf("Schema() returned errors: %v", schemaResp.Diagnostics)
			}

			// Convert config to tftypes.Value
			configValue := tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"domain":                tftypes.String,
						"skip_tls_verification": tftypes.Bool,
						"service_account_key":   tftypes.String,
					},
				},
				map[string]tftypes.Value{
					"domain":                convertTypesStringToTFType(tt.config.Domain),
					"skip_tls_verification": convertTypesBoolToTFType(tt.config.SkipTlsVerification),
					"service_account_key":   convertTypesStringToTFType(tt.config.ServiceAccountKey),
				},
			)

			// Create config
			config := tfsdk.Config{
				Schema: schemaResp.Schema,
				Raw:    configValue,
			}

			// Create configure request
			req := provider.ConfigureRequest{
				Config: config,
				ClientCapabilities: provider.ConfigureProviderClientCapabilities{
					DeferralAllowed: true,
				},
			}

			// Create configure response
			resp := &provider.ConfigureResponse{}

			// Call Configure
			p.Configure(ctx, req, resp)

			// Check for errors
			if tt.expectError {
				if !resp.Diagnostics.HasError() {
					t.Errorf("Expected error but got none")
				}
				if tt.errorContains != "" {
					found := false
					for _, diag := range resp.Diagnostics.Errors() {
						if strings.Contains(diag.Summary(), tt.errorContains) || strings.Contains(diag.Detail(), tt.errorContains) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected error containing '%s', but got: %v", tt.errorContains, resp.Diagnostics)
					}
				}
			} else {
				if resp.Diagnostics.HasError() {
					t.Errorf("Unexpected error: %v", resp.Diagnostics)
				}
			}

			// With lazy initialization, Configure should always succeed (no deferral)
			// and should always set ClientInfo
			if !tt.expectError {
				if resp.Deferred != nil {
					t.Errorf("Did not expect deferred response with lazy initialization, but got: %v", resp.Deferred)
				}

				// Verify ClientInfo was set
				if resp.DataSourceData == nil {
					t.Error("Expected DataSourceData (ClientInfo) to be set but was nil")
				} else {
					clientInfo, ok := resp.DataSourceData.(*client.ClientInfo)
					if !ok {
						t.Error("Expected DataSourceData to be *client.ClientInfo")
					} else {
						if clientInfo.Config == nil {
							t.Error("Expected ClientInfo.Config to be set but was nil")
						}
						if clientInfo.ClientFactory == nil && tt.clientFactory != nil {
							t.Error("Expected ClientInfo.ClientFactory to be set but was nil")
						}
					}
				}

				if resp.ResourceData == nil {
					t.Error("Expected ResourceData (ClientInfo) to be set but was nil")
				} else {
					clientInfo, ok := resp.ResourceData.(*client.ClientInfo)
					if !ok {
						t.Error("Expected ResourceData to be *client.ClientInfo")
					} else {
						if clientInfo.Config == nil {
							t.Error("Expected ClientInfo.Config to be set but was nil")
						}
						if clientInfo.ClientFactory == nil && tt.clientFactory != nil {
							t.Error("Expected ClientInfo.ClientFactory to be set but was nil")
						}
					}
				}
			}
		})
	}
}

// TestZitactlProvider_Configure_StoresConfigForLaterUse verifies that the provider
// stores configuration even when values are unknown, allowing lazy client initialization.
func TestZitactlProvider_Configure_StoresConfigForLaterUse(t *testing.T) {
	ctx := context.Background()
	p := &ZitactlProvider{
		version:       "test",
		clientFactory: client.MockSuccessClientFactory,
	}

	schemaReq := provider.SchemaRequest{}
	schemaResp := &provider.SchemaResponse{}
	p.Schema(ctx, schemaReq, schemaResp)

	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("Schema() returned errors: %v", schemaResp.Diagnostics)
	}

	t.Log("Phase 1: Configure with unknown service account key value")

	unknownConfig := client.ZitactlProviderModel{
		Domain:              types.StringValue(testDomain),
		SkipTlsVerification: types.BoolValue(false),
		ServiceAccountKey:   types.StringUnknown(),
	}

	unknownConfigValue := tftypes.NewValue(
		tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"domain":                tftypes.String,
				"skip_tls_verification": tftypes.Bool,
				"service_account_key":   tftypes.String,
			},
		},
		map[string]tftypes.Value{
			"domain":                convertTypesStringToTFType(unknownConfig.Domain),
			"skip_tls_verification": convertTypesBoolToTFType(unknownConfig.SkipTlsVerification),
			"service_account_key":   convertTypesStringToTFType(unknownConfig.ServiceAccountKey),
		},
	)

	req1 := provider.ConfigureRequest{
		Config: tfsdk.Config{
			Schema: schemaResp.Schema,
			Raw:    unknownConfigValue,
		},
		ClientCapabilities: provider.ConfigureProviderClientCapabilities{
			DeferralAllowed: true,
		},
	}

	resp1 := &provider.ConfigureResponse{}
	p.Configure(ctx, req1, resp1)

	if resp1.Diagnostics.HasError() {
		t.Fatalf("Phase 1 returned unexpected error: %v", resp1.Diagnostics)
	}

	// With lazy initialization, no deferral should occur
	if resp1.Deferred != nil {
		t.Error("Phase 1: Did not expect deferred response with lazy initialization")
	}

	// ClientInfo should be set even with unknown values
	if resp1.DataSourceData == nil {
		t.Fatal("Phase 1: DataSourceData should be set even with unknown values")
	}

	clientInfo1, ok := resp1.DataSourceData.(*client.ClientInfo)
	if !ok {
		t.Fatal("Phase 1: DataSourceData should be *client.ClientInfo")
	}

	if clientInfo1.Config == nil {
		t.Error("Phase 1: ClientInfo.Config should be set")
	}

	t.Log("Phase 1: Successfully stored configuration with unknown values")

	t.Log("Phase 2: Configure with known service account key value")

	knownConfig := client.ZitactlProviderModel{
		Domain:              types.StringValue(testDomain),
		SkipTlsVerification: types.BoolValue(false),
		ServiceAccountKey:   types.StringValue(testServiceAccountKey),
	}

	knownConfigValue := tftypes.NewValue(
		tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"domain":                tftypes.String,
				"skip_tls_verification": tftypes.Bool,
				"service_account_key":   tftypes.String,
			},
		},
		map[string]tftypes.Value{
			"domain":                convertTypesStringToTFType(knownConfig.Domain),
			"skip_tls_verification": convertTypesBoolToTFType(knownConfig.SkipTlsVerification),
			"service_account_key":   convertTypesStringToTFType(knownConfig.ServiceAccountKey),
		},
	)

	req2 := provider.ConfigureRequest{
		Config: tfsdk.Config{
			Schema: schemaResp.Schema,
			Raw:    knownConfigValue,
		},
		ClientCapabilities: provider.ConfigureProviderClientCapabilities{
			DeferralAllowed: true,
		},
	}

	resp2 := &provider.ConfigureResponse{}
	p.Configure(ctx, req2, resp2)

	if resp2.Deferred != nil {
		t.Error("Phase 2: Should not be deferred")
	}

	if resp2.Diagnostics.HasError() {
		t.Errorf("Phase 2: Unexpected error: %v", resp2.Diagnostics)
	}

	if resp2.DataSourceData == nil {
		t.Fatal("Phase 2: DataSourceData should be set")
	}

	clientInfo2, ok := resp2.DataSourceData.(*client.ClientInfo)
	if !ok {
		t.Fatal("Phase 2: DataSourceData should be *client.ClientInfo")
	}

	if clientInfo2.Config == nil {
		t.Error("Phase 2: ClientInfo.Config should be set")
	}

	// Verify the config has known values now
	if clientInfo2.Config.ServiceAccountKey.IsUnknown() {
		t.Error("Phase 2: ServiceAccountKey should be known now")
	}

	t.Log("Phase 2: Successfully stored configuration with known values")
}

// TestZitactlProvider_NewWithClientFactory verifies that NewWithClientFactory
// correctly injects a custom client factory into the provider.
func TestZitactlProvider_NewWithClientFactory(t *testing.T) {
	customFactory := client.MockFailureClientFactory
	providerFunc := NewWithClientFactory("test-version", customFactory)

	p := providerFunc()
	zp, ok := p.(*ZitactlProvider)
	if !ok {
		t.Fatal("Expected provider to be of type *ZitactlProvider")
	}

	if zp.version != "test-version" {
		t.Errorf("Expected version 'test-version', got '%s'", zp.version)
	}

	if zp.clientFactory == nil {
		t.Fatal("Expected clientFactory to be set, but was nil")
	}

	// Verify the factory is stored in ClientInfo
	ctx := context.Background()

	schemaReq := provider.SchemaRequest{}
	schemaResp := &provider.SchemaResponse{}
	zp.Schema(ctx, schemaReq, schemaResp)

	config := client.ZitactlProviderModel{
		Domain:              types.StringValue(testDomain),
		SkipTlsVerification: types.BoolValue(false),
		ServiceAccountKey:   types.StringValue(testServiceAccountKey),
	}

	configValue := tftypes.NewValue(
		tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"domain":                tftypes.String,
				"skip_tls_verification": tftypes.Bool,
				"service_account_key":   tftypes.String,
			},
		},
		map[string]tftypes.Value{
			"domain":                convertTypesStringToTFType(config.Domain),
			"skip_tls_verification": convertTypesBoolToTFType(config.SkipTlsVerification),
			"service_account_key":   convertTypesStringToTFType(config.ServiceAccountKey),
		},
	)

	req := provider.ConfigureRequest{
		Config: tfsdk.Config{
			Schema: schemaResp.Schema,
			Raw:    configValue,
		},
		ClientCapabilities: provider.ConfigureProviderClientCapabilities{
			DeferralAllowed: false,
		},
	}

	resp := &provider.ConfigureResponse{}
	zp.Configure(ctx, req, resp)

	// With lazy initialization, Configure should succeed
	if resp.Diagnostics.HasError() {
		t.Errorf("Unexpected error during Configure: %v", resp.Diagnostics)
	}

	// Verify ClientInfo contains the custom factory
	if resp.DataSourceData == nil {
		t.Fatal("Expected DataSourceData to be set")
	}

	clientInfo, ok := resp.DataSourceData.(*client.ClientInfo)
	if !ok {
		t.Fatal("Expected DataSourceData to be *client.ClientInfo")
	}

	if clientInfo.ClientFactory == nil {
		t.Error("Expected ClientInfo.ClientFactory to be set")
	}
}

// clearTestEnvVars clears all environment variables used in tests.
func clearTestEnvVars() {
	_ = os.Unsetenv("ZITACTL_DOMAIN")
	_ = os.Unsetenv("ZITACTL_SKIP_TLS_VERIFICATION")
	_ = os.Unsetenv("ZITACTL_SERVICE_ACCOUNT_KEY")
}
