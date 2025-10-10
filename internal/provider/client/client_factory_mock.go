// Copyright (c) Igor Voronin
// SPDX-License-Identifier: MIT

package client

import (
	"context"
	"errors"

	"github.com/zitadel/zitadel-go/v3/pkg/client"
)

// MockSuccessClientFactory creates a client factory that always succeeds.
// Used for testing successful provider configuration scenarios.
func MockSuccessClientFactory(ctx context.Context, domain string, skipTlsVerification bool, serviceAccountKeyJSON string) (*client.Client, error) {
	return nil, nil // No client, no error
}

// MockFailureClientFactory creates a client factory that always fails.
// Used for testing error handling when client creation fails.
func MockFailureClientFactory(ctx context.Context, domain string, skipTlsVerification bool, serviceAccountKeyJSON string) (*client.Client, error) {
	return nil, errors.New("mock connection failure")
}
