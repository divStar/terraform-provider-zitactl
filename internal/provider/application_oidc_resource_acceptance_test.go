// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccApplicationOIDCResource_Basic tests the full CRUD lifecycle of an OIDC application.
func TestAccApplicationOIDCResource_Basic(t *testing.T) {
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
				Config: testAccApplicationOIDCResourceConfig(orgName, "test-oidc-app", false, []string{"https://example.com/callback"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("zitactl_application_oidc.test", "name", "test-oidc-app"),
					resource.TestCheckResourceAttr("zitactl_application_oidc.test", "dev_mode", "false"),
					resource.TestCheckResourceAttr("zitactl_application_oidc.test", "redirect_uris.#", "1"),
					resource.TestCheckResourceAttr("zitactl_application_oidc.test", "redirect_uris.0", "https://example.com/callback"),
					resource.TestCheckResourceAttr("zitactl_application_oidc.test", "grant_types.#", "2"),
					resource.TestCheckResourceAttr("zitactl_application_oidc.test", "response_types.#", "1"),
					resource.TestCheckResourceAttr("zitactl_application_oidc.test", "app_type", "OIDC_APP_TYPE_WEB"),
					resource.TestCheckResourceAttr("zitactl_application_oidc.test", "auth_method_type", "OIDC_AUTH_METHOD_TYPE_BASIC"),
					resource.TestCheckResourceAttrSet("zitactl_application_oidc.test", "id"),
					resource.TestCheckResourceAttrSet("zitactl_application_oidc.test", "project_id"),
					resource.TestCheckResourceAttrSet("zitactl_application_oidc.test", "client_id"),
					resource.TestCheckResourceAttrSet("zitactl_application_oidc.test", "client_secret"),
				),
			},
			// Update testing - enable dev mode and add redirect URIs
			{
				Config: testAccApplicationOIDCResourceConfig(orgName, "test-oidc-app", true, []string{"https://example.com/callback", "https://example.com/callback2"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("zitactl_application_oidc.test", "name", "test-oidc-app"),
					resource.TestCheckResourceAttr("zitactl_application_oidc.test", "dev_mode", "true"),
					resource.TestCheckResourceAttr("zitactl_application_oidc.test", "redirect_uris.#", "2"),
					resource.TestCheckResourceAttrSet("zitactl_application_oidc.test", "id"),
				),
			},
			// Update testing - rename application
			{
				Config: testAccApplicationOIDCResourceConfig(orgName, "test-oidc-app-renamed", true, []string{"https://example.com/callback", "https://example.com/callback2"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("zitactl_application_oidc.test", "name", "test-oidc-app-renamed"),
					resource.TestCheckResourceAttrSet("zitactl_application_oidc.test", "id"),
				),
			},
			// Delete testing automatically occurs at the end
		},
	})
}

// TestAccApplicationOIDCResource_WithOptionalFields tests creation with optional fields.
func TestAccApplicationOIDCResource_WithOptionalFields(t *testing.T) {
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
				Config: testAccApplicationOIDCResourceConfigWithOptionalFields(orgName, "test-oidc-full"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("zitactl_application_oidc.test", "name", "test-oidc-full"),
					resource.TestCheckResourceAttr("zitactl_application_oidc.test", "access_token_role_assertion", "true"),
					resource.TestCheckResourceAttr("zitactl_application_oidc.test", "id_token_role_assertion", "true"),
					resource.TestCheckResourceAttr("zitactl_application_oidc.test", "id_token_userinfo_assertion", "true"),
					resource.TestCheckResourceAttr("zitactl_application_oidc.test", "additional_origins.#", "1"),
					resource.TestCheckResourceAttr("zitactl_application_oidc.test", "additional_origins.0", "https://example.com"),
					resource.TestCheckResourceAttr("zitactl_application_oidc.test", "post_logout_redirect_uris.#", "1"),
					resource.TestCheckResourceAttr("zitactl_application_oidc.test", "post_logout_redirect_uris.0", "https://example.com/logout"),
					resource.TestCheckResourceAttrSet("zitactl_application_oidc.test", "id"),
				),
			},
		},
	})
}

