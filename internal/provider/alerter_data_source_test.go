// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

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
					resource.TestCheckResourceAttr("data.komodo_alerter.test", "endpoint_type", "Custom"),
					resource.TestCheckResourceAttr("data.komodo_alerter.test", "custom_endpoint.url", "http://localhost:7002"),
				),
			},
		},
	})
}

func testAccAlerterDataSourceConfig_basic(name, url string) string {
	return fmt.Sprintf(`
resource "komodo_alerter" "src" {
  name          = %q
  endpoint_type = "Custom"
  custom_endpoint = {
    url = %q
  }
}

data "komodo_alerter" "test" {
  id         = komodo_alerter.src.id
  depends_on = [komodo_alerter.src]
}
`, name, url)
}
