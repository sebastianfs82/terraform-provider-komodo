// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNetworkDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkDataSourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_network.test", "name"),
					resource.TestCheckResourceAttrSet("data.komodo_network.test", "server_id"),
				),
			},
		},
	})
}

func TestAccNetworkDataSource_fields(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkDataSourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_network.test", "network_id"),
					resource.TestCheckResourceAttrSet("data.komodo_network.test", "driver"),
					resource.TestCheckResourceAttrSet("data.komodo_network.test", "scope"),
					resource.TestCheckResourceAttrSet("data.komodo_network.test", "enable_ipv6"),
					resource.TestCheckResourceAttrSet("data.komodo_network.test", "internal"),
					resource.TestCheckResourceAttrSet("data.komodo_network.test", "attachable"),
					resource.TestCheckResourceAttrSet("data.komodo_network.test", "ingress"),
					resource.TestCheckResourceAttrSet("data.komodo_network.test", "in_use"),
				),
			},
		},
	})
}

// testAccNetworkDataSourceConfig_basic discovers the first server via komodo_servers,
// then reads its first listed network via komodo_networks and komodo_network.
func testAccNetworkDataSourceConfig_basic() string {
	return `
data "komodo_servers" "all" {}

data "komodo_networks" "all" {
  server_id  = data.komodo_servers.all.servers[0].name
  depends_on = [data.komodo_servers.all]
}

data "komodo_network" "test" {
  server_id  = data.komodo_servers.all.servers[0].name
  name       = data.komodo_networks.all.networks[0].name
  depends_on = [data.komodo_networks.all]
}
`
}
