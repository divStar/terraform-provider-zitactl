// Copyright (c) Igor Voronin
// SPDX-License-Identifier: MIT

package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccProjectResource_Basic tests the full CRUD lifecycle of a project.
func TestAccProjectResource_Basic(t *testing.T) {
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
			// Create and Read testing
			{
				Config: testAccProjectResourceConfig(orgName, "test-project", false, false, false, "PRIVATE_LABELING_SETTING_UNSPECIFIED"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("zitactl_project.test", "name", "test-project"),
					resource.TestCheckResourceAttr("zitactl_project.test", "project_role_assertion", "false"),
					resource.TestCheckResourceAttr("zitactl_project.test", "project_role_check", "false"),
					resource.TestCheckResourceAttr("zitactl_project.test", "has_project_check", "false"),
					resource.TestCheckResourceAttr("zitactl_project.test", "private_labeling_setting", "PRIVATE_LABELING_SETTING_UNSPECIFIED"),
					resource.TestCheckResourceAttrSet("zitactl_project.test", "id"),
					resource.TestCheckResourceAttrSet("zitactl_project.test", "org_id"),
					resource.TestCheckResourceAttrSet("zitactl_project.test", "state"),
				),
			},
			// Update testing - change booleans
			{
				Config: testAccProjectResourceConfig(orgName, "test-project", true, true, false, "PRIVATE_LABELING_SETTING_UNSPECIFIED"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("zitactl_project.test", "name", "test-project"),
					resource.TestCheckResourceAttr("zitactl_project.test", "project_role_assertion", "true"),
					resource.TestCheckResourceAttr("zitactl_project.test", "project_role_check", "true"),
					resource.TestCheckResourceAttr("zitactl_project.test", "has_project_check", "false"),
					resource.TestCheckResourceAttrSet("zitactl_project.test", "id"),
				),
			},
			// Update testing - change private labeling setting
			{
				Config: testAccProjectResourceConfig(orgName, "test-project", true, true, false, "PRIVATE_LABELING_SETTING_ENFORCE_PROJECT_RESOURCE_OWNER_POLICY"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("zitactl_project.test", "name", "test-project"),
					resource.TestCheckResourceAttr("zitactl_project.test", "private_labeling_setting", "PRIVATE_LABELING_SETTING_ENFORCE_PROJECT_RESOURCE_OWNER_POLICY"),
					resource.TestCheckResourceAttrSet("zitactl_project.test", "id"),
				),
			},
			// Update testing - change all booleans
			{
				Config: testAccProjectResourceConfig(orgName, "test-project", false, false, true, "PRIVATE_LABELING_SETTING_ALLOW_LOGIN_USER_RESOURCE_OWNER_POLICY"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("zitactl_project.test", "name", "test-project"),
					resource.TestCheckResourceAttr("zitactl_project.test", "project_role_assertion", "false"),
					resource.TestCheckResourceAttr("zitactl_project.test", "project_role_check", "false"),
					resource.TestCheckResourceAttr("zitactl_project.test", "has_project_check", "true"),
					resource.TestCheckResourceAttr("zitactl_project.test", "private_labeling_setting", "PRIVATE_LABELING_SETTING_ALLOW_LOGIN_USER_RESOURCE_OWNER_POLICY"),
					resource.TestCheckResourceAttrSet("zitactl_project.test", "id"),
				),
			},
			// Update testing - rename project
			{
				Config: testAccProjectResourceConfig(orgName, "test-project-renamed", false, false, true, "PRIVATE_LABELING_SETTING_ALLOW_LOGIN_USER_RESOURCE_OWNER_POLICY"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("zitactl_project.test", "name", "test-project-renamed"),
					resource.TestCheckResourceAttrSet("zitactl_project.test", "id"),
				),
			},
			// Delete testing automatically occurs at the end
		},
	})
}

