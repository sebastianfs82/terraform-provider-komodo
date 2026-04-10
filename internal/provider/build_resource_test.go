// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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

func TestAccBuildResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBuildResourceConfig("tf-acc-build-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_build.test", "name", "tf-acc-build-basic"),
					resource.TestCheckResourceAttrSet("komodo_build.test", "id"),
				),
			},
		},
	})
}

func TestAccBuildResource_update(t *testing.T) {
	const name = "tf-acc-build-update"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBuildResourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_build.test", "name", name),
				),
			},
			{
				Config: testAccBuildResourceConfigWithRepo(name, "myorg/myrepo", "main"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_build.test", "repo", "myorg/myrepo"),
					resource.TestCheckResourceAttr("komodo_build.test", "branch", "main"),
				),
			},
		},
	})
}

func TestAccBuildResource_importState(t *testing.T) {
	var buildID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBuildResourceConfig("tf-acc-build-import"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_build.test", "id"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["komodo_build.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						buildID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config:            testAccBuildResourceConfig("tf-acc-build-import"),
				ResourceName:      "komodo_build.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return buildID, nil
				},
				ImportStateVerifyIgnore: []string{"webhook_secret", "secret_args"},
			},
		},
	})
}

func TestAccBuildResource_disappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBuildResourceConfig("tf-acc-build-disappears"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_build.test", "id"),
					testAccBuildDisappears("komodo_build.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccBuildDisappears(resourceName string) resource.TestCheckFunc {
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
		return c.DeleteBuild(context.Background(), rs.Primary.ID)
	}
}

func testAccBuildResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_build" "test" {
  name = %q
}
`, name)
}

func testAccBuildResourceConfigWithRepo(name, repo, branch string) string {
	return fmt.Sprintf(`
resource "komodo_build" "test" {
  name   = %q
  repo   = %q
  branch = %q
}
`, name, repo, branch)
}
