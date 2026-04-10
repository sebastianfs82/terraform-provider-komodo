package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTagDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTagDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_tag.example", "name"),
					resource.TestCheckResourceAttrSet("data.komodo_tag.example", "color"),
				),
			},
		},
	})
}

const testAccTagDataSourceConfig = `
resource "komodo_tag" "example" {
  name  = "tf_tag_ds"
  color = "Blue"
}

data "komodo_tag" "example" {
  name = komodo_tag.example.name
}
`

func TestAccTagDataSource_fields(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTagDataSourceConfig_fields,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_tag.example", "name", "tf-acc-tag-ds-fields"),
					resource.TestCheckResourceAttr("data.komodo_tag.example", "color", "Purple"),
					resource.TestCheckResourceAttrSet("data.komodo_tag.example", "owner"),
				),
			},
		},
	})
}

func TestAccTagDataSource_byID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTagDataSourceConfig_byID,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_tag.byid", "id"),
					resource.TestCheckResourceAttrSet("data.komodo_tag.byid", "name"),
					resource.TestCheckResourceAttrSet("data.komodo_tag.byid", "color"),
				),
			},
		},
	})
}

const testAccTagDataSourceConfig_fields = `
resource "komodo_tag" "src" {
  name  = "tf-acc-tag-ds-fields"
  color = "Purple"
}

data "komodo_tag" "example" {
  name = komodo_tag.src.name
}
`

const testAccTagDataSourceConfig_byID = `
resource "komodo_tag" "src" {
  name  = "tf-acc-tag-ds-byid"
  color = "Red"
}

data "komodo_tag" "byid" {
  id = komodo_tag.src.id
}
`
