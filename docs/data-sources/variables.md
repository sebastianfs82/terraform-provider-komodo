---
page_title: "komodo_variables Data Source - terraform-provider-komodo"
subcategory: ""
description: |-
  Lists all Komodo variables visible to the authenticated user.
---

# komodo_variables (Data Source)

Lists all Komodo variables visible to the authenticated user.

~> **Note** Secret variable values are returned as an empty string by the API.

## Example Usage

```terraform
data "komodo_variables" "all" {}

output "variable_names" {
  value = [for v in data.komodo_variables.all.variables : v.name]
}
```

## Schema

### Read-Only

- `variables` (List of Object) The list of variables. (see [below for nested schema](#nestedatt--variables))

<a id="nestedatt--variables"></a>
### Nested Schema for `variables`

Read-Only:

- `name` (String) The variable name.
- `value` (String) The variable value. Empty string for secret variables.
- `description` (String) An optional description of the variable.
- `is_secret` (Boolean) Whether the variable is treated as a secret.
