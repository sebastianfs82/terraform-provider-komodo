package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccVariableDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVariableDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_variable.example", "name", "tf_var_ds"),
					resource.TestCheckResourceAttrSet("data.komodo_variable.example", "value"),
				),
			},
		},
	})
}

const testAccVariableDataSourceConfig = `
resource "komodo_variable" "example" {
  name        = "tf_var_ds"
  value       = "ds-value"
  description = "Data source test variable"
  is_secret   = false
}

data "komodo_variable" "example" {
  name = komodo_variable.example.name
}
`

func TestAccVariableDataSource_fields(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVariableDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_variable.example", "name", "tf_var_ds"),
					resource.TestCheckResourceAttr("data.komodo_variable.example", "value", "ds-value"),
					resource.TestCheckResourceAttr("data.komodo_variable.example", "description", "Data source test variable"),
					resource.TestCheckResourceAttr("data.komodo_variable.example", "is_secret", "false"),
				),
			},
		},
	})
}

func TestAccVariableDataSource_secret(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "komodo_variable" "secret" {
  name      = "tf_var_ds_secret"
  value     = "topsecret"
  is_secret = true
}

data "komodo_variable" "secret" {
  name = komodo_variable.secret.name
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_variable.secret", "name", "tf_var_ds_secret"),
					resource.TestCheckResourceAttr("data.komodo_variable.secret", "is_secret", "true"),
				),
			},
		},
	})
}
