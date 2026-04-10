resource "komodo_alerter" "slack" {
  name          = "slack-alerts"
  endpoint_type = "Slack"
  enabled       = true

  slack_endpoint = {
    url = var.slack_webhook_url
  }
}

resource "komodo_alerter" "custom" {
  name          = "custom-webhook"
  endpoint_type = "Custom"

  custom_endpoint = {
    url = "https://my-webhook.example.com/alert"
  }
}
