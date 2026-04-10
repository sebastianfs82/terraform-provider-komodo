// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccProviderAccountResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderAccountResourceConfig_basic("github.com", true, "testuser", "mytoken123"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_provider_account.test", "domain", "github.com"),
					resource.TestCheckResourceAttr("komodo_provider_account.test", "https_enabled", "true"),
					resource.TestCheckResourceAttr("komodo_provider_account.test", "username", "testuser"),
					resource.TestCheckResourceAttr("komodo_provider_account.test", "token", "mytoken123"),
					resource.TestCheckResourceAttrSet("komodo_provider_account.test", "id"),
				),
			},
		},
	})
}

func TestAccProviderAccountResource_updateToken(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderAccountResourceConfig_basic("gitlab.com", true, "updateuser", "original-token"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_provider_account.test", "domain", "gitlab.com"),
					resource.TestCheckResourceAttr("komodo_provider_account.test", "username", "updateuser"),
					resource.TestCheckResourceAttr("komodo_provider_account.test", "token", "original-token"),
				),
			},
			{
				Config: testAccProviderAccountResourceConfig_basic("gitlab.com", true, "updateuser", "updated-token"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_provider_account.test", "domain", "gitlab.com"),
					resource.TestCheckResourceAttr("komodo_provider_account.test", "username", "updateuser"),
					resource.TestCheckResourceAttr("komodo_provider_account.test", "token", "updated-token"),
				),
			},
		},
	})
}

func TestAccProviderAccountResource_updateHttps(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderAccountResourceConfig_basic("bitbucket.org", false, "httpsuser", "token-abc"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_provider_account.test", "https_enabled", "false"),
				),
			},
			{
				Config: testAccProviderAccountResourceConfig_basic("bitbucket.org", true, "httpsuser", "token-abc"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_provider_account.test", "https_enabled", "true"),
				),
			},
		},
	})
}

func TestAccProviderAccountResource_import(t *testing.T) {
	var accountID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderAccountResourceConfig_basic("github.com", true, "importuser", "import-token"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_provider_account.test", "id"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["komodo_provider_account.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						accountID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config:                  testAccProviderAccountResourceConfig_basic("github.com", true, "importuser", "import-token"),
				ResourceName:            "komodo_provider_account.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       func(_ *terraform.State) (string, error) { return accountID, nil },
				ImportStateVerifyIgnore: []string{"token"},
			},
		},
	})
}

func testAccProviderAccountResourceConfig_basic(domain string, https bool, username, token string) string {
	return fmt.Sprintf(`
resource "komodo_provider_account" "test" {
  domain        = %q
  https_enabled = %t
  username      = %q
  token         = %q
}
`, domain, https, username, token)
}

func TestAccProviderAccountResource_disappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderAccountResourceConfig_basic("github.com", true, "disappearuser", "disappear-token"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_provider_account.test", "id"),
					testAccProviderAccountDisappears("komodo_provider_account.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// testAccProviderAccountDisappears deletes the git provider account directly via the
// API (simulating manual deletion outside Terraform) so that the next plan detects drift.
func testAccProviderAccountDisappears(resourceName string) resource.TestCheckFunc {
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
		return c.DeleteGitProviderAccount(context.Background(), rs.Primary.ID)
	}
}
