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

// ── Resource tests ────────────────────────────────────────────────────────────

func TestAccResourceSyncResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSyncResourceConfig("tf-acc-rsync-basic", "# managed by terraform"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_resource_sync.test", "name", "tf-acc-rsync-basic"),
					resource.TestCheckResourceAttr("komodo_resource_sync.test", "file_contents", "# managed by terraform"),
					resource.TestCheckResourceAttrSet("komodo_resource_sync.test", "id"),
				),
			},
		},
	})
}

func TestAccResourceSyncResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSyncResourceConfig("tf-acc-rsync-update", "# v1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_resource_sync.test", "file_contents", "# v1"),
				),
			},
			{
				Config: testAccResourceSyncResourceConfig("tf-acc-rsync-update", "# v2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_resource_sync.test", "file_contents", "# v2"),
				),
			},
		},
	})
}

func TestAccResourceSyncResource_import(t *testing.T) {
	var syncID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSyncResourceConfig("tf-acc-rsync-import", "# import test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_resource_sync.test", "id"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["komodo_resource_sync.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						syncID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config:            testAccResourceSyncResourceConfig("tf-acc-rsync-import", "# import test"),
				ResourceName:      "komodo_resource_sync.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(_ *terraform.State) (string, error) { return syncID, nil },
			},
		},
	})
}

func TestAccResourceSyncResource_disappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSyncResourceConfig("tf-acc-rsync-disappears", "# disappears"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_resource_sync.test", "id"),
					testAccResourceSyncDisappears("komodo_resource_sync.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccResourceSyncDisappears(resourceName string) resource.TestCheckFunc {
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
		return c.DeleteResourceSync(context.Background(), rs.Primary.ID)
	}
}

func testAccResourceSyncResourceConfig(name, fileContents string) string {
	return fmt.Sprintf(`
resource "komodo_resource_sync" "test" {
  name          = %q
  file_contents = %q
}
`, name, fileContents)
}

// ── Data source test ──────────────────────────────────────────────────────────

func TestAccResourceSyncDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSyncDataSourceConfig("tf-acc-rsync-ds", "# ds test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.komodo_resource_sync.test", "id",
						"komodo_resource_sync.test", "id",
					),
					resource.TestCheckResourceAttr("data.komodo_resource_sync.test", "name", "tf-acc-rsync-ds"),
					resource.TestCheckResourceAttr("data.komodo_resource_sync.test", "file_contents", "# ds test"),
				),
			},
		},
	})
}

func testAccResourceSyncDataSourceConfig(name, fileContents string) string {
	return fmt.Sprintf(`
resource "komodo_resource_sync" "test" {
  name          = %q
  file_contents = %q
}

data "komodo_resource_sync" "test" {
  id = komodo_resource_sync.test.id
}
`, name, fileContents)
}
