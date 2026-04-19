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

func TestAccProceduresDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProceduresDataSourceConfig("tf-acc-procedures-ds-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_procedures.all", "procedures.#"),
				),
			},
		},
	})
}

func TestAccProceduresDataSource_containsCreated(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProceduresDataSourceConfig("tf-acc-procedures-ds-find"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.komodo_procedures.all",
						"procedures.*",
						map[string]string{
							"name": "tf-acc-procedures-ds-find",
						},
					),
				),
			},
		},
	})
}

func testAccProceduresDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_procedure" "test" {
  name = %q
}

data "komodo_procedures" "all" {
  depends_on = [komodo_procedure.test]
}
`, name)
}

func TestUnitProceduresDataSource_configure(t *testing.T) {
d := &ProceduresDataSource{}
resp := &datasource.ConfigureResponse{}
d.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: "wrong"}, resp)
if !resp.Diagnostics.HasError() {
t.Fatal("expected diagnostic error for wrong provider data type")
}
}
