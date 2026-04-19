resource "komodo_resource_sync" "example" {
  name = "my-resource-sync"

  source {
    path           = "myorg/infra"
    branch         = "main"
    resource_paths = ["resources/"]
  }

  scope = ["resources", "variables"]

  managed_mode {
    enabled    = true
    tag_filter = ["prod"]
  }
}
