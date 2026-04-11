// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// ─── StartDeployment ─────────────────────────────────────────────────────────

func TestAccStartDeploymentAction_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStartDeploymentActionConfig("tf-test-start-deploy"),
			},
		},
	})
}

func testAccStartDeploymentActionConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_deployment" "test" {
  name = %q
}

action "komodo_start_deployment" "test" {
  config {
    id = komodo_deployment.test.name
  }
}
`, name)
}

// ─── PullDeployment ──────────────────────────────────────────────────────────

func TestAccPullDeploymentAction_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPullDeploymentActionConfig("tf-test-pull-deploy"),
			},
		},
	})
}

func testAccPullDeploymentActionConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_deployment" "test" {
  name = %q
}

action "komodo_pull_deployment" "test" {
  config {
    id = komodo_deployment.test.name
  }
}
`, name)
}

// ─── RunAction ───────────────────────────────────────────────────────────────

func TestAccRunActionAction_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRunActionActionConfig("tf-test-run-action"),
			},
		},
	})
}

func testAccRunActionActionConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_action" "test" {
  name = %q
}

action "komodo_run_action" "test" {
  config {
    id = komodo_action.test.name
  }
}
`, name)
}

// ─── RunBuild ────────────────────────────────────────────────────────────────

func TestAccRunBuildAction_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRunBuildActionConfig("tf-test-run-build"),
			},
		},
	})
}

func testAccRunBuildActionConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_build" "test" {
  name = %q
}

action "komodo_run_build" "test" {
  config {
    id = komodo_build.test.name
  }
}
`, name)
}

// ─── RunProcedure ────────────────────────────────────────────────────────────

func TestAccRunProcedureAction_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRunProcedureActionConfig("tf-test-run-procedure"),
			},
		},
	})
}

func testAccRunProcedureActionConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_procedure" "test" {
  name = %q
}

action "komodo_run_procedure" "test" {
  config {
    id = komodo_procedure.test.name
  }
}
`, name)
}
