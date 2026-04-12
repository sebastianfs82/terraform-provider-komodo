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

func TestAccActionResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccActionResourceConfig("tf-acc-action-basic", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_action.test", "name", "tf-acc-action-basic"),
					resource.TestCheckResourceAttrSet("komodo_action.test", "id"),
				),
			},
		},
	})
}

func TestAccActionResource_update(t *testing.T) {
	const name = "tf-acc-action-update"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccActionResourceConfig(name, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_action.test", "name", name),
				),
			},
			{
				Config: testAccActionResourceConfigWithFileContents(name, "console.log('hello');"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_action.test", "file_contents", "console.log('hello');"),
				),
			},
		},
	})
}

func TestAccActionResource_importState(t *testing.T) {
	var actionID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccActionResourceConfig("tf-acc-action-import", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_action.test", "id"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["komodo_action.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						actionID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config:            testAccActionResourceConfig("tf-acc-action-import", ""),
				ResourceName:      "komodo_action.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return actionID, nil
				},
				ImportStateVerifyIgnore: []string{"webhook_secret"},
			},
		},
	})
}

func TestAccActionResource_disappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccActionResourceConfig("tf-acc-action-disappears", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_action.test", "id"),
					testAccActionDisappears("komodo_action.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccActionResource_rename(t *testing.T) {
	var savedID string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccActionResourceConfig("tf-acc-action-rename-orig", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_action.test", "name", "tf-acc-action-rename-orig"),
					resource.TestCheckResourceAttrSet("komodo_action.test", "id"),
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["komodo_action.test"]
						savedID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config: testAccActionResourceConfig("tf-acc-action-rename-new", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_action.test", "name", "tf-acc-action-rename-new"),
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["komodo_action.test"]
						if rs.Primary.ID != savedID {
							return fmt.Errorf("resource was recreated: ID changed from %q to %q", savedID, rs.Primary.ID)
						}
						return nil
					},
				),
			},
		},
	})
}

func testAccActionDisappears(resourceName string) resource.TestCheckFunc {
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
		return c.DeleteAction(context.Background(), rs.Primary.ID)
	}
}

func testAccActionResourceConfig(name, _ string) string {
	return fmt.Sprintf(`
resource "komodo_action" "test" {
  name = %q
}
`, name)
}

func testAccActionResourceConfigWithFileContents(name, fileContents string) string {
	return fmt.Sprintf(`
resource "komodo_action" "test" {
  name          = %q
  file_contents = %q
}
`, name, fileContents)
}
