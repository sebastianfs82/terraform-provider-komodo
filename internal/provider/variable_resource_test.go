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
					resource.TestCheckResourceAttr("komodo_variable.test", "secret_enabled", "false"),
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
					resource.TestCheckResourceAttr("komodo_variable.test", "secret_enabled", "true"),
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
  secret_enabled = %t
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

// TestAccVariableResource_update verifies that updating value does not mark
// secret_enabled as unknown — it must remain at its known state value.
func TestAccVariableResource_update(t *testing.T) {
	const name = "tf_acc_variable_update"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVariableResourceConfig_basic(name, "value-one", "desc", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_variable.test", "value", "value-one"),
					resource.TestCheckResourceAttr("komodo_variable.test", "secret_enabled", "false"),
				),
			},
			{
				Config: testAccVariableResourceConfig_basic(name, "value-two", "desc", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_variable.test", "value", "value-two"),
					// secret_enabled must stay known (false) after updating an unrelated field
					resource.TestCheckResourceAttr("komodo_variable.test", "secret_enabled", "false"),
				),
			},
		},
	})
}

// TestAccVariableResource_descriptionDefault verifies that omitting description
// from config results in an empty string (not unknown after apply).
func TestAccVariableResource_descriptionDefault(t *testing.T) {
	const name = "tf_acc_variable_desc_default"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVariableResourceConfig_noDescription(name, "hello"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_variable.test", "name", name),
					// description must be "" when omitted, not unknown
					resource.TestCheckResourceAttr("komodo_variable.test", "description", ""),
				),
			},
		},
	})
}

// TestAccVariableResource_descriptionUpdate verifies that changing the description
// value is applied and reflected in state correctly.
func TestAccVariableResource_descriptionUpdate(t *testing.T) {
	const name = "tf_acc_variable_desc_update"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVariableResourceConfig_basic(name, "val", "initial description", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_variable.test", "description", "initial description"),
				),
			},
			{
				Config: testAccVariableResourceConfig_basic(name, "val", "updated description", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_variable.test", "description", "updated description"),
				),
			},
		},
	})
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

func testAccVariableResourceConfig_noDescription(name, value string) string {
	return fmt.Sprintf(`
resource "komodo_variable" "test" {
  name  = %q
  value = %q
}
`, name, value)
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
