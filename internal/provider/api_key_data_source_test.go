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

func TestAccApiKeyDataSource_byKey(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccApiKeyDataSourceConfig_byKey("tf-acc-apikey-ds-bykey"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_api_key.test", "name", "tf-acc-apikey-ds-bykey"),
					resource.TestCheckResourceAttrSet("data.komodo_api_key.test", "key"),
					resource.TestCheckResourceAttrSet("data.komodo_api_key.test", "user_id"),
					resource.TestCheckResourceAttrSet("data.komodo_api_key.test", "created_at"),
					resource.TestCheckResourceAttr("data.komodo_api_key.test", "expires_at", ""),
				),
			},
		},
	})
}

func TestAccApiKeyDataSource_byName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccApiKeyDataSourceConfig_byName("tf-acc-apikey-ds-byname"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_api_key.test", "name", "tf-acc-apikey-ds-byname"),
					resource.TestCheckResourceAttrSet("data.komodo_api_key.test", "key"),
					resource.TestCheckResourceAttrSet("data.komodo_api_key.test", "user_id"),
				),
			},
		},
	})
}

func testAccApiKeyDataSourceConfig_byKey(name string) string {
	return fmt.Sprintf(`
resource "komodo_api_key" "test" {
  name = %[1]q
}

data "komodo_api_key" "test" {
  key        = komodo_api_key.test.key
  depends_on = [komodo_api_key.test]
}
`, name)
}

func testAccApiKeyDataSourceConfig_byName(name string) string {
	return fmt.Sprintf(`
resource "komodo_api_key" "test" {
  name = %[1]q
}

data "komodo_api_key" "test" {
  name       = %[1]q
  depends_on = [komodo_api_key.test]
}
`, name)
}

func TestUnitApiKeyDataSource_configure(t *testing.T) {
d := &ApiKeyDataSource{}
resp := &datasource.ConfigureResponse{}
d.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: "wrong"}, resp)
if !resp.Diagnostics.HasError() {
t.Fatal("expected diagnostic error for wrong provider data type")
}
}
