# List all deployments
data "komodo_deployments" "all" {}

# List deployments running on a specific server
data "komodo_deployments" "on_server" {
  server_id = "6627c3e4f1a2b3c4d5e6f7a8"
}
