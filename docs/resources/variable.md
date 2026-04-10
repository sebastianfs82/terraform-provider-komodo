---
page_title: "komodo_variable Resource - komodo"
subcategory: ""
description: |-
  Manages a Komodo variable.
---

# komodo_variable (Resource)

Manages a Komodo variable.

## Example Usage

```terraform
resource "komodo_variable" "example" {
  name        = "MY_VAR"
  value       = "my-value"
  description = "A test variable"
  is_secret   = false
}

# Secret variable
resource "komodo_variable" "secret" {
  name      = "DB_PASSWORD"
  value     = "supersecret"
  is_secret = true
}
```

## Schema

### Required

- `name` (String) The variable name. Changing this value forces recreation of the resource.

### Optional

- `value` (String) The variable value.
- `description` (String) A description for the variable.
- `is_secret` (Boolean) Whether the variable is treated as a secret.

### Read-Only

- `id` (String) The variable identifier.

## Import

Import is supported using the following syntax:

```shell
terraform import komodo_variable.example MY_VAR
```