// TestAccApplicationOIDCResource_ProjectIdChangeRequiresReplace tests that changing project_id forces replacement.
func TestAccApplicationOIDCResource_ProjectIdChangeRequiresReplace(t *testing.T) {
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
			// Create initial application
			{
				Config: testAccApplicationOIDCResourceConfig(orgName, "test-app-project-change", false, []string{"https://example.com/callback"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("zitactl_application_oidc.test", "name", "test-app-project-change"),
					resource.TestCheckResourceAttrSet("zitactl_application_oidc.test", "id"),
					resource.TestCheckResourceAttrSet("zitactl_application_oidc.test", "project_id"),
				),
			},
			// Attempt to change project_id (should force replacement)
			{
				Config:             testAccApplicationOIDCResourceConfigDifferentProject(orgName, "test-app-project-change"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestAccApplicationOIDCResource_Import tests the import functionality
func TestAccApplicationOIDCResource_Import(t *testing.T) {
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
				Config: testAccApplicationOIDCResourceConfig(orgName, "test-oidc-import", false, []string{"https://example.com/callback"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("zitactl_application_oidc.test", "name", "test-oidc-import"),
					resource.TestCheckResourceAttrSet("zitactl_application_oidc.test", "id"),
					resource.TestCheckResourceAttrSet("zitactl_application_oidc.test", "project_id"),
				),
			},
			// Test import
			{
				// Import ID format: project_id:app_id
				ResourceName:      "zitactl_application_oidc.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["zitactl_application_oidc.test"]
					if !ok {
						return "", fmt.Errorf("resource not found")
					}
					projectId := rs.Primary.Attributes["project_id"]
					appId := rs.Primary.ID
					return fmt.Sprintf("%s:%s", projectId, appId), nil
				},
				ImportStateVerifyIgnore: []string{"client_secret", "client_id"},
			},
		},
	})
}

// testAccApplicationOIDCResourceConfig returns the Terraform configuration for basic OIDC app test.
func testAccApplicationOIDCResourceConfig(orgName, appName string, devMode bool, redirectUris []string) string {
	redirectUrisHCL := ""
	for _, uri := range redirectUris {
		redirectUrisHCL += fmt.Sprintf("    %q,\n", uri)
	}

	return fmt.Sprintf(`
data "zitactl_orgs" "test" {
  name = %[1]q
}

resource "zitactl_project" "test" {
  name   = "test-project-for-oidc"
  org_id = data.zitactl_orgs.test.ids[0]
}

resource "zitactl_application_oidc" "test" {
  name       = %[2]q
  project_id = zitactl_project.test.id

  redirect_uris = [
%[3]s  ]

  grant_types = [
    "OIDC_GRANT_TYPE_AUTHORIZATION_CODE",
    "OIDC_GRANT_TYPE_REFRESH_TOKEN"
  ]

  response_types = [
    "OIDC_RESPONSE_TYPE_CODE"
  ]

  app_type         = "OIDC_APP_TYPE_WEB"
  auth_method_type = "OIDC_AUTH_METHOD_TYPE_BASIC"
  dev_mode         = %[4]t
}
`, orgName, appName, redirectUrisHCL, devMode)
}

// testAccApplicationOIDCResourceConfigWithOptionalFields includes all optional fields.
func testAccApplicationOIDCResourceConfigWithOptionalFields(orgName, appName string) string {
	return fmt.Sprintf(`
data "zitactl_orgs" "test" {
  name = %[1]q
}

resource "zitactl_project" "test" {
  name   = "test-project-for-oidc"
  org_id = data.zitactl_orgs.test.ids[0]
}

resource "zitactl_application_oidc" "test" {
  name       = %[2]q
  project_id = zitactl_project.test.id

  redirect_uris = ["https://example.com/callback"]

  grant_types = [
    "OIDC_GRANT_TYPE_AUTHORIZATION_CODE",
    "OIDC_GRANT_TYPE_REFRESH_TOKEN"
  ]

  response_types = [
    "OIDC_RESPONSE_TYPE_CODE"
  ]

  app_type         = "OIDC_APP_TYPE_WEB"
  auth_method_type = "OIDC_AUTH_METHOD_TYPE_BASIC"
  dev_mode         = false

  access_token_role_assertion  = true
  id_token_role_assertion      = true
  id_token_userinfo_assertion  = true
  
  additional_origins = ["https://example.com"]
  post_logout_redirect_uris = ["https://example.com/logout"]

  access_token_type = "OIDC_TOKEN_TYPE_BEARER"
  version           = "OIDC_VERSION_1_0"
}
`, orgName, appName)
}

// testAccApplicationOIDCResourceConfigDifferentProject creates app with a different project.
func testAccApplicationOIDCResourceConfigDifferentProject(orgName, appName string) string {
	return fmt.Sprintf(`
data "zitactl_orgs" "test" {
  name = %[1]q
}

resource "zitactl_project" "test" {
  name   = "test-project-for-oidc"
  org_id = data.zitactl_orgs.test.ids[0]
}

resource "zitactl_project" "test2" {
  name   = "test-project-for-oidc-2"
  org_id = data.zitactl_orgs.test.ids[0]
}

resource "zitactl_application_oidc" "test" {
  name       = %[2]q
  project_id = zitactl_project.test2.id

  redirect_uris = ["https://example.com/callback"]

  grant_types = [
    "OIDC_GRANT_TYPE_AUTHORIZATION_CODE",
    "OIDC_GRANT_TYPE_REFRESH_TOKEN"
  ]

  response_types = [
    "OIDC_RESPONSE_TYPE_CODE"
  ]

  app_type         = "OIDC_APP_TYPE_WEB"
  auth_method_type = "OIDC_AUTH_METHOD_TYPE_BASIC"
}
`, orgName, appName)
}
