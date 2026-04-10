---
page_title: "komodo_registry_account Resource - komodo"
subcategory: ""
description: |-
  Manages a Komodo docker registry account.
---

# komodo_registry_account (Resource)

Manages a Komodo docker registry account. Docker registry accounts store credentials used to pull images from private container registries.

## Example Usage

```terraform
# Docker Hub account
resource "komodo_registry_account" "dockerhub" {
  domain   = "docker.io"
  username = "myuser"
  token    = "my-dockerhub-token"
}

# GitHub Container Registry
resource "komodo_registry_account" "ghcr" {
  domain   = "ghcr.io"
  username = "myuser"
  token    = "ghp_mytoken"
}

# Self-hosted registry
resource "komodo_registry_account" "private" {
  domain   = "registry.example.com"
  username = "admin"
  token    = "registry-password"
}
```

## Schema

### Required

- `username` (String) The account username.
- `token` (String, Sensitive) The plaintext access token (password) for the account.

### Optional

- `domain` (String) The registry domain (e.g. `registry.example.com`). Leave empty or omit for Docker Hub (`docker.io`).

### Read-Only

- `id` (String) The docker registry account identifier (ObjectId).

## Import

Docker registry accounts can be imported using the account ObjectId:

```shell
terraform import komodo_registry_account.example <id>
```
