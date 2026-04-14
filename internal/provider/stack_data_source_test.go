// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccStackDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStackDataSourceConfig_basic("tf-test-stack-ds"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_stack.test", "name", "tf-test-stack-ds"),
					resource.TestCheckResourceAttrSet("data.komodo_stack.test", "id"),
				),
			},
		},
	})
}

func TestAccStackDataSource_withGit(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStackDataSourceConfig_withGit("tf-test-stack-ds-git"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_stack.test", "name", "tf-test-stack-ds-git"),
					resource.TestCheckResourceAttrSet("data.komodo_stack.test", "source.url"),
					resource.TestCheckResourceAttr("data.komodo_stack.test", "source.path", "owner/my-stack-repo"),
					resource.TestCheckResourceAttr("data.komodo_stack.test", "source.branch", "main"),
					resource.TestCheckResourceAttrSet("data.komodo_stack.test", "id"),
				),
			},
		},
	})
}

func testAccStackDataSourceConfig_basic(name string) string {
	return `
resource "komodo_stack" "test" {
  name = "` + name + `"
}

data "komodo_stack" "test" {
  name       = komodo_stack.test.name
  depends_on = [komodo_stack.test]
}
`
}

func testAccStackDataSourceConfig_withGit(name string) string {
	return `
resource "komodo_stack" "test" {
  name = "` + name + `"

  source {
    url    = "https://github.com"
    path   = "owner/my-stack-repo"
    branch = "main"
  }
}

data "komodo_stack" "test" {
  name       = komodo_stack.test.name
  depends_on = [komodo_stack.test]
}
`
}
