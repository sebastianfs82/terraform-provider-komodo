resource "komodo_procedure" "example" {
  name = "my-procedure"

  lifecycle {
    action_trigger {
      events = [
        after_create,
        after_update,
      ]
      actions = [
        action.komodo_procedure.run
      ]
    }
  }
}

action "komodo_procedure" "run" {
  config {
    id     = komodo_procedure.example.id
    action = "run"
  }
}
