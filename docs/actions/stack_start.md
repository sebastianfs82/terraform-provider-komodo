---
page_title: "komodo_stack_start Action - terraform-provider-komodo"
subcategory: ""
description: |-
  Triggers a docker compose start on the target Komodo stack.
---

# komodo_stack_start (Action)

Triggers a `docker compose start` on the target Komodo stack.

## Example Usage

Actions are not executed automatically — they must be triggered either by an
`action_trigger` block in a resource lifecycle, or invoked explicitly on the CLI:

```shell
terraform apply -invoke action.komodo_stack_start.example
```

### Trigger after stack resource is created

```terraform
resource "komodo_stack" "my_app" {
  name = "my-app"
  # ...
}

action "komodo_stack_start" "example" {
  config {
    stack = komodo_stack.my_app.name
  }
}

resource "terraform_data" "start_trigger" {
  input = komodo_stack.my_app.id

  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.komodo_stack_start.example]
    }
  }
}
```

### Start only specific services

```terraform
action "komodo_stack_start" "partial" {
  config {
    stack    = "my-app"
    services = ["worker"]
  }
}
```

## Schema

### Config

#### Required

- `stack` (String) Id or name of the stack to start.

#### Optional

- `services` (List of String) Filter to only start specific services. If empty, starts all services.
