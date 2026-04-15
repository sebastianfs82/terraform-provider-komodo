// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// testAccServerActionServerID returns the ID of the first available server in the Komodo instance.
// Falls back to KOMODO_TEST_SERVER_ID if set. Skips the test if no servers are found.
func testAccServerActionServerID(t *testing.T) string {
	t.Helper()
	return testAccLookupServerID(t, "server action acceptance tests")
}

// ─── PruneBuildx ─────────────────────────────────────────────────────────────

func TestAccServerPruneBuildxAction_basic(t *testing.T) {
	serverID := testAccServerActionServerID(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServerPruneBuildxActionConfig(serverID),
			},
		},
	})
}

func testAccServerPruneBuildxActionConfig(serverID string) string {
	return fmt.Sprintf(`
action "komodo_server_prune_buildx" "test" {
  config {
    id = %q
  }
}
`, serverID)
}

// ─── PruneContainers ─────────────────────────────────────────────────────────

func TestAccServerPruneContainersAction_basic(t *testing.T) {
	serverID := testAccServerActionServerID(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServerPruneContainersActionConfig(serverID),
			},
		},
	})
}

func testAccServerPruneContainersActionConfig(serverID string) string {
	return fmt.Sprintf(`
action "komodo_server_prune_containers" "test" {
  config {
    id = %q
  }
}
`, serverID)
}

// ─── PruneBuilders ───────────────────────────────────────────────────────────

func TestAccServerPruneBuildersAction_basic(t *testing.T) {
	serverID := testAccServerActionServerID(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServerPruneBuildersActionConfig(serverID),
			},
		},
	})
}

func testAccServerPruneBuildersActionConfig(serverID string) string {
	return fmt.Sprintf(`
action "komodo_server_prune_builders" "test" {
  config {
    id = %q
  }
}
`, serverID)
}

// ─── PruneImages ─────────────────────────────────────────────────────────────

func TestAccServerPruneImagesAction_basic(t *testing.T) {
	serverID := testAccServerActionServerID(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServerPruneImagesActionConfig(serverID),
			},
		},
	})
}

func testAccServerPruneImagesActionConfig(serverID string) string {
	return fmt.Sprintf(`
action "komodo_server_prune_images" "test" {
  config {
    id = %q
  }
}
`, serverID)
}

// ─── PruneNetworks ───────────────────────────────────────────────────────────

func TestAccServerPruneNetworksAction_basic(t *testing.T) {
	serverID := testAccServerActionServerID(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServerPruneNetworksActionConfig(serverID),
			},
		},
	})
}

func testAccServerPruneNetworksActionConfig(serverID string) string {
	return fmt.Sprintf(`
action "komodo_server_prune_networks" "test" {
  config {
    id = %q
  }
}
`, serverID)
}

// ─── PruneSystem ─────────────────────────────────────────────────────────────

func TestAccServerPruneSystemAction_basic(t *testing.T) {
	serverID := testAccServerActionServerID(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServerPruneSystemActionConfig(serverID),
			},
		},
	})
}

func testAccServerPruneSystemActionConfig(serverID string) string {
	return fmt.Sprintf(`
action "komodo_server_prune_system" "test" {
  config {
    id = %q
  }
}
`, serverID)
}

// ─── PruneVolumes ────────────────────────────────────────────────────────────

func TestAccServerPruneVolumesAction_basic(t *testing.T) {
	serverID := testAccServerActionServerID(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServerPruneVolumesActionConfig(serverID),
			},
		},
	})
}

func testAccServerPruneVolumesActionConfig(serverID string) string {
	return fmt.Sprintf(`
action "komodo_server_prune_volumes" "test" {
  config {
    id = %q
  }
}
`, serverID)
}
