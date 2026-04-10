---
page_title: "komodo_users Data Source - terraform-provider-komodo"
subcategory: ""
description: |-
  Lists all Komodo users visible to the authenticated admin.
---

# komodo_users (Data Source)

Lists all Komodo users visible to the authenticated admin.

## Example Usage

```terraform
data "komodo_users" "all" {}

output "admin_usernames" {
  value = [for u in data.komodo_users.all.users : u.username if u.admin]
}
```

## Schema

### Read-Only

- `users` (List of Object) The list of users. (see [below for nested schema](#nestedatt--users))

<a id="nestedatt--users"></a>
### Nested Schema for `users`

Read-Only:

- `id` (String) The user identifier (ObjectId).
- `username` (String) The globally unique username.
- `enabled` (Boolean) Whether the user is enabled and able to access the API.
- `admin` (Boolean) Whether the user has global admin permissions.
- `create_servers` (Boolean) Whether the user can create servers.
- `create_builds` (Boolean) Whether the user can create builds.
