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

func TestAccDeploymentResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentResourceConfig("tf-acc-deployment-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_deployment.test", "name", "tf-acc-deployment-basic"),
					resource.TestCheckResourceAttrSet("komodo_deployment.test", "id"),
				),
			},
		},
	})
}

func TestAccDeploymentResource_update(t *testing.T) {
	const name = "tf-acc-deployment-update"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentResourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_deployment.test", "name", name),
				),
			},
			{
				Config: testAccDeploymentResourceConfigWithImage(name, "nginx:latest"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_deployment.test", "name", name),
					resource.TestCheckResourceAttr("komodo_deployment.test", "image.type", "Image"),
					resource.TestCheckResourceAttr("komodo_deployment.test", "image.image", "nginx:latest"),
				),
			},
		},
	})
}

func TestAccDeploymentResource_importState(t *testing.T) {
	var deploymentID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentResourceConfig("tf-acc-deployment-import"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_deployment.test", "id"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["komodo_deployment.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						deploymentID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config:            testAccDeploymentResourceConfig("tf-acc-deployment-import"),
				ResourceName:      "komodo_deployment.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return deploymentID, nil
				},
			},
		},
	})
}

func TestAccDeploymentResource_disappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentResourceConfig("tf-acc-deployment-disappears"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_deployment.test", "id"),
					testAccDeploymentDisappears("komodo_deployment.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDeploymentResource_rename(t *testing.T) {
	var savedID string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentResourceConfig("tf-acc-deployment-rename-orig"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_deployment.test", "name", "tf-acc-deployment-rename-orig"),
					resource.TestCheckResourceAttrSet("komodo_deployment.test", "id"),
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["komodo_deployment.test"]
						savedID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config: testAccDeploymentResourceConfig("tf-acc-deployment-rename-new"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_deployment.test", "name", "tf-acc-deployment-rename-new"),
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["komodo_deployment.test"]
						if rs.Primary.ID != savedID {
							return fmt.Errorf("resource was recreated: ID changed from %q to %q", savedID, rs.Primary.ID)
						}
						return nil
					},
				),
			},
		},
	})
}

func testAccDeploymentDisappears(resourceName string) resource.TestCheckFunc {
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
		return c.DeleteDeployment(context.Background(), rs.Primary.ID)
	}
}

func testAccDeploymentResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_deployment" "test" {
  name = %q
}
`, name)
}

func testAccDeploymentResourceConfigWithImage(name, image string) string {
	return fmt.Sprintf(`
resource "komodo_deployment" "test" {
  name = %q
  image = {
    type  = "Image"
    image = %q
  }
}
`, name, image)
}

func TestAccDeploymentResource_tags(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentWithTagConfig("tf-acc-deployment-tags"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_deployment.test", "tags.#", "1"),
					resource.TestCheckResourceAttrPair("komodo_deployment.test", "tags.0", "komodo_tag.test", "id"),
				),
			},
			{
				Config: testAccDeploymentClearTagsConfig("tf-acc-deployment-tags"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_deployment.test", "tags.#", "0"),
				),
			},
		},
	})
}

func testAccDeploymentWithTagConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_tag" "test" {
  name  = "tf-acc-tag-deployment"
  color = "Green"
}

resource "komodo_deployment" "test" {
  name = %q
  tags = [komodo_tag.test.id]
}
`, name)
}

func testAccDeploymentClearTagsConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_deployment" "test" {
  name = %q
  tags = []
}
`, name)
}
