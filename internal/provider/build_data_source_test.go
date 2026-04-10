// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccBuildDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBuildDataSourceConfig_basic("tf-acc-build-ds-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_build.test", "name", "tf-acc-build-ds-basic"),
					resource.TestCheckResourceAttrSet("data.komodo_build.test", "id"),
				),
			},
		},
	})
}

func TestAccBuildDataSource_fields(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBuildDataSourceConfig_fields("tf-acc-build-ds-fields"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_build.test", "repo", "myorg/myrepo"),
					resource.TestCheckResourceAttr("data.komodo_build.test", "branch", "main"),
					resource.TestCheckResourceAttr("data.komodo_build.test", "include_latest_tag", "true"),
				),
			},
		},
	})
}

func testAccBuildDataSourceConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "komodo_build" "test" {
  name = %q
}

data "komodo_build" "test" {
  id = komodo_build.test.id
}
`, name)
}

func testAccBuildDataSourceConfig_fields(name string) string {
	return fmt.Sprintf(`
resource "komodo_build" "test" {
  name               = %q
  repo               = "myorg/myrepo"
  branch             = "main"
  include_latest_tag = true
}

data "komodo_build" "test" {
  id = komodo_build.test.id
}
`, name)
}
