# List all repos
data "komodo_repos" "all" {}

# List repos cloned on a specific server
data "komodo_repos" "on_server" {
  server_id = "6627c3e4f1a2b3c4d5e6f7a8"
}

# List repos using a specific builder
data "komodo_repos" "by_builder" {
  builder_id = "6627c3e4f1a2b3c4d5e6f7a8"
}
