// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceSyncsDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSyncsDataSourceConfig("tf-acc-rsyncs-ds-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_resource_syncs.all", "resource_syncs.#"),
				),
			},
		},
	})
}

func TestAccResourceSyncsDataSource_containsCreated(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSyncsDataSourceConfig("tf-acc-rsyncs-ds-find"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.komodo_resource_syncs.all",
						"resource_syncs.*",
						map[string]string{
							"name": "tf-acc-rsyncs-ds-find",
						},
					),
				),
			},
		},
	})
}

func testAccResourceSyncsDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_resource_sync" "test" {
  name          = %q
  file_contents = "# list ds test"
}

data "komodo_resource_syncs" "all" {
  depends_on = [komodo_resource_sync.test]
}
`, name)
}