// TestAccProjectResource_InvalidOrgId tests that creating a project with invalid org_id fails.
func TestAccProjectResource_InvalidOrgId(t *testing.T) {
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("Acceptance test - set TF_ACC=1 to run")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { TestAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccProjectResourceConfigWithInvalidOrgId("test-project-invalid-org", "invalid-org-id-123456"),
				ExpectError: regexp.MustCompile(`Error creating project|rpc error|invalid`),
			},
		},
	})
}

// TestAccProjectResource_MissingOrgId tests that org_id is required.
func TestAccProjectResource_MissingOrgId(t *testing.T) {
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("Acceptance test - set TF_ACC=1 to run")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { TestAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccProjectResourceConfigWithoutOrgId("test-project-no-org"),
				ExpectError: regexp.MustCompile(`Missing required argument|The argument "org_id" is required`),
			},
		},
	})
}

// TestAccProjectResource_OrgIdChangeRequiresReplace tests that changing org_id forces replacement.
func TestAccProjectResource_OrgIdChangeRequiresReplace(t *testing.T) {
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
			// Create initial project
			{
				Config: testAccProjectResourceConfig(orgName, "test-project-org-change", false, false, false, "PRIVATE_LABELING_SETTING_UNSPECIFIED"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("zitactl_project.test", "name", "test-project-org-change"),
					resource.TestCheckResourceAttrSet("zitactl_project.test", "id"),
					resource.TestCheckResourceAttrSet("zitactl_project.test", "org_id"),
				),
			},
			// Attempt to change org_id (should force replacement and fail on invalid org)
			{
				Config:      testAccProjectResourceConfigWithInvalidOrgId("test-project-org-change", "different-org-id"),
				ExpectError: regexp.MustCompile(`Invalid Organization ID|does not exist`),
			},
		},
	})
}

// TestAccProjectResource_Import tests the import functionality
func TestAccProjectResource_Import(t *testing.T) {
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
			// Create the resource first
			{
				Config: testAccProjectResourceConfig(orgName, "test-project-import", false, false, false, "PRIVATE_LABELING_SETTING_UNSPECIFIED"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("zitactl_project.test", "name", "test-project-import"),
					resource.TestCheckResourceAttrSet("zitactl_project.test", "id"),
					resource.TestCheckResourceAttrSet("zitactl_project.test", "org_id"),
				),
			},
			// Test import
			{
				ResourceName:      "zitactl_project.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// testAccProjectResourceConfig returns the Terraform configuration for the project resource test.
func testAccProjectResourceConfig(orgName, projectName string, roleAssertion, roleCheck, projectCheck bool, privateLabelingSetting string) string {
	return fmt.Sprintf(`
data "zitactl_orgs" "test" {
  name = %[1]q
}

resource "zitactl_project" "test" {
  name                     = %[2]q
  org_id                   = data.zitactl_orgs.test.ids[0]
  project_role_assertion   = %[3]t
  project_role_check       = %[4]t
  has_project_check        = %[5]t
  private_labeling_setting = %[6]q
}
`, orgName, projectName, roleAssertion, roleCheck, projectCheck, privateLabelingSetting)
}

// testAccProjectResourceConfigWithInvalidOrgId returns configuration with an invalid org_id.
func testAccProjectResourceConfigWithInvalidOrgId(projectName, orgId string) string {
	return fmt.Sprintf(`
resource "zitactl_project" "test" {
  name                   = %[1]q
  org_id                 = %[2]q
  project_role_assertion = false
  project_role_check     = false
  has_project_check      = false
}
`, projectName, orgId)
}

// testAccProjectResourceConfigWithoutOrgId returns configuration without org_id (should fail validation).
func testAccProjectResourceConfigWithoutOrgId(projectName string) string {
	return fmt.Sprintf(`
resource "zitactl_project" "test" {
  name                   = %[1]q
  project_role_assertion = false
  project_role_check     = false
  has_project_check      = false
}
`, projectName)
}
