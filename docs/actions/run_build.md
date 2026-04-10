---
page_title: "komodo_run_build Action - terraform-provider-komodo"
subcategory: ""
description: |-
  Runs the target build.
---

# komodo_run_build (Action)

Runs the target build.

## Example Usage

```shell
terraform apply -invoke action.komodo_run_build.example
```

```terraform
action "komodo_run_build" "example" {
  config {
    build = "my-build"
  }
}
```

## Schema

### Config

#### Required

- `build` (String) Id or name of the build to run.
