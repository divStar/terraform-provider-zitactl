// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package client

import "github.com/hashicorp/terraform-plugin-framework/types"

// ZitactlProviderModel describes the provider configuration data model.
// This struct belongs to the `client` package rather than the `provider` package
// to allow lazy client initialization and to avoid a circular dependency hell.
type ZitactlProviderModel struct {
	Domain              types.String `tfsdk:"domain"`
	SkipTlsVerification types.Bool   `tfsdk:"skip_tls_verification"`
	ServiceAccountKey   types.String `tfsdk:"service_account_key"`
}
