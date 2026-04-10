// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserGroupDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupDataSourceConfig_basic("tf-test-ds-group"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_user_group.test", "name", "tf-test-ds-group"),
					resource.TestCheckResourceAttr("data.komodo_user_group.test", "everyone", "false"),
					resource.TestCheckResourceAttrSet("data.komodo_user_group.test", "id"),
					resource.TestCheckResourceAttrSet("data.komodo_user_group.test", "updated_at"),
					resource.TestCheckResourceAttr("data.komodo_user_group.test", "users.#", "0"),
				),
			},
		},
	})
}

func TestAccUserGroupDataSource_withUsers(t *testing.T) {
	userID := os.Getenv("KOMODO_TEST_USER_ID")
	if userID == "" {
		t.Skip("KOMODO_TEST_USER_ID must be set to run user membership tests")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupDataSourceConfig_withUser("tf-test-ds-group-users", userID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_user_group.test", "name", "tf-test-ds-group-users"),
					resource.TestCheckResourceAttr("data.komodo_user_group.test", "users.#", "1"),
					resource.TestCheckResourceAttr("data.komodo_user_group.test", "users.0", userID),
				),
			},
		},
	})
}

func testAccUserGroupDataSourceConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "komodo_user_group" "test" {
  name = %q
}

data "komodo_user_group" "test" {
  name = komodo_user_group.test.name
}
`, name)
}

func testAccUserGroupDataSourceConfig_withUser(name, userID string) string {
	return fmt.Sprintf(`
resource "komodo_user_group" "test" {
  name  = %q
  users = [%q]
}

data "komodo_user_group" "test" {
  name = komodo_user_group.test.name
}
`, name, userID)
}
