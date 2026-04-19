// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"testing"

	datasource "github.com/hashicorp/terraform-plugin-framework/datasource"
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
  name = %q

  source {
    contents = "# list ds test"
  }
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
  name = %q
  source {
    repo_id = komodo_repo.test.id
  }
}

data "komodo_resource_syncs" "filtered" {
  repo_id    = komodo_repo.test.id
  depends_on = [komodo_resource_sync.test]
}
`, name, name)
}

func TestUnitResourceSyncsDataSource_configure(t *testing.T) {
	d := &ResourceSyncsDataSource{}
	resp := &datasource.ConfigureResponse{}
	d.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: "wrong"}, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for wrong provider data type")
	}
}
