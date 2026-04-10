// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccServiceUserDataSource_byUsername(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceUserDataSourceConfig_byUsername("tf-svc-ds-name"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_service_user.test", "username", "tf-svc-ds-name"),
					resource.TestCheckResourceAttrSet("data.komodo_service_user.test", "id"),
					resource.TestCheckResourceAttr("data.komodo_service_user.test", "enabled", "true"),
					resource.TestCheckResourceAttr("data.komodo_service_user.test", "admin", "false"),
					resource.TestCheckResourceAttr("data.komodo_service_user.test", "create_servers", "false"),
					resource.TestCheckResourceAttr("data.komodo_service_user.test", "create_builds", "false"),
				),
			},
		},
	})
}

func TestAccServiceUserDataSource_byID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceUserDataSourceConfig_byID("tf-svc-ds-id"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_service_user.test", "username", "tf-svc-ds-id"),
					resource.TestCheckResourceAttrPair(
						"data.komodo_service_user.test", "id",
						"komodo_service_user.test", "id",
					),
				),
			},
		},
	})
}

func TestAccServiceUserDataSource_withPermissions(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceUserDataSourceConfig_withPermissions("tf-svc-ds-perms", true, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_service_user.test", "username", "tf-svc-ds-perms"),
					resource.TestCheckResourceAttr("data.komodo_service_user.test", "create_servers", "true"),
					resource.TestCheckResourceAttr("data.komodo_service_user.test", "create_builds", "true"),
				),
			},
		},
	})
}

// Config helpers

func testAccServiceUserDataSourceConfig_byUsername(username string) string {
	return fmt.Sprintf(`
resource "komodo_service_user" "test" {
  username = %[1]q
}

data "komodo_service_user" "test" {
  username = komodo_service_user.test.username
}
`, username)
}

func testAccServiceUserDataSourceConfig_byID(username string) string {
	return fmt.Sprintf(`
resource "komodo_service_user" "test" {
  username = %[1]q
}

data "komodo_service_user" "test" {
  id = komodo_service_user.test.id
}
`, username)
}

func testAccServiceUserDataSourceConfig_withPermissions(username string, createServers, createBuilds bool) string {
	return fmt.Sprintf(`
resource "komodo_service_user" "test" {
  username       = %[1]q
  create_servers = %[2]t
  create_builds  = %[3]t
}

data "komodo_service_user" "test" {
  username = komodo_service_user.test.username
}
`, username, createServers, createBuilds)
}
