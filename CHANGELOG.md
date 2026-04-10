## 0.2.0 (April 10, 2026)

FEATURES:

* **`komodo_builds` data source:** Added `builder_id` and `repo_id` filter arguments to limit results to builds using a specific builder or sourced from a specific repo.
* **`komodo_deployments` data source:** Added `server_id` filter argument to limit results to deployments running on a specific server.
* **`komodo_repos` data source:** Added `builder_id` and `server_id` filter arguments to limit results to repos built or cloned on specific resources.
* **`komodo_resource_syncs` data source:** Added `repo_id` filter argument to limit results to resource syncs sourced from a specific repo.
* **`komodo_stacks` data source:** Added `repo_id` and `server_id` filter arguments to limit results to stacks sourced from or running on specific resources.

BREAKING CHANGES:

* **`komodo_repo` resource / data source:** The `source.url` attribute has been removed. It is replaced by the separate `source.domain` (e.g. `github.com`) and `source.https_enabled` attributes. When `source.account_id` is set, `domain` and `https_enabled` are derived automatically and must not be specified.
* **`komodo_stack` resource / data source:** The `compose.files` block has been renamed to `compose` and its `paths` attribute has been renamed to `file_paths`. Update all configurations and state references accordingly.

ENHANCEMENTS:

* **`komodo_repo` resource:** `source.branch` now defaults to `main` when not specified.
* **Documentation:** All resource, data source, and action pages are now grouped by domain (Stacks, Repos, Builds, Deployments, Procedures, Actions, Servers, Networks, Alerters, Resource Syncs, Users & Access, Configuration) on the Terraform Registry sidebar via the `subcategory` frontmatter field.

BUG FIXES:

* **`komodo_repo` resource:** Fixed "Provider produced inconsistent result" error when `source.account_id` is set — `domain` and `https_enabled` are now correctly stored as null in state when derived from the account.
* **`komodo_repo` resource:** Fixed `source.https_enabled` incorrectly showing as "(known after apply)" due to a stray `Computed: true` in the schema.

---

## 0.1.1 (April 10, 2026)

BUG FIXES:

* **`komodo_stack` resource:** Fixed trailing newline inconsistency in `pre_deploy` and `post_deploy` fields causing unnecessary diffs on plan.

---

## 0.1.0 (April 10, 2026)

FEATURES:

* **New Resource:** `komodo_action`
* **New Resource:** `komodo_alerter`
* **New Resource:** `komodo_api_key`
* **New Resource:** `komodo_build`
* **New Resource:** `komodo_builder`
* **New Resource:** `komodo_deployment`
* **New Resource:** `komodo_network`
* **New Resource:** `komodo_onboarding_key`
* **New Resource:** `komodo_procedure`
* **New Resource:** `komodo_provider_account`
* **New Resource:** `komodo_registry_account`
* **New Resource:** `komodo_repo`
* **New Resource:** `komodo_resource_sync`
* **New Resource:** `komodo_server`
* **New Resource:** `komodo_service_user`
* **New Resource:** `komodo_stack`
* **New Resource:** `komodo_tag`
* **New Resource:** `komodo_user`
* **New Resource:** `komodo_user_group`
* **New Resource:** `komodo_user_group_membership`
* **New Resource:** `komodo_variable`
* **New Data Source:** `komodo_action`
* **New Data Source:** `komodo_actions`
* **New Data Source:** `komodo_alerter`
* **New Data Source:** `komodo_alerters`
* **New Data Source:** `komodo_build`
* **New Data Source:** `komodo_builds`
* **New Data Source:** `komodo_builder`
* **New Data Source:** `komodo_builders`
* **New Data Source:** `komodo_deployment`
* **New Data Source:** `komodo_deployments`
* **New Data Source:** `komodo_network`
* **New Data Source:** `komodo_networks`
* **New Data Source:** `komodo_onboarding_key`
* **New Data Source:** `komodo_procedure`
* **New Data Source:** `komodo_procedures`
* **New Data Source:** `komodo_provider_account`
* **New Data Source:** `komodo_provider_accounts`
* **New Data Source:** `komodo_registry_account`
* **New Data Source:** `komodo_registry_accounts`
* **New Data Source:** `komodo_repo`
* **New Data Source:** `komodo_repos`
* **New Data Source:** `komodo_resource_sync`
* **New Data Source:** `komodo_resource_syncs`
* **New Data Source:** `komodo_server`
* **New Data Source:** `komodo_servers`
* **New Data Source:** `komodo_service_user`
* **New Data Source:** `komodo_service_users`
* **New Data Source:** `komodo_stack`
* **New Data Source:** `komodo_stacks`
* **New Data Source:** `komodo_tag`
* **New Data Source:** `komodo_tags`
* **New Data Source:** `komodo_user`
* **New Data Source:** `komodo_users`
* **New Data Source:** `komodo_user_group`
* **New Data Source:** `komodo_user_groups`
* **New Data Source:** `komodo_variable`
* **New Data Source:** `komodo_variables`
