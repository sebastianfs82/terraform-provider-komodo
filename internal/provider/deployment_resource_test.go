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

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
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
	serverID := testAccLookupServerID(t, "deployment rename tests (rename requires a known container state)")
	var savedID string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentResourceConfigWithServer("tf-acc-deployment-rename-orig", serverID),
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
				Config: testAccDeploymentResourceConfigWithServer("tf-acc-deployment-rename-new", serverID),
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

func testAccDeploymentResourceConfigWithServer(name, serverID string) string {
	return fmt.Sprintf(`
resource "komodo_deployment" "test" {
  name      = %q
  server_id = %q
}
`, name, serverID)
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

// ─── Unit tests ──────────────────────────────────────────────────────────────

func wrongRawDeploymentPlan(t *testing.T, r *DeploymentResource) tfsdk.Plan {
	t.Helper()
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schemaResp.Schema,
	}
}

func wrongRawDeploymentState(t *testing.T, r *DeploymentResource) tfsdk.State {
	t.Helper()
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	return tfsdk.State{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schemaResp.Schema,
	}
}

func TestUnitDeploymentResource_configure(t *testing.T) {
	t.Run("wrong_type", func(t *testing.T) {
		r := &DeploymentResource{}
		req := fwresource.ConfigureRequest{ProviderData: "not-a-client"}
		resp := &fwresource.ConfigureResponse{}
		r.Configure(context.Background(), req, resp)
		if !resp.Diagnostics.HasError() {
			t.Fatal("expected diagnostic error for wrong ProviderData type")
		}
	})
}

func TestUnitDeploymentResource_createPlanGetError(t *testing.T) {
	r := &DeploymentResource{client: &client.Client{}}
	req := fwresource.CreateRequest{Plan: wrongRawDeploymentPlan(t, r)}
	resp := &fwresource.CreateResponse{}
	r.Create(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed plan")
	}
}

func TestUnitDeploymentResource_readStateGetError(t *testing.T) {
	r := &DeploymentResource{client: &client.Client{}}
	req := fwresource.ReadRequest{State: wrongRawDeploymentState(t, r)}
	resp := &fwresource.ReadResponse{}
	r.Read(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed state")
	}
}

func TestUnitDeploymentResource_updatePlanGetError(t *testing.T) {
	r := &DeploymentResource{client: &client.Client{}}
	req := fwresource.UpdateRequest{Plan: wrongRawDeploymentPlan(t, r)}
	resp := &fwresource.UpdateResponse{}
	r.Update(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed plan")
	}
}

func TestUnitDeploymentResource_deleteStateGetError(t *testing.T) {
	r := &DeploymentResource{client: &client.Client{}}
	req := fwresource.DeleteRequest{State: wrongRawDeploymentState(t, r)}
	resp := &fwresource.DeleteResponse{}
	r.Delete(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed state")
	}
}

func TestUnitDeploymentResource_imageModelToClient_buildNoVersion(t *testing.T) {
	m := &DeploymentImageModel{
		Type:    types.StringValue("Build"),
		BuildID: types.StringValue("build-abc"),
		Version: types.StringNull(),
		Image:   types.StringNull(),
	}
	result := imageModelToClient(m)
	if result.Build == nil {
		t.Fatal("expected Build to be set")
	}
	if result.Build.BuildID != "build-abc" {
		t.Fatalf("unexpected BuildID: %s", result.Build.BuildID)
	}
	if result.Image != nil {
		t.Fatal("expected Image to be nil for Build type")
	}
}

func TestUnitDeploymentResource_imageModelToClient_buildWithVersion(t *testing.T) {
	m := &DeploymentImageModel{
		Type:    types.StringValue("Build"),
		BuildID: types.StringValue("build-xyz"),
		Version: types.StringValue("2.3.4"),
		Image:   types.StringNull(),
	}
	result := imageModelToClient(m)
	if result.Build == nil {
		t.Fatal("expected Build to be set")
	}
	if result.Build.Version.Major != 2 || result.Build.Version.Minor != 3 || result.Build.Version.Patch != 4 {
		t.Fatalf("unexpected version: %+v", result.Build.Version)
	}
}

