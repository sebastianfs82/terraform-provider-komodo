// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"context"
	"testing"
	datasource "github.com/hashicorp/terraform-plugin-framework/datasource"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDeploymentsDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentsDataSourceConfig("tf-acc-deployments-ds-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_deployments.all", "deployments.#"),
				),
			},
		},
	})
}

func TestAccDeploymentsDataSource_containsCreated(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentsDataSourceConfig("tf-acc-deployments-ds-find"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.komodo_deployments.all",
						"deployments.*",
						map[string]string{
							"name": "tf-acc-deployments-ds-find",
						},
					),
				),
			},
		},
	})
}

func TestAccDeploymentsDataSource_filteredByServerID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentsDataSourceConfig_filteredByServerID("tf-acc-deps-ds-server"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.komodo_deployments.filtered",
						"deployments.*",
						map[string]string{
							"name": "tf-acc-deps-ds-server",
						},
					),
				),
			},
		},
	})
}

func testAccDeploymentsDataSourceConfig(name string) string {
	return fmt.Sprintf(`
data "komodo_servers" "all" {}

resource "komodo_deployment" "test" {
  name      = %q
  server_id = data.komodo_servers.all.servers[0].id
  image {
    type = "Image"
    name = "nginx:latest"
  }
}

data "komodo_deployments" "all" {
  depends_on = [komodo_deployment.test]
}
`, name)
}

func testAccDeploymentsDataSourceConfig_filteredByServerID(name string) string {
	return fmt.Sprintf(`
data "komodo_servers" "all" {}

resource "komodo_deployment" "test" {
  name      = %q
  server_id = data.komodo_servers.all.servers[0].id
  image {
    type = "Image"
    name = "nginx:latest"
  }
}

data "komodo_deployments" "filtered" {
  server_id  = data.komodo_servers.all.servers[0].id
  depends_on = [komodo_deployment.test]
}
`, name)
}

func TestUnitDeploymentsDataSource_configure(t *testing.T) {
d := &DeploymentsDataSource{}
resp := &datasource.ConfigureResponse{}
d.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: "wrong"}, resp)
if !resp.Diagnostics.HasError() {
t.Fatal("expected diagnostic error for wrong provider data type")
}
}
