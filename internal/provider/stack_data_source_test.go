// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	datasource "github.com/hashicorp/terraform-plugin-framework/datasource"
	"regexp"
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

func TestUnitStackDataSource_configure(t *testing.T) {
	d := &StackDataSource{}
	resp := &datasource.ConfigureResponse{}
	d.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: "wrong"}, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for wrong provider data type")
	}
}

func TestAccStackDataSource_bothSet_isError(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccStackDataSourceConfig_bothSet(),
				ExpectError: regexp.MustCompile(`Only one of`),
			},
		},
	})
}

func TestAccStackDataSource_neitherSet_isError(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccStackDataSourceConfig_neitherSet(),
				ExpectError: regexp.MustCompile(`One of`),
			},
		},
	})
}

func testAccStackDataSourceConfig_bothSet() string {
	return `
data "komodo_stack" "test" {
  id   = "some-id"
  name = "some-name"
}
`
}

func testAccStackDataSourceConfig_neitherSet() string {
	return `
data "komodo_stack" "test" {}
`
}