func TestUnitDeploymentResource_imageModelToClient_imageType(t *testing.T) {
	m := &DeploymentImageModel{
		Type:    types.StringValue("Image"),
		Image:   types.StringValue("nginx:latest"),
		BuildID: types.StringNull(),
		Version: types.StringNull(),
	}
	result := imageModelToClient(m)
	if result.Image == nil {
		t.Fatal("expected Image to be set")
	}
	if result.Image.Image != "nginx:latest" {
		t.Fatalf("unexpected image: %s", result.Image.Image)
	}
	if result.Build != nil {
		t.Fatal("expected Build to be nil for Image type")
	}
}

func TestUnitDeploymentResource_deploymentToModel_externalImage(t *testing.T) {
	ctx := context.Background()
	c := &client.Client{}

	d := &client.Deployment{
		ID:   client.OID{OID: "dep123"},
		Name: "my-deployment",
		Tags: []string{},
		Config: client.DeploymentConfig{
			ServerID: "server-1",
			Image: client.DeploymentImage{
				Image: &client.DeploymentImageExternal{Image: "nginx:latest"},
			},
		},
	}
	// Pre-set Image block to trigger the "update existing image" path in deploymentToModel.
	data := &DeploymentResourceModel{
		Tags:  types.ListValueMust(types.StringType, nil),
		Image: &DeploymentImageModel{},
	}

	deploymentToModel(ctx, c, d, data)

	if data.ID.ValueString() != "dep123" {
		t.Fatalf("unexpected id: %s", data.ID.ValueString())
	}
	if data.Name.ValueString() != "my-deployment" {
		t.Fatalf("unexpected name: %s", data.Name.ValueString())
	}
	if data.ServerID.ValueString() != "server-1" {
		t.Fatalf("unexpected server_id: %s", data.ServerID.ValueString())
	}
	if data.Image == nil {
		t.Fatal("expected image block")
	}
	if data.Image.Type.ValueString() != "Image" {
		t.Fatalf("expected image type=Image, got %s", data.Image.Type.ValueString())
	}
	if data.Image.Image.ValueString() != "nginx:latest" {
		t.Fatalf("expected image=nginx:latest, got %s", data.Image.Image.ValueString())
	}
}

func TestUnitDeploymentResource_deploymentToModel_buildImage(t *testing.T) {
	ctx := context.Background()
	c := &client.Client{}

	d := &client.Deployment{
		ID:   client.OID{OID: "dep456"},
		Name: "build-deployment",
		Tags: []string{},
		Config: client.DeploymentConfig{
			Image: client.DeploymentImage{
				Build: &client.DeploymentImageBuild{
					BuildID: "build-xyz",
					Version: client.BuildVersion{Major: 1, Minor: 2, Patch: 3},
				},
			},
		},
	}
	data := &DeploymentResourceModel{
		Tags:  types.ListValueMust(types.StringType, nil),
		Image: &DeploymentImageModel{Version: types.StringNull()},
	}

	deploymentToModel(ctx, c, d, data)

	if data.Image == nil {
		t.Fatal("expected image block")
	}
	if data.Image.Type.ValueString() != "Build" {
		t.Fatalf("expected image type=Build, got %s", data.Image.Type.ValueString())
	}
	if data.Image.BuildID.ValueString() != "build-xyz" {
		t.Fatalf("expected build_id=build-xyz, got %s", data.Image.BuildID.ValueString())
	}
	if data.Image.Version.ValueString() != "1.2.3" {
		t.Fatalf("expected version=1.2.3, got %s", data.Image.Version.ValueString())
	}
}

