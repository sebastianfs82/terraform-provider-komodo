resource "komodo_action" "example" {
  name = "my-action"

  lifecycle {
    action_trigger {
      events = [
        after_create,
        after_update,
      ]
      actions = [
        action.komodo_action.run
      ]
    }
  }
}

action "komodo_action" "run" {
  config {
    id     = komodo_action.example.id
    action = "run"
  }
}
