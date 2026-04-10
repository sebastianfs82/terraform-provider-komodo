// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTagsDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTagsDataSourceConfig("tf-acc-tags-ds-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_tags.all", "tags.#"),
				),
			},
		},
	})
}

func TestAccTagsDataSource_containsCreated(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTagsDataSourceConfig("tf-acc-tags-ds-find"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.komodo_tags.all",
						"tags.*",
						map[string]string{
							"name": "tf-acc-tags-ds-find",
						},
					),
				),
			},
		},
	})
}

func testAccTagsDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_tag" "test" {
  name = %q
}

data "komodo_tags" "all" {
  depends_on = [komodo_tag.test]
}
`, name)
}