func TestUnitDeploymentResource_partialDeploymentConfigFromModel_serverOnly(t *testing.T) {
	ctx := context.Background()
	c := &client.Client{}

	data := &DeploymentResourceModel{
		ServerID:                       types.StringValue("srv-1"),
		SwarmID:                        types.StringNull(),
		SkipSecretInterpolationEnabled: types.BoolValue(true),
		PollForUpdatesEnabled:          types.BoolValue(false),
		AutoUpdateEnabled:              types.BoolValue(false),
		SendAlertsEnabled:              types.BoolValue(true),
		Image:                          nil,
		Container:                      nil,
		Termination:                    nil,
	}

	cfg := partialDeploymentConfigFromModel(ctx, c, data)

	if cfg.ServerID == nil || *cfg.ServerID != "srv-1" {
		t.Fatalf("expected server_id=srv-1, got %v", cfg.ServerID)
	}
	// SkipSecretInterpolation is the logical inverse of secret_interpolation_enabled.
	if cfg.SkipSecretInterpolation == nil || *cfg.SkipSecretInterpolation {
		t.Fatal("expected SkipSecretInterpolation=false when secret_interpolation_enabled=true")
	}
	if cfg.Image != nil {
		t.Fatal("expected no image in config when data.Image=nil")
	}
	if cfg.SendAlerts == nil || !*cfg.SendAlerts {
		t.Fatal("expected send_alerts=true")
	}
}

func TestUnitDeploymentResource_partialDeploymentConfigFromModel_withImageAndContainer(t *testing.T) {
	ctx := context.Background()
	c := &client.Client{}

	data := &DeploymentResourceModel{
		ServerID:                       types.StringValue("srv-1"),
		SwarmID:                        types.StringNull(),
		SkipSecretInterpolationEnabled: types.BoolNull(),
		PollForUpdatesEnabled:          types.BoolNull(),
		AutoUpdateEnabled:              types.BoolNull(),
		SendAlertsEnabled:              types.BoolNull(),
		Image: &DeploymentImageModel{
			Type:            types.StringValue("Image"),
			Image:           types.StringValue("redis:7"),
			BuildID:         types.StringNull(),
			Version:         types.StringNull(),
			RegistryAccount: types.StringNull(),
			RedeployEnabled: types.BoolValue(false),
		},
		Container: &DeploymentContainerModel{
			Network:        types.StringValue("bridge"),
			Restart:        types.StringValue("unless-stopped"),
			Command:        types.StringNull(),
			Replicas:       types.Int64Value(2),
			ExtraArguments: types.ListValueMust(types.StringType, nil),
			Ports:          types.ListValueMust(types.StringType, nil),
			Volumes:        types.ListValueMust(types.StringType, nil),
			Environment:    types.MapNull(types.StringType),
			Labels:         types.ListValueMust(types.StringType, nil),
			Links:          types.ListValueMust(types.StringType, nil),
		},
		Termination: nil,
	}

	cfg := partialDeploymentConfigFromModel(ctx, c, data)

	if cfg.Image == nil {
		t.Fatal("expected image in config")
	}
	if cfg.Network == nil || *cfg.Network != "bridge" {
		t.Fatalf("expected network=bridge, got %v", cfg.Network)
	}
	if cfg.Restart == nil || *cfg.Restart != "unless-stopped" {
		t.Fatalf("expected restart=unless-stopped, got %v", cfg.Restart)
	}
	if cfg.Replicas == nil || *cfg.Replicas != 2 {
		t.Fatalf("expected replicas=2, got %v", cfg.Replicas)
	}
}

