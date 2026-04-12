# Preferred: authenticate with an API key.
# Generate one via the Komodo UI or the komodo_api_key resource,
# then supply the values through environment variables or a secret manager.
provider "komodo" {
  endpoint   = "https://komodo.example.com"
  api_key    = var.komodo_api_key    # or KOMODO_API_KEY env var
  api_secret = var.komodo_api_secret # or KOMODO_API_SECRET env var
}

# Alternative: username / password (suitable for local development only).
# provider "komodo" {
#   endpoint = "https://komodo.example.com"
#   username = var.komodo_username   # or KOMODO_USERNAME env var
#   password = var.komodo_password   # or KOMODO_PASSWORD env var
# }

