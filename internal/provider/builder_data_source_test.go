// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"context"
	"regexp"
	"testing"
	datasource "github.com/hashicorp/terraform-plugin-framework/datasource"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccBuilderDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBuilderDataSourceConfig_byName("tf-acc-builder-ds-basic", "http://localhost:8120"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_builder.by_name", "id"),
					resource.TestCheckResourceAttr("data.komodo_builder.by_name", "name", "tf-acc-builder-ds-basic"),
				),
			},
		},
	})
}

func TestAccBuilderDataSource_fields(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBuilderDataSourceConfig_byName("tf-acc-builder-ds-fields", "http://localhost:8121"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_builder.by_name", "name", "tf-acc-builder-ds-fields"),
					resource.TestCheckResourceAttr("data.komodo_builder.by_name", "type", "Url"),
					resource.TestCheckResourceAttr("data.komodo_builder.by_name", "url_config.address", "http://localhost:8121"),
				),
			},
		},
	})
}

func TestAccBuilderDataSource_byID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBuilderDataSourceConfig_byID("tf-acc-builder-ds-byid", "http://localhost:8122"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_builder.by_id", "id"),
					resource.TestCheckResourceAttr("data.komodo_builder.by_id", "name", "tf-acc-builder-ds-byid"),
					resource.TestCheckResourceAttr("data.komodo_builder.by_id", "type", "Url"),
				),
			},
		},
	})
}

func testAccBuilderDataSourceConfig_byName(name, address string) string {
	return fmt.Sprintf(`
resource "komodo_builder" "src" {
  name       = %q
  type       = "Url"
  url_config {
    address = %q
  }
}

data "komodo_builder" "by_name" {
  name = komodo_builder.src.name
}
`, name, address)
}

func testAccBuilderDataSourceConfig_byID(name, address string) string {
	return fmt.Sprintf(`
resource "komodo_builder" "src" {
  name       = %q
  type       = "Url"
  url_config {
    address = %q
  }
}

data "komodo_builder" "by_id" {
  id = komodo_builder.src.id
}
`, name, address)
}

func TestUnitBuilderDataSource_configure(t *testing.T) {
d := &BuilderDataSource{}
resp := &datasource.ConfigureResponse{}
d.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: "wrong"}, resp)
if !resp.Diagnostics.HasError() {
t.Fatal("expected diagnostic error for wrong provider data type")
}
}

func TestAccBuilderDataSource_bothSet_isError(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccBuilderDataSourceConfig_bothSet(),
				ExpectError: regexp.MustCompile(`Only one of`),
			},
		},
	})
}

func TestAccBuilderDataSource_neitherSet_isError(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccBuilderDataSourceConfig_neitherSet(),
				ExpectError: regexp.MustCompile(`One of`),
			},
		},
	})
}

func testAccBuilderDataSourceConfig_bothSet() string {
	return `
data "komodo_builder" "test" {
  id   = "some-id"
  name = "some-name"
}
`
}

func testAccBuilderDataSourceConfig_neitherSet() string {
	return `
data "komodo_builder" "test" {}
`
}
