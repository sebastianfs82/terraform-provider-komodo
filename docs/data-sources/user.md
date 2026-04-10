---
page_title: "komodo_user Data Source - komodo"
subcategory: ""
description: |-
  Reads an existing Komodo user by username or id.
---

# komodo_user (Data Source)

Reads an existing Komodo user by username or id.

## Example Usage

```terraform
# Look up by username
data "komodo_user" "alice" {
  username = "alice"
}

# Look up by ObjectId
data "komodo_user" "by_id" {
  id = "6627c3e4f1a2b3c4d5e6f7a8"
}
```

## Schema

### Optional

- `id` (String) The user identifier (ObjectId). If set, takes precedence over `username`.
- `username` (String) The globally unique username.

### Read-Only

- `enabled` (Boolean) Whether the user is enabled and able to access the API.
- `admin` (Boolean) Whether the user has global admin permissions.
- `create_servers` (Boolean) Whether the user can create servers.
- `create_builds` (Boolean) Whether the user can create builds.
