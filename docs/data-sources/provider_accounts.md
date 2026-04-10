---
page_title: "komodo_provider_accounts Data Source - terraform-provider-komodo"
subcategory: ""
description: |-
  Lists all Komodo git provider accounts visible to the authenticated user.
---

# komodo_provider_accounts (Data Source)

Lists all Komodo git provider accounts visible to the authenticated user.

## Example Usage

```terraform
data "komodo_provider_accounts" "all" {}

output "github_accounts" {
  value = [for a in data.komodo_provider_accounts.all.provider_accounts : a.username if a.domain == "github.com"]
}
```

## Schema

### Read-Only

- `provider_accounts` (List of Object) The list of git provider accounts. (see [below for nested schema](#nestedatt--provider_accounts))

<a id="nestedatt--provider_accounts"></a>
### Nested Schema for `provider_accounts`

Read-Only:

- `id` (String) The git provider account identifier (ObjectId).
- `domain` (String) The git provider domain (e.g. `github.com`).
- `username` (String) The git account username.
- `https` (Boolean) Whether HTTPS is used for this git provider.
