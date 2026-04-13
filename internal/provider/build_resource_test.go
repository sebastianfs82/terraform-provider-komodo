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

// ---------------------------------------------------------------------------
// Basic lifecycle
// ---------------------------------------------------------------------------

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

func TestAccBuildResource_rename(t *testing.T) {
	var savedID string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBuildResourceConfig("tf-acc-build-rename-orig"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_build.test", "name", "tf-acc-build-rename-orig"),
					resource.TestCheckResourceAttrSet("komodo_build.test", "id"),
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["komodo_build.test"]
						savedID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config: testAccBuildResourceConfig("tf-acc-build-rename-new"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_build.test", "name", "tf-acc-build-rename-new"),
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["komodo_build.test"]
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

func TestAccBuildResource_importState(t *testing.T) {
	var buildID string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBuildResourceWithSourceConfig("tf-acc-build-import", "main"),
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
				Config:            testAccBuildResourceWithSourceConfig("tf-acc-build-import", "main"),
				ResourceName:      "komodo_build.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(_ *terraform.State) (string, error) { return buildID, nil },
				// webhook.secret and build.secret_args are sensitive and may not round-trip
				ImportStateVerifyIgnore: []string{"webhook.secret", "build.secret_args"},
			},
		},
	})
}

// ---------------------------------------------------------------------------
// Tags
// ---------------------------------------------------------------------------

func TestAccBuildResource_tags(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBuildWithTagConfig("tf-acc-build-tags"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_build.test", "tags.#", "1"),
					resource.TestCheckResourceAttrPair("komodo_build.test", "tags.0", "komodo_tag.test", "id"),
				),
			},
			{
				Config: testAccBuildClearTagsConfig("tf-acc-build-tags"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_build.test", "tags.#", "0"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// source block
// ---------------------------------------------------------------------------

func TestAccBuildResource_sourceBlock(t *testing.T) {
	const name = "tf-acc-build-source"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Add source block
			{
				Config: testAccBuildResourceWithSourceConfig(name, "main"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_build.test", "source.path", "myorg/myrepo"),
					resource.TestCheckResourceAttr("komodo_build.test", "source.branch", "main"),
				),
			},
			// Update source block
			{
				Config: testAccBuildResourceWithSourceConfig(name, "develop"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_build.test", "source.path", "myorg/myrepo"),
					resource.TestCheckResourceAttr("komodo_build.test", "source.branch", "develop"),
				),
			},
			// Remove source block
			{
				Config: testAccBuildResourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("komodo_build.test", "source.path"),
					resource.TestCheckNoResourceAttr("komodo_build.test", "source.branch"),
				),
			},
		},
	})
}

