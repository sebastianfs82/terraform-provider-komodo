// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccProviderAccountsDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderAccountsDataSourceConfig("github.com", true, "tf-pa-ds-basic", "token-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_provider_accounts.all", "provider_accounts.#"),
				),
			},
		},
	})
}

func TestAccProviderAccountsDataSource_containsCreated(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderAccountsDataSourceConfig("github.com", true, "tf-pa-ds-find", "token-find"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.komodo_provider_accounts.all",
						"provider_accounts.*",
						map[string]string{
							"domain":   "github.com",
							"username": "tf-pa-ds-find",
						},
					),
				),
			},
		},
	})
}

func testAccProviderAccountsDataSourceConfig(domain string, https bool, username, token string) string {
	return fmt.Sprintf(`
resource "komodo_provider_account" "test" {
  domain        = %q
  https_enabled = %t
  username      = %q
  token         = %q
}

data "komodo_provider_accounts" "all" {
  depends_on = [komodo_provider_account.test]
}
`, domain, https, username, token)
}
