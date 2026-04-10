// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccProcedureDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProcedureDataSourceConfig_basic("tf-acc-procedure-ds-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_procedure.test", "id"),
					resource.TestCheckResourceAttr("data.komodo_procedure.test", "name", "tf-acc-procedure-ds-basic"),
				),
			},
		},
	})
}

func TestAccProcedureDataSource_fields(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProcedureDataSourceConfig_withSchedule("tf-acc-procedure-ds-fields"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_procedure.test", "name", "tf-acc-procedure-ds-fields"),
					resource.TestCheckResourceAttr("data.komodo_procedure.test", "schedule_format", "Cron"),
					resource.TestCheckResourceAttr("data.komodo_procedure.test", "schedule", "0 * * * *"),
					resource.TestCheckResourceAttr("data.komodo_procedure.test", "schedule_enabled", "true"),
					resource.TestCheckResourceAttrSet("data.komodo_procedure.test", "failure_alert"),
					resource.TestCheckResourceAttrSet("data.komodo_procedure.test", "webhook_enabled"),
				),
			},
		},
	})
}

func testAccProcedureDataSourceConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "komodo_procedure" "src" {
  name = %q
}

data "komodo_procedure" "test" {
  id         = komodo_procedure.src.id
  depends_on = [komodo_procedure.src]
}
`, name)
}

func testAccProcedureDataSourceConfig_withSchedule(name string) string {
	return fmt.Sprintf(`
resource "komodo_procedure" "src" {
  name             = %q
  schedule_format  = "Cron"
  schedule         = "0 * * * *"
  schedule_enabled = true
}

data "komodo_procedure" "test" {
  id         = komodo_procedure.src.id
  depends_on = [komodo_procedure.src]
}
`, name)
}
