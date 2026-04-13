## 0.8.0 (April 13, 2026)

BREAKING CHANGES:

* **Configuration block syntax change across multiple resources:** The following optional single-nested configuration blocks have been converted from `SingleNestedAttribute` (attribute-assignment style: `block = { ... }`) to `SingleNestedBlock` (HCL block style: `block { ... }`). All existing configurations using the `=` assignment form must be updated to use bare block syntax:
  * `komodo_action`: `schedule`, `webhook`
  * `komodo_alerter`: `endpoint`
  * `komodo_builder`: `url_config`, `server_config`, `aws_config`
  * `komodo_procedure`: `schedule`, `webhook`
  * `komodo_repo`: `source`, `webhook`, `on_clone`, `on_pull`, `environment`
  * `komodo_resource_sync`: `webhook`
  * `komodo_server`: `alerts`
  * `komodo_stack`: `source`, `compose`, `environment`, `build`, `webhook`, `pre_deploy`, `post_deploy`
* **`komodo_build` resource / data source:** Complete schema overhaul:
  * The flat `version` block (`major`, `minor`, `patch` as integer attributes) and the top-level `auto_increment_version` attribute have been replaced by a `version` block with two sub-attributes: `value` (string, e.g. `"1.0.0"`) and `auto_increment_enabled` (bool).
  * Flat source attributes (`git_provider`, `git_https`, `git_account`, `linked_repo`, `repo`, `branch`, `commit`) have been replaced by a `source` block.
  * Flat image attributes (`image_name`, `image_tag`, `include_latest_tag`, `include_version_tags`, `include_commit_tag`) have been replaced by an `image` block.
  * Flat build attributes (`build_path`, `build_args`, `secret_args`, `extra_args`) have been replaced by a `build` block.
  * `image_registry[].account` has been renamed to `account_id`; `image_registry[].domain` has been removed.
* **`komodo_deployment` resource / data source:** `image.version` is no longer a nested block with `major`, `minor`, and `patch` integer attributes. It is now a plain string attribute formatted as `"MAJOR.MINOR.PATCH"` (e.g. `version = "1.0.0"`). Omit or set to `""` for latest.

FEATURES:

* **`komodo_stack` resource:** New attributes: `auto_update_enabled`, `auto_update_scope`, `poll_updates_enabled`, `alerts_enabled`, `extra_arguments`, `compose_cmd_wrapper`, `compose_cmd_wrapper_include`, `ignore_services`, `links`, and `destroy_enforced`.
* **`komodo_repo` resource:** New `links` attribute for quick links in the Komodo UI.
* **`komodo_build` resource:** New `links` attribute for quick links in the Komodo UI.
* **`komodo_deployment` resource:** New `links` attribute for quick links in the Komodo UI.

BUG FIXES:

* **Git account resolution:** `ResolveGitAccountID` now tries an exact `domain + username` match first and falls back to `username`-only matching, fixing resolution failures for custom-hosted git providers.
* **Docker registry account resolution:** New `ResolveDockerRegistryAccountID` client method uses the same two-step matching logic, replacing the previous best-effort approach.
* **`komodo_builder` data source / resource:** `ListBuilders` now calls the `ListFullBuilders` API endpoint to retrieve complete builder configuration data.
* **`komodo_build`, `komodo_deployment` resources:** Removed incorrect text-based "not found" detection from `GetBuild` and `GetDeployment` API response parsing that could suppress real errors.
* **`komodo_server` resource:** The `alerts` block is now only populated when it was already present in the plan, preventing spurious diffs for servers configured without an explicit `alerts` block.
* **`komodo_stack`, `komodo_repo` resources:** `webhook.secret` is now correctly marked as `Sensitive`, preventing the value from appearing in plain text in plan output and logs.
* **`komodo_build` acceptance tests:** Simplified `testAccBuildResourceWithSourceConfig` by removing the redundant `path` parameter (now hardcoded inside the helper), aligning call sites with the updated schema.

---

## 0.7.0 (April 13, 2026)

BREAKING CHANGES:

* **`komodo_procedure` resource / data source:**
  * The `stages` JSON string attribute has been replaced by a `stage` nested list block. Each `stage` block has a `name` attribute and one or more `execution` nested blocks with `type`, `parameters` (map), and `enabled` attributes.
  * `failure_alert` has been renamed to `failure_alert_enabled` for consistency with the naming convention used across other resources.
