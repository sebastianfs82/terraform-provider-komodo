# List all resource syncs
data "komodo_resource_syncs" "all" {}

# List resource syncs sourced from a specific linked repo
data "komodo_resource_syncs" "from_repo" {
  repo_id = "my-repo-name"
}
