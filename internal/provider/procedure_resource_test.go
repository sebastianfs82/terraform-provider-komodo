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

func TestAccProcedureResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProcedureResourceConfig("tf-acc-procedure-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_procedure.test", "name", "tf-acc-procedure-basic"),
					resource.TestCheckResourceAttrSet("komodo_procedure.test", "id"),
				),
			},
		},
	})
}

func TestAccProcedureResource_update(t *testing.T) {
	const name = "tf-acc-procedure-update"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProcedureResourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_procedure.test", "name", name),
				),
			},
			{
				Config: testAccProcedureResourceConfigWithSchedule(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_procedure.test", "schedule.format", "Cron"),
					resource.TestCheckResourceAttr("komodo_procedure.test", "schedule.expression", "0 * * * *"),
					resource.TestCheckResourceAttr("komodo_procedure.test", "schedule.enabled", "true"),
				),
			},
		},
	})
}

func TestAccProcedureResource_importState(t *testing.T) {
	var procedureID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProcedureResourceConfig("tf-acc-procedure-import"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_procedure.test", "id"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["komodo_procedure.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						procedureID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config:            testAccProcedureResourceConfig("tf-acc-procedure-import"),
				ResourceName:      "komodo_procedure.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return procedureID, nil
				},
				ImportStateVerifyIgnore: []string{"webhook"},
			},
		},
	})
}

func TestAccProcedureResource_disappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProcedureResourceConfig("tf-acc-procedure-disappears"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_procedure.test", "id"),
					testAccProcedureDisappears("komodo_procedure.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccProcedureResource_rename(t *testing.T) {
	var savedID string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProcedureResourceConfig("tf-acc-procedure-rename-orig"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_procedure.test", "name", "tf-acc-procedure-rename-orig"),
					resource.TestCheckResourceAttrSet("komodo_procedure.test", "id"),
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["komodo_procedure.test"]
						savedID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config: testAccProcedureResourceConfig("tf-acc-procedure-rename-new"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_procedure.test", "name", "tf-acc-procedure-rename-new"),
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["komodo_procedure.test"]
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

func testAccProcedureDisappears(resourceName string) resource.TestCheckFunc {
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
		return c.DeleteProcedure(context.Background(), rs.Primary.ID)
	}
}

func testAccProcedureResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_procedure" "test" {
  name = %q
}
`, name)
}

func testAccProcedureResourceConfigWithSchedule(name string) string {
	return fmt.Sprintf(`
resource "komodo_procedure" "test" {
  name = %q
  schedule {
    format     = "Cron"
    expression = "0 * * * *"
    enabled    = true
  }
}
`, name)
}

func TestAccProcedureResource_tags(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProcedureWithTagConfig("tf-acc-procedure-tags"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_procedure.test", "tags.#", "1"),
					resource.TestCheckResourceAttrPair("komodo_procedure.test", "tags.0", "komodo_tag.test", "id"),
				),
			},
			{
				Config: testAccProcedureClearTagsConfig("tf-acc-procedure-tags"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_procedure.test", "tags.#", "0"),
				),
			},
		},
	})
}

func testAccProcedureWithTagConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_tag" "test" {
  name  = "tf-acc-tag-procedure"
  color = "Green"
}

resource "komodo_procedure" "test" {
  name = %q
  tags = [komodo_tag.test.id]
}
`, name)
}

func testAccProcedureClearTagsConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_procedure" "test" {
  name = %q
  tags = []
}
`, name)
}

func TestAccProcedureResource_scheduleAlertEnabled(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProcedureResourceConfigWithFullSchedule("tf-acc-procedure-schedule-alert"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_procedure.test", "schedule.format", "Cron"),
					resource.TestCheckResourceAttr("komodo_procedure.test", "schedule.expression", "0 * * * *"),
					resource.TestCheckResourceAttr("komodo_procedure.test", "schedule.alert_enabled", "true"),
					resource.TestCheckResourceAttr("komodo_procedure.test", "schedule.timezone", "Europe/Berlin"),
				),
			},
		},
	})
}

func TestAccProcedureResource_scheduleTimezone(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProcedureResourceConfigWithFullSchedule("tf-acc-procedure-schedule-tz"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_procedure.test", "schedule.timezone", "Europe/Berlin"),
				),
			},
		},
	})
}

func TestAccProcedureResource_scheduleDefaults(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProcedureResourceConfigWithMinimalSchedule("tf-acc-procedure-schedule-defaults"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_procedure.test", "schedule.format", "Cron"),
					resource.TestCheckResourceAttr("komodo_procedure.test", "schedule.expression", "0 * * * *"),
					// enabled and alert_enabled default to true; timezone defaults to ""
					resource.TestCheckResourceAttr("komodo_procedure.test", "schedule.enabled", "true"),
					resource.TestCheckResourceAttr("komodo_procedure.test", "schedule.alert_enabled", "true"),
					resource.TestCheckResourceAttr("komodo_procedure.test", "schedule.timezone", ""),
				),
			},
		},
	})
}

func testAccProcedureResourceConfigWithFullSchedule(name string) string {
	return fmt.Sprintf(`
resource "komodo_procedure" "test" {
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

func testAccProcedureResourceConfigWithMinimalSchedule(name string) string {
	return fmt.Sprintf(`
resource "komodo_procedure" "test" {
  name = %q
  schedule {
    format     = "Cron"
    expression = "0 * * * *"
  }
}
`, name)
}

// ---------------------------------------------------------------------------
// failure_alert_enabled
// ---------------------------------------------------------------------------

func TestAccProcedureResource_failureAlertDefault(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProcedureResourceConfig("tf-acc-procedure-failure-alert-default"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// default is true
					resource.TestCheckResourceAttr("komodo_procedure.test", "failure_alert_enabled", "true"),
				),
			},
		},
	})
}

func TestAccProcedureResource_failureAlertDisabled(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProcedureResourceConfigWithFailureAlert("tf-acc-procedure-failure-alert-off", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_procedure.test", "failure_alert_enabled", "false"),
				),
			},
			{
				Config: testAccProcedureResourceConfigWithFailureAlert("tf-acc-procedure-failure-alert-off", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_procedure.test", "failure_alert_enabled", "true"),
				),
			},
		},
	})
}

func testAccProcedureResourceConfigWithFailureAlert(name string, enabled bool) string {
	return fmt.Sprintf(`
resource "komodo_procedure" "test" {
  name                  = %q
  failure_alert_enabled = %t
}
`, name, enabled)
}

// ---------------------------------------------------------------------------
// stage / execution blocks (native HCL, no jsonencode)
// ---------------------------------------------------------------------------

func TestAccProcedureResource_stages(t *testing.T) {
	const name = "tf-acc-procedure-stages"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with one stage containing one execution
			{
				Config: testAccProcedureResourceConfigWithStage(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_procedure.test", "stage.#", "1"),
					resource.TestCheckResourceAttr("komodo_procedure.test", "stage.0.name", "Deploy"),
					resource.TestCheckResourceAttr("komodo_procedure.test", "stage.0.execution.#", "1"),
					resource.TestCheckResourceAttr("komodo_procedure.test", "stage.0.execution.0.type", "RunProcedure"),
					resource.TestCheckResourceAttr("komodo_procedure.test", "stage.0.execution.0.enabled", "true"),
				),
			},
			// Update to two stages
			{
				Config: testAccProcedureResourceConfigWithTwoStages(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_procedure.test", "stage.#", "2"),
					resource.TestCheckResourceAttr("komodo_procedure.test", "stage.0.name", "Stage1"),
					resource.TestCheckResourceAttr("komodo_procedure.test", "stage.1.name", "Stage2"),
				),
			},
			// Remove all stages
			{
				Config: testAccProcedureResourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_procedure.test", "stage.#", "0"),
				),
			},
		},
	})
}

func TestAccProcedureResource_executionEnabled(t *testing.T) {
	const name = "tf-acc-procedure-exec-enabled"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProcedureResourceConfigWithDisabledExecution(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_procedure.test", "stage.0.execution.0.enabled", "false"),
				),
			},
		},
	})
}

