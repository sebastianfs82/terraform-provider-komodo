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

func TestAccServiceUsersDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceUsersDataSourceConfig("tf-acc-svcusers-ds-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_service_users.all", "service_users.#"),
				),
			},
		},
	})
}

func TestAccServiceUsersDataSource_containsCreated(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceUsersDataSourceConfig("tf-acc-svcusers-ds-find"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.komodo_service_users.all",
						"service_users.*",
						map[string]string{
							"username": "tf-acc-svcusers-ds-find",
						},
					),
				),
			},
		},
	})
}

func testAccServiceUsersDataSourceConfig(username string) string {
	return fmt.Sprintf(`
resource "komodo_service_user" "test" {
  username = %q
}

data "komodo_service_users" "all" {
  depends_on = [komodo_service_user.test]
}
`, username)
}

func TestUnitServiceUsersDataSource_configure(t *testing.T) {
d := &ServiceUsersDataSource{}
resp := &datasource.ConfigureResponse{}
d.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: "wrong"}, resp)
if !resp.Diagnostics.HasError() {
t.Fatal("expected diagnostic error for wrong provider data type")
}
}
