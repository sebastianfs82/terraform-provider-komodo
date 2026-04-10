resource "komodo_resource_sync" "example" {
  name   = "my-resource-sync"
  repo   = "myorg/infra"
  branch = "main"

  resource_path     = ["resources/"]
  managed           = true
  include_variables = true
}
