data "komodo_onboarding_key" "by_name" {
  name = "new-server-key"
}

data "komodo_onboarding_key" "by_public_key" {
  public_key = "ssh-ed25519 AAAA..."
}
