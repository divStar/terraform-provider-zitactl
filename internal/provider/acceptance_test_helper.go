// Copyright (c) Igor Voronin
// SPDX-License-Identifier: MIT

package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
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
		t.Fatal("ZITACTL_DOMAIN must be set for acceptance tests")
	}
	if v := os.Getenv("ZITACTL_SERVICE_ACCOUNT_KEY"); v == "" {
		t.Fatal("ZITACTL_SERVICE_ACCOUNT_KEY must be set for acceptance tests")
	}
}
