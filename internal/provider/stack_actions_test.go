// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// ─── DeployStack ─────────────────────────────────────────────────────────────

func TestAccStackDeployAction_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStackDeployActionConfig_basic("tf-test-deploy"),
			},
		},
	})
}

func TestAccStackDeployAction_withServices(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStackDeployActionConfig_withServices("tf-test-deploy-svc"),
			},
		},
	})
}

func TestAccStackDeployAction_withStopTime(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStackDeployActionConfig_withStopTime("tf-test-deploy-timeout"),
			},
		},
	})
}

// ─── StartStack ──────────────────────────────────────────────────────────────

func TestAccStackStartAction_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStackStartActionConfig_basic("tf-test-start"),
			},
		},
	})
}

func TestAccStackStartAction_withServices(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStackStartActionConfig_withServices("tf-test-start-svc"),
			},
		},
	})
}

// ─── StopStack ───────────────────────────────────────────────────────────────

func TestAccStackStopAction_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStackStopActionConfig_basic("tf-test-stop"),
			},
		},
	})
}

func TestAccStackStopAction_withStopTime(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStackStopActionConfig_withStopTime("tf-test-stop-timeout"),
			},
		},
	})
}

// ─── PauseStack ──────────────────────────────────────────────────────────────

func TestAccStackPauseAction_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStackPauseActionConfig_basic("tf-test-pause"),
			},
		},
	})
}

func TestAccStackPauseAction_withServices(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStackPauseActionConfig_withServices("tf-test-pause-svc"),
			},
		},
	})
}

// ─── DestroyStack ─────────────────────────────────────────────────────────────

func TestAccStackDestroyAction_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStackDestroyActionConfig_basic("tf-test-destroy"),
			},
		},
	})
}

func TestAccStackDestroyAction_withOptions(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStackDestroyActionConfig_withOptions("tf-test-destroy-opts"),
			},
		},
	})
}

// ─── Config helpers ──────────────────────────────────────────────────────────

func testAccStackWithComposeConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_stack" "test" {
  name = "%s"

  files = {
    contents = <<-EOT
      services:
        web:
          image: nginx:latest
    EOT
  }
}
`, name)
}

func testAccStackDeployActionConfig_basic(name string) string {
	return testAccStackWithComposeConfig(name) + `
action "komodo_stack_deploy" "test" {
  config {
    stack = komodo_stack.test.name
  }
}
`
}

func testAccStackDeployActionConfig_withServices(name string) string {
	return testAccStackWithComposeConfig(name) + `
action "komodo_stack_deploy" "test" {
  config {
    stack    = komodo_stack.test.name
    services = ["web"]
  }
}
`
}

func testAccStackDeployActionConfig_withStopTime(name string) string {
	return testAccStackWithComposeConfig(name) + `
action "komodo_stack_deploy" "test" {
  config {
    stack     = komodo_stack.test.name
    stop_time = 30
  }
}
`
}

func testAccStackStartActionConfig_basic(name string) string {
	return testAccStackWithComposeConfig(name) + `
action "komodo_stack_start" "test" {
  config {
    stack = komodo_stack.test.name
  }
}
`
}

func testAccStackStartActionConfig_withServices(name string) string {
	return testAccStackWithComposeConfig(name) + `
action "komodo_stack_start" "test" {
  config {
    stack    = komodo_stack.test.name
    services = ["web"]
  }
}
`
}

func testAccStackStopActionConfig_basic(name string) string {
	return testAccStackWithComposeConfig(name) + `
action "komodo_stack_stop" "test" {
  config {
    stack = komodo_stack.test.name
  }
}
`
}

func testAccStackStopActionConfig_withStopTime(name string) string {
	return testAccStackWithComposeConfig(name) + `
action "komodo_stack_stop" "test" {
  config {
    stack     = komodo_stack.test.name
    stop_time = 30
  }
}
`
}

func testAccStackPauseActionConfig_basic(name string) string {
	return testAccStackWithComposeConfig(name) + `
action "komodo_stack_pause" "test" {
  config {
    stack = komodo_stack.test.name
  }
}
`
}

func testAccStackPauseActionConfig_withServices(name string) string {
	return testAccStackWithComposeConfig(name) + `
action "komodo_stack_pause" "test" {
  config {
    stack    = komodo_stack.test.name
    services = ["web"]
  }
}
`
}

func testAccStackDestroyActionConfig_basic(name string) string {
	return testAccStackWithComposeConfig(name) + `
action "komodo_stack_destroy" "test" {
  config {
    stack = komodo_stack.test.name
  }
}
`
}

func testAccStackDestroyActionConfig_withOptions(name string) string {
	return testAccStackWithComposeConfig(name) + `
action "komodo_stack_destroy" "test" {
  config {
    stack          = komodo_stack.test.name
    remove_orphans = true
    stop_time      = 30
  }
}
`
}
