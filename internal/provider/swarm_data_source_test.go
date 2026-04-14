// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSwarmDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSwarmDataSourceConfig_viaResource("tf-acc-swarm-ds-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_swarm.lookup", "id"),
					resource.TestCheckResourceAttr("data.komodo_swarm.lookup", "name", "tf-acc-swarm-ds-basic"),
					resource.TestCheckResourceAttrSet("data.komodo_swarm.lookup", "alerts_enabled"),
					resource.TestCheckResourceAttr("data.komodo_swarm.lookup", "server_ids.#", "0"),
					resource.TestCheckResourceAttr("data.komodo_swarm.lookup", "links.#", "0"),
				),
			},
		},
	})
}

func TestAccSwarmDataSource_byName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSwarmDataSourceConfig_fromList(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_swarm.test", "id"),
					resource.TestCheckResourceAttrSet("data.komodo_swarm.test", "name"),
				),
			},
		},
	})
}

func TestAccSwarmDataSource_byID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSwarmDataSourceConfig_byID("tf-acc-swarm-ds-byid"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_swarm.by_id", "id"),
					resource.TestCheckResourceAttr("data.komodo_swarm.by_id", "name", "tf-acc-swarm-ds-byid"),
				),
			},
		},
	})
}

func TestAccSwarmDataSource_bothSet_isError(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSwarmDataSourceConfig_bothSet(),
				ExpectError: regexp.MustCompile(`Only one of`),
			},
		},
	})
}

func TestAccSwarmDataSource_neitherSet_isError(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSwarmDataSourceConfig_neitherSet(),
				ExpectError: regexp.MustCompile(`One of`),
			},
		},
	})
}

// --- Config helpers ---

func testAccSwarmDataSourceConfig_viaResource(name string) string {
	return fmt.Sprintf(`
resource "komodo_swarm" "src" {
  name = %q
}

data "komodo_swarm" "lookup" {
  name       = komodo_swarm.src.name
  depends_on = [komodo_swarm.src]
}
`, name)
}

func testAccSwarmDataSourceConfig_fromList() string {
	return `
resource "komodo_swarm" "src" {
  name = "tf-acc-swarm-ds-byname"
}

data "komodo_swarms" "all" {
  depends_on = [komodo_swarm.src]
}

data "komodo_swarm" "test" {
  name       = komodo_swarm.src.name
  depends_on = [data.komodo_swarms.all]
}
`
}

func testAccSwarmDataSourceConfig_byID(name string) string {
	return fmt.Sprintf(`
resource "komodo_swarm" "src" {
  name = %q
}

data "komodo_swarm" "by_id" {
  id         = komodo_swarm.src.id
  depends_on = [komodo_swarm.src]
}
`, name)
}

func testAccSwarmDataSourceConfig_bothSet() string {
	return `
data "komodo_swarm" "test" {
  id   = "some-id"
  name = "some-name"
}
`
}

func testAccSwarmDataSourceConfig_neitherSet() string {
	return `
data "komodo_swarm" "test" {}
`
}
