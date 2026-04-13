# Terraform Provider for Komodo

A Terraform provider for managing [Komodo](https://komo.do/) resources.
Komodo is a self-hosted tool for building and deploying Docker containers at scale.

## Table of Contents

- [Requirements](#requirements)
- [Authentication](#authentication)
- [Quick Start](#quick-start)
- [End-to-End Example](#end-to-end-example)
- [Resources](#resources)
- [Data Sources](#data-sources)
- [Actions](#actions)
- [Developing the Provider](#developing-the-provider)

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- A running [Komodo](https://komo.do/) instance **v2.0.0 or above**

## Authentication

The provider supports two mutually exclusive authentication methods.
Configure one via the provider block or the corresponding environment variables.

### API key (preferred)

API keys are scoped to a single Komodo user and do not expose the account password.
Generate one through the Komodo UI or with the `komodo_api_key` resource,
then store the key and secret as secrets in your CI system.

```hcl
provider "komodo" {
  endpoint   = "https://komodo.example.com"
  api_key    = var.komodo_api_key
  api_secret = var.komodo_api_secret
}
```

| Environment variable  | Description      |
|-----------------------|------------------|
| `KOMODO_ENDPOINT`     | API endpoint URL |
| `KOMODO_API_KEY`      | API key value    |
| `KOMODO_API_SECRET`   | API key secret   |

### Username / password

Fine for interactive local development, but avoid committing plain-text passwords to version control.

```hcl
provider "komodo" {
  endpoint = "https://komodo.example.com"
  username = var.komodo_username
  password = var.komodo_password
}
```

| Environment variable | Description      |
|----------------------|------------------|
| `KOMODO_ENDPOINT`    | API endpoint URL |
| `KOMODO_USERNAME`    | Komodo username  |
| `KOMODO_PASSWORD`    | Komodo password  |

## Quick Start

**1. Add the provider to your configuration:**

```hcl
terraform {
  required_providers {
    komodo = {
      source  = "sebastianfs82/komodo"
      version = "~> 0.7"
    }
  }
}

provider "komodo" {
  # All values read from environment variables
}
```

**2. Export credentials:**

```bash
export KOMODO_ENDPOINT="https://komodo.example.com"
export KOMODO_API_KEY="K-your-api-key"
export KOMODO_API_SECRET="your-api-secret"
```

**3. Run Terraform:**

```bash
terraform init
terraform plan
terraform apply
```

## End-to-End Example

This example covers a realistic production setup:

1. Authenticate with an API key
2. Register GitHub credentials as a provider account
3. Clone a private Git repository into Komodo
4. Deploy a Docker Compose stack whose `compose.yaml` lives in that repository
5. Tag the stack for grouping and filtering in the Komodo UI

```hcl
terraform {
  required_providers {
    komodo = {
      source  = "sebastianfs82/komodo"
      version = "~> 0.7"
    }
  }
}

# ─── Provider ────────────────────────────────────────────────────────────────
# Preferred: authenticate with an API key so no account password is exposed.
# Pass values via environment variables (KOMODO_API_KEY / KOMODO_API_SECRET)
# or Terraform variables backed by a secret manager.
provider "komodo" {
  endpoint   = "https://komodo.example.com"
  api_key    = var.komodo_api_key
  api_secret = var.komodo_api_secret
}

# Uncomment to use username / password instead:
#
# provider "komodo" {
#   endpoint = "https://komodo.example.com"
#   username = var.komodo_username
#   password = var.komodo_password
# }

variable "komodo_api_key"    { sensitive = true }
variable "komodo_api_secret" { sensitive = true }
variable "github_token"      { sensitive = true }
variable "server_id"         {}

# ─── Git credentials ─────────────────────────────────────────────────────────
# komodo_provider_account registers a GitHub personal access token (PAT) with
# Komodo so it can authenticate against github.com for every clone and pull.
resource "komodo_provider_account" "github" {
  domain        = "github.com"
  https_enabled = true
  username      = "myorg"
  token         = var.github_token
}

# ─── Tag ─────────────────────────────────────────────────────────────────────
# Tags let you group and filter resources in the Komodo UI.
resource "komodo_tag" "app" {
  name  = "my-app"
  color = "#4f46e5"
}

# ─── Repository ──────────────────────────────────────────────────────────────
# Registering the repo gives Komodo a named handle for cloning, pulling, and
# triggering builds or deploys via webhooks.
resource "komodo_repo" "app" {
  name      = "my-app"
  server_id = var.server_id
  links     = [komodo_tag.app.id]

  source = {
    path       = "myorg/my-app"   # <owner>/<repo> on GitHub
    branch     = "main"
    account_id = komodo_provider_account.github.id
  }
}

# ─── Stack ───────────────────────────────────────────────────────────────────
# The stack sources its compose file from the repository registered above.
# Using repo_id delegates authentication and clone details to the komodo_repo
# resource — no need to repeat provider account credentials here.
resource "komodo_stack" "app" {
  name      = "my-app"
  server_id = var.server_id
  links     = [komodo_tag.app.id]

  source = {
    repo_id = komodo_repo.app.id
    path    = "docker-compose.yml"
    branch  = "main"
  }

  environment = {
    variables = {
      APP_ENV  = "production"
      APP_PORT = "8080"
    }
  }

  auto_pull_enabled = true
}
```

> See the [examples/](./examples/) directory for individual resource and data source examples.

## Resources

| Resource | Description |
|----------|-------------|
| [`komodo_action`](docs/resources/action.md) | Custom executable action |
| [`komodo_alerter`](docs/resources/alerter.md) | Alert channel (Slack, webhook, etc.) |
| [`komodo_api_key`](docs/resources/api_key.md) | User or service-user API key |
| [`komodo_build`](docs/resources/build.md) | Docker image build definition |
| [`komodo_builder`](docs/resources/builder.md) | Build agent / runner |
| [`komodo_deployment`](docs/resources/deployment.md) | Single-container deployment |
| [`komodo_network`](docs/resources/network.md) | Docker network on a server |
| [`komodo_onboarding_key`](docs/resources/onboarding_key.md) | One-time server onboarding key |
| [`komodo_procedure`](docs/resources/procedure.md) | Ordered sequence of actions |
| [`komodo_provider_account`](docs/resources/provider_account.md) | Git provider credentials (GitHub, GitLab, …) |
| [`komodo_registry_account`](docs/resources/registry_account.md) | Docker registry credentials |
| [`komodo_repo`](docs/resources/repo.md) | Git repository registered in Komodo |
| [`komodo_resource_sync`](docs/resources/resource_sync.md) | Syncs Komodo resources from a git repo |
| [`komodo_service_user`](docs/resources/service_user.md) | Non-human service account |
| [`komodo_stack`](docs/resources/stack.md) | Docker Compose stack |
| [`komodo_tag`](docs/resources/tag.md) | Label for grouping resources |
| [`komodo_terminal`](docs/resources/terminal.md) | Terminal session on a target resource |
| [`komodo_user`](docs/resources/user.md) | Human user account |
| [`komodo_user_group`](docs/resources/user_group.md) | User group for permission management |
| [`komodo_user_group_membership`](docs/resources/user_group.md) | User ↔ group membership |
| [`komodo_variable`](docs/resources/variable.md) | Global variable / secret |

## Data Sources

All singular data sources look up a single resource by `id` or `name`.
Plural data sources return a filtered list.

### Singular

| Data Source | Description |
|-------------|-------------|
| [`komodo_action`](docs/data-sources/action.md) | Look up an action |
| [`komodo_alerter`](docs/data-sources/alerter.md) | Look up an alerter |
| [`komodo_build`](docs/data-sources/build.md) | Look up a build |
| [`komodo_builder`](docs/data-sources/builder.md) | Look up a builder |
| [`komodo_deployment`](docs/data-sources/deployment.md) | Look up a deployment |
| [`komodo_network`](docs/data-sources/network.md) | Look up a Docker network |
| [`komodo_onboarding_key`](docs/data-sources/onboarding_key.md) | Look up an onboarding key |
| [`komodo_procedure`](docs/data-sources/procedure.md) | Look up a procedure |
| [`komodo_provider_account`](docs/data-sources/provider_account.md) | Look up a git provider account |
| [`komodo_registry_account`](docs/data-sources/registry_account.md) | Look up a registry account |
| [`komodo_repo`](docs/data-sources/repo.md) | Look up a repository |
| [`komodo_resource_sync`](docs/data-sources/resource_sync.md) | Look up a resource sync |
| [`komodo_server`](docs/data-sources/server.md) | Look up a server |
| [`komodo_service_user`](docs/data-sources/service_user.md) | Look up a service user |
| [`komodo_stack`](docs/data-sources/stack.md) | Look up a stack |
| [`komodo_tag`](docs/data-sources/tag.md) | Look up a tag |
| [`komodo_terminal`](docs/data-sources/terminal.md) | Look up a terminal session |
| [`komodo_user`](docs/data-sources/user.md) | Look up a user |
| [`komodo_user_group`](docs/data-sources/user_group.md) | Look up a user group |
| [`komodo_variable`](docs/data-sources/variable.md) | Look up a variable |

### Plural

| Data Source | Description |
|-------------|-------------|
| [`komodo_actions`](docs/data-sources/actions.md) | List / filter actions |
| [`komodo_alerters`](docs/data-sources/alerters.md) | List / filter alerters |
| [`komodo_builds`](docs/data-sources/builds.md) | List / filter builds |
| [`komodo_builders`](docs/data-sources/builders.md) | List / filter builders |
| [`komodo_deployments`](docs/data-sources/deployments.md) | List / filter deployments |
| [`komodo_networks`](docs/data-sources/networks.md) | List / filter networks |
| [`komodo_procedures`](docs/data-sources/procedures.md) | List / filter procedures |
| [`komodo_provider_accounts`](docs/data-sources/provider_accounts.md) | List / filter provider accounts |
| [`komodo_registry_accounts`](docs/data-sources/registry_accounts.md) | List / filter registry accounts |
| [`komodo_repos`](docs/data-sources/repos.md) | List / filter repositories |
| [`komodo_resource_syncs`](docs/data-sources/resource_syncs.md) | List / filter resource syncs |
| [`komodo_servers`](docs/data-sources/servers.md) | List / filter servers |
| [`komodo_service_users`](docs/data-sources/service_users.md) | List / filter service users |
| [`komodo_stacks`](docs/data-sources/stacks.md) | List / filter stacks |
| [`komodo_tags`](docs/data-sources/tags.md) | List / filter tags |
| [`komodo_terminals`](docs/data-sources/terminals.md) | List terminal sessions |
| [`komodo_users`](docs/data-sources/users.md) | List / filter users |
| [`komodo_user_groups`](docs/data-sources/user_groups.md) | List / filter user groups |
| [`komodo_variables`](docs/data-sources/variables.md) | List / filter variables |

## Actions

Terraform actions let you trigger imperative Komodo operations as part of a plan/apply cycle.

| Action | Description |
|--------|-------------|
| `komodo_repo_clone` | Clone a registered repository |
| `komodo_repo_pull` | Pull latest changes for a repository |
| `komodo_repo_build` | Trigger a repository build |
| `komodo_stack_deploy` | Deploy a stack |
| `komodo_stack_destroy` | Destroy a running stack |
| `komodo_stack_start` | Start a stopped stack |
| `komodo_stack_stop` | Stop a running stack |
| `komodo_stack_pause` | Pause a running stack |

## Developing the Provider

**Requirements:** [Go](https://golang.org/doc/install) >= 1.24

```bash
# Clone
git clone https://github.com/sebastianfs82/terraform-provider-komodo
cd terraform-provider-komodo

# Build and install to $GOPATH/bin
go install .

# Run unit tests
go test ./internal/...

# Run acceptance tests against a live Komodo instance
export TF_ACC=1
export KOMODO_ENDPOINT="http://localhost:9120/"
export KOMODO_USERNAME="admin"
export KOMODO_PASSWORD="admin"
go test -v -timeout 120m ./internal/provider/...

# Regenerate documentation
make generate
```

### Local Dev Override

To use the locally built binary with Terraform, add a dev override to `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "sebastianfs82/komodo" = "/path/to/your/GOPATH/bin"
  }
  direct {}
}
```

## License

[Mozilla Public License 2.0](./LICENSE)
