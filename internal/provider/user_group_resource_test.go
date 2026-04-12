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

func TestAccUserGroupResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupResourceConfig_basic("tf-test-group"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user_group.test", "name", "tf-test-group"),
					resource.TestCheckResourceAttr("komodo_user_group.test", "everyone_enabled", "false"),
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

// TestAccUserGroupResource_unmanagedUsersNoDrift verifies that when users is
// not configured, externally-added users do not cause a non-empty plan (no drift).
func TestAccUserGroupResource_unmanagedUsersNoDrift(t *testing.T) {
	userID := os.Getenv("KOMODO_TEST_USER_ID")
	if userID == "" {
		t.Skip("KOMODO_TEST_USER_ID must be set to run user membership tests")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create the group without specifying users
				Config: testAccUserGroupResourceConfig_basic("tf-test-group-unmanaged"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_user_group.test", "id"),
					// Add a user out-of-band via the API
					testAccUserGroupAddUser("komodo_user_group.test", userID),
				),
				// No diff should be produced despite the external user addition
				ExpectNonEmptyPlan: false,
			},
			{
				// Re-apply the same config — must produce an empty plan
				Config:   testAccUserGroupResourceConfig_basic("tf-test-group-unmanaged"),
				PlanOnly: true,
			},
		},
	})
}

// TestAccUserGroupResource_unmanagedUsersNotRemoved verifies that switching from
// a managed users list back to no users config removes the previously-managed users
// once, then stops tracking the list (future manual additions are not touched).
func TestAccUserGroupResource_unmanagedUsersNotRemoved(t *testing.T) {
	userID := os.Getenv("KOMODO_TEST_USER_ID")
	if userID == "" {
		t.Skip("KOMODO_TEST_USER_ID must be set to run user membership tests")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Start with managed user list
				Config: testAccUserGroupResourceConfig_withUser("tf-test-group-unmanage-transition", userID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user_group.test", "users.#", "1"),
				),
			},
			{
				// Remove users from config — Terraform should remove the previously-managed
				// user once, then stop managing the list.
				Config: testAccUserGroupResourceConfig_basic("tf-test-group-unmanage-transition"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// users attribute is now null in state (unmanaged)
					resource.TestCheckNoResourceAttr("komodo_user_group.test", "users"),
					// The previously-managed user has been removed from the group
					testAccUserGroupNotHasMember("komodo_user_group.test", userID),
				),
			},
			{
				// Add the user back out-of-band — should produce no plan diff (truly unmanaged now)
				Config: testAccUserGroupResourceConfig_basic("tf-test-group-unmanage-transition"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccUserGroupAddUser("komodo_user_group.test", userID),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				// Re-apply the same config — must produce an empty plan (out-of-band user ignored)
				Config:   testAccUserGroupResourceConfig_basic("tf-test-group-unmanage-transition"),
				PlanOnly: true,
			},
		},
	})
}

// TestAccUserGroupResource_managedUsersFullControl verifies that when users is
// specified, Terraform enforces the exact list and removes unlisted members.
func TestAccUserGroupResource_managedUsersFullControl(t *testing.T) {
	userID := os.Getenv("KOMODO_TEST_USER_ID")
	if userID == "" {
		t.Skip("KOMODO_TEST_USER_ID must be set to run user membership tests")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Start with managed user
				Config: testAccUserGroupResourceConfig_withUser("tf-test-group-full-control", userID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user_group.test", "users.#", "1"),
				),
			},
			{
				// Manage with empty list — user must be removed
				Config: testAccUserGroupResourceConfig_emptyUsers("tf-test-group-full-control"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user_group.test", "users.#", "0"),
					testAccUserGroupNotHasMember("komodo_user_group.test", userID),
				),
			},
		},
	})
}

func testAccUserGroupAddUser(resourceName, userID string) resource.TestCheckFunc {
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
		_, err := c.AddUserToUserGroup(context.Background(), client.AddUserToUserGroupRequest{
			UserGroup: rs.Primary.ID,
			User:      userID,
		})
		return err
	}
}

func testAccUserGroupNotHasMember(resourceName, userID string) resource.TestCheckFunc {
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
		group, err := c.GetUserGroup(context.Background(), rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("unable to fetch group: %s", err)
		}
		for _, u := range group.Users {
			if u == userID {
				return fmt.Errorf("expected user %s to NOT be a member of group %s, but was", userID, rs.Primary.ID)
			}
		}
		return nil
	}
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

func testAccUserGroupResourceConfig_emptyUsers(name string) string {
	return fmt.Sprintf(`
resource "komodo_user_group" "test" {
  name  = %q
  users = []
}
`, name)
}

func testAccUserGroupResourceConfig_everyoneEnabled(name string, enabled bool) string {
	return fmt.Sprintf(`
resource "komodo_user_group" "test" {
  name             = %q
  everyone_enabled = %t
}
`, name, enabled)
}

func testAccUserGroupResourceConfig_everyoneAndUsers(name string) string {
	return fmt.Sprintf(`
resource "komodo_user_group" "test" {
  name     = %q
  everyone_enabled = true
  users    = ["some-user-id"]
}
`, name)
}

// TestAccUserGroupResource_everyoneEnabledDefault verifies that omitting
// everyone_enabled results in false in state (not unknown after apply).
func TestAccUserGroupResource_everyoneEnabledDefault(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupResourceConfig_basic("tf-test-group-everyone-default"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user_group.test", "everyone_enabled", "false"),
				),
			},
		},
	})
}

// TestAccUserGroupResource_everyoneEnabledUpdate verifies that toggling
// everyone_enabled is applied and reflected correctly in state.
func TestAccUserGroupResource_everyoneEnabledUpdate(t *testing.T) {
	const name = "tf-test-group-everyone-update"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupResourceConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user_group.test", "everyone_enabled", "false"),
				),
			},
			{
				Config: testAccUserGroupResourceConfig_everyoneEnabled(name, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user_group.test", "everyone_enabled", "true"),
				),
			},
			{
				Config: testAccUserGroupResourceConfig_everyoneEnabled(name, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user_group.test", "everyone_enabled", "false"),
				),
			},
		},
	})
}

// TestAccUserGroupResource_everyoneEnabledDrift verifies that an external change
// to everyone_enabled is detected as drift (non-empty plan).
func TestAccUserGroupResource_everyoneEnabledDrift(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupResourceConfig_basic("tf-test-group-everyone-drift"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user_group.test", "everyone_enabled", "false"),
					// Simulate external change: set everyone_enabled = true out-of-band
					testAccUserGroupSetEveryoneEnabled("komodo_user_group.test", true),
				),
				// After the external change the plan must show a diff
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccUserGroupSetEveryoneEnabled(resourceName string, enabled bool) resource.TestCheckFunc {
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
		_, err := c.SetEveryoneUserGroup(context.Background(), client.SetEveryoneUserGroupRequest{
			UserGroup: rs.Primary.ID,
			Everyone:  enabled,
		})
		return err
	}
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
