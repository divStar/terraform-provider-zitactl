// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccOrgsDataSource_Basic tests the case where the org ID can be retrieved successfully.
// This test demonstrates lazy client initialization: the provider stores configuration in Configure(),
// and the Zitadel client is created lazily when the data source's Read() method calls GetClient().
func TestAccOrgsDataSource_Basic(t *testing.T) {
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("Acceptance test - set TF_ACC=1 to run")
	}

	orgName := os.Getenv("ZITACTL_TEST_ORG_NAME")
	if orgName == "" {
		orgName = "Sanctum"
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { TestAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOrgsDataSourceConfig(orgName, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.zitactl_orgs.test", "name", orgName),
					resource.TestCheckResourceAttrSet("data.zitactl_orgs.test", "ids.#"),
				),
			},
		},
	})
}

// TestAccOrgsDataSource_WithMethod tests different query methods.
func TestAccOrgsDataSource_WithMethod(t *testing.T) {
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("Acceptance test - set TF_ACC=1 to run")
	}

	orgName := os.Getenv("ZITACTL_TEST_ORG_NAME")
	if orgName == "" {
		orgName = "Sanctum"
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { TestAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test with explicit equals method
			{
				Config: testAccOrgsDataSourceConfig(orgName, "TEXT_QUERY_METHOD_EQUALS"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.zitactl_orgs.test", "name", orgName),
					resource.TestCheckResourceAttr("data.zitactl_orgs.test", "name_method", "TEXT_QUERY_METHOD_EQUALS"),
					resource.TestCheckResourceAttrSet("data.zitactl_orgs.test", "ids.#"),
				),
			},
			// Test with contains ignore case method
			{
				Config: testAccOrgsDataSourceConfig(orgName, "TEXT_QUERY_METHOD_CONTAINS_IGNORE_CASE"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.zitactl_orgs.test", "name", orgName),
					resource.TestCheckResourceAttr("data.zitactl_orgs.test", "name_method", "TEXT_QUERY_METHOD_CONTAINS_IGNORE_CASE"),
					resource.TestCheckResourceAttrSet("data.zitactl_orgs.test", "ids.#"),
				),
			},
		},
	})
}

// TestAccOrgsDataSource_InvalidMethod tests invalid query method.
func TestAccOrgsDataSource_InvalidMethod(t *testing.T) {
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("Acceptance test - set TF_ACC=1 to run")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { TestAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccOrgsDataSourceConfig("Sanctum", "INVALID_TEXT_QUERY_METHOD"),
				ExpectError: regexp.MustCompile(`Invalid name_method`),
			},
		},
	})
}

// TestAccOrgsDataSource_NotFound tests the case where the org does not exist.
func TestAccOrgsDataSource_NotFound(t *testing.T) {
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("Acceptance test - set TF_ACC=1 to run")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { TestAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOrgsDataSourceConfig("NonExistentOrg123456", ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.zitactl_orgs.test", "name", "NonExistentOrg123456"),
					resource.TestCheckResourceAttr("data.zitactl_orgs.test", "ids.#", "0"),
				),
			},
		},
	})
}

// testAccOrgsDataSourceConfig returns the Terraform configuration for this acceptance test.
func testAccOrgsDataSourceConfig(orgName string, nameMethod string) string {
	methodConfig := ""
	if nameMethod != "" {
		methodConfig = fmt.Sprintf("\n  name_method = %q", nameMethod)
	}

	return fmt.Sprintf(`
data "zitactl_orgs" "test" {
  name = %q%s
}
`, orgName, methodConfig)
}
