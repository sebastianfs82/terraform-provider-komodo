// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// testAccServerActionServerID returns the server ID/name to use for server action acceptance tests.
// Set KOMODO_TEST_SERVER_ID to the name or ID of an available server in your Komodo instance.
func testAccServerActionServerID(t *testing.T) string {
	t.Helper()
	v := os.Getenv("KOMODO_TEST_SERVER_ID")
	if v == "" {
		t.Skip("KOMODO_TEST_SERVER_ID must be set for server action acceptance tests")
	}
	return v
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
