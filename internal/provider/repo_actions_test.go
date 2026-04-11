// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// ─── Build ───────────────────────────────────────────────────────────────────

func TestAccRepoBuildAction_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRepoBuildActionConfig("tf-test-repo-build"),
			},
		},
	})
}

func testAccRepoBuildActionConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_repo" "test" {
  name = %q
}

action "komodo_repo_build" "test" {
  config {
    id = komodo_repo.test.name
  }
}
`, name)
}

// ─── Clone ───────────────────────────────────────────────────────────────────

func TestAccRepoCloneAction_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRepoCloneActionConfig("tf-test-repo-clone"),
			},
		},
	})
}

func testAccRepoCloneActionConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_repo" "test" {
  name = %q
}

action "komodo_repo_clone" "test" {
  config {
    id = komodo_repo.test.name
  }
}
`, name)
}

// ─── Pull ────────────────────────────────────────────────────────────────────

func TestAccRepoPullAction_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRepoPullActionConfig("tf-test-repo-pull"),
			},
		},
	})
}

func testAccRepoPullActionConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_repo" "test" {
  name = %q
}

action "komodo_repo_pull" "test" {
  config {
    id = komodo_repo.test.name
  }
}
`, name)
}
