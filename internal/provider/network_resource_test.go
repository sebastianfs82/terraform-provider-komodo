// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// testAccNetworkServerID returns the server ID/name to use for network acceptance tests.
// Set KOMODO_TEST_SERVER_ID to the name or ID of an available server in your Komodo instance.
func testAccNetworkServerID(t *testing.T) string {
	t.Helper()
	v := os.Getenv("KOMODO_TEST_SERVER_ID")
	if v == "" {
		t.Skip("KOMODO_TEST_SERVER_ID must be set for network acceptance tests")
	}
	return v
}

// --- Resource tests ---

func TestAccNetworkResource_basic(t *testing.T) {
	serverID := testAccNetworkServerID(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkResourceConfig(serverID, "tf-test-network"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_network.test", "server_id", serverID),
					resource.TestCheckResourceAttr("komodo_network.test", "name", "tf-test-network"),
					resource.TestCheckResourceAttrSet("komodo_network.test", "id"),
				),
			},
			// Verify state is stable on a second plan/apply (no-diff).
			{
				Config:   testAccNetworkResourceConfig(serverID, "tf-test-network"),
				PlanOnly: true,
			},
		},
	})
}

func TestAccNetworkResource_importState(t *testing.T) {
	serverID := testAccNetworkServerID(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkResourceConfig(serverID, "tf-import-network"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_network.test", "id"),
				),
			},
			{
				ResourceName:      "komodo_network.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNetworkResourceConfig(serverID, name string) string {
	return fmt.Sprintf(`
resource "komodo_network" "test" {
  server_id = %q
  name      = %q
}
`, serverID, name)
}
