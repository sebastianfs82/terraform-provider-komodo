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

func TestAccBuilderResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBuilderResourceUrlConfig("tf-acc-builder-basic", "http://localhost:8120"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_builder.test", "name", "tf-acc-builder-basic"),
					resource.TestCheckResourceAttr("komodo_builder.test", "type", "Url"),
					resource.TestCheckResourceAttr("komodo_builder.test", "url_config.address", "http://localhost:8120"),
					resource.TestCheckResourceAttrSet("komodo_builder.test", "id"),
				),
			},
		},
	})
}

func TestAccBuilderResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBuilderResourceUrlConfig("tf-acc-builder-update", "http://localhost:8120"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_builder.test", "url_config.address", "http://localhost:8120"),
				),
			},
			{
				Config: testAccBuilderResourceUrlConfig("tf-acc-builder-update", "http://localhost:9000"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_builder.test", "url_config.address", "http://localhost:9000"),
				),
			},
		},
	})
}

func TestAccBuilderResource_import(t *testing.T) {
	var builderID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBuilderResourceUrlConfig("tf-acc-builder-import", "http://localhost:8120"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_builder.test", "id"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["komodo_builder.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						builderID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config:            testAccBuilderResourceUrlConfig("tf-acc-builder-import", "http://localhost:8120"),
				ResourceName:      "komodo_builder.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(_ *terraform.State) (string, error) { return builderID, nil },
			},
		},
	})
}

func TestAccBuilderResource_disappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBuilderResourceUrlConfig("tf-acc-builder-disappears", "http://localhost:8120"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_builder.test", "id"),
					testAccBuilderDisappears("komodo_builder.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccBuilderResource_rename(t *testing.T) {
	var savedID string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBuilderResourceUrlConfig("tf-acc-builder-rename-orig", "http://localhost:8120"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_builder.test", "name", "tf-acc-builder-rename-orig"),
					resource.TestCheckResourceAttrSet("komodo_builder.test", "id"),
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["komodo_builder.test"]
						savedID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config: testAccBuilderResourceUrlConfig("tf-acc-builder-rename-new", "http://localhost:8120"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_builder.test", "name", "tf-acc-builder-rename-new"),
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["komodo_builder.test"]
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

func testAccBuilderDisappears(resourceName string) resource.TestCheckFunc {
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
		return c.DeleteBuilder(context.Background(), rs.Primary.ID)
	}
}

func testAccBuilderResourceUrlConfig(name, address string) string {
	return fmt.Sprintf(`
resource "komodo_builder" "test" {
  name         = %q
  builder_type = "Url"
  url_config = {
    address = %q
  }
}
`, name, address)
}

func TestAccBuilderResource_tags(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBuilderWithTagConfig("tf-acc-builder-tags"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_builder.test", "tags.#", "1"),
					resource.TestCheckResourceAttrPair("komodo_builder.test", "tags.0", "komodo_tag.test", "id"),
				),
			},
			{
				Config: testAccBuilderClearTagsConfig("tf-acc-builder-tags"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_builder.test", "tags.#", "0"),
				),
			},
		},
	})
}

func testAccBuilderWithTagConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_tag" "test" {
  name  = "tf-acc-tag-builder"
  color = "Green"
}

resource "komodo_builder" "test" {
  name         = %q
  type         = "Url"
  url_config = {
    address = "http://localhost:8120"
  }
  tags = [komodo_tag.test.id]
}
`, name)
}

func testAccBuilderClearTagsConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_builder" "test" {
  name         = %q
  type         = "Url"
  url_config = {
    address = "http://localhost:8120"
  }
  tags = []
}
`, name)
}
