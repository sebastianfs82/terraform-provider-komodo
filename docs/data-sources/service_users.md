---
page_title: "komodo_service_users Data Source - terraform-provider-komodo"
subcategory: ""
description: |-
  Lists all Komodo service users visible to the authenticated admin.
---

# komodo_service_users (Data Source)

Lists all Komodo service users visible to the authenticated admin.

## Example Usage

```terraform
data "komodo_service_users" "all" {}

output "service_user_names" {
  value = [for u in data.komodo_service_users.all.service_users : u.username]
}
```

## Schema

### Read-Only

- `service_users` (List of Object) The list of service users. (see [below for nested schema](#nestedatt--service_users))

<a id="nestedatt--service_users"></a>
### Nested Schema for `service_users`

Read-Only:

- `id` (String) The service user identifier (ObjectId).
- `username` (String) The globally unique username of the service user.
- `enabled` (Boolean) Whether the service user is enabled and able to access the API.
- `admin` (Boolean) Whether the service user has global admin permissions.
- `create_servers` (Boolean) Whether the service user can create servers.
- `create_builds` (Boolean) Whether the service user can create builds.
