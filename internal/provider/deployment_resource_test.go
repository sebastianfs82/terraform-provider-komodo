// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"regexp"
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
					resource.TestCheckResourceAttr("komodo_deployment.test", "image.name", "nginx:latest"),
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
	serverID := os.Getenv("KOMODO_TEST_SERVER_ID")
	if serverID == "" {
		t.Skip("KOMODO_TEST_SERVER_ID must be set for deployment rename tests (rename requires a known container state)")
	}
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
  image {
    type = "Image"
    name = %q
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

// ---------------------------------------------------------------------------
// Validator error tests — image block
// These tests verify plan-time validation rejects invalid configurations.
// No Komodo API calls are made; errors are raised before apply.
// ---------------------------------------------------------------------------

// TestAccDeploymentResource_imageTypeLowercase verifies that a lowercase image
// type (e.g. "image") is rejected; the attribute only accepts "Image" or "Build".
func TestAccDeploymentResource_imageTypeLowercase(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "komodo_deployment" "test" {
  name = "tf-acc-validator-test"
  image {
    type = "image"
    name = "nginx:latest"
  }
}`,
				ExpectError: regexp.MustCompile(`value must be one of`),
			},
		},
	})
}

// TestAccDeploymentResource_imageNameRequired verifies that image.name is
// required when image.type is "Image".
func TestAccDeploymentResource_imageNameRequired(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "komodo_deployment" "test" {
  name = "tf-acc-validator-test"
  image {
    type = "Image"
  }
}`,
				ExpectError: regexp.MustCompile(`image\.name is required`),
			},
		},
	})
}

// TestAccDeploymentResource_imageBuildIDRequired verifies that image.build_id is
// required when image.type is "Build".
func TestAccDeploymentResource_imageBuildIDRequired(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "komodo_deployment" "test" {
  name = "tf-acc-validator-test"
  image {
    type = "Build"
  }
}`,
				ExpectError: regexp.MustCompile(`image\.build_id is required`),
			},
		},
	})
}

// TestAccDeploymentResource_imageBuildIDForbiddenWithTypeImage verifies that
// image.build_id cannot be set together with image.type = "Image".
func TestAccDeploymentResource_imageBuildIDForbiddenWithTypeImage(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "komodo_deployment" "test" {
  name = "tf-acc-validator-test"
  image {
    type     = "Image"
    name     = "nginx:latest"
    build_id = "some-build-id"
  }
}`,
				ExpectError: regexp.MustCompile(`image\.build_id not allowed`),
			},
		},
	})
}

// TestAccDeploymentResource_imageNameForbiddenWithTypeBuild verifies that
// image.name cannot be set together with image.type = "Build".
func TestAccDeploymentResource_imageNameForbiddenWithTypeBuild(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "komodo_deployment" "test" {
  name = "tf-acc-validator-test"
  image {
    type     = "Build"
    build_id = "some-build-id"
    name     = "nginx:latest"
  }
}`,
				ExpectError: regexp.MustCompile(`image\.name not allowed`),
			},
		},
	})
}

// TestAccDeploymentResource_imageVersionForbiddenWithTypeImage verifies that
// image.version cannot be set when image.type is "Image".
func TestAccDeploymentResource_imageVersionForbiddenWithTypeImage(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "komodo_deployment" "test" {
  name = "tf-acc-validator-test"
  image {
    type    = "Image"
    name    = "nginx:latest"
    version = "1.0"
  }
}`,
				ExpectError: regexp.MustCompile(`image\.version not allowed`),
			},
		},
	})
}

// TestAccDeploymentResource_imageRedeployEnabledForbiddenWithTypeImage verifies
// that image.redeploy_enabled = true cannot be set when image.type is "Image".
func TestAccDeploymentResource_imageRedeployEnabledForbiddenWithTypeImage(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "komodo_deployment" "test" {
  name = "tf-acc-validator-test"
  image {
    type             = "Image"
    name             = "nginx:latest"
    redeploy_enabled = true
  }
}`,
				ExpectError: regexp.MustCompile(`image\.redeploy_enabled not allowed`),
			},
		},
	})
}

// TestAccDeploymentResource_autoUpdateRequiresPollUpdates verifies that
// auto_update_enabled = true is rejected when poll_updates_enabled is
// explicitly set to false.
func TestAccDeploymentResource_autoUpdateRequiresPollUpdates(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "komodo_deployment" "test" {
  name                 = "tf-acc-validator-test"
  auto_update_enabled  = true
  poll_updates_enabled = false
}`,
				ExpectError: regexp.MustCompile(`poll_updates_enabled required`),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// Acceptance tests — container fields and renamed/negated attributes
// ---------------------------------------------------------------------------

// TestAccDeploymentResource_containerEnvironment verifies that the container
// environment map is written and read back correctly (no trailing-newline drift,
// no empty extra entries).
func TestAccDeploymentResource_containerEnvironment(t *testing.T) {
	const name = "tf-acc-deployment-env"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentContainerEnvConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_deployment.test", "name", name),
					resource.TestCheckResourceAttr("komodo_deployment.test", "container.environment.APP_ENV", "production"),
					resource.TestCheckResourceAttr("komodo_deployment.test", "container.environment.LOG_LEVEL", "info"),
				),
			},
			// Second step with identical config — verifies no perpetual diff
			// caused by trailing newlines on read-back.
			{
				Config:   testAccDeploymentContainerEnvConfig(name),
				PlanOnly: true,
			},
		},
	})
}

