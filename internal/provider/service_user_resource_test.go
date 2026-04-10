package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"terraform-provider-komodo/internal/client"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccServiceUserResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceUserResourceConfig_basic("tf-svc-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_service_user.test", "username", "tf-svc-basic"),
					resource.TestCheckResourceAttrSet("komodo_service_user.test", "id"),
					resource.TestCheckResourceAttr("komodo_service_user.test", "enabled", "true"),
					resource.TestCheckResourceAttr("komodo_service_user.test", "admin", "false"),
					resource.TestCheckResourceAttr("komodo_service_user.test", "create_servers", "false"),
					resource.TestCheckResourceAttr("komodo_service_user.test", "create_builds", "false"),
				),
			},
		},
	})
}

func TestAccServiceUserResource_withDescription(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceUserResourceConfig_withDescription("tf-svc-desc", "initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_service_user.test", "username", "tf-svc-desc"),
					resource.TestCheckResourceAttr("komodo_service_user.test", "description", "initial description"),
				),
			},
			// Update description
			{
				Config: testAccServiceUserResourceConfig_withDescription("tf-svc-desc", "updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_service_user.test", "description", "updated description"),
				),
			},
		},
	})
}

func TestAccServiceUserResource_withPermissions(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceUserResourceConfig_withPermissions("tf-svc-perms", true, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_service_user.test", "username", "tf-svc-perms"),
					resource.TestCheckResourceAttr("komodo_service_user.test", "create_servers", "true"),
					resource.TestCheckResourceAttr("komodo_service_user.test", "create_builds", "true"),
				),
			},
			// Update permissions
			{
				Config: testAccServiceUserResourceConfig_withPermissions("tf-svc-perms", false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_service_user.test", "create_servers", "false"),
					resource.TestCheckResourceAttr("komodo_service_user.test", "create_builds", "false"),
				),
			},
		},
	})
}

func TestAccServiceUserResource_import(t *testing.T) {
	var userID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceUserResourceConfig_basic("tf-svc-import"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_service_user.test", "username", "tf-svc-import"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["komodo_service_user.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						userID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config:            testAccServiceUserResourceConfig_basic("tf-svc-import"),
				ResourceName:      "komodo_service_user.test",
				ImportState:       true,
				ImportStateVerify: true,
				// description is not returned by the API on Read
				ImportStateVerifyIgnore: []string{"description"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return userID, nil
				},
			},
		},
	})
}

func TestAccServiceUserResource_withApiKey(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceUserResourceConfig_withApiKey("tf-svc-apikey", "svc-user-key"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_service_user.test", "username", "tf-svc-apikey"),
					resource.TestCheckResourceAttr("komodo_api_key.svc_key", "name", "svc-user-key"),
					resource.TestCheckResourceAttrSet("komodo_api_key.svc_key", "key"),
					resource.TestCheckResourceAttrSet("komodo_api_key.svc_key", "secret"),
					resource.TestCheckResourceAttr("komodo_api_key.svc_key", "expires", "0"),
					resource.TestCheckResourceAttrPair(
						"komodo_api_key.svc_key", "service_user_id",
						"komodo_service_user.test", "id",
					),
				),
			},
		},
	})
}

// Config helpers

func testAccServiceUserResourceConfig_basic(username string) string {
	return fmt.Sprintf(`
resource "komodo_service_user" "test" {
  username = %[1]q
}
`, username)
}

func testAccServiceUserResourceConfig_withDescription(username, description string) string {
	return fmt.Sprintf(`
resource "komodo_service_user" "test" {
  username    = %[1]q
  description = %[2]q
}
`, username, description)
}

func testAccServiceUserResourceConfig_withPermissions(username string, createServers, createBuilds bool) string {
	return fmt.Sprintf(`
resource "komodo_service_user" "test" {
  username       = %[1]q
  create_servers = %[2]t
  create_builds  = %[3]t
}
`, username, createServers, createBuilds)
}

func testAccServiceUserResourceConfig_withApiKey(username, keyName string) string {
	return fmt.Sprintf(`
resource "komodo_service_user" "test" {
  username = %[1]q
}

resource "komodo_api_key" "svc_key" {
  name            = %[2]q
  service_user_id = komodo_service_user.test.id
}
`, username, keyName)
}

func TestAccServiceUserResource_disappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceUserResourceConfig_basic("disappear-svc-user"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_service_user.test", "id"),
					testAccServiceUserDisappears("komodo_service_user.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccServiceUserDisappears(resourceName string) resource.TestCheckFunc {
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
