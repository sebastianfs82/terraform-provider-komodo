---
page_title: "komodo_provider_account Data Source - komodo"
subcategory: ""
description: |-
  Reads an existing Komodo git provider account by id.
---

# komodo_provider_account (Data Source)

Reads an existing Komodo git provider account by id.

## Example Usage

```terraform
data "komodo_provider_account" "github" {
  id = "6627c3e4f1a2b3c4d5e6f7a8"
}

output "git_domain" {
  value = data.komodo_provider_account.github.domain
}
```

## Schema

### Required

- `id` (String) The git provider account identifier (ObjectId).

### Read-Only

- `domain` (String) The git provider domain without a protocol prefix (e.g. `github.com`).
- `https` (Boolean) Whether HTTPS is used for cloning.
- `username` (String) The account username.
- `token` (String, Sensitive) The plaintext access token for the account.