func TestUnitDeploymentResource_deploymentToModel_containerBlock(t *testing.T) {
	ctx := context.Background()
	c := &client.Client{}

	d := &client.Deployment{
		ID:   client.OID{OID: "dep-container"},
		Name: "container-dep",
		Config: client.DeploymentConfig{
			ServerID:    "srv-1",
			Network:     "bridge",
			Restart:     "unless-stopped",
			Command:     "./start.sh",
			Replicas:    3,
			Ports:       "8080:80\n443:443",
			Volumes:     "/data:/app/data",
			Environment: "DATABASE_URL=postgres://localhost/db\nDEBUG=false",
			Labels:      "app=myapp",
			Image:       client.DeploymentImage{Image: &client.DeploymentImageExternal{Image: "nginx:latest"}},
		},
	}
	data := &DeploymentResourceModel{
		Tags: types.ListValueMust(types.StringType, nil),
	}
	deploymentToModel(ctx, c, d, data)

	if data.Container == nil {
		t.Fatal("expected non-nil container block")
	}
	if data.Container.Network.ValueString() != "bridge" {
		t.Fatalf("expected network=bridge, got %s", data.Container.Network.ValueString())
	}
	if data.Container.Restart.ValueString() != "unless-stopped" {
		t.Fatalf("expected restart=unless-stopped, got %s", data.Container.Restart.ValueString())
	}
	if data.Container.Command.ValueString() != "./start.sh" {
		t.Fatalf("expected command=./start.sh, got %s", data.Container.Command.ValueString())
	}
	if data.Container.Replicas.ValueInt64() != 3 {
		t.Fatalf("expected replicas=3, got %d", data.Container.Replicas.ValueInt64())
	}
	if len(data.Container.Ports.Elements()) != 2 {
		t.Fatalf("expected 2 ports, got %d", len(data.Container.Ports.Elements()))
	}
	if len(data.Container.Volumes.Elements()) != 1 {
		t.Fatalf("expected 1 volume, got %d", len(data.Container.Volumes.Elements()))
	}
	if len(data.Container.Environment.Elements()) != 2 {
		t.Fatalf("expected 2 env vars, got %d", len(data.Container.Environment.Elements()))
	}
	if len(data.Container.Labels.Elements()) != 1 {
		t.Fatalf("expected 1 label, got %d", len(data.Container.Labels.Elements()))
	}
}

func TestUnitDeploymentResource_deploymentToModel_terminationBlock(t *testing.T) {
	ctx := context.Background()
	c := &client.Client{}

	d := &client.Deployment{
		ID:   client.OID{OID: "dep-term"},
		Name: "term-dep",
		Config: client.DeploymentConfig{
			TerminationSignal:       "SIGTERM",
			TerminationTimeout:      30,
			TerminationSignalLabels: "traefik.enable=false",
			Image:                   client.DeploymentImage{Image: &client.DeploymentImageExternal{}},
		},
	}
	data := &DeploymentResourceModel{
		Tags: types.ListValueMust(types.StringType, nil),
	}
	deploymentToModel(ctx, c, d, data)

	if data.Termination == nil {
		t.Fatal("expected non-nil termination block")
	}
	if data.Termination.Signal.ValueString() != "SIGTERM" {
		t.Fatalf("expected signal=SIGTERM, got %s", data.Termination.Signal.ValueString())
	}
	if data.Termination.Timeout.ValueInt64() != 30 {
		t.Fatalf("expected timeout=30, got %d", data.Termination.Timeout.ValueInt64())
	}
	if data.Termination.SignalLabels.ValueString() != "traefik.enable=false" {
		t.Fatalf("expected signal_labels=traefik.enable=false, got %s", data.Termination.SignalLabels.ValueString())
	}
}

func TestUnitDeploymentResource_deploymentToModel_imagePreserveExisting(t *testing.T) {
	ctx := context.Background()
	c := &client.Client{}

	// API returns neither Build nor Image (both nil) — prior model has an image block.
	// Expect: RegistryAccount and RedeployEnabled updated, rest preserved.
	d := &client.Deployment{
		ID:   client.OID{OID: "dep-preserve"},
		Name: "preserve-dep",
		Config: client.DeploymentConfig{
			RedeployOnBuild: true,
		},
	}
	data := &DeploymentResourceModel{
		Tags: types.ListValueMust(types.StringType, nil),
		Image: &DeploymentImageModel{
			Type:            types.StringValue("Image"),
			Image:           types.StringValue("nginx:latest"),
			BuildID:         types.StringNull(),
			Version:         types.StringNull(),
			RegistryAccount: types.StringNull(),
			RedeployEnabled: types.BoolValue(false),
		},
	}
	deploymentToModel(ctx, c, d, data)

	if data.Image == nil {
		t.Fatal("expected image block preserved when API returns neither variant")
	}
	if !data.Image.RedeployEnabled.ValueBool() {
		t.Fatal("expected RedeployEnabled updated to true from API")
	}
	if data.Image.Type.ValueString() != "Image" {
		t.Fatal("expected image type preserved")
	}
}

