// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccReposDataSource_filteredByServerID(t *testing.T) {
	serverID := os.Getenv("KOMODO_TEST_SERVER_ID")
	if serverID == "" {
		t.Skip("KOMODO_TEST_SERVER_ID must be set to run server_id filter tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccReposDataSourceConfig_filteredByServerID("tf-acc-repos-ds-server", serverID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.komodo_repos.filtered",
						"repositories.*",
						map[string]string{
							"name": "tf-acc-repos-ds-server",
						},
					),
				),
			},
		},
	})
}

func TestAccReposDataSource_filteredByBuilderID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccReposDataSourceConfig_filteredByBuilderID("tf-acc-repos-ds-bld"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.komodo_repos.filtered",
						"repositories.*",
						map[string]string{
							"name": "tf-acc-repos-ds-bld",
						},
					),
				),
			},
		},
	})
}

func testAccReposDataSourceConfig_filteredByServerID(name, serverID string) string {
	return fmt.Sprintf(`
resource "komodo_repo" "test" {
  name      = %q
  server_id = %q
}

data "komodo_repos" "filtered" {
  server_id  = %q
  depends_on = [komodo_repo.test]
}
`, name, serverID, serverID)
}

func testAccReposDataSourceConfig_filteredByBuilderID(name string) string {
	return fmt.Sprintf(`
resource "komodo_builder" "test" {
  name = %q
  type = "Url"
  url_config {
    address = "https://builder.example.com"
  }
}

resource "komodo_repo" "test" {
  name       = %q
  builder_id = komodo_builder.test.id
}

data "komodo_repos" "filtered" {
  builder_id = komodo_builder.test.id
  depends_on = [komodo_repo.test]
}
`, name, name)
}
