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

func TestAccUserResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserResourceConfig_basic("tf-user-basic", "Password1!"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user.test", "username", "tf-user-basic"),
					resource.TestCheckResourceAttrSet("komodo_user.test", "id"),
					resource.TestCheckResourceAttr("komodo_user.test", "enabled", "true"),
					resource.TestCheckResourceAttr("komodo_user.test", "admin_enabled", "false"),
					resource.TestCheckResourceAttr("komodo_user.test", "create_server_enabled", "false"),
					resource.TestCheckResourceAttr("komodo_user.test", "create_build_enabled", "false"),
				),
			},
		},
	})
}

func TestAccUserResource_withPermissions(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserResourceConfig_withPermissions("tf-user-perms", "Password1!", true, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user.test", "username", "tf-user-perms"),
					resource.TestCheckResourceAttr("komodo_user.test", "create_server_enabled", "true"),
					resource.TestCheckResourceAttr("komodo_user.test", "create_build_enabled", "true"),
				),
			},
			// Update: revoke permissions
			{
				Config: testAccUserResourceConfig_withPermissions("tf-user-perms", "Password1!", false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user.test", "create_server_enabled", "false"),
					resource.TestCheckResourceAttr("komodo_user.test", "create_build_enabled", "false"),
				),
			},
		},
	})
}

func TestAccUserResource_disableEnable(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserResourceConfig_withEnabled("tf-user-toggle", "Password1!", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user.test", "enabled", "true"),
				),
			},
			{
				Config: testAccUserResourceConfig_withEnabled("tf-user-toggle", "Password1!", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user.test", "enabled", "false"),
				),
			},
		},
	})
}

func TestAccUserResource_import(t *testing.T) {
	var userID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserResourceConfig_basic("tf-user-import", "Password1!"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user.test", "username", "tf-user-import"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["komodo_user.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						userID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config:            testAccUserResourceConfig_basic("tf-user-import", "Password1!"),
				ResourceName:      "komodo_user.test",
				ImportState:       true,
				ImportStateVerify: true,
				// password is not returned by the API on Read
				ImportStateVerifyIgnore: []string{"password"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return userID, nil
				},
			},
		},
	})
}

// Config helpers

func testAccUserResourceConfig_basic(username, password string) string {
	return fmt.Sprintf(`
resource "komodo_user" "test" {
  username = %[1]q
  password = %[2]q
}
`, username, password)
}

func testAccUserResourceConfig_withPermissions(username, password string, createServers, createBuilds bool) string {
	return fmt.Sprintf(`
resource "komodo_user" "test" {
  username       = %[1]q
  password       = %[2]q
  create_server_enabled = %[3]t
  create_build_enabled  = %[4]t
}
`, username, password, createServers, createBuilds)
}

func testAccUserResourceConfig_withEnabled(username, password string, enabled bool) string {
	return fmt.Sprintf(`
resource "komodo_user" "test" {
  username = %[1]q
  password = %[2]q
  enabled  = %[3]t
}
`, username, password, enabled)
}

func TestAccUserResource_adminEnabled(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with admin_enabled = true
			{
				Config: testAccUserResourceConfig_withAdmin("tf-user-admin", "Password1!", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user.test", "admin_enabled", "true"),
					resource.TestCheckResourceAttr("komodo_user.test", "enabled", "true"),
				),
			},
			// Update: revoke admin
			{
				Config: testAccUserResourceConfig_withAdmin("tf-user-admin", "Password1!", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user.test", "admin_enabled", "false"),
				),
			},
		},
	})
}

func TestAccUserResource_enabledDefault(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// enabled not set explicitly — should default to true
			{
				Config: testAccUserResourceConfig_basic("tf-user-enabled-default", "Password1!"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user.test", "enabled", "true"),
				),
			},
			// Explicitly disable
			{
				Config: testAccUserResourceConfig_withEnabled("tf-user-enabled-default", "Password1!", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user.test", "enabled", "false"),
				),
			},
			// Remove enabled from config — should plan a change back to true and apply it
			{
				Config: testAccUserResourceConfig_basic("tf-user-enabled-default", "Password1!"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user.test", "enabled", "true"),
				),
			},
		},
	})
}

func TestAccUserResource_adminConflictWithCreateServer(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccUserResourceConfig_adminWithCreateServer("tf-user-conflict-srv", "Password1!"),
				ExpectError: regexp.MustCompile(`create_server_enabled cannot be set to true alongside admin_enabled = true`),
			},
		},
	})
}

func TestAccUserResource_adminConflictWithCreateBuild(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccUserResourceConfig_adminWithCreateBuild("tf-user-conflict-bld", "Password1!"),
				ExpectError: regexp.MustCompile(`create_build_enabled cannot be set to true alongside admin_enabled = true`),
			},
		},
	})
}

func testAccUserResourceConfig_withAdmin(username, password string, admin bool) string {
	return fmt.Sprintf(`
resource "komodo_user" "test" {
  username      = %[1]q
  password      = %[2]q
  admin_enabled = %[3]t
}
`, username, password, admin)
}

func testAccUserResourceConfig_adminWithCreateServer(username, password string) string {
	return fmt.Sprintf(`
resource "komodo_user" "test" {
  username              = %[1]q
  password              = %[2]q
  admin_enabled         = true
  create_server_enabled = true
}
`, username, password)
}

func testAccUserResourceConfig_adminWithCreateBuild(username, password string) string {
	return fmt.Sprintf(`
resource "komodo_user" "test" {
  username             = %[1]q
  password             = %[2]q
  admin_enabled        = true
  create_build_enabled = true
}
`, username, password)
}

func TestAccUserResource_disappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserResourceConfig_basic("disappear-user", "Password123!"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_user.test", "id"),
					testAccUserDisappears("komodo_user.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccUserDisappears(resourceName string) resource.TestCheckFunc {
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
		return c.DeleteUser(context.Background(), rs.Primary.ID)
	}
}
