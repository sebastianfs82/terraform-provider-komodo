// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccActionResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccActionResourceConfig("tf-acc-action-basic", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_action.test", "name", "tf-acc-action-basic"),
					resource.TestCheckResourceAttrSet("komodo_action.test", "id"),
				),
			},
		},
	})
}

func TestAccActionResource_update(t *testing.T) {
	const name = "tf-acc-action-update"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccActionResourceConfig(name, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_action.test", "name", name),
				),
			},
			{
				Config: testAccActionResourceConfigWithFileContents(name, "console.log('hello');"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_action.test", "file_contents", "console.log('hello');"),
				),
			},
		},
	})
}

func TestAccActionResource_importState(t *testing.T) {
	var actionID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccActionResourceConfig("tf-acc-action-import", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_action.test", "id"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["komodo_action.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						actionID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config:            testAccActionResourceConfig("tf-acc-action-import", ""),
				ResourceName:      "komodo_action.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return actionID, nil
				},
				ImportStateVerifyIgnore: []string{"webhook"},
			},
		},
	})
}

func TestAccActionResource_disappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccActionResourceConfig("tf-acc-action-disappears", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_action.test", "id"),
					testAccActionDisappears("komodo_action.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccActionResource_rename(t *testing.T) {
	var savedID string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccActionResourceConfig("tf-acc-action-rename-orig", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_action.test", "name", "tf-acc-action-rename-orig"),
					resource.TestCheckResourceAttrSet("komodo_action.test", "id"),
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["komodo_action.test"]
						savedID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config: testAccActionResourceConfig("tf-acc-action-rename-new", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_action.test", "name", "tf-acc-action-rename-new"),
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["komodo_action.test"]
						if rs.Primary.ID != savedID {
							return fmt.Errorf("resource was recreated: ID changed from %q to %q", savedID, rs.Primary.ID)
						}
						return nil
					},
				),
			},
		},
	})
}

func testAccActionDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found in state: %s", resourceName)
		}
		c := client.NewClient(
			os.Getenv("KOMODO_ENDPOINT"),
			os.Getenv("KOMODO_USERNAME"),
			os.Getenv("KOMODO_PASSWORD"),
		)
		return c.DeleteAction(context.Background(), rs.Primary.ID)
	}
}

func testAccActionResourceConfig(name, _ string) string {
	return fmt.Sprintf(`
resource "komodo_action" "test" {
  name = %q
}
`, name)
}

func testAccActionResourceConfigWithFileContents(name, fileContents string) string {
	return fmt.Sprintf(`
resource "komodo_action" "test" {
  name          = %q
  file_contents = %q
}
`, name, fileContents)
}

func TestAccActionResource_tags(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccActionWithTagConfig("tf-acc-action-tags"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_action.test", "tags.#", "1"),
					resource.TestCheckResourceAttrPair("komodo_action.test", "tags.0", "komodo_tag.test", "id"),
				),
			},
			{
				Config: testAccActionClearTagsConfig("tf-acc-action-tags"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_action.test", "tags.#", "0"),
				),
			},
		},
	})
}

func testAccActionWithTagConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_tag" "test" {
  name  = "tf-acc-tag-action"
  color = "Green"
}

resource "komodo_action" "test" {
  name = %q
  tags = [komodo_tag.test.id]
}
`, name)
}

func testAccActionClearTagsConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_action" "test" {
  name = %q
  tags = []
}
`, name)
}

func TestAccActionResource_schedule(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccActionResourceConfigWithFullSchedule("tf-acc-action-schedule"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_action.test", "schedule.format", "Cron"),
					resource.TestCheckResourceAttr("komodo_action.test", "schedule.expression", "0 * * * *"),
					resource.TestCheckResourceAttr("komodo_action.test", "schedule.enabled", "true"),
					resource.TestCheckResourceAttr("komodo_action.test", "schedule.alert_enabled", "true"),
					resource.TestCheckResourceAttr("komodo_action.test", "schedule.timezone", "Europe/Berlin"),
				),
			},
		},
	})
}

func TestAccActionResource_scheduleDefaults(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccActionResourceConfigWithMinimalSchedule("tf-acc-action-schedule-defaults"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_action.test", "schedule.format", "Cron"),
					resource.TestCheckResourceAttr("komodo_action.test", "schedule.expression", "0 * * * *"),
					// enabled and alert_enabled default to true; timezone defaults to ""
					resource.TestCheckResourceAttr("komodo_action.test", "schedule.enabled", "true"),
					resource.TestCheckResourceAttr("komodo_action.test", "schedule.alert_enabled", "true"),
					resource.TestCheckResourceAttr("komodo_action.test", "schedule.timezone", ""),
				),
			},
		},
	})
}

