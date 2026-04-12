// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccUserGroupMembershipResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupMembershipResourceConfig_basic("tf-test-membership-group", "tf-test-membership-svc"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user_group_membership.test", "user_group_id", "tf-test-membership-group"),
					resource.TestCheckResourceAttr("komodo_user_group_membership.test", "user_id", "tf-test-membership-svc"),
					resource.TestCheckResourceAttr("komodo_user_group_membership.test", "id", "tf-test-membership-group/tf-test-membership-svc"),
				),
			},
		},
	})
}

func TestAccUserGroupMembershipResource_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupMembershipResourceConfig_basic("tf-test-membership-import-group", "tf-test-membership-import-svc"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_user_group_membership.test", "id"),
				),
			},
			{
				Config:            testAccUserGroupMembershipResourceConfig_basic("tf-test-membership-import-group", "tf-test-membership-import-svc"),
				ResourceName:      "komodo_user_group_membership.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "tf-test-membership-import-group/tf-test-membership-import-svc",
			},
		},
	})
}

func TestAccUserGroupMembershipResource_disappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupMembershipResourceConfig_basic("tf-test-membership-disappears-group", "tf-test-membership-disappears-svc"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user_group_membership.test", "user_group_id", "tf-test-membership-disappears-group"),
					testAccUserGroupMembershipDisappears("komodo_user_group_membership.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccUserGroupMembershipDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		userGroup := rs.Primary.Attributes["user_group_id"]
		user := rs.Primary.Attributes["user_id"]

		c := client.NewClient(
			os.Getenv("KOMODO_ENDPOINT"),
			os.Getenv("KOMODO_USERNAME"),
			os.Getenv("KOMODO_PASSWORD"),
		)

		_, err := c.RemoveUserFromUserGroup(context.Background(), client.RemoveUserFromUserGroupRequest{
			UserGroup: userGroup,
			User:      user,
		})
		return err
	}
}

// Config helpers

func testAccUserGroupMembershipResourceConfig_basic(groupName, svcUserName string) string {
	return fmt.Sprintf(`
resource "komodo_user_group" "test" {
  name = %q
}

resource "komodo_service_user" "test" {
  username    = %q
  description = "test service user for membership"
}

resource "komodo_user_group_membership" "test" {
  user_group_id = komodo_user_group.test.name
  user_id       = komodo_service_user.test.username

  depends_on = [komodo_user_group.test, komodo_service_user.test]
}
`, groupName, svcUserName)
}

func TestAccUserGroupMembershipResource_everyoneEnabledBlocked(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccUserGroupMembershipResourceConfig_everyoneEnabled("tf-test-membership-everyone-group", "tf-test-membership-everyone-svc"),
				ExpectError: regexp.MustCompile(`everyone_enabled.*is true`),
			},
		},
	})
}

func testAccUserGroupMembershipResourceConfig_everyoneEnabled(groupName, svcUserName string) string {
	return fmt.Sprintf(`
resource "komodo_user_group" "test" {
  name             = %q
  everyone_enabled = true
}

resource "komodo_service_user" "test" {
  username    = %q
  description = "test service user for membership everyone check"
}

resource "komodo_user_group_membership" "test" {
  user_group_id = komodo_user_group.test.name
  user_id       = komodo_service_user.test.username

  depends_on = [komodo_user_group.test, komodo_service_user.test]
}
`, groupName, svcUserName)
}
