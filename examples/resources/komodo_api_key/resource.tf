resource "komodo_api_key" "example" {
  name = "my-api-key"
}

resource "komodo_api_key" "expiring" {
  name       = "expiring-key"
  expires_at = "2030-01-01T00:00:00Z"
}

# API key for a service user
resource "komodo_api_key" "svc" {
  name            = "ci-key"
  service_user_id = komodo_service_user.example.id
}
