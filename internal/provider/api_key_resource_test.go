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

func TestAccApiKeyResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccApiKeyResourceConfig("tf-acc-api-key-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_api_key.test", "name", "tf-acc-api-key-basic"),
					resource.TestCheckResourceAttr("komodo_api_key.test", "expires", "0"),
					resource.TestCheckResourceAttrSet("komodo_api_key.test", "key"),
					resource.TestCheckResourceAttrSet("komodo_api_key.test", "secret"),
					resource.TestCheckResourceAttrSet("komodo_api_key.test", "user_id"),
					resource.TestCheckResourceAttrSet("komodo_api_key.test", "created_at"),
				),
			},
		},
	})
}

func TestAccApiKeyResource_withExpiration(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccApiKeyResourceConfigWithExpiration("tf-acc-api-key-expiring", 1893456000000),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_api_key.test", "name", "tf-acc-api-key-expiring"),
					resource.TestCheckResourceAttr("komodo_api_key.test", "expires", "1893456000000"),
					resource.TestCheckResourceAttrSet("komodo_api_key.test", "key"),
					resource.TestCheckResourceAttrSet("komodo_api_key.test", "secret"),
				),
			},
		},
	})
}

func TestAccApiKeyResource_serviceUser(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccApiKeyResourceConfigServiceUser("tf-acc-svc-apikey", "tf-acc-svc-apikey-key"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_api_key.test", "name", "tf-acc-svc-apikey-key"),
					resource.TestCheckResourceAttr("komodo_api_key.test", "expires", "0"),
					resource.TestCheckResourceAttrSet("komodo_api_key.test", "key"),
					resource.TestCheckResourceAttrSet("komodo_api_key.test", "secret"),
					resource.TestCheckResourceAttrPair(
						"komodo_api_key.test", "service_user_id",
						"komodo_service_user.svc", "id",
					),
				),
			},
		},
	})
}

func TestAccApiKeyResource_importState(t *testing.T) {
	var keyValue string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccApiKeyResourceConfig("tf-acc-api-key-import"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_api_key.test", "key"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["komodo_api_key.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						keyValue = rs.Primary.Attributes["key"]
						return nil
					},
				),
			},
			{
				Config:            testAccApiKeyResourceConfig("tf-acc-api-key-import"),
				ResourceName:      "komodo_api_key.test",
				ImportState:       true,
				ImportStateVerify: true,
				// Secret is only available on creation and cannot be retrieved via import.
				ImportStateVerifyIgnore: []string{"secret"},
				ImportStateIdFunc: func(_ *terraform.State) (string, error) {
					return keyValue, nil
				},
			},
		},
	})
}

func TestAccApiKeyResource_disappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccApiKeyResourceConfig("tf-acc-api-key-disappears"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_api_key.test", "key"),
					testAccApiKeyDisappears("komodo_api_key.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccApiKeyDisappears(resourceName string) resource.TestCheckFunc {
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
		return c.DeleteApiKey(context.Background(), client.DeleteApiKeyRequest{Key: rs.Primary.Attributes["key"]})
	}
}

func testAccApiKeyResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_api_key" "test" {
  name = %q
}
`, name)
}

func testAccApiKeyResourceConfigWithExpiration(name string, expires int64) string {
	return fmt.Sprintf(`
resource "komodo_api_key" "test" {
  name    = %q
  expires = %d
}
`, name, expires)
}

func testAccApiKeyResourceConfigServiceUser(username, keyName string) string {
	return fmt.Sprintf(`
resource "komodo_service_user" "svc" {
  username = %q
}

resource "komodo_api_key" "test" {
  name            = %q
  service_user_id = komodo_service_user.svc.id
}
`, username, keyName)
}
