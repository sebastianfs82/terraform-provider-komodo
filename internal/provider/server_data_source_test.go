// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccServerDataSource_basic reads a server by name via komodo_servers list then passes it to komodo_server.
func TestAccServerDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServerDataSourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_server.test", "id"),
					resource.TestCheckResourceAttrSet("data.komodo_server.test", "name"),
				),
			},
		},
	})
}

func TestAccServerDataSource_fields(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServerDataSourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_server.test", "id"),
					resource.TestCheckResourceAttrSet("data.komodo_server.test", "address"),
					resource.TestCheckResourceAttrSet("data.komodo_server.test", "enabled"),
					resource.TestCheckResourceAttrSet("data.komodo_server.test", "auto_prune"),
					resource.TestCheckResourceAttrSet("data.komodo_server.test", "stats_monitoring"),
				),
			},
		},
	})
}

func TestAccServerDataSource_byID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServerDataSourceConfig_byID(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_server.by_id", "id"),
					resource.TestCheckResourceAttrSet("data.komodo_server.by_id", "name"),
				),
			},
		},
	})
}

func TestAccServerDataSource_viaResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServerDataSourceConfig_viaResource("tf-acc-server-ds-lookup"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_server.lookup", "id"),
					resource.TestCheckResourceAttr("data.komodo_server.lookup", "name", "tf-acc-server-ds-lookup"),
					resource.TestCheckResourceAttrSet("data.komodo_server.lookup", "enabled"),
					resource.TestCheckResourceAttrSet("data.komodo_server.lookup", "cpu_warning"),
					resource.TestCheckResourceAttrSet("data.komodo_server.lookup", "mem_critical"),
				),
			},
		},
	})
}

func TestAccServersDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServersDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_servers.all", "servers.#"),
				),
			},
		},
	})
}

func TestAccServersDataSource_hasFields(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServersDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_servers.all", "servers.#"),
					resource.TestCheckResourceAttrSet("data.komodo_servers.all", "servers.0.id"),
					resource.TestCheckResourceAttrSet("data.komodo_servers.all", "servers.0.name"),
					resource.TestCheckResourceAttrSet("data.komodo_servers.all", "servers.0.enabled"),
				),
			},
		},
	})
}

func testAccServerDataSourceConfig_basic() string {
	return `
data "komodo_servers" "all" {}

data "komodo_server" "test" {
  name       = data.komodo_servers.all.servers[0].name
  depends_on = [data.komodo_servers.all]
}
`
}

func testAccServerDataSourceConfig_byID() string {
	return `
data "komodo_servers" "all" {}

data "komodo_server" "by_id" {
  id         = data.komodo_servers.all.servers[0].id
  depends_on = [data.komodo_servers.all]
}
`
}

func testAccServerDataSourceConfig_viaResource(name string) string {
	return fmt.Sprintf(`
resource "komodo_server" "src" {
  name = %q
}

data "komodo_server" "lookup" {
  name       = komodo_server.src.name
  depends_on = [komodo_server.src]
}
`, name)
}

func testAccServersDataSourceConfig() string {
	return `
data "komodo_servers" "all" {}
`
}
