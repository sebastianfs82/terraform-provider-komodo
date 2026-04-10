resource "komodo_api_key" "example" {
  name    = "my-api-key"
  expires = 0
}

resource "komodo_api_key" "expiring" {
  name    = "expiring-key"
  expires = 1735689600000
}

# API key for a service user
resource "komodo_api_key" "svc" {
  name            = "ci-key"
  service_user_id = komodo_service_user.example.id
}