func TestAccActionResource_scheduleUpdate(t *testing.T) {
	const name = "tf-acc-action-schedule-update"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccActionResourceConfigWithFullSchedule(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_action.test", "schedule.expression", "0 * * * *"),
					resource.TestCheckResourceAttr("komodo_action.test", "schedule.enabled", "true"),
					resource.TestCheckResourceAttr("komodo_action.test", "schedule.alert_enabled", "true"),
					resource.TestCheckResourceAttr("komodo_action.test", "schedule.timezone", "Europe/Berlin"),
				),
			},
			{
				// Update expression only; omit enabled/alert_enabled/timezone → defaults applied
				Config: testAccActionResourceConfigWithMinimalSchedule(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_action.test", "schedule.expression", "0 * * * *"),
					resource.TestCheckResourceAttr("komodo_action.test", "schedule.enabled", "true"),
					resource.TestCheckResourceAttr("komodo_action.test", "schedule.alert_enabled", "true"),
					resource.TestCheckResourceAttr("komodo_action.test", "schedule.timezone", ""),
				),
			},
		},
	})
}

func TestAccActionResource_runOnStartupEnabled(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccActionResourceConfigWithRunOnStartup("tf-acc-action-run-startup"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_action.test", "run_on_startup_enabled", "true"),
				),
			},
		},
	})
}

func TestAccActionResource_reloadDependenciesEnabled(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccActionResourceConfigWithReloadDeps("tf-acc-action-reload-deps"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_action.test", "reload_dependencies_enabled", "true"),
				),
			},
		},
	})
}

func testAccActionResourceConfigWithFullSchedule(name string) string {
	return fmt.Sprintf(`
resource "komodo_action" "test" {
  name = %q
  schedule {
    format        = "Cron"
    expression    = "0 * * * *"
    enabled       = true
    alert_enabled = true
    timezone      = "Europe/Berlin"
  }
}
`, name)
}

func testAccActionResourceConfigWithMinimalSchedule(name string) string {
	return fmt.Sprintf(`
resource "komodo_action" "test" {
  name = %q
  schedule {
    format     = "Cron"
    expression = "0 * * * *"
  }
}
`, name)
}

func testAccActionResourceConfigWithRunOnStartup(name string) string {
	return fmt.Sprintf(`
resource "komodo_action" "test" {
  name                   = %q
  run_on_startup_enabled = true
}
`, name)
}

func testAccActionResourceConfigWithReloadDeps(name string) string {
	return fmt.Sprintf(`
resource "komodo_action" "test" {
  name                        = %q
  reload_dependencies_enabled = true
}
`, name)
}

func TestAccActionResource_arguments(t *testing.T) {
	const name = "tf-acc-action-arguments"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccActionResourceConfigWithArguments(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_action.test", "argument.#", "2"),
					resource.TestCheckResourceAttr("komodo_action.test", "argument.0.name", "MY_TEST_VAR"),
					resource.TestCheckResourceAttr("komodo_action.test", "argument.0.value", "Hello World"),
					resource.TestCheckResourceAttr("komodo_action.test", "argument.1.name", "MY_TEST_VAR2"),
					resource.TestCheckResourceAttr("komodo_action.test", "argument.1.value", "Hello World2"),
				),
			},
			{
				// Update: change a value and remove one argument
				Config: testAccActionResourceConfigWithSingleArgument(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_action.test", "argument.#", "1"),
					resource.TestCheckResourceAttr("komodo_action.test", "argument.0.name", "MY_TEST_VAR"),
					resource.TestCheckResourceAttr("komodo_action.test", "argument.0.value", "Updated Value"),
				),
			},
			{
				// Remove all arguments
				Config: testAccActionResourceConfig(name, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_action.test", "argument.#", "0"),
				),
			},
		},
	})
}

func testAccActionResourceConfigWithArguments(name string) string {
	return fmt.Sprintf(`
resource "komodo_action" "test" {
  name = %q

  argument {
    name  = "MY_TEST_VAR"
    value = "Hello World"
  }

  argument {
    name  = "MY_TEST_VAR2"
    value = "Hello World2"
  }
}
`, name)
}

func testAccActionResourceConfigWithSingleArgument(name string) string {
	return fmt.Sprintf(`
resource "komodo_action" "test" {
  name = %q

  argument {
    name  = "MY_TEST_VAR"
    value = "Updated Value"
  }
}
`, name)
}
