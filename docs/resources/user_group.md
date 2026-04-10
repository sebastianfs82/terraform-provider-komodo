---
page_title: "komodo_user_group Resource - komodo"
subcategory: ""
description: |-
  Manages a Komodo user group.
---

# komodo_user_group (Resource)

Manages a Komodo user group.

## Example Usage

```terraform
resource "komodo_user_group" "developers" {
  name  = "Developers"
  users = [komodo_user.alice.id, komodo_user.bob.id]
}

# Group that includes every user automatically
resource "komodo_user_group" "everyone" {
  name     = "AllUsers"
  everyone = true
}
```

## Schema

### Required

- `name` (String) The user group name.

### Optional

- `everyone` (Boolean) When `true`, the group automatically includes every user. Mutually exclusive with `users`.
- `users` (List of String) List of user IDs to include in the group. Mutually exclusive with `everyone`.
- `all` (Map of String) Additional key/value metadata for the group.

### Read-Only

- `id` (String) The user group identifier (ObjectId).
- `updated_at` (Number) Last update timestamp (milliseconds since epoch).

## Import

Import is supported using the following syntax:

```shell
terraform import komodo_user_group.example Developers
```
