resource "komodo_resource_sync" "example" {
  name = "my-sync"

  lifecycle {
    action_trigger {
      events = [
        after_create,
        after_update,
      ]
      actions = [
        action.komodo_resource_sync.run
      ]
    }
  }
}

action "komodo_resource_sync" "run" {
  config {
    id     = komodo_resource_sync.example.id
    action = "run"
  }
}
