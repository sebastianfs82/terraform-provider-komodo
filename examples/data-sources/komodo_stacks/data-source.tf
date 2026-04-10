# List all stacks
data "komodo_stacks" "all" {}

# List stacks running on a specific server
data "komodo_stacks" "on_server" {
  server_id = "6627c3e4f1a2b3c4d5e6f7a8"
}

# List stacks sourced from a specific linked repo
data "komodo_stacks" "from_repo" {
  repo_id = "my-repo-name"
}