func TestAccBuildResource_sourceURL(t *testing.T) {
	const name = "tf-acc-build-source-url"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "komodo_build" "test" {
  name = %q
  source {
    url    = "https://github.com"
    path   = "myorg/myrepo"
    branch = "main"
  }
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_build.test", "source.url", "https://github.com"),
					resource.TestCheckResourceAttr("komodo_build.test", "source.path", "myorg/myrepo"),
					resource.TestCheckResourceAttr("komodo_build.test", "source.branch", "main"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// version block
// ---------------------------------------------------------------------------

func TestAccBuildResource_versionBlock(t *testing.T) {
	const name = "tf-acc-build-version"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Add version block
			{
				Config: fmt.Sprintf(`
resource "komodo_build" "test" {
  name = %q
  version {
    value                  = "1.2.3"
    auto_increment_enabled = false
  }
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_build.test", "version.value", "1.2.3"),
					resource.TestCheckResourceAttr("komodo_build.test", "version.auto_increment_enabled", "false"),
				),
			},
			// Update version block
			{
				Config: fmt.Sprintf(`
resource "komodo_build" "test" {
  name = %q
  version {
    value                  = "2.0.0"
    auto_increment_enabled = true
  }
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_build.test", "version.value", "2.0.0"),
					resource.TestCheckResourceAttr("komodo_build.test", "version.auto_increment_enabled", "true"),
				),
			},
			// Remove version block - API resets to 0.0.0 + auto_increment=true
			{
				Config: testAccBuildResourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("komodo_build.test", "version.value"),
					resource.TestCheckNoResourceAttr("komodo_build.test", "version.auto_increment_enabled"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// image block
// ---------------------------------------------------------------------------

func TestAccBuildResource_imageBlock(t *testing.T) {
	const name = "tf-acc-build-image"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Add image block with all flags
			{
				Config: fmt.Sprintf(`
resource "komodo_build" "test" {
  name = %q
  image {
    name                          = "myorg/my-service"
    tag                           = "stable"
    include_latest_tag_enabled    = true
    include_version_tags_enabled  = false
    include_commit_tag_enabled    = true
  }
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_build.test", "image.name", "myorg/my-service"),
					resource.TestCheckResourceAttr("komodo_build.test", "image.tag", "stable"),
					resource.TestCheckResourceAttr("komodo_build.test", "image.include_latest_tag_enabled", "true"),
					resource.TestCheckResourceAttr("komodo_build.test", "image.include_version_tags_enabled", "false"),
					resource.TestCheckResourceAttr("komodo_build.test", "image.include_commit_tag_enabled", "true"),
				),
			},
			// Update image block
			{
				Config: fmt.Sprintf(`
resource "komodo_build" "test" {
  name = %q
  image {
    name                         = "myorg/updated"
    include_latest_tag_enabled   = false
    include_version_tags_enabled = true
    include_commit_tag_enabled   = false
  }
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_build.test", "image.name", "myorg/updated"),
					resource.TestCheckResourceAttr("komodo_build.test", "image.include_latest_tag_enabled", "false"),
					resource.TestCheckResourceAttr("komodo_build.test", "image.include_version_tags_enabled", "true"),
					resource.TestCheckResourceAttr("komodo_build.test", "image.include_commit_tag_enabled", "false"),
				),
			},
			// Remove image block
			{
				Config: testAccBuildResourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("komodo_build.test", "image.name"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// image.registry block (account_id resolves to DockerRegistryAccount OID)
// ---------------------------------------------------------------------------

func TestAccBuildResource_imageRegistry(t *testing.T) {
	const name = "tf-acc-build-img-registry"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create build with one registry entry
			{
				Config: testAccBuildWithRegistryConfig(name, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_build.test", "image.registry.#", "1"),
					resource.TestCheckResourceAttrPair(
						"komodo_build.test", "image.registry.0.account_id",
						"komodo_registry_account.reg1", "id",
					),
					resource.TestCheckResourceAttr("komodo_build.test", "image.registry.0.organization", "myorg"),
				),
			},
			// Add second registry entry
			{
				Config: testAccBuildWithRegistryConfig(name, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_build.test", "image.registry.#", "2"),
					resource.TestCheckResourceAttrPair(
						"komodo_build.test", "image.registry.0.account_id",
						"komodo_registry_account.reg1", "id",
					),
					resource.TestCheckResourceAttrPair(
						"komodo_build.test", "image.registry.1.account_id",
						"komodo_registry_account.reg2", "id",
					),
				),
			},
			// Remove back to one entry
			{
				Config: testAccBuildWithRegistryConfig(name, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_build.test", "image.registry.#", "1"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// build block
// ---------------------------------------------------------------------------

func TestAccBuildResource_buildBlock(t *testing.T) {
	const name = "tf-acc-build-block"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Add build block
			{
				Config: fmt.Sprintf(`
resource "komodo_build" "test" {
  name = %q
  build {
    path       = "./app"
    extra_args = ["--no-cache"]
    args       = "ENV=production"
  }
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_build.test", "build.path", "./app"),
					resource.TestCheckResourceAttr("komodo_build.test", "build.extra_args.#", "1"),
					resource.TestCheckResourceAttr("komodo_build.test", "build.extra_args.0", "--no-cache"),
					resource.TestCheckResourceAttr("komodo_build.test", "build.args", "ENV=production"),
				),
			},
			// Update build block
			{
				Config: fmt.Sprintf(`
resource "komodo_build" "test" {
  name = %q
  build {
    path = "."
  }
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_build.test", "build.path", "."),
					resource.TestCheckResourceAttr("komodo_build.test", "build.extra_args.#", "0"),
				),
			},
			// Remove build block
			{
				Config: testAccBuildResourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("komodo_build.test", "build.path"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// webhook block
// ---------------------------------------------------------------------------

func TestAccBuildResource_webhookBlock(t *testing.T) {
	const name = "tf-acc-build-webhook"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Add webhook block
			{
				Config: fmt.Sprintf(`
resource "komodo_build" "test" {
  name = %q
  webhook {
    enabled = true
  }
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_build.test", "webhook.enabled", "true"),
				),
			},
			// Disable webhook
			{
				Config: fmt.Sprintf(`
resource "komodo_build" "test" {
  name = %q
  webhook {
    enabled = false
  }
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_build.test", "webhook.enabled", "false"),
				),
			},
			// Remove webhook block - provider sends enabled=false, secret="" to clear
			{
				Config: testAccBuildResourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("komodo_build.test", "webhook.enabled"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// pre_build block
// ---------------------------------------------------------------------------

func TestAccBuildResource_preBuildBlock(t *testing.T) {
	const name = "tf-acc-build-prebuild"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Add pre_build block
			{
				Config: fmt.Sprintf(`
resource "komodo_build" "test" {
  name = %q
  pre_build {
    path    = "/app"
    command = "make test"
  }
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_build.test", "pre_build.path", "/app"),
					resource.TestCheckResourceAttr("komodo_build.test", "pre_build.command", "make test"),
				),
			},
			// Update pre_build command
			{
				Config: fmt.Sprintf(`
resource "komodo_build" "test" {
  name = %q
  pre_build {
    path    = "/app"
    command = "make lint"
  }
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_build.test", "pre_build.command", "make lint"),
				),
			},
			// Remove pre_build block - provider sends empty path+command to clear
			{
				Config: testAccBuildResourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("komodo_build.test", "pre_build.command"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

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

func testAccBuildResourceWithSourceConfig(name, branch string) string {
	return fmt.Sprintf(`
resource "komodo_build" "test" {
  name = %q
  source {
    path   = "myorg/myrepo"
    branch = %q
  }
}
`, name, branch)
}

func testAccBuildWithTagConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_tag" "test" {
  name  = "tf-acc-tag-build"
  color = "Green"
}

resource "komodo_build" "test" {
  name = %q
  tags = [komodo_tag.test.id]
}
`, name)
}

func testAccBuildClearTagsConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_build" "test" {
  name = %q
  tags = []
}
`, name)
}

// testAccBuildWithRegistryConfig creates a build with 1 or 2 registry blocks.
// It always creates both komodo_registry_account resources so references are valid.
func testAccBuildWithRegistryConfig(name string, count int) string {
	registryBlocks := `
  registry {
    account_id   = komodo_registry_account.reg1.id
    organization = "myorg"
  }`
	if count >= 2 {
		registryBlocks += `
  registry {
    account_id = komodo_registry_account.reg2.id
  }`
	}
	return fmt.Sprintf(`
resource "komodo_registry_account" "reg1" {
  domain   = "registry.example.com"
  username = "tf-acc-reg1-user"
  token    = "reg1-token"
}

resource "komodo_registry_account" "reg2" {
  domain   = "ghcr.io"
  username = "tf-acc-reg2-user"
  token    = "reg2-token"
}

resource "komodo_build" "test" {
  name = %q
  image {%s
  }
}
`, name, registryBlocks)
}
