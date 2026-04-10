---
page_title: "komodo_provider_account Resource - komodo"
subcategory: ""
description: |-
  Manages a Komodo git provider account.
---

# komodo_provider_account (Resource)

Manages a Komodo git provider account. A git provider account stores authentication credentials for a specific git host (e.g. GitHub, GitLab).

~> **Note** The combination of `domain` and `username` must be unique in Komodo. Attempting to create a second account with the same domain and username will fail.

## Example Usage

```terraform
resource "komodo_provider_account" "github" {
  domain   = "github.com"
  https    = true
  username = "myuser"
  token    = var.github_token
}

resource "komodo_provider_account" "gitlab_self_hosted" {
  domain   = "git.internal.example.com"
  https    = true
  username = "ci-bot"
  token    = var.gitlab_token
}
```

## Schema

### Required

- `domain` (String) The git provider domain without a protocol prefix (e.g. `github.com`, `gitlab.com`).
- `username` (String) The account username.
- `token` (String, Sensitive) The plaintext access token for the account.

### Optional

- `https` (Boolean) Whether to use HTTPS (`true`) or HTTP (`false`) when cloning. Defaults to `true` when not specified.

### Read-Only

- `id` (String) The git provider account identifier (ObjectId).

## Import

Import is supported using the following syntax:

```shell
# Import by ObjectId
terraform import komodo_provider_account.example 6627c3e4f1a2b3c4d5e6f7a8
```

~> **Note** The `token` attribute is not returned by the read API after creation. After importing, you must set `token` in your configuration to avoid a permanent diff; add it to `lifecycle.ignore_changes` if the token is managed outside Terraform.
