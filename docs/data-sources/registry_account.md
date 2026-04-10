---
page_title: "komodo_registry_account Data Source - komodo"
subcategory: ""
description: |-
  Reads an existing Komodo docker registry account.
---

# komodo_registry_account (Data Source)

Reads an existing Komodo docker registry account by its ObjectId.

## Example Usage

```terraform
data "komodo_registry_account" "example" {
  id = "507f1f77bcf86cd799439011"
}
```

## Schema

### Required

- `id` (String) The docker registry account identifier (ObjectId).

### Read-Only

- `domain` (String) The registry domain (e.g. `registry.example.com`). Empty string indicates Docker Hub (`docker.io`).
- `username` (String) The account username.
- `token` (String, Sensitive) The plaintext access token (password) for the account.
