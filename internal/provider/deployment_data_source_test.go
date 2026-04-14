// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDeploymentDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentDataSourceConfig_basic("tf-acc-deployment-ds-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_deployment.test", "name", "tf-acc-deployment-ds-basic"),
					resource.TestCheckResourceAttrSet("data.komodo_deployment.test", "id"),
				),
			},
		},
	})
}

func TestAccDeploymentDataSource_fields(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentDataSourceConfig_fields("tf-acc-deployment-ds-fields"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_deployment.test", "image.type", "Image"),
					resource.TestCheckResourceAttr("data.komodo_deployment.test", "image.name", "nginx:latest"),
				),
			},
		},
	})
}

func testAccDeploymentDataSourceConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "komodo_deployment" "test" {
  name = %q
}

data "komodo_deployment" "test" {
  id = komodo_deployment.test.id
}
`, name)
}

func testAccDeploymentDataSourceConfig_fields(name string) string {
	return fmt.Sprintf(`
resource "komodo_deployment" "test" {
  name = %q
  image {
    type = "Image"
    name = "nginx:latest"
  }
}

data "komodo_deployment" "test" {
  id = komodo_deployment.test.id
}
`, name)
}
