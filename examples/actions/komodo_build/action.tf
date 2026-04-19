resource "komodo_build" "example" {
  name       = "my-build"
  builder_id = komodo_builder.example.id

  source {
    provider = "Github"
    repo     = "myorg/myrepo"
    branch   = "main"
  }

  lifecycle {
    action_trigger {
      events = [
        after_create,
        after_update,
      ]
      actions = [
        action.komodo_build.run
      ]
    }
  }
}

action "komodo_build" "run" {
  config {
    id     = komodo_build.example.id
    action = "run"
  }
}
