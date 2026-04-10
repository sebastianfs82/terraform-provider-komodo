// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccBuildsDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBuildsDataSourceConfig("tf-acc-builds-ds-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_builds.all", "builds.#"),
				),
			},
		},
	})
}

func TestAccBuildsDataSource_containsCreated(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBuildsDataSourceConfig("tf-acc-builds-ds-find"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.komodo_builds.all",
						"builds.*",
						map[string]string{
							"name": "tf-acc-builds-ds-find",
						},
					),
				),
			},
		},
	})
}

func TestAccBuildsDataSource_filteredByBuilderID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBuildsDataSourceConfig_filteredByBuilderID("tf-acc-builds-ds-bld"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.komodo_builds.filtered",
						"builds.*",
						map[string]string{
							"name": "tf-acc-builds-ds-bld",
						},
					),
				),
			},
		},
	})
}

func TestAccBuildsDataSource_filteredByRepoID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBuildsDataSourceConfig_filteredByRepoID("tf-acc-builds-ds-repo"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.komodo_builds.filtered",
						"builds.*",
						map[string]string{
							"name": "tf-acc-builds-ds-repo",
						},
					),
				),
			},
		},
	})
}

func testAccBuildsDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_build" "test" {
  name       = %q
  image_name = "test-image"
}

data "komodo_builds" "all" {
  depends_on = [komodo_build.test]
}
`, name)
}

func testAccBuildsDataSourceConfig_filteredByBuilderID(name string) string {
	return fmt.Sprintf(`
resource "komodo_builder" "test" {
  name         = %q
  builder_type = "Url"
  url_config = {
    address = "https://builder.example.com"
  }
}

resource "komodo_build" "test" {
  name       = %q
  builder_id = komodo_builder.test.id
}

data "komodo_builds" "filtered" {
  builder_id = komodo_builder.test.id
  depends_on = [komodo_build.test]
}
`, name, name)
}

func testAccBuildsDataSourceConfig_filteredByRepoID(name string) string {
	return fmt.Sprintf(`
resource "komodo_repo" "test" {
  name = %q
}

resource "komodo_build" "test" {
  name        = %q
  linked_repo = komodo_repo.test.name
}

data "komodo_builds" "filtered" {
  repo_id    = komodo_repo.test.name
  depends_on = [komodo_build.test]
}
`, name, name)
}
