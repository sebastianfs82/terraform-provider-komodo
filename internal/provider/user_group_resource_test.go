// Copyright (c) HashiCorp, Inc.
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

func TestAccUserGroupResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupResourceConfig_basic("tf-test-group"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user_group.test", "name", "tf-test-group"),
					resource.TestCheckResourceAttr("komodo_user_group.test", "everyone", "false"),
					resource.TestCheckResourceAttrSet("komodo_user_group.test", "id"),
					resource.TestCheckResourceAttrSet("komodo_user_group.test", "updated_at"),
				),
			},
		},
	})
}

func TestAccUserGroupResource_rename(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupResourceConfig_basic("tf-test-group-original"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user_group.test", "name", "tf-test-group-original"),
				),
			},
			{
				Config: testAccUserGroupResourceConfig_basic("tf-test-group-renamed"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user_group.test", "name", "tf-test-group-renamed"),
					resource.TestCheckResourceAttrSet("komodo_user_group.test", "id"),
				),
			},
		},
	})
}

func TestAccUserGroupResource_withUsers(t *testing.T) {
	userID := os.Getenv("KOMODO_TEST_USER_ID")
	if userID == "" {
		t.Skip("KOMODO_TEST_USER_ID must be set to run user membership tests")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupResourceConfig_withUser("tf-test-group-users", userID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user_group.test", "name", "tf-test-group-users"),
					resource.TestCheckResourceAttr("komodo_user_group.test", "users.#", "1"),
					resource.TestCheckResourceAttr("komodo_user_group.test", "users.0", userID),
				),
			},
		},
	})
}

func TestAccUserGroupResource_addRemoveUser(t *testing.T) {
	userID := os.Getenv("KOMODO_TEST_USER_ID")
	if userID == "" {
		t.Skip("KOMODO_TEST_USER_ID must be set to run user membership tests")
	}
	var groupID string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupResourceConfig_withUser("tf-test-group-add-remove", userID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user_group.test", "users.#", "1"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["komodo_user_group.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						groupID = rs.Primary.ID
						return nil
					},
				),
			},
			// Remove the user
			{
				Config: testAccUserGroupResourceConfig_basic("tf-test-group-add-remove"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("komodo_user_group.test", "users"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["komodo_user_group.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						if rs.Primary.ID != groupID {
							return fmt.Errorf("expected same group ID after user removal, got %s", rs.Primary.ID)
						}
						return nil
					},
				),
			},
		},
	})
}

func TestAccUserGroupResource_import(t *testing.T) {
	var groupID string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupResourceConfig_basic("tf-test-group-import"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_user_group.test", "id"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["komodo_user_group.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						groupID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config:            testAccUserGroupResourceConfig_basic("tf-test-group-import"),
				ResourceName:      "komodo_user_group.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(*terraform.State) (string, error) {
					return groupID, nil
				},
			},
		},
	})
}

func TestAccUserGroupResource_everyoneConflictsWithUsers(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccUserGroupResourceConfig_everyoneAndUsers("tf-test-group-invalid"),
				ExpectError: regexp.MustCompile("Conflicting configuration"),
			},
		},
	})
}

// Config helpers

func testAccUserGroupResourceConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "komodo_user_group" "test" {
  name = %q
}
`, name)
}

func testAccUserGroupResourceConfig_withUser(name, userID string) string {
	return fmt.Sprintf(`
resource "komodo_user_group" "test" {
  name  = %q
  users = [%q]
}
`, name, userID)
}

func testAccUserGroupResourceConfig_everyoneAndUsers(name string) string {
	return fmt.Sprintf(`
resource "komodo_user_group" "test" {
  name     = %q
  everyone = true
  users    = ["some-user-id"]
}
`, name)
}

func TestAccUserGroupResource_disappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupResourceConfig_basic("tf-disappear-group"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_user_group.test", "id"),
					testAccUserGroupDisappears("komodo_user_group.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccUserGroupDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found in state: %s", resourceName)
		}
		c := client.NewClient(
			os.Getenv("KOMODO_ENDPOINT"),
			os.Getenv("KOMODO_USERNAME"),
			os.Getenv("KOMODO_PASSWORD"),
		)
		return c.DeleteUserGroup(context.Background(), rs.Primary.ID)
	}
}
