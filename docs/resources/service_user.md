---
page_title: "komodo_service_user Resource - komodo"
subcategory: ""
description: |-
  Manages a Komodo service user.
---

# komodo_service_user (Resource)

Manages a Komodo service user. Service users have no password and authenticate exclusively via API keys.

~> **Note** Changing `username` forces recreation of the resource.

## Example Usage

```terraform
resource "komodo_service_user" "ci" {
  username    = "ci-service"
  description = "CI/CD pipeline service account"
}

# Service user with API key
resource "komodo_service_user" "deployer" {
  username       = "deployer"
  create_servers = true
  create_builds  = true
}

resource "komodo_api_key" "deployer_key" {
  name            = "deployer-key"
  service_user_id = komodo_service_user.deployer.id
}
```

## Schema

### Required

- `username` (String) The globally unique username for the service user. Changing this value forces recreation of the resource.

### Optional

- `description` (String) A description for the service user.
- `enabled` (Boolean) Whether the service user is enabled and able to access the API. Defaults to `true`.
- `admin` (Boolean) Whether the service user has global admin permissions.
- `create_servers` (Boolean) Whether the service user can create servers. Defaults to `false`.
- `create_builds` (Boolean) Whether the service user can create builds. Defaults to `false`.

### Read-Only

- `id` (String) The service user identifier (ObjectId).

## Import

Import is supported using the following syntax:

```shell
# Import by username or ObjectId
terraform import komodo_service_user.example ci-service
```
