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

func TestAccAlerterDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAlerterDataSourceConfig_basic("tf-acc-alerter-ds-basic", "http://localhost:7001"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_alerter.test", "id"),
					resource.TestCheckResourceAttr("data.komodo_alerter.test", "name", "tf-acc-alerter-ds-basic"),
				),
			},
		},
	})
}

func TestAccAlerterDataSource_fields(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAlerterDataSourceConfig_basic("tf-acc-alerter-ds-fields", "http://localhost:7002"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_alerter.test", "name", "tf-acc-alerter-ds-fields"),
					resource.TestCheckResourceAttr("data.komodo_alerter.test", "endpoint.type", "Custom"),
					resource.TestCheckResourceAttr("data.komodo_alerter.test", "endpoint.url", "http://localhost:7002"),
				),
			},
		},
	})
}

func testAccAlerterDataSourceConfig_basic(name, url string) string {
	return fmt.Sprintf(`
resource "komodo_alerter" "src" {
  name = %q
  endpoint {
    type = "Custom"
    url  = %q
  }
}

data "komodo_alerter" "test" {
  id         = komodo_alerter.src.id
  depends_on = [komodo_alerter.src]
}
`, name, url)
}

func TestUnitAlerterDataSource_configure(t *testing.T) {
d := &AlerterDataSource{}
resp := &datasource.ConfigureResponse{}
d.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: "wrong"}, resp)
if !resp.Diagnostics.HasError() {
t.Fatal("expected diagnostic error for wrong provider data type")
}
}

func TestAccAlerterDataSource_bothSet_isError(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccAlerterDataSourceConfig_bothSet(),
				ExpectError: regexp.MustCompile(`Only one of`),
			},
		},
	})
}

func TestAccAlerterDataSource_neitherSet_isError(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccAlerterDataSourceConfig_neitherSet(),
				ExpectError: regexp.MustCompile(`One of`),
			},
		},
	})
}

func testAccAlerterDataSourceConfig_bothSet() string {
	return `
data "komodo_alerter" "test" {
  id   = "some-id"
  name = "some-name"
}
`
}

func testAccAlerterDataSourceConfig_neitherSet() string {
	return `
data "komodo_alerter" "test" {}
`
}
