// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"context"
	"regexp"
	"testing"
	datasource "github.com/hashicorp/terraform-plugin-framework/datasource"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccActionDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccActionDataSourceConfig_basic("tf-acc-action-ds-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_action.test", "name", "tf-acc-action-ds-basic"),
					resource.TestCheckResourceAttrSet("data.komodo_action.test", "id"),
				),
			},
		},
	})
}

func TestAccActionDataSource_fields(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccActionDataSourceConfig_fields("tf-acc-action-ds-fields"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_action.test", "webhook.enabled", "true"),
					resource.TestCheckResourceAttr("data.komodo_action.test", "failure_alert_enabled", "true"),
					resource.TestCheckResourceAttr("data.komodo_action.test", "file_contents", "console.log('test');"),
				),
			},
		},
	})
}

func testAccActionDataSourceConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "komodo_action" "test" {
  name = %q
}

data "komodo_action" "test" {
  id = komodo_action.test.id
}
`, name)
}

func testAccActionDataSourceConfig_fields(name string) string {
	return fmt.Sprintf(`
resource "komodo_action" "test" {
  name            = %q
  webhook {
    enabled = true
  }
  failure_alert_enabled   = true
  file_contents   = "console.log('test');"
}

data "komodo_action" "test" {
  id = komodo_action.test.id
}
`, name)
}

func TestUnitActionDataSource_configure(t *testing.T) {
d := &ActionDataSource{}
resp := &datasource.ConfigureResponse{}
d.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: "wrong"}, resp)
if !resp.Diagnostics.HasError() {
t.Fatal("expected diagnostic error for wrong provider data type")
}
}

func TestAccActionDataSource_bothSet_isError(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccActionDataSourceConfig_bothSet(),
				ExpectError: regexp.MustCompile(`Only one of`),
			},
		},
	})
}

func TestAccActionDataSource_neitherSet_isError(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccActionDataSourceConfig_neitherSet(),
				ExpectError: regexp.MustCompile(`One of`),
			},
		},
	})
}

func testAccActionDataSourceConfig_bothSet() string {
	return `
data "komodo_action" "test" {
  id   = "some-id"
  name = "some-name"
}
`
}

func testAccActionDataSourceConfig_neitherSet() string {
	return `
data "komodo_action" "test" {}
`
}
