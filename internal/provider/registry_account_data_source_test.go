// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRegistryAccountDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryAccountDataSourceConfig("registry.example.com", "ds-user", "ds-token"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_registry_account.test", "id"),
					resource.TestCheckResourceAttr("data.komodo_registry_account.test", "domain", "registry.example.com"),
					resource.TestCheckResourceAttr("data.komodo_registry_account.test", "username", "ds-user"),
				),
			},
		},
	})
}

func TestAccRegistryAccountDataSource_fields(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryAccountDataSourceConfig("ghcr.io", "ds-fields-user", "ds-fields-token"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_registry_account.test", "domain", "ghcr.io"),
					resource.TestCheckResourceAttr("data.komodo_registry_account.test", "username", "ds-fields-user"),
					resource.TestCheckResourceAttrSet("data.komodo_registry_account.test", "token"),
				),
			},
		},
	})
}

func testAccRegistryAccountDataSourceConfig(domain, username, token string) string {
	return testAccRegistryAccountResourceConfig(domain, username, token) + `

data "komodo_registry_account" "test" {
  id = komodo_registry_account.test.id
}
`
}
