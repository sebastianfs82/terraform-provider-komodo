resource "komodo_alerter" "slack" {
  name    = "slack-alerts"
  enabled = true

  endpoint {
    type = "Slack"
    url  = var.slack_webhook_url
  }
}

resource "komodo_alerter" "discord" {
  name = "discord-alerts"

  endpoint {
    type = "Discord"
    url  = "https://discord.com/api/webhooks/000000000000000000/xxxxxxxxxxxx"
  }
}

resource "komodo_alerter" "custom" {
  name = "custom-webhook"

  endpoint {
    type = "Custom"
    url  = "https://my-webhook.example.com/alert"
  }
}
