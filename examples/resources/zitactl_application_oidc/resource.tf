resource "zitactl_application_oidc" "this" {
  project_id = zitactl_project.this.id

  name                      = "myproject"
  redirect_uris             = ["https://myproject.${var.cluster.domain}/oauth2/authorize"]
  response_types            = ["OIDC_RESPONSE_TYPE_CODE"]
  grant_types               = ["OIDC_GRANT_TYPE_AUTHORIZATION_CODE"]
  app_type                  = "OIDC_APP_TYPE_WEB"
  auth_method_type          = "OIDC_AUTH_METHOD_TYPE_BASIC"
  post_logout_redirect_uris = ["https://myproject.${var.cluster.domain}/"]
}