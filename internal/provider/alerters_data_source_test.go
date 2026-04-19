// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"testing"

	datasource "github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAlertersDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAlertersDataSourceConfig("tf-acc-alerters-ds-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_alerters.all", "alerters.#"),
				),
			},
		},
	})
}

func TestAccAlertersDataSource_containsCreated(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAlertersDataSourceConfig("tf-acc-alerters-ds-find"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.komodo_alerters.all",
						"alerters.*",
						map[string]string{
							"name": "tf-acc-alerters-ds-find",
						},
					),
				),
			},
		},
	})
}

func testAccAlertersDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_alerter" "test" {
  name = %q
  endpoint {
    type = "Custom"
    url  = "https://example.com/hook"
  }
}

data "komodo_alerters" "all" {
  depends_on = [komodo_alerter.test]
}
`, name)
}

func TestUnitAlertersDataSource_configure(t *testing.T) {
	d := &AlertersDataSource{}
	resp := &datasource.ConfigureResponse{}
	d.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: "wrong"}, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for wrong provider data type")
	}
}
