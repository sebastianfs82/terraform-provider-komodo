---
page_title: "komodo_registry_accounts Data Source - terraform-provider-komodo"
subcategory: ""
description: |-
  Lists all Komodo Docker registry accounts visible to the authenticated user.
---

# komodo_registry_accounts (Data Source)

Lists all Komodo Docker registry accounts visible to the authenticated user.

## Example Usage

```terraform
data "komodo_registry_accounts" "all" {}

output "registry_domains" {
  value = [for a in data.komodo_registry_accounts.all.registry_accounts : a.domain]
}
```

## Schema

### Read-Only

- `registry_accounts` (List of Object) The list of Docker registry accounts. (see [below for nested schema](#nestedatt--registry_accounts))

<a id="nestedatt--registry_accounts"></a>
### Nested Schema for `registry_accounts`

Read-Only:

- `id` (String) The registry account identifier (ObjectId).
- `domain` (String) The registry domain (e.g. `docker.io`). Empty string indicates Docker Hub.
- `username` (String) The registry account username.
