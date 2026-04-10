---
page_title: "komodo_variable Data Source - komodo"
subcategory: ""
description: |-
  Reads an existing Komodo variable.
---

# komodo_variable (Data Source)

Reads an existing Komodo variable.

## Example Usage

```terraform
data "komodo_variable" "example" {
  name = "MY_VAR"
}

output "var_value" {
  value     = data.komodo_variable.example.value
  sensitive = data.komodo_variable.example.is_secret
}
```

## Schema

### Required

- `name` (String) The variable name.

### Read-Only

- `value` (String) The variable value.
- `description` (String) The variable description.
- `is_secret` (Boolean) Whether the variable is treated as a secret.
