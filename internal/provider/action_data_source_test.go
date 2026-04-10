// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccActionDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccActionDataSourceConfig_basic("tf-acc-action-ds-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_action.test", "name", "tf-acc-action-ds-basic"),
					resource.TestCheckResourceAttrSet("data.komodo_action.test", "id"),
				),
			},
		},
	})
}

func TestAccActionDataSource_fields(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccActionDataSourceConfig_fields("tf-acc-action-ds-fields"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_action.test", "webhook_enabled", "true"),
					resource.TestCheckResourceAttr("data.komodo_action.test", "failure_alert", "true"),
					resource.TestCheckResourceAttr("data.komodo_action.test", "file_contents", "console.log('test');"),
				),
			},
		},
	})
}

func testAccActionDataSourceConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "komodo_action" "test" {
  name = %q
}

data "komodo_action" "test" {
  id = komodo_action.test.id
}
`, name)
}

func testAccActionDataSourceConfig_fields(name string) string {
	return fmt.Sprintf(`
resource "komodo_action" "test" {
  name            = %q
  webhook_enabled = true
  failure_alert   = true
  file_contents   = "console.log('test');"
}

data "komodo_action" "test" {
  id = komodo_action.test.id
}
`, name)
}
