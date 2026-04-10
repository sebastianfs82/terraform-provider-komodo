---
page_title: "komodo_user Resource - komodo"
subcategory: ""
description: |-
  Manages a Komodo local user.
---

# komodo_user (Resource)

Manages a Komodo local user. Local users authenticate with a username and password.

~> **Note** Changing `username` or `password` forces recreation of the resource.

## Example Usage

```terraform
resource "komodo_user" "alice" {
  username = "alice"
  password = "securePassword1!"
}

# User with elevated permissions
resource "komodo_user" "ci_user" {
  username       = "ci-runner"
  password       = "securePassword1!"
  create_servers = true
  create_builds  = true
}
```

## Schema

### Required

- `username` (String) The globally unique username. Changing this value forces recreation of the resource.
- `password` (String, Sensitive) The password for the local user. Changing this value forces recreation of the resource.

### Optional

- `enabled` (Boolean) Whether the user is enabled and able to access the API. Defaults to `true`.
- `admin` (Boolean) Whether the user has global admin permissions. Cannot be combined with `create_servers` or `create_builds`.
- `create_servers` (Boolean) Whether the user can create servers. Cannot be set when `admin` is `true`. Defaults to `false`.
- `create_builds` (Boolean) Whether the user can create builds. Cannot be set when `admin` is `true`. Defaults to `false`.

### Read-Only

- `id` (String) The user identifier (ObjectId).

## Import

Import is supported using the following syntax:

```shell
# Import by username or ObjectId
terraform import komodo_user.example alice
```
