resource "komodo_alerter" "example" {
  name = "slack-alerts"

  endpoint {
    type = "Slack"
    url  = var.slack_webhook_url
  }

  lifecycle {
    action_trigger {
      events = [
        after_create,
        after_update,
      ]
      actions = [
        action.komodo_alerter.test
      ]
    }
  }
}

action "komodo_alerter" "test" {
  config {
    id     = komodo_alerter.example.id
    action = "test"
  }
}