// TestAccDeploymentResource_containerListFields verifies that ports, volumes,
// and labels (all list-of-string) round-trip correctly without drift.
func TestAccDeploymentResource_containerListFields(t *testing.T) {
	const name = "tf-acc-deployment-lists"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentContainerListsConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_deployment.test", "name", name),
					resource.TestCheckResourceAttr("komodo_deployment.test", "container.ports.#", "2"),
					resource.TestCheckResourceAttr("komodo_deployment.test", "container.ports.0", "8080:80"),
					resource.TestCheckResourceAttr("komodo_deployment.test", "container.ports.1", "9090:90"),
					resource.TestCheckResourceAttr("komodo_deployment.test", "container.volumes.#", "1"),
					resource.TestCheckResourceAttr("komodo_deployment.test", "container.volumes.0", "/tmp/data:/data"),
					resource.TestCheckResourceAttr("komodo_deployment.test", "container.labels.#", "1"),
					resource.TestCheckResourceAttr("komodo_deployment.test", "container.labels.0", "com.example.env=test"),
				),
			},
			// No-drift check.
			{
				Config:   testAccDeploymentContainerListsConfig(name),
				PlanOnly: true,
			},
		},
	})
}

// TestAccDeploymentResource_containerLinks verifies that the links field lives
// inside the container block (not at the resource root) and round-trips cleanly.
func TestAccDeploymentResource_containerLinks(t *testing.T) {
	const name = "tf-acc-deployment-links"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentContainerLinksConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_deployment.test", "name", name),
					resource.TestCheckResourceAttr("komodo_deployment.test", "container.links.#", "1"),
					resource.TestCheckResourceAttr("komodo_deployment.test", "container.links.0", "sidecar:alias"),
				),
			},
			{
				Config:   testAccDeploymentContainerLinksConfig(name),
				PlanOnly: true,
			},
		},
	})
}

// TestAccDeploymentResource_secretInterpolationEnabled verifies that the
// secret_interpolation_enabled flag (which is stored as its inverse
// skip_secret_interpolation in the API) round-trips correctly.
func TestAccDeploymentResource_secretInterpolationEnabled(t *testing.T) {
	const name = "tf-acc-deployment-secret"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentSecretInterpolationConfig(name, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_deployment.test", "name", name),
					resource.TestCheckResourceAttr("komodo_deployment.test", "secret_interpolation_enabled", "true"),
				),
			},
			{
				Config: testAccDeploymentSecretInterpolationConfig(name, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_deployment.test", "secret_interpolation_enabled", "false"),
				),
			},
			{
				Config:   testAccDeploymentSecretInterpolationConfig(name, false),
				PlanOnly: true,
			},
		},
	})
}

// TestAccDeploymentResource_imageName verifies that the renamed image.name
// attribute (previously image.image) is written and read back correctly.
func TestAccDeploymentResource_imageName(t *testing.T) {
	const name = "tf-acc-deployment-imagename"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentResourceConfigWithImage(name, "nginx:latest"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_deployment.test", "image.type", "Image"),
					resource.TestCheckResourceAttr("komodo_deployment.test", "image.name", "nginx:latest"),
				),
			},
			{
				Config:   testAccDeploymentResourceConfigWithImage(name, "nginx:latest"),
				PlanOnly: true,
			},
		},
	})
}

// ---------------------------------------------------------------------------
// Config helpers for acceptance tests
// ---------------------------------------------------------------------------

func testAccDeploymentContainerEnvConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_deployment" "test" {
  name = %q
  container = {
    environment = {
      APP_ENV   = "production"
      LOG_LEVEL = "info"
    }
  }
}
`, name)
}

func testAccDeploymentContainerListsConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_deployment" "test" {
  name = %q
  container = {
    ports   = ["8080:80", "9090:90"]
    volumes = ["/tmp/data:/data"]
    labels  = ["com.example.env=test"]
  }
}
`, name)
}

func testAccDeploymentContainerLinksConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_deployment" "test" {
  name = %q
  container = {
    links = ["sidecar:alias"]
  }
}
`, name)
}

func testAccDeploymentSecretInterpolationConfig(name string, enabled bool) string {
	return fmt.Sprintf(`
resource "komodo_deployment" "test" {
  name                         = %q
  secret_interpolation_enabled = %v
}
`, name, enabled)
}
