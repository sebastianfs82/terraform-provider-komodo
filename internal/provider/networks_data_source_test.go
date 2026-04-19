// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"
	datasource "github.com/hashicorp/terraform-plugin-framework/datasource"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNetworksDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworksDataSourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_networks.all", "networks.#"),
				),
			},
		},
	})
}

func TestAccNetworksDataSource_hasFields(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworksDataSourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_networks.all", "networks.#"),
					resource.TestCheckResourceAttrSet("data.komodo_networks.all", "networks.0.name"),
					resource.TestCheckResourceAttrSet("data.komodo_networks.all", "networks.0.network_id"),
					resource.TestCheckResourceAttrSet("data.komodo_networks.all", "networks.0.driver"),
					resource.TestCheckResourceAttrSet("data.komodo_networks.all", "networks.0.scope"),
				),
			},
		},
	})
}

func testAccNetworksDataSourceConfig_basic() string {
	return `
data "komodo_servers" "all" {}

data "komodo_networks" "all" {
  server_id  = data.komodo_servers.all.servers[0].name
  depends_on = [data.komodo_servers.all]
}
`
}

func TestUnitNetworksDataSource_configure(t *testing.T) {
d := &NetworksDataSource{}
resp := &datasource.ConfigureResponse{}
d.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: "wrong"}, resp)
if !resp.Diagnostics.HasError() {
t.Fatal("expected diagnostic error for wrong provider data type")
}
}
