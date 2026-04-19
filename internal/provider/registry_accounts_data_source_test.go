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

func TestAccRegistryAccountsDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryAccountsDataSourceConfig("registry.example.com", "tf-ra-ds-basic", "token-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_registry_accounts.all", "registry_accounts.#"),
				),
			},
		},
	})
}

func TestAccRegistryAccountsDataSource_containsCreated(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryAccountsDataSourceConfig("registry.example.com", "tf-ra-ds-find", "token-find"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.komodo_registry_accounts.all",
						"registry_accounts.*",
						map[string]string{
							"domain":   "registry.example.com",
							"username": "tf-ra-ds-find",
						},
					),
				),
			},
		},
	})
}

func testAccRegistryAccountsDataSourceConfig(domain, username, token string) string {
	return fmt.Sprintf(`
resource "komodo_registry_account" "test" {
  domain   = %q
  username = %q
  token    = %q
}

data "komodo_registry_accounts" "all" {
  depends_on = [komodo_registry_account.test]
}
`, domain, username, token)
}

func TestUnitRegistryAccountsDataSource_configure(t *testing.T) {
d := &RegistryAccountsDataSource{}
resp := &datasource.ConfigureResponse{}
d.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: "wrong"}, resp)
if !resp.Diagnostics.HasError() {
t.Fatal("expected diagnostic error for wrong provider data type")
}
}
