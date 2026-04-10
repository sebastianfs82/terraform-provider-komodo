// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccStacksDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStacksDataSourceConfig("tf-acc-stacks-ds-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_stacks.all", "stacks.#"),
				),
			},
		},
	})
}

func TestAccStacksDataSource_containsCreated(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStacksDataSourceConfig("tf-acc-stacks-ds-find"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.komodo_stacks.all",
						"stacks.*",
						map[string]string{
							"name": "tf-acc-stacks-ds-find",
						},
					),
				),
			},
		},
	})
}

func TestAccStacksDataSource_filteredByServerID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStacksDataSourceConfig_filteredByServerID("tf-acc-stacks-ds-server"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.komodo_stacks.filtered",
						"stacks.*",
						map[string]string{
							"name": "tf-acc-stacks-ds-server",
						},
					),
				),
			},
		},
	})
}

func testAccStacksDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_stack" "test" {
  name = %q
}

data "komodo_stacks" "all" {
  depends_on = [komodo_stack.test]
}
`, name)
}

func testAccStacksDataSourceConfig_filteredByServerID(name string) string {
	return fmt.Sprintf(`
data "komodo_servers" "all" {}

resource "komodo_stack" "test" {
  name      = %q
  server_id = data.komodo_servers.all.servers[0].id
}

data "komodo_stacks" "filtered" {
  server_id  = data.komodo_servers.all.servers[0].id
  depends_on = [komodo_stack.test]
}
`, name)
}
