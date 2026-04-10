// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"terraform-provider-komodo/internal/client"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccRegistryAccountResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryAccountResourceConfig("registry.example.com", "testuser", "mytoken123"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_registry_account.test", "domain", "registry.example.com"),
					resource.TestCheckResourceAttr("komodo_registry_account.test", "username", "testuser"),
					resource.TestCheckResourceAttr("komodo_registry_account.test", "token", "mytoken123"),
					resource.TestCheckResourceAttrSet("komodo_registry_account.test", "id"),
				),
			},
		},
	})
}

func TestAccRegistryAccountResource_updateToken(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryAccountResourceConfig("ghcr.io", "updateuser", "original-token"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_registry_account.test", "domain", "ghcr.io"),
					resource.TestCheckResourceAttr("komodo_registry_account.test", "username", "updateuser"),
					resource.TestCheckResourceAttr("komodo_registry_account.test", "token", "original-token"),
				),
			},
			{
				Config: testAccRegistryAccountResourceConfig("ghcr.io", "updateuser", "updated-token"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_registry_account.test", "domain", "ghcr.io"),
					resource.TestCheckResourceAttr("komodo_registry_account.test", "username", "updateuser"),
					resource.TestCheckResourceAttr("komodo_registry_account.test", "token", "updated-token"),
				),
			},
		},
	})
}

func TestAccRegistryAccountResource_import(t *testing.T) {
	var accountID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryAccountResourceConfig("registry.example.com", "importuser", "import-token"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_registry_account.test", "id"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["komodo_registry_account.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						accountID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config:                  testAccRegistryAccountResourceConfig("registry.example.com", "importuser", "import-token"),
				ResourceName:            "komodo_registry_account.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       func(_ *terraform.State) (string, error) { return accountID, nil },
				ImportStateVerifyIgnore: []string{"token"},
			},
		},
	})
}

func TestAccRegistryAccountResource_disappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryAccountResourceConfig("registry.example.com", "disappearuser", "disappear-token"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_registry_account.test", "id"),
					testAccRegistryAccountDisappears("komodo_registry_account.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccRegistryAccountDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found in state: %s", resourceName)
		}
		c := client.NewClient(
			os.Getenv("KOMODO_ENDPOINT"),
			os.Getenv("KOMODO_USERNAME"),
			os.Getenv("KOMODO_PASSWORD"),
		)
		return c.DeleteDockerRegistryAccount(context.Background(), rs.Primary.ID)
	}
}

func testAccRegistryAccountResourceConfig(domain, username, token string) string {
	return fmt.Sprintf(`
resource "komodo_registry_account" "test" {
  domain   = %q
  username = %q
  token    = %q
}
`, domain, username, token)
}