* **`komodo_builds` data source:** The `repo_id` and `builder_id` filter attributes have been removed. The data source now always returns all builds.
* **`komodo_deployments` data source:** The `server_id` filter attribute has been removed. The data source now always returns all deployments.
* **`komodo_networks` data source:** The previously required `server_id` attribute has been removed. The data source now lists Docker networks across all Komodo-managed servers instead of a single server.
* **`komodo_repos` data source:** The `server_id` and `builder_id` filter attributes have been removed. The data source now always returns all repositories.
* **`komodo_resource_syncs` data source:** The `repo_id` filter attribute has been removed. The data source now always returns all resource syncs.
* **`komodo_stacks` data source:** The `server_id` and `repo_id` filter attributes have been removed. The data source now always returns all stacks.

FEATURES:

* **`komodo_terminal` resource:** Manages a Komodo terminal session attached to a target resource (server, deployment, or stack). Supports `target_type`, `target_id`, `container`, `service`, `mode`, and `command` attributes.
* **`komodo_terminal` data source:** Reads an existing Komodo terminal session by name.
* **`komodo_terminals` data source:** Lists all Komodo terminal sessions visible to the authenticated user.

---

## 0.6.0 (April 12, 2026)

BREAKING CHANGES:

* **`komodo_action` resource / data source:** Several attributes have been restructured:
  * Flat schedule attributes (`schedule_format`, `schedule`, `schedule_enabled`, `schedule_timezone`, `schedule_alert`) have been replaced by a nested `schedule` block with sub-attributes `format`, `expression`, `enabled`, `timezone`, and `alert_enabled`.
  * Flat webhook attributes (`webhook_enabled`, `webhook_secret`) have been replaced by a nested `webhook` block with `enabled` and `secret`.
  * `failure_alert` → `failure_alert_enabled`
  * `run_at_startup` → `run_on_startup_enabled`
  * `reload_deno_deps` → `reload_dependencies_enabled`
  * The `arguments` string attribute and `arguments_format` have been replaced by an `argument` list block with `name` and `value` sub-attributes.
* **`komodo_procedure` resource / data source:** Flat schedule and webhook attributes have been replaced by nested `schedule` and `webhook` blocks (same structure as `komodo_action`).
* **`komodo_build` resource / data source:** Flat webhook attributes (`webhook_enabled`, `webhook_secret`) have been replaced by a nested `webhook` block with `enabled` and `secret`.
* **`komodo_resource_sync` resource / data source:** Flat webhook attributes (`webhook_enabled`, `webhook_secret`) have been replaced by a nested `webhook` block with `enabled` and `secret`.
* **`komodo_builder` resource / data source:** The `builder_type` attribute has been renamed to `type`.
* **`komodo_user` resource:** `create_servers` → `create_server_enabled` and `create_builds` → `create_build_enabled`, for consistency with the naming convention introduced for `komodo_service_user` in 0.4.0.

FEATURES:

* **Tags support across all major resources:** A `tags` attribute (list of tag IDs) has been added to `komodo_action`, `komodo_alerter`, `komodo_build`, `komodo_builder`, `komodo_deployment`, `komodo_procedure`, `komodo_repo`, `komodo_resource_sync`, `komodo_server`, and `komodo_stack`. Use `komodo_tag.<name>.id` to reference tags.

---

## 0.5.0 (April 12, 2026)

FEATURES:

* **In-place rename for most resources:** The following resources no longer force replacement when `name` is changed. Instead, Terraform calls the Komodo `Rename*` API and updates in place: `komodo_action`, `komodo_alerter`, `komodo_build`, `komodo_builder`, `komodo_deployment`, `komodo_procedure`, `komodo_repo`, `komodo_resource_sync`, `komodo_server`, `komodo_stack`.

BREAKING CHANGES:

* **`komodo_alerter` resource / data source:** The endpoint configuration has been redesigned for simplicity:
  * The `endpoint_type` attribute and the five separate endpoint blocks (`custom_endpoint`, `slack_endpoint`, `discord_endpoint`, `ntfy_endpoint`, `pushover_endpoint`) have been replaced by a single `endpoint` block with `type` (required), `url` (required), and `email` (optional) attributes.
  * The `alert_types` attribute has been renamed to `types`.
* **`komodo_tag` resource:** Changing the tag `name` no longer forces replacement. Rename is now handled by updating the tag in place via the `UpdateTag` API.

ENHANCEMENTS:

* **`komodo_alerter` resource:** Added `resource` list block to filter alerts to specific resources (include or exclude individual resources by type and id).
* **`komodo_alerter` resource:** Added `maintenance` list block to configure scheduled maintenance windows during which alerts from the alerter are suppressed. Supports `Daily`, `Weekly`, and `OneTime` schedule types.
* **`komodo_alerter` resource:** Added `ValidateConfig` to enforce that at most one `endpoint` block is present.

DOCUMENTATION:

