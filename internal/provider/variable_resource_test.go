// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccVariableResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVariableResourceConfig_basic("MY_VAR", "my-value", "desc", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_variable.test", "name", "MY_VAR"),
					resource.TestCheckResourceAttr("komodo_variable.test", "value", "my-value"),
					resource.TestCheckResourceAttr("komodo_variable.test", "description", "desc"),
					resource.TestCheckResourceAttr("komodo_variable.test", "is_secret", "false"),
					resource.TestCheckResourceAttrSet("komodo_variable.test", "id"),
				),
			},
		},
	})
}

func TestAccVariableResource_secret(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVariableResourceConfig_basic("SECRET_VAR", "supersecret", "desc", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_variable.test", "is_secret", "true"),
				),
			},
		},
	})
}

func TestAccVariableResource_multiple(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVariableResourceConfig_multiple("VAR1", "VAR2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_variable.test1", "name", "VAR1"),
					resource.TestCheckResourceAttr("komodo_variable.test2", "name", "VAR2"),
				),
			},
		},
	})
}

func TestAccVariableResource_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVariableResourceConfig_basic("IMPORT_VAR", "import-value", "desc", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_variable.test", "name", "IMPORT_VAR"),
				),
			},
			{
				Config:            testAccVariableResourceConfig_basic("IMPORT_VAR", "import-value", "desc", false),
				ResourceName:      "komodo_variable.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "IMPORT_VAR",
			},
		},
	})
}

// Test configuration functions

func testAccVariableResourceConfig_basic(name, value, description string, isSecret bool) string {
	return fmt.Sprintf(`
resource "komodo_variable" "test" {
  name        = "%s"
  value       = "%s"
  description = "%s"
  is_secret   = %t
}
`, name, value, description, isSecret)
}

func testAccVariableResourceConfig_multiple(name1, name2 string) string {
	return fmt.Sprintf(`
resource "komodo_variable" "test1" {
  name = "%s"
}

resource "komodo_variable" "test2" {
  name = "%s"
}
`, name1, name2)
}
func TestAccVariableResource_disappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVariableResourceConfig_basic("DISAPPEAR_VAR", "val", "", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_variable.test", "id"),
					testAccVariableDisappears("komodo_variable.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccVariableDisappears(resourceName string) resource.TestCheckFunc {
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
		return c.DeleteVariable(context.Background(), client.DeleteVariableRequest{ID: rs.Primary.ID})
	}
}
