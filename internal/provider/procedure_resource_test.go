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
					resource.TestCheckResourceAttr("komodo_procedure.test", "schedule_format", "Cron"),
					resource.TestCheckResourceAttr("komodo_procedure.test", "schedule", "0 * * * *"),
					resource.TestCheckResourceAttr("komodo_procedure.test", "schedule_enabled", "true"),
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
				ImportStateVerifyIgnore: []string{"webhook_secret"},
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
  name             = %q
  schedule_format  = "Cron"
  schedule         = "0 * * * *"
  schedule_enabled = true
}
`, name)
}
