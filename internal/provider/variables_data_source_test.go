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

func TestAccVariablesDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVariablesDataSourceConfig("tf_acc_vars_ds_basic", "list-test-value"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_variables.all", "variables.#"),
				),
			},
		},
	})
}

func TestAccVariablesDataSource_containsCreated(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVariablesDataSourceConfig("tf_acc_vars_ds_find", "find-value"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.komodo_variables.all",
						"variables.*",
						map[string]string{
							"name":  "tf_acc_vars_ds_find",
							"value": "find-value",
						},
					),
				),
			},
		},
	})
}

func testAccVariablesDataSourceConfig(name, value string) string {
	return fmt.Sprintf(`
resource "komodo_variable" "test" {
  name  = %q
  value = %q
}

data "komodo_variables" "all" {
  depends_on = [komodo_variable.test]
}
`, name, value)
}

func TestUnitVariablesDataSource_configure(t *testing.T) {
d := &VariablesDataSource{}
resp := &datasource.ConfigureResponse{}
d.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: "wrong"}, resp)
if !resp.Diagnostics.HasError() {
t.Fatal("expected diagnostic error for wrong provider data type")
}
}
