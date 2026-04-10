// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceSyncsDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSyncsDataSourceConfig("tf-acc-rsyncs-ds-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_resource_syncs.all", "resource_syncs.#"),
				),
			},
		},
	})
}

func TestAccResourceSyncsDataSource_containsCreated(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSyncsDataSourceConfig("tf-acc-rsyncs-ds-find"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.komodo_resource_syncs.all",
						"resource_syncs.*",
						map[string]string{
							"name": "tf-acc-rsyncs-ds-find",
						},
					),
				),
			},
		},
	})
}

func TestAccResourceSyncsDataSource_filteredByRepoID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSyncsDataSourceConfig_filteredByRepoID("tf-acc-rsyncs-ds-repo"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.komodo_resource_syncs.filtered",
						"resource_syncs.*",
						map[string]string{
							"name": "tf-acc-rsyncs-ds-repo",
						},
					),
				),
			},
		},
	})
}

func testAccResourceSyncsDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_resource_sync" "test" {
  name          = %q
  file_contents = "# list ds test"
}

data "komodo_resource_syncs" "all" {
  depends_on = [komodo_resource_sync.test]
}
`, name)
}

func testAccResourceSyncsDataSourceConfig_filteredByRepoID(name string) string {
	return fmt.Sprintf(`
resource "komodo_repo" "test" {
  name = %q
}

resource "komodo_resource_sync" "test" {
  name        = %q
  linked_repo = komodo_repo.test.name
}

data "komodo_resource_syncs" "filtered" {
  repo_id    = komodo_repo.test.name
  depends_on = [komodo_resource_sync.test]
}
`, name, name)
}
