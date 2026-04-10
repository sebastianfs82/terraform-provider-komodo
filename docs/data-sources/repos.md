---
page_title: "komodo_repos Data Source - komodo"
subcategory: ""
description: |-
  Lists all Komodo git repositories visible to the authenticated user.
---

# komodo_repos (Data Source)

Lists all Komodo git repositories visible to the authenticated user.

## Example Usage

```terraform
data "komodo_repos" "all" {}

output "repo_names" {
  value = [for r in data.komodo_repos.all.repositories : r.name]
}
```

## Attributes Reference

- `repositories` - List of git repositories. Each item has the following attributes:
  - `id` - The git repository identifier (ObjectId).
  - `name` - The name of the git repository.
  - `server_id` - The ID of the server the repo is cloned on.
  - `builder_id` - The ID of the attached builder.
  - `path` - The folder on the server the repo is cloned into.
  - `links` - Quick links associated with this repository.
  - `source` - Git source configuration. See [source](#nestedatt--repositories--source) below.
  - `webhook` - Webhook configuration (null when webhooks are not configured). See [webhook](#nestedatt--repositories--webhook) below.
  - `on_clone` - The command run after the repository is cloned. See [on_clone / on_pull](#nestedatt--repositories--on_clone) below.
  - `on_pull` - The command run after the repository is pulled. See [on_clone / on_pull](#nestedatt--repositories--on_pull) below.
  - `environment` - Environment variable configuration (null when not configured). See [environment](#nestedatt--repositories--environment) below.

<a id="nestedatt--repositories--source"></a>
### Nested Schema for `repositories.source`

- `url` - The URL of the git provider, e.g. `https://github.com`.
- `account_id` - The git account used for private repositories.
- `path` - The repository path, e.g. `owner/repo`.
- `branch` - The branch checked out.
- `commit` - The specific commit hash checked out.

<a id="nestedatt--repositories--webhook"></a>
### Nested Schema for `repositories.webhook`

- `enabled` - Whether webhooks trigger an action on this repository.
- `secret` - The alternate webhook secret.

<a id="nestedatt--repositories--on_clone"></a>
<a id="nestedatt--repositories--on_pull"></a>
### Nested Schema for `repositories.on_clone` / `repositories.on_pull`

- `path` - The working directory for the command.
- `command` - The shell command to run.

<a id="nestedatt--repositories--environment"></a>
### Nested Schema for `repositories.environment`

- `file_path` - Path to the environment file.
- `variables` - Map of environment variables injected. Keys are uppercased.
