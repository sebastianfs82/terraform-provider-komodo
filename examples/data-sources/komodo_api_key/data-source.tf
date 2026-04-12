data "komodo_api_key" "by_key" {
  key = "K-EDXr0hotlGVGIM67mPSXpPpkvp3j6e92fxybUsNJ"
}

data "komodo_api_key" "by_name" {
  name = "my-api-key"
}

# Look up an API key belonging to a service user
data "komodo_api_key" "svc_key" {
  name            = "ci-key"
  service_user_id = komodo_service_user.example.id
}
