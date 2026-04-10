// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOnboardingKeyDataSource_byName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOnboardingKeyDataSourceConfig_byName("tf-acc-ok-ds-name"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_onboarding_key.test", "name", "tf-acc-ok-ds-name"),
					resource.TestCheckResourceAttrSet("data.komodo_onboarding_key.test", "public_key"),
					resource.TestCheckResourceAttrSet("data.komodo_onboarding_key.test", "enabled"),
				),
			},
		},
	})
}

func TestAccOnboardingKeyDataSource_byPublicKey(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOnboardingKeyDataSourceConfig_byPublicKey("tf-acc-ok-ds-pubkey"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_onboarding_key.by_key", "name", "tf-acc-ok-ds-pubkey"),
					resource.TestCheckResourceAttrSet("data.komodo_onboarding_key.by_key", "public_key"),
				),
			},
		},
	})
}

func testAccOnboardingKeyDataSourceConfig_byName(name string) string {
	return fmt.Sprintf(`
resource "komodo_onboarding_key" "test" {
  name = %[1]q
}

data "komodo_onboarding_key" "test" {
  name       = %[1]q
  depends_on = [komodo_onboarding_key.test]
}
`, name)
}

func testAccOnboardingKeyDataSourceConfig_byPublicKey(name string) string {
	return fmt.Sprintf(`
resource "komodo_onboarding_key" "src" {
  name = %[1]q
}

data "komodo_onboarding_key" "by_key" {
  public_key = komodo_onboarding_key.src.public_key
  depends_on = [komodo_onboarding_key.src]
}
`, name)
}
