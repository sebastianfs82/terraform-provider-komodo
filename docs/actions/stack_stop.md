---
page_title: "komodo_stack_stop Action - terraform-provider-komodo"
subcategory: ""
description: |-
  Triggers a docker compose stop on the target Komodo stack.
---

# komodo_stack_stop (Action)

Triggers a `docker compose stop` on the target Komodo stack.

## Example Usage

Actions are not executed automatically — they must be triggered either by an
`action_trigger` block in a resource lifecycle, or invoked explicitly on the CLI:

```shell
terraform apply -invoke action.komodo_stack_stop.example
```

### Trigger to stop a stack on demand

```terraform
resource "komodo_stack" "my_app" {
  name = "my-app"
  # ...
}

action "komodo_stack_stop" "example" {
  config {
    stack = komodo_stack.my_app.name
  }
}

# Change `input` to a new value (e.g. a timestamp) and re-apply to re-stop.
resource "terraform_data" "stop_trigger" {
  input = komodo_stack.my_app.id

  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.komodo_stack_stop.example]
    }
  }
}
```

### Stop specific services with a custom termination timeout

```terraform
action "komodo_stack_stop" "graceful" {
  config {
    stack     = "my-app"
    services  = ["web"]
    stop_time = 30
  }
}
```

## Schema

### Config

#### Required

- `stack` (String) Id or name of the stack to stop.

#### Optional

- `services` (List of String) Filter to only stop specific services. If empty, stops all services.
- `stop_time` (Number) Override the default termination max time in seconds.
