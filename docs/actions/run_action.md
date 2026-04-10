---
page_title: "komodo_run_action Action - terraform-provider-komodo"
subcategory: ""
description: |-
  Runs the target action.
---

# komodo_run_action (Action)

Runs the target action.

## Example Usage

```shell
terraform apply -invoke action.komodo_run_action.example
```

```terraform
resource "komodo_action" "my_action" {
  name = "my-action"
  # ...
}

action "komodo_run_action" "example" {
  config {
    action = komodo_action.my_action.name
  }
}
```

## Schema

### Config

#### Required

- `action` (String) Id or name of the action to run.
