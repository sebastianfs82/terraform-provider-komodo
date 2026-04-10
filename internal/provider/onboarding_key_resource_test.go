// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccOnboardingKeyResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccOnboardingKeyResourceConfig_basic("tf-onboarding-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_onboarding_key.test", "name", "tf-onboarding-basic"),
					resource.TestCheckResourceAttrSet("komodo_onboarding_key.test", "public_key"),
					resource.TestCheckResourceAttrSet("komodo_onboarding_key.test", "private_key"),
					resource.TestCheckResourceAttr("komodo_onboarding_key.test", "enabled", "true"),
					resource.TestCheckResourceAttr("komodo_onboarding_key.test", "expires", "0"),
					resource.TestCheckResourceAttr("komodo_onboarding_key.test", "privileged", "false"),
					resource.TestCheckResourceAttr("komodo_onboarding_key.test", "create_builder", "false"),
				),
			},
		},
	})
}

func TestAccOnboardingKeyResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create
			{
				Config: testAccOnboardingKeyResourceConfig_basic("tf-onboarding-update"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_onboarding_key.test", "name", "tf-onboarding-update"),
					resource.TestCheckResourceAttr("komodo_onboarding_key.test", "enabled", "true"),
				),
			},
			// Update: disable and rename
			{
				Config: testAccOnboardingKeyResourceConfig_disabled("tf-onboarding-updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_onboarding_key.test", "name", "tf-onboarding-updated"),
					resource.TestCheckResourceAttr("komodo_onboarding_key.test", "enabled", "false"),
				),
			},
			// Re-enable
			{
				Config: testAccOnboardingKeyResourceConfig_basic("tf-onboarding-updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_onboarding_key.test", "name", "tf-onboarding-updated"),
					resource.TestCheckResourceAttr("komodo_onboarding_key.test", "enabled", "true"),
				),
			},
		},
	})
}

func TestAccOnboardingKeyResource_import(t *testing.T) {
	var publicKey string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and capture public_key
			{
				Config: testAccOnboardingKeyResourceConfig_basic("tf-onboarding-import"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_onboarding_key.test", "name", "tf-onboarding-import"),
					resource.TestCheckResourceAttrSet("komodo_onboarding_key.test", "public_key"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["komodo_onboarding_key.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						publicKey = rs.Primary.Attributes["public_key"]
						return nil
					},
				),
			},
			// Import by public_key
			{
				Config:                               testAccOnboardingKeyResourceConfig_basic("tf-onboarding-import"),
				ResourceName:                         "komodo_onboarding_key.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "public_key",
				// private_key is only available on creation
				ImportStateVerifyIgnore: []string{"private_key"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return publicKey, nil
				},
			},
		},
	})
}

func TestAccOnboardingKeyResource_disappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOnboardingKeyResourceConfig_basic("tf-onboarding-disappear"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_onboarding_key.test", "name", "tf-onboarding-disappear"),
				),
			},
		},
	})
}

// Test configuration helpers

func testAccOnboardingKeyResourceConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "komodo_onboarding_key" "test" {
  name = %[1]q
}
`, name)
}

func testAccOnboardingKeyResourceConfig_disabled(name string) string {
	return fmt.Sprintf(`
resource "komodo_onboarding_key" "test" {
  name    = %[1]q
  enabled = false
}
`, name)
}
