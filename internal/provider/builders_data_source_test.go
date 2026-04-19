// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"context"
	"testing"
	datasource "github.com/hashicorp/terraform-plugin-framework/datasource"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccBuildersDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBuildersDataSourceConfig("tf-acc-builders-ds-basic", "https://builder.example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_builders.all", "builders.#"),
				),
			},
		},
	})
}

func TestAccBuildersDataSource_containsCreated(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBuildersDataSourceConfig("tf-acc-builders-ds-find", "https://builder.example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.komodo_builders.all",
						"builders.*",
						map[string]string{
							"name": "tf-acc-builders-ds-find",
							"type": "Url",
						},
					),
				),
			},
		},
	})
}

func testAccBuildersDataSourceConfig(name, address string) string {
	return fmt.Sprintf(`
resource "komodo_builder" "test" {
  name = %q
  type = "Url"
  url_config {
    address = %q
  }
}

data "komodo_builders" "all" {
  depends_on = [komodo_builder.test]
}
`, name, address)
}

func TestUnitBuildersDataSource_configure(t *testing.T) {
d := &BuildersDataSource{}
resp := &datasource.ConfigureResponse{}
d.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: "wrong"}, resp)
if !resp.Diagnostics.HasError() {
t.Fatal("expected diagnostic error for wrong provider data type")
}
}