func TestUnitDeploymentResource_deploymentToModel_imageNilWhenNoConfig(t *testing.T) {
	ctx := context.Background()
	c := &client.Client{}

	// API returns no image config and prior data.Image is nil — image stays nil.
	d := &client.Deployment{
		ID:     client.OID{OID: "dep-noimg"},
		Name:   "noimg-dep",
		Config: client.DeploymentConfig{},
	}
	data := &DeploymentResourceModel{
		Tags:  types.ListValueMust(types.StringType, nil),
		Image: nil,
	}
	deploymentToModel(ctx, c, d, data)

	if data.Image != nil {
		t.Fatal("expected nil image when API returns no image and prior is nil")
	}
}

func TestUnitDeploymentResource_partialDeploymentConfigFromModel_terminationBlock(t *testing.T) {
	ctx := context.Background()
	c := &client.Client{}

	data := &DeploymentResourceModel{
		Tags: types.ListValueMust(types.StringType, nil),
		Termination: &DeploymentTerminationModel{
			Signal:       types.StringValue("SIGTERM"),
			Timeout:      types.Int64Value(30),
			SignalLabels: types.StringValue("foo=bar"),
		},
	}
	cfg := partialDeploymentConfigFromModel(ctx, c, data)
	if cfg.TerminationSignal == nil || *cfg.TerminationSignal != "SIGTERM" {
		t.Fatalf("unexpected TerminationSignal: %v", cfg.TerminationSignal)
	}
	if cfg.TerminationTimeout == nil || *cfg.TerminationTimeout != 30 {
		t.Fatalf("unexpected TerminationTimeout: %v", cfg.TerminationTimeout)
	}
	if cfg.TerminationSignalLabels == nil || *cfg.TerminationSignalLabels != "foo=bar" {
		t.Fatalf("unexpected TerminationSignalLabels: %v", cfg.TerminationSignalLabels)
	}
}

func TestUnitDeploymentResource_partialDeploymentConfigFromModel_containerLinks(t *testing.T) {
	ctx := context.Background()
	c := &client.Client{}

	links, _ := types.ListValueFrom(ctx, types.StringType, []string{"container1", "container2"})
	data := &DeploymentResourceModel{
		Tags: types.ListValueMust(types.StringType, nil),
		Container: &DeploymentContainerModel{
			Links: links,
		},
	}
	cfg := partialDeploymentConfigFromModel(ctx, c, data)
	if cfg.Links == nil || len(*cfg.Links) != 2 {
		t.Fatalf("unexpected Links: %v", cfg.Links)
	}
}

func TestUnitDeploymentResource_imageValidator_typeRequired(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "komodo_deployment" "test" {
  name      = "tf-test-dep"
  server_id = "srv"
  image {}
}`,
				ExpectError: regexp.MustCompile(`image\.type is required`),
			},
		},
	})
}

func TestUnitDeploymentResource_imageValidator_nameRequiredForImage(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "komodo_deployment" "test" {
  name      = "tf-test-dep"
  server_id = "srv"
  image {
    type = "Image"
  }
}`,
				ExpectError: regexp.MustCompile(`image\.name is required`),
			},
		},
	})
}

