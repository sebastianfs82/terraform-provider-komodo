// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

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
