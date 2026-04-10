# List all builds
data "komodo_builds" "all" {}

# List builds using a specific builder
data "komodo_builds" "by_builder" {
  builder_id = "6627c3e4f1a2b3c4d5e6f7a8"
}

# List builds sourced from a specific linked repo
data "komodo_builds" "from_repo" {
  repo_id = "my-repo-name"
}
