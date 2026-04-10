// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

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
					resource.TestCheckResourceAttr("data.komodo_builder.by_name", "builder_type", "Url"),
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
					resource.TestCheckResourceAttr("data.komodo_builder.by_id", "builder_type", "Url"),
				),
			},
		},
	})
}

func testAccBuilderDataSourceConfig_byName(name, address string) string {
	return fmt.Sprintf(`
resource "komodo_builder" "src" {
  name         = %q
  builder_type = "Url"
  url_config = {
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
  name         = %q
  builder_type = "Url"
  url_config = {
    address = %q
  }
}

data "komodo_builder" "by_id" {
  id = komodo_builder.src.id
}
`, name, address)
}
