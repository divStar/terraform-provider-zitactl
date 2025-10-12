// Copyright (c) Igor Voronin
// SPDX-License-Identifier: MIT

package client

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/zitadel/zitadel-go/v3/pkg/client"
)

// ClientInfo contains provider configuration and factory for lazy client creation.
type ClientInfo struct {
	Config        *ZitactlProviderModel
	ClientFactory ClientFactory
	Client        *client.Client // if a Client is already created, it will be returned
}

// GetClient creates or returns the Zitadel client, only when all config values are known.
func (ci *ClientInfo) GetClient(ctx context.Context) (*client.Client, error) {
	// Check for missing config
	if ci.Config == nil {
		return nil, fmt.Errorf("provider is not configured")
	}

	// Return client if already created
	if ci.Client != nil {
		return ci.Client, nil
	}

	// Check for unknown values
	if ci.Config.Domain.IsUnknown() || ci.Config.SkipTlsVerification.IsUnknown() || ci.Config.ServiceAccountKey.IsUnknown() {
		unknownFields := getUnknownFieldNames(*ci.Config)
		return nil, fmt.Errorf("provider configuration contains unknown values: %s", strings.Join(unknownFields, ", "))
	}

	// Get configuration values
	domain := ci.Config.Domain.ValueString()
	if domain == "" {
		domain = os.Getenv("ZITACTL_DOMAIN")
	}
	if domain == "" {
		return nil, fmt.Errorf("the 'domain' attribute must be set")
	}

	skipTlsVerification := ci.Config.SkipTlsVerification.ValueBool()
	if ci.Config.SkipTlsVerification.IsNull() {
		skipTlsVerificationEnv := os.Getenv("ZITACTL_SKIP_TLS_VERIFICATION")
		skipTlsVerification = skipTlsVerificationEnv == "true" || skipTlsVerificationEnv == "1"
	}

	serviceAccountKey := ci.Config.ServiceAccountKey.ValueString()
	if serviceAccountKey == "" {
		serviceAccountKey = os.Getenv("ZITACTL_SERVICE_ACCOUNT_KEY")
	}
	if serviceAccountKey == "" {
		return nil, fmt.Errorf("the 'service_account_key' attribute must be set")
	}

	// Create client
	clientFactory := ci.ClientFactory
	if clientFactory == nil {
		clientFactory = DefaultClientFactory
	}

	zitadelClient, err := clientFactory(ctx, domain, skipTlsVerification, serviceAccountKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create Zitadel client: %w", err)
	}
	return zitadelClient, nil
}

// getUnknownFieldNames returns the names of any unknown fields in the provider configuration.
func getUnknownFieldNames(data ZitactlProviderModel) []string {
	var unknownFields []string

	if data.Domain.IsUnknown() {
		unknownFields = append(unknownFields, "domain")
	}
	if data.SkipTlsVerification.IsUnknown() {
		unknownFields = append(unknownFields, "skip_tls_verification")
	}
	if data.ServiceAccountKey.IsUnknown() {
		unknownFields = append(unknownFields, "service_account_key")
	}

	return unknownFields
}
