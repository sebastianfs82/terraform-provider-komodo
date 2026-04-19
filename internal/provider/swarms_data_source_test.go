// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"
	datasource "github.com/hashicorp/terraform-plugin-framework/datasource"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSwarmsDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSwarmsDataSourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_swarms.all", "swarms.#"),
				),
			},
		},
	})
}

func TestAccSwarmsDataSource_hasFields(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSwarmsDataSourceConfig_withSwarm("tf-acc-swarms-ds-fields"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_swarms.all", "swarms.#"),
					resource.TestCheckResourceAttrSet("data.komodo_swarms.all", "swarms.0.id"),
					resource.TestCheckResourceAttrSet("data.komodo_swarms.all", "swarms.0.name"),
					resource.TestCheckResourceAttrSet("data.komodo_swarms.all", "swarms.0.alerts_enabled"),
					resource.TestCheckResourceAttrSet("data.komodo_swarms.all", "swarms.0.server_ids.#"),
					resource.TestCheckResourceAttrSet("data.komodo_swarms.all", "swarms.0.links.#"),
				),
			},
		},
	})
}

// --- Config helpers ---

func testAccSwarmsDataSourceConfig_basic() string {
	return `
data "komodo_swarms" "all" {}
`
}

func testAccSwarmsDataSourceConfig_withSwarm(name string) string {
	return testAccSwarmResourceConfig(name) + `
data "komodo_swarms" "all" {
  depends_on = [komodo_swarm.test]
}
`
}

func TestUnitSwarmsDataSource_configure(t *testing.T) {
d := &SwarmsDataSource{}
resp := &datasource.ConfigureResponse{}
d.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: "wrong"}, resp)
if !resp.Diagnostics.HasError() {
t.Fatal("expected diagnostic error for wrong provider data type")
}
}
