---
page_title: "komodo_user_group Data Source - komodo"
subcategory: ""
description: |-
  Reads an existing Komodo user group by name.
---

# komodo_user_group (Data Source)

Reads an existing Komodo user group by name.

## Example Usage

```terraform
data "komodo_user_group" "developers" {
  name = "Developers"
}

output "group_members" {
  value = data.komodo_user_group.developers.users
}
```

## Schema

### Required

- `name` (String) The user group name.

### Read-Only

- `id` (String) The user group identifier (ObjectId).
- `everyone` (Boolean) Whether the group automatically includes every user.
- `users` (List of String) List of user IDs in the group.
- `all` (Map of String) Additional key/value metadata for the group.
- `updated_at` (Number) Last update timestamp (milliseconds since epoch).
