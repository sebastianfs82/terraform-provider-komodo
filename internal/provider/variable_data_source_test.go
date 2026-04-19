// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	datasource "github.com/hashicorp/terraform-plugin-framework/datasource"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccVariableDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVariableDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_variable.example", "name", "tf_var_ds"),
					resource.TestCheckResourceAttrSet("data.komodo_variable.example", "value"),
				),
			},
		},
	})
}

const testAccVariableDataSourceConfig = `
resource "komodo_variable" "example" {
  name        = "tf_var_ds"
  value       = "ds-value"
  description = "Data source test variable"
  secret_enabled = false
}

data "komodo_variable" "example" {
  name = komodo_variable.example.name
}
`

func TestAccVariableDataSource_fields(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVariableDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_variable.example", "name", "tf_var_ds"),
					resource.TestCheckResourceAttr("data.komodo_variable.example", "value", "ds-value"),
					resource.TestCheckResourceAttr("data.komodo_variable.example", "description", "Data source test variable"),
					resource.TestCheckResourceAttr("data.komodo_variable.example", "secret_enabled", "false"),
				),
			},
		},
	})
}

func TestAccVariableDataSource_secret(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "komodo_variable" "secret" {
  name      = "tf_var_ds_secret"
  value     = "topsecret"
  secret_enabled = true
}

data "komodo_variable" "secret" {
  name = komodo_variable.secret.name
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_variable.secret", "name", "tf_var_ds_secret"),
					resource.TestCheckResourceAttr("data.komodo_variable.secret", "secret_enabled", "true"),
				),
			},
		},
	})
}

func TestUnitVariableDataSource_configure(t *testing.T) {
	d := &VariableDataSource{}
	resp := &datasource.ConfigureResponse{}
	d.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: "wrong"}, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for wrong provider data type")
	}
}

func TestAccVariableDataSource_bothSet_isError(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccVariableDataSourceConfig_bothSet(),
				ExpectError: regexp.MustCompile(`Only one of`),
			},
		},
	})
}

func TestAccVariableDataSource_neitherSet_isError(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccVariableDataSourceConfig_neitherSet(),
				ExpectError: regexp.MustCompile(`One of`),
			},
		},
	})
}

func testAccVariableDataSourceConfig_bothSet() string {
	return `
data "komodo_variable" "test" {
  id   = "some-id"
  name = "some-name"
}
`
}

func testAccVariableDataSourceConfig_neitherSet() string {
	return `
data "komodo_variable" "test" {}
`
}
