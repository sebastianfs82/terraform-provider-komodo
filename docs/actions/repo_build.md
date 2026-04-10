---
page_title: "komodo_repo_build Action - terraform-provider-komodo"
subcategory: ""
description: |-
  Builds the target repo using its attached builder.
---

# komodo_repo_build (Action)

Builds the target repo using its attached builder.

The repo must have a builder attached (`builder_id`). For AWS-type builders, Komodo
will spawn the builder instance, clone the repo, run any `on_clone`/`on_pull` hooks,
and then execute the build.

## Example Usage

Actions are not executed automatically — they must be triggered either by an
`action_trigger` block in a resource lifecycle, or invoked explicitly on the CLI:

```shell
terraform apply -invoke action.komodo_repo_build.example
```

### Trigger after repo resource changes

```terraform
resource "komodo_repo" "my_repo" {
  name = "my-repo"
  # ...
}

action "komodo_repo_build" "example" {
  config {
    repo = komodo_repo.my_repo.name
  }
}

# Automatically build whenever the repo resource is created or updated.
resource "terraform_data" "build_trigger" {
  input = komodo_repo.my_repo.id

  lifecycle {
    action_trigger {
      events  = [after_create, after_update]
      actions = [action.komodo_repo_build.example]
    }
  }
}
```

## Schema

### Config

#### Required

- `repo` (String) Id or name of the repo to build.
