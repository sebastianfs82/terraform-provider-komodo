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

func TestAccAlerterResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAlerterResourceCustomConfig("tf-acc-alerter-basic", "http://localhost:7000"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_alerter.test", "name", "tf-acc-alerter-basic"),
					resource.TestCheckResourceAttr("komodo_alerter.test", "endpoint_type", "Custom"),
					resource.TestCheckResourceAttr("komodo_alerter.test", "custom_endpoint.url", "http://localhost:7000"),
					resource.TestCheckResourceAttrSet("komodo_alerter.test", "id"),
				),
			},
		},
	})
}

func TestAccAlerterResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAlerterResourceCustomConfig("tf-acc-alerter-update", "http://localhost:7000"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_alerter.test", "custom_endpoint.url", "http://localhost:7000"),
				),
			},
			{
				Config: testAccAlerterResourceCustomConfig("tf-acc-alerter-update", "http://localhost:8000"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_alerter.test", "custom_endpoint.url", "http://localhost:8000"),
				),
			},
		},
	})
}

func TestAccAlerterResource_import(t *testing.T) {
	var alerterID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAlerterResourceCustomConfig("tf-acc-alerter-import", "http://localhost:7000"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_alerter.test", "id"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["komodo_alerter.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						alerterID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config:            testAccAlerterResourceCustomConfig("tf-acc-alerter-import", "http://localhost:7000"),
				ResourceName:      "komodo_alerter.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(_ *terraform.State) (string, error) { return alerterID, nil },
			},
		},
	})
}

func TestAccAlerterResource_disappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAlerterResourceCustomConfig("tf-acc-alerter-disappears", "http://localhost:7000"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_alerter.test", "id"),
					testAccAlerterDisappears("komodo_alerter.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAlerterDisappears(resourceName string) resource.TestCheckFunc {
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
		return c.DeleteAlerter(context.Background(), rs.Primary.ID)
	}
}

func testAccAlerterResourceCustomConfig(name, url string) string {
	return fmt.Sprintf(`
resource "komodo_alerter" "test" {
  name          = %q
  endpoint_type = "Custom"
  custom_endpoint = {
    url = %q
  }
}
`, name, url)
}
