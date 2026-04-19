// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"context"
	"testing"
	datasource "github.com/hashicorp/terraform-plugin-framework/datasource"

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

// TestAccProviderAccountDataSource_existingAccount reads an account that was
// created via the komodo_provider_account resource and verifies the data source
// returns the expected attributes.
func TestAccProviderAccountDataSource_existingAccount(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderAccountDataSourceConfig_existingAccount(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.komodo_provider_account.test", "id",
						"komodo_provider_account.test", "id",
					),
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

func testAccProviderAccountDataSourceConfig_existingAccount() string {
	return `
resource "komodo_provider_account" "test" {
  domain        = "github.com"
  https_enabled = true
  username      = "tf-acc-existing-account"
  token         = "acc-test-token"
}

data "komodo_provider_account" "test" {
  id = komodo_provider_account.test.id
}
`
}

func TestUnitProviderAccountDataSource_configure(t *testing.T) {
d := &ProviderAccountDataSource{}
resp := &datasource.ConfigureResponse{}
d.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: "wrong"}, resp)
if !resp.Diagnostics.HasError() {
t.Fatal("expected diagnostic error for wrong provider data type")
}
}
