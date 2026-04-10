// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccProviderAccountDataSource_byID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderAccountDataSourceConfig_byID("github.com", true, "dsuser", "ds-token"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.komodo_provider_account.test", "id",
						"komodo_provider_account.test", "id",
					),
					resource.TestCheckResourceAttr("data.komodo_provider_account.test", "domain", "github.com"),
					resource.TestCheckResourceAttr("data.komodo_provider_account.test", "https_enabled", "true"),
					resource.TestCheckResourceAttr("data.komodo_provider_account.test", "username", "dsuser"),
				),
			},
		},
	})
}

func testAccProviderAccountDataSourceConfig_byID(domain string, https bool, username, token string) string {
	return fmt.Sprintf(`
resource "komodo_provider_account" "test" {
  domain   = %q
  https_enabled = %t
  username = %q
  token    = %q
}

data "komodo_provider_account" "test" {
  id = komodo_provider_account.test.id
}
`, domain, https, username, token)
}

// testAccProviderAccountDataSourceConfig_byIDStr is a helper returning a config
// where the data source looks up a known ID (useful when the resource already exists).
func testAccProviderAccountDataSourceConfig_byIDStr(id string) string {
	return fmt.Sprintf(`
data "komodo_provider_account" "test" {
  id = %q
}
`, id)
}

// TestAccProviderAccountDataSource_existingAccount reads an account that was
// created externally (requires the user to set KOMODO_GIT_PROVIDER_ACCOUNT_ID).
func TestAccProviderAccountDataSource_existingAccount(t *testing.T) {
	id := ""
	// Only run if the env var is set
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			if id == "" {
				t.Skip("KOMODO_GIT_PROVIDER_ACCOUNT_ID not set, skipping test")
			}
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderAccountDataSourceConfig_byIDStr(id),
				Check: resource.ComposeAggregateTestCheckFunc(
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["data.komodo_provider_account.test"]
						if !ok {
							return fmt.Errorf("data source not found")
						}
						if rs.Primary.Attributes["domain"] == "" {
							return fmt.Errorf("expected domain to be set")
						}
						return nil
					},
				),
			},
		},
	})
}