* **README:** Expanded provider documentation with additional configuration examples and usage guidance.

BUG FIXES:

* **`komodo_user_group` acceptance tests:** Removed unused `testAccUserGroupHasMember` helper.
* **`komodo_repo` acceptance tests:** Simplified `testAccRepoResourceConfig_withConfig` by removing the redundant `domain` parameter (now hardcoded inside the helper), aligning tests with the `0.2.0` schema change that dropped `source.url`.

---

## 0.4.0 (April 12, 2026)

FEATURES:

* **`komodo_api_key` data source:** Reads an existing Komodo API key by key ID or name. Supports looking up keys belonging to service users via `service_user_id`.

BREAKING CHANGES:

* **`komodo_api_key` resource / data source:** The `expires` attribute has been renamed to `expires_at` for consistency with other timestamp fields across the provider. Update all configuration references accordingly.
* **`komodo_api_key` resource / data source:** The `expires_at` attribute (formerly `expires`) now accepts an RFC3339 string (e.g. `"2030-01-01T00:00:00Z"`) instead of a Unix millisecond integer. Use `""` (empty string) for no expiration.
* **`komodo_onboarding_key` resource / data source:** The `expires` and `created_at` attributes now return RFC3339 strings instead of Unix millisecond integers.
* **`komodo_user_group` resource / data sources:** The `updated_at` attribute now returns an RFC3339 string instead of a Unix millisecond integer.

ENHANCEMENTS:

* **`komodo_api_key` resource:** Plan-time validation now rejects `expires_at` values that are not valid RFC3339 timestamps, or that represent a date already in the past.
* **`komodo_user_group_membership` resource:** Creating a membership now fails at apply time with a clear error if the target user group has `everyone_enabled = true`.
* **`komodo_user_group` resource:** Config validation blocks setting `everyone_enabled = true` together with a non-empty `users` list.
* **`komodo_service_user` resource:** The `admin`, `create_servers`, and `create_builds` attributes have been renamed to `admin_enabled`, `create_server_enabled`, and `create_build_enabled` respectively, matching the naming convention used by `komodo_user`. Both `enabled` (default `true`) and `admin_enabled` (default `false`) now have static defaults, eliminating "(known after apply)" noise on first plan.
* **`komodo_service_user` resource:** Config validation blocks setting `admin_enabled = true` together with `create_server_enabled = true` or `create_build_enabled = true`.

---

## 0.3.0 (April 11, 2026)

FEATURES:

* **`komodo_version` data source:** Reads the running Komodo Core API server version string.
* **`komodo_stack_deploy_if_changed` action:** Deploys a stack only when its compose definition has changed since the last deployment.
* **`komodo_stack_run_service` action:** Runs a one-off service in a stack via `docker compose run`.

BREAKING CHANGES:

* **All action resources:** The resource-specific identifier field (e.g. `deployment`, `stack`, `build`, `repo`, `server`, `procedure`, `action`, `alerter`, `sync`) has been renamed to `id` for consistency. Update all action resource configurations accordingly.
* **`komodo_server` resource / data source:** Several fields have been renamed or restructured:
  * `insecure_tls` → `certificate_verification_enabled` (semantics inverted — `true` means TLS verification is **on**)
  * `auto_rotate_keys` → `auto_rotate_keys_enabled`
  * `auto_prune` → `auto_prune_images_enabled`
  * `stats_monitoring` → `historical_system_statistics_enabled`
  * `ignore_mounts` → `ignored_disk_mounts`
  * The individual alert flag attributes (`send_unreachable_alerts`, `send_cpu_alerts`, `send_mem_alerts`, `send_disk_alerts`, `send_version_mismatch_alerts`) and the separate threshold attributes (`cpu_warning`, `cpu_critical`, `mem_warning`, `mem_critical`, `disk_warning`, `disk_critical`) have been replaced by a single nested `alerts` block with `enabled`, `types` (set), and a `thresholds` sub-block.

ENHANCEMENTS:

* **`komodo_server` resource:** Added `public_key` (Computed) attribute that exposes the server's public key.
* **`komodo_server` resource:** Added `maintenance` list block for configuring scheduled maintenance windows (name, schedule type, day/date, hour, minute, duration, timezone, enabled).
* **`komodo_server` resource:** Config validation now rejects plans where `alerts.enabled = true` but `alerts.types` is empty, providing a clear error message.
* **`komodo_onboarding_key` resource:** Added a plan-time version guard — plans fail with a descriptive error when the connected Komodo Core server is older than v2.0.0.
* **`komodo_user` resource:** Improved `admin` attribute description to document that promoting a user to admin requires the provider to be authenticated as a superuser (root/init admin).

---

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
