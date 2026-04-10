---
page_title: "komodo_run_procedure Action - terraform-provider-komodo"
subcategory: ""
description: |-
  Runs the target procedure.
---

# komodo_run_procedure (Action)

Runs the target procedure.

## Example Usage

```shell
terraform apply -invoke action.komodo_run_procedure.example
```

```terraform
resource "komodo_procedure" "my_procedure" {
  name = "my-procedure"
  # ...
}

action "komodo_run_procedure" "example" {
  config {
    procedure = komodo_procedure.my_procedure.name
  }
}
```

## Schema

### Config

#### Required

- `procedure` (String) Id or name of the procedure to run.
