// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	datasource "github.com/hashicorp/terraform-plugin-framework/datasource"
	"regexp"
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
					resource.TestCheckResourceAttrSet("data.komodo_server.test", "enabled"),
					resource.TestCheckResourceAttrSet("data.komodo_server.test", "auto_prune_enabled"),
					resource.TestCheckResourceAttrSet("data.komodo_server.test", "alerts.enabled"),
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
					resource.TestCheckResourceAttrSet("data.komodo_server.lookup", "alerts.thresholds.cpu_warning"),
					resource.TestCheckResourceAttrSet("data.komodo_server.lookup", "alerts.thresholds.memory_critical"),
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

// TestAccServerDataSource_bothSet_isError verifies that supplying both id and name is a validation error.
func TestAccServerDataSource_bothSet_isError(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccServerDataSourceConfig_bothSet(),
				ExpectError: regexp.MustCompile(`Only one of`),
			},
		},
	})
}

// TestAccServerDataSource_neitherSet_isError verifies that omitting both id and name is a validation error.
func TestAccServerDataSource_neitherSet_isError(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccServerDataSourceConfig_neitherSet(),
				ExpectError: regexp.MustCompile(`One of`),
			},
		},
	})
}

// TestAccServerDataSource_exactValues creates a resource with explicit thresholds and region,
// then reads it back via the data source and asserts those exact values are returned.
func TestAccServerDataSource_exactValues(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServerDataSourceConfig_exactValues("tf-acc-server-ds-exact", "eu-west", 72.5, 90.0),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_server.exact", "name", "tf-acc-server-ds-exact"),
					resource.TestCheckResourceAttr("data.komodo_server.exact", "region", "eu-west"),
					resource.TestCheckResourceAttr("data.komodo_server.exact", "alerts.thresholds.cpu_warning", "72.5"),
					resource.TestCheckResourceAttr("data.komodo_server.exact", "alerts.thresholds.cpu_critical", "90"),
				),
			},
		},
	})
}

// TestAccServersDataSource_alertFields verifies that the servers list exposes all renamed and nested alert fields.
func TestAccServersDataSource_alertFields(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServersDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_servers.all", "servers.#"),
					resource.TestCheckResourceAttrSet("data.komodo_servers.all", "servers.0.tls_ignored"),
					resource.TestCheckResourceAttrSet("data.komodo_servers.all", "servers.0.auto_prune_enabled"),
					resource.TestCheckResourceAttrSet("data.komodo_servers.all", "servers.0.auto_rotate_keys_enabled"),
					resource.TestCheckResourceAttrSet("data.komodo_servers.all", "servers.0.alerts.enabled"),
					resource.TestCheckResourceAttrSet("data.komodo_servers.all", "servers.0.alerts.thresholds.cpu_warning"),
					resource.TestCheckResourceAttrSet("data.komodo_servers.all", "servers.0.alerts.thresholds.memory_critical"),
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

func testAccServerDataSourceConfig_bothSet() string {
	return `
data "komodo_server" "both" {
  id   = "000000000000000000000000"
  name = "some-server"
}
`
}

func testAccServerDataSourceConfig_neitherSet() string {
	return `
data "komodo_server" "neither" {}
`
}

func testAccServerDataSourceConfig_exactValues(name, region string, cpuWarn, cpuCrit float64) string {
	return fmt.Sprintf(`
resource "komodo_server" "exact" {
  name   = %q
  region = %q
  alerts {
    thresholds {
      cpu_warning  = %g
      cpu_critical = %g
    }
  }
}

data "komodo_server" "exact" {
  name       = komodo_server.exact.name
  depends_on = [komodo_server.exact]
}
`, name, region, cpuWarn, cpuCrit)
}

func TestUnitServerDataSource_configure(t *testing.T) {
	d := &ServerDataSource{}
	resp := &datasource.ConfigureResponse{}
	d.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: "wrong"}, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for wrong provider data type")
	}
}
