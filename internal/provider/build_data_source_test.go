// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// ---------------------------------------------------------------------------
// Basic lookups
// ---------------------------------------------------------------------------

func TestAccBuildDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBuildDataSourceConfig_basic("tf-acc-build-ds-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_build.test", "name", "tf-acc-build-ds-basic"),
					resource.TestCheckResourceAttrSet("data.komodo_build.test", "id"),
				),
			},
		},
	})
}

func TestAccBuildDataSource_byName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBuildDataSourceConfig_byName("tf-acc-build-ds-byname"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_build.test", "name", "tf-acc-build-ds-byname"),
					resource.TestCheckResourceAttrSet("data.komodo_build.test", "id"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// source block reflected in data source
// ---------------------------------------------------------------------------

func TestAccBuildDataSource_sourceBlock(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBuildDataSourceConfig_source("tf-acc-build-ds-source", "myorg/myrepo", "main"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_build.test", "source.path", "myorg/myrepo"),
					resource.TestCheckResourceAttr("data.komodo_build.test", "source.branch", "main"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// image block reflected in data source
// ---------------------------------------------------------------------------

func TestAccBuildDataSource_imageBlock(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBuildDataSourceConfig_image("tf-acc-build-ds-image"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_build.test", "image.name", "myorg/my-service"),
					resource.TestCheckResourceAttr("data.komodo_build.test", "image.include_latest_tag_enabled", "true"),
					resource.TestCheckResourceAttr("data.komodo_build.test", "image.include_version_tags_enabled", "false"),
					resource.TestCheckResourceAttr("data.komodo_build.test", "image.include_commit_tag_enabled", "true"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// image.registry blocks reflected in data source
// ---------------------------------------------------------------------------

func TestAccBuildDataSource_imageRegistry(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBuildDataSourceConfig_registry("tf-acc-build-ds-registry"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_build.test", "image.registry.#", "1"),
					resource.TestCheckResourceAttrPair(
						"data.komodo_build.test", "image.registry.0.account_id",
						"komodo_registry_account.test", "id",
					),
					resource.TestCheckResourceAttr("data.komodo_build.test", "image.registry.0.organization", "myorg"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// version block reflected in data source
// ---------------------------------------------------------------------------

func TestAccBuildDataSource_versionBlock(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBuildDataSourceConfig_version("tf-acc-build-ds-version"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_build.test", "version.value"),
					resource.TestCheckResourceAttrSet("data.komodo_build.test", "version.auto_increment_enabled"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// Config helpers
// ---------------------------------------------------------------------------

func testAccBuildDataSourceConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "komodo_build" "test" {
  name = %q
}

data "komodo_build" "test" {
  id = komodo_build.test.id
}
`, name)
}

func testAccBuildDataSourceConfig_byName(name string) string {
	return fmt.Sprintf(`
resource "komodo_build" "test" {
  name = %q
}

data "komodo_build" "test" {
  name       = komodo_build.test.name
  depends_on = [komodo_build.test]
}
`, name)
}

func testAccBuildDataSourceConfig_source(name, path, branch string) string {
	return fmt.Sprintf(`
resource "komodo_build" "test" {
  name = %q
  source {
    path   = %q
    branch = %q
  }
}

data "komodo_build" "test" {
  id = komodo_build.test.id
}
`, name, path, branch)
}

func testAccBuildDataSourceConfig_image(name string) string {
	return fmt.Sprintf(`
resource "komodo_build" "test" {
  name = %q
  image {
    name                         = "myorg/my-service"
    include_latest_tag_enabled   = true
    include_version_tags_enabled = false
    include_commit_tag_enabled   = true
  }
}

data "komodo_build" "test" {
  id = komodo_build.test.id
}
`, name)
}

func testAccBuildDataSourceConfig_registry(name string) string {
	return fmt.Sprintf(`
resource "komodo_registry_account" "test" {
  domain   = "registry.example.com"
  username = "tf-acc-ds-reg-user"
  token    = "ds-reg-token"
}

resource "komodo_build" "test" {
  name = %q
  image {
    registry {
      account_id   = komodo_registry_account.test.id
      organization = "myorg"
    }
  }
}

data "komodo_build" "test" {
  id = komodo_build.test.id
}
`, name)
}

func testAccBuildDataSourceConfig_version(name string) string {
	return fmt.Sprintf(`
resource "komodo_build" "test" {
  name = %q
  version {
    value                  = "1.0.0"
    auto_increment_enabled = true
  }
}

data "komodo_build" "test" {
  id = komodo_build.test.id
}
`, name)
}
