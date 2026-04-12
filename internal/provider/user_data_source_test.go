// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserDataSource_byUsername(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserDataSourceConfig_byUsername("tf-user-ds-name", "Password1!"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_user.test", "username", "tf-user-ds-name"),
					resource.TestCheckResourceAttrSet("data.komodo_user.test", "id"),
					resource.TestCheckResourceAttr("data.komodo_user.test", "enabled", "true"),
					resource.TestCheckResourceAttr("data.komodo_user.test", "admin_enabled", "false"),
					resource.TestCheckResourceAttr("data.komodo_user.test", "create_server_enabled", "false"),
					resource.TestCheckResourceAttr("data.komodo_user.test", "create_build_enabled", "false"),
				),
			},
		},
	})
}

func TestAccUserDataSource_byID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserDataSourceConfig_byID("tf-user-ds-id", "Password1!"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_user.test", "username", "tf-user-ds-id"),
					resource.TestCheckResourceAttrSet("data.komodo_user.test", "id"),
					resource.TestCheckResourceAttrPair(
						"data.komodo_user.test", "id",
						"komodo_user.test", "id",
					),
				),
			},
		},
	})
}

func TestAccUserDataSource_withPermissions(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserDataSourceConfig_withPermissions("tf-user-ds-perms", "Password1!"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_user.test", "username", "tf-user-ds-perms"),
					resource.TestCheckResourceAttr("data.komodo_user.test", "create_server_enabled", "true"),
					resource.TestCheckResourceAttr("data.komodo_user.test", "create_build_enabled", "true"),
				),
			},
		},
	})
}

// Config helpers

func testAccUserDataSourceConfig_byUsername(username, password string) string {
	return fmt.Sprintf(`
resource "komodo_user" "test" {
  username = %[1]q
  password = %[2]q
}

data "komodo_user" "test" {
  username = komodo_user.test.username
}
`, username, password)
}

func testAccUserDataSourceConfig_byID(username, password string) string {
	return fmt.Sprintf(`
resource "komodo_user" "test" {
  username = %[1]q
  password = %[2]q
}

data "komodo_user" "test" {
  id = komodo_user.test.id
}
`, username, password)
}

func testAccUserDataSourceConfig_withPermissions(username, password string) string {
	return fmt.Sprintf(`
resource "komodo_user" "test" {
  username       = %[1]q
  password       = %[2]q
  create_server_enabled = true
  create_build_enabled  = true
}

data "komodo_user" "test" {
  username = komodo_user.test.username
}
`, username, password)
}
