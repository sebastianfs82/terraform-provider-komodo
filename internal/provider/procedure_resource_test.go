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
