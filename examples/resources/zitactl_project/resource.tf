resource "zitactl_project" "this" {
  name                   = "myproject"
  org_id                 = local.myproject_id
  project_role_assertion = true
  project_role_check     = true
}