resource "komodo_repo" "example" {
  name      = "my-repo"
  server_id = komodo_server.example.id

  source {
    path   = "myorg/myrepo"
    branch = "main"
  }

  lifecycle {
    action_trigger {
      events = [
        after_create,
        after_update,
      ]
      actions = [
        action.komodo_repo.build
      ]
    }
  }
}

# Build the repo whenever it is created or updated.
action "komodo_repo" "build" {
  config {
    id     = komodo_repo.example.id
    action = "build"
  }
}

# Clone the repo onto its attached server.
action "komodo_repo" "clone" {
  config {
    id     = komodo_repo.example.id
    action = "clone"
  }
}

# Pull the latest commits.
action "komodo_repo" "pull" {
  config {
    id     = komodo_repo.example.id
    action = "pull"
  }
}