func TestAccProcedureResource_stagesImport(t *testing.T) {
	var procedureID string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProcedureResourceConfigWithStage("tf-acc-procedure-stages-import"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_procedure.test", "id"),
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["komodo_procedure.test"]
						procedureID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config:            testAccProcedureResourceConfigWithStage("tf-acc-procedure-stages-import"),
				ResourceName:      "komodo_procedure.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return procedureID, nil
				},
				ImportStateVerifyIgnore: []string{"webhook"},
			},
		},
	})
}

// ---------------------------------------------------------------------------
// execution parameters (native HCL map, no jsonencode)
// ---------------------------------------------------------------------------

func TestAccProcedureResource_executionParameters(t *testing.T) {
	const name = "tf-acc-procedure-exec-params"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProcedureResourceConfigWithParametersV1(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_procedure.test", "stage.0.execution.0.parameters.%", "1"),
					resource.TestCheckResourceAttrSet("komodo_procedure.test", "stage.0.execution.0.parameters.procedure"),
				),
			},
			// Update parameter value — should be detected as a change (drift detection)
			{
				Config: testAccProcedureResourceConfigWithParametersV2(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_procedure.test", "stage.0.execution.0.parameters.procedure"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// Config helpers
// ---------------------------------------------------------------------------

func testAccProcedureResourceConfigWithStage(name string) string {
	return fmt.Sprintf(`
resource "komodo_procedure" "child" {
  name = "%s-child"
}

resource "komodo_procedure" "test" {
  name = %q

  stage {
    name = "Deploy"

    execution {
      type = "RunProcedure"
      parameters = {
        procedure = komodo_procedure.child.id
      }
    }
  }
}
`, name, name)
}

func testAccProcedureResourceConfigWithTwoStages(name string) string {
	return fmt.Sprintf(`
resource "komodo_procedure" "child" {
  name = "%s-child"
}

resource "komodo_procedure" "test" {
  name = %q

  stage {
    name = "Stage1"

    execution {
      type = "RunProcedure"
      parameters = {
        procedure = komodo_procedure.child.id
      }
    }
  }

  stage {
    name = "Stage2"

    execution {
      type = "RunProcedure"
      parameters = {
        procedure = komodo_procedure.child.id
      }
    }
  }
}
`, name, name)
}

func testAccProcedureResourceConfigWithDisabledExecution(name string) string {
	return fmt.Sprintf(`
resource "komodo_procedure" "child" {
  name = "%s-child"
}

resource "komodo_procedure" "test" {
  name = %q

  stage {
    name = "Stage1"

    execution {
      type    = "RunProcedure"
      enabled = false
      parameters = {
        procedure = komodo_procedure.child.id
      }
    }
  }
}
`, name, name)
}

func testAccProcedureResourceConfigWithParameters(name, procID string) string {
	return fmt.Sprintf(`
resource "komodo_procedure" "test" {
  name = %q

  stage {
    name = "Run"

    execution {
      type = "RunProcedure"
      parameters = {
        procedure = %q
      }
    }
  }
}
`, name, procID)
}

func testAccProcedureResourceConfigWithParametersV1(name string) string {
	return fmt.Sprintf(`
resource "komodo_procedure" "child1" {
  name = "%s-child1"
}

resource "komodo_procedure" "child2" {
  name = "%s-child2"
}

resource "komodo_procedure" "test" {
  name = %q

  stage {
    name = "Run"

    execution {
      type = "RunProcedure"
      parameters = {
        procedure = komodo_procedure.child1.id
      }
    }
  }
}
`, name, name, name)
}

func testAccProcedureResourceConfigWithParametersV2(name string) string {
	return fmt.Sprintf(`
resource "komodo_procedure" "child1" {
  name = "%s-child1"
}

resource "komodo_procedure" "child2" {
  name = "%s-child2"
}

resource "komodo_procedure" "test" {
  name = %q

  stage {
    name = "Run"

    execution {
      type = "RunProcedure"
      parameters = {
        procedure = komodo_procedure.child2.id
      }
    }
  }
}
`, name, name, name)
}
