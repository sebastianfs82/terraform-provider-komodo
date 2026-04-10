resource "komodo_service_user" "example" {
  username    = "ci-service"
  description = "CI/CD pipeline service account"
}

resource "komodo_api_key" "example" {
  name            = "ci-key"
  service_user_id = komodo_service_user.example.id
}
