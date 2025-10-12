// Copyright (c) Igor Voronin
// SPDX-License-Identifier: MIT

package client

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"

	oidcClient "github.com/zitadel/oidc/v3/pkg/client"
	"github.com/zitadel/oidc/v3/pkg/oidc"
	"github.com/zitadel/zitadel-go/v3/pkg/client"
	"github.com/zitadel/zitadel-go/v3/pkg/zitadel"
)

// ClientFactory is a function type for creating Zitadel clients.
// This allows for dependency injection in tests.
type ClientFactory func(ctx context.Context, domain string, skipTlsVerification bool, serviceAccountKeyJSON string) (*client.Client, error)

// DefaultClientFactory creates a real Zitadel client using service account authentication.
func DefaultClientFactory(ctx context.Context, domain string, skipTlsVerification bool, serviceAccountKeyJSON string) (*client.Client, error) {
	var zitadelOpts []zitadel.Option
	if skipTlsVerification {
		zitadelOpts = append(zitadelOpts, zitadel.WithInsecureSkipVerifyTLS())
		// Workaround for https://github.com/zitadel/zitadel-go/issues/405: set default http client to also ignore TLS
		if transport, ok := http.DefaultTransport.(*http.Transport); ok {
			transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		}
	}

	// Validate and parse JSON into a KeyFile struct
	// TODO: KeyFile is deprecated, waiting on https://github.com/zitadel/oidc/issues/806
	var keyJson oidcClient.KeyFile //nolint:staticcheck
	if err := json.Unmarshal([]byte(serviceAccountKeyJSON), &keyJson); err != nil {
		return nil, fmt.Errorf("invalid service account key JSON: %w", err)
	}

	// Use JWTAuthentication with the parsed KeyFile
	options := client.WithAuth(
		client.JWTAuthentication(
			&keyJson,
			oidc.ScopeOpenID,
			client.ScopeZitadelAPI(),
		),
	)

	return client.New(ctx, zitadel.New(domain, zitadelOpts...), options)
}
