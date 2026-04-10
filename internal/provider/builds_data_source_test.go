// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccBuildsDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBuildsDataSourceConfig("tf-acc-builds-ds-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_builds.all", "builds.#"),
				),
			},
		},
	})
}

func TestAccBuildsDataSource_containsCreated(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBuildsDataSourceConfig("tf-acc-builds-ds-find"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.komodo_builds.all",
						"builds.*",
						map[string]string{
							"name": "tf-acc-builds-ds-find",
						},
					),
				),
			},
		},
	})
}

func testAccBuildsDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_build" "test" {
  name       = %q
  image_name = "test-image"
}

data "komodo_builds" "all" {
  depends_on = [komodo_build.test]
}
`, name)
}