func TestUnitDeploymentResource_imageValidator_buildIDRequiredForBuild(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "komodo_deployment" "test" {
  name      = "tf-test-dep"
  server_id = "srv"
  image {
    type = "Build"
  }
}`,
				ExpectError: regexp.MustCompile(`image\.build_id is required`),
			},
		},
	})
}

// ─── Mock-server unit tests ───────────────────────────────────────────────────

// mockValidDeploymentJSON is a minimal but complete Deployment JSON that satisfies
// deploymentToModel without panicking. network="host" and restart="no" ensure the
// container block is populated so the state round-trips cleanly.
const mockValidDeploymentJSON = `{"_id":{"$oid":"507f1f77bcf86cd799439011"},"name":"tf-mock-dep","tags":[],"config":{"server_id":"","swarm_id":"","image":{},"image_registry_account":"","skip_secret_interp":false,"redeploy_on_build":false,"poll_for_updates":false,"auto_update":false,"send_alerts":false,"links":[],"network":"host","restart":"no","command":"","replicas":1,"termination_signal":"SIGTERM","termination_timeout":10,"extra_args":[],"term_signal_labels":"","ports":"","volumes":"","environment":"","labels":""}}`

func TestUnitDeploymentResource_createClientError(t *testing.T) {
	srv := newStatefulUserMockServer(t, func(typ string, _ int) (int, string) {
		if typ == "CreateDeployment" {
			return 500, `"create error"`
		}
		return 200, mockValidDeploymentJSON
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockUserProviderConfig(srv.URL) + `resource "komodo_deployment" "test" { name = "tf-mock-dep" }`,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

func TestUnitDeploymentResource_createMissingID(t *testing.T) {
	srv := newStatefulUserMockServer(t, func(typ string, _ int) (int, string) {
		if typ == "CreateDeployment" {
			return 200, `{"_id":{"$oid":""},"name":"tf-mock-dep","tags":[],"config":{}}`
		}
		return 200, mockValidDeploymentJSON
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockUserProviderConfig(srv.URL) + `resource "komodo_deployment" "test" { name = "tf-mock-dep" }`,
				ExpectError: regexp.MustCompile(`(?i)missing ID`),
			},
		},
	})
}

func TestUnitDeploymentResource_deleteClientError(t *testing.T) {
	// Only the first DeleteDeployment call fails; subsequent cleanup calls succeed
	// so the framework's post-test destroy doesn't leave dangling resources.
	srv := newStatefulUserMockServer(t, func(typ string, n int) (int, string) {
		if typ == "DeleteDeployment" && n == 1 {
			return 500, `"delete error"`
		}
		return 200, mockValidDeploymentJSON
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockUserProviderConfig(srv.URL) + `resource "komodo_deployment" "test" { name = "tf-mock-dep" }`,
			},
			{
				Config:      mockUserProviderConfig(srv.URL),
				Destroy:     true,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

func TestUnitDeploymentResource_updateRenameError(t *testing.T) {
	srv := newStatefulUserMockServer(t, func(typ string, _ int) (int, string) {
		if typ == "RenameDeployment" {
			return 500, `"rename error"`
		}
		return 200, mockValidDeploymentJSON
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockUserProviderConfig(srv.URL) + `resource "komodo_deployment" "test" { name = "tf-mock-dep" }`,
			},
			{
				Config:      mockUserProviderConfig(srv.URL) + `resource "komodo_deployment" "test" { name = "tf-mock-dep-v2" }`,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

func TestUnitDeploymentResource_updateDeploymentError(t *testing.T) {
	srv := newStatefulUserMockServer(t, func(typ string, _ int) (int, string) {
		if typ == "UpdateDeployment" {
			return 500, `"update error"`
		}
		return 200, mockValidDeploymentJSON
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockUserProviderConfig(srv.URL) + `resource "komodo_deployment" "test" { name = "tf-mock-dep" }`,
			},
			{
				// Same name (no rename), but explicit server_id triggers UpdateDeployment.
				Config: mockUserProviderConfig(srv.URL) + `
resource "komodo_deployment" "test" {
  name      = "tf-mock-dep"
  server_id = "my-server"
}`,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

func TestUnitDeploymentResource_autoUpdateValidator_requiresPollUpdates(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "komodo_deployment" "test" {
  name                  = "tf-test-dep"
  server_id             = "srv"
  auto_update_enabled   = true
  poll_updates_enabled  = false
}`,
				ExpectError: regexp.MustCompile(`poll_updates_enabled required`),
			},
		},
	})
}
