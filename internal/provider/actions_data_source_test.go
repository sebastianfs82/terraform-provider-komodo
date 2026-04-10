// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccActionsDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccActionsDataSourceConfig("tf-acc-actions-ds-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_actions.all", "actions.#"),
				),
			},
		},
	})
}

func TestAccActionsDataSource_containsCreated(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccActionsDataSourceConfig("tf-acc-actions-ds-find"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.komodo_actions.all",
						"actions.*",
						map[string]string{
							"name": "tf-acc-actions-ds-find",
						},
					),
				),
			},
		},
	})
}

func testAccActionsDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_action" "test" {
  name = %q
}

data "komodo_actions" "all" {
  depends_on = [komodo_action.test]
}
`, name)
}
