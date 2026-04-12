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

func TestAccTagResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTagResourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_tag.example", "name"),
					resource.TestCheckResourceAttrSet("komodo_tag.example", "color"),
				),
			},
		},
	})
}

const testAccTagResourceConfig = `
resource "komodo_tag" "example" {
  name  = "tf_tag"
  color = "Green"
}
`

func TestAccTagResource_disappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTagResourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_tag.example", "id"),
					testAccTagDisappears("komodo_tag.example"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccTagResource_updateColor(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTagResourceConfigWithColor("tf-acc-tag-color-update", "Green"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_tag.test", "color", "Green"),
				),
			},
			{
				Config: testAccTagResourceConfigWithColor("tf-acc-tag-color-update", "Blue"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_tag.test", "color", "Blue"),
				),
			},
		},
	})
}

func TestAccTagResource_ownerStableOnUpdate(t *testing.T) {
	var savedOwner string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTagResourceConfigWithColor("tf-acc-tag-owner-stable", "Green"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_tag.test", "owner"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["komodo_tag.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						savedOwner = rs.Primary.Attributes["owner"]
						return nil
					},
				),
			},
			{
				Config: testAccTagResourceConfigWithColor("tf-acc-tag-owner-stable", "Blue"),
				Check: resource.ComposeAggregateTestCheckFunc(
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["komodo_tag.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						if rs.Primary.Attributes["owner"] != savedOwner {
							return fmt.Errorf("owner changed after color update: was %q, got %q", savedOwner, rs.Primary.Attributes["owner"])
						}
						return nil
					},
				),
			},
		},
	})
}

func TestAccTagResource_importState(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTagResourceConfigWithColor("tf-acc-tag-import", "Red"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_tag.test", "id"),
				),
			},
			{
				ResourceName:      "komodo_tag.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccTagResourceConfigWithColor(name, color string) string {
	return fmt.Sprintf(`
resource "komodo_tag" "test" {
  name  = %q
  color = %q
}
`, name, color)
}

func testAccTagDisappears(resourceName string) resource.TestCheckFunc {
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
		return c.DeleteTag(context.Background(), rs.Primary.Attributes["name"])
	}
}
