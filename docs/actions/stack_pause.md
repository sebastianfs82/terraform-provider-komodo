---
page_title: "komodo_stack_pause Action - terraform-provider-komodo"
subcategory: ""
description: |-
  Triggers a docker compose pause on the target Komodo stack.
---

# komodo_stack_pause (Action)

Triggers a `docker compose pause` on the target Komodo stack.

## Example Usage

Actions are not executed automatically — they must be triggered either by an
`action_trigger` block in a resource lifecycle, or invoked explicitly on the CLI:

```shell
terraform apply -invoke action.komodo_stack_pause.example
```

### Manually pause a stack

```terraform
action "komodo_stack_pause" "example" {
  config {
    stack = "my-app"
  }
}
```

Run `terraform apply -invoke action.komodo_stack_pause.example` to pause the stack on demand.

### Pause only specific services

```terraform
action "komodo_stack_pause" "partial" {
  config {
    stack    = "my-app"
    services = ["worker", "scheduler"]
  }
}
```

## Schema

### Config

#### Required

- `stack` (String) Id or name of the stack to pause.

#### Optional

- `services` (List of String) Filter to only pause specific services. If empty, pauses all services.
