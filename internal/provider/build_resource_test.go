// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// ---------------------------------------------------------------------------
// Unit tests – trailing-newline trimming (no API required)
// ---------------------------------------------------------------------------

func TestUnitBuildResource_trailingNewlineTrim(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"no newline", "KEY=value", "KEY=value"},
		{"trailing LF", "KEY=value\n", "KEY=value"},
		{"trailing CRLF", "KEY=value\r\n", "KEY=value"},
		{"trailing multi LF", "KEY=value\n\n", "KEY=value"},
		{"multiline LF", "A=1\nB=2\n", "A=1\nB=2"},
		{"empty string", "", ""},
		{"only newline", "\n", ""},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := strings.TrimRight(tc.input, "\n\r")
			if got != tc.want {
				t.Errorf("TrimRight(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestUnitBuildResource_parseBuildArguments(t *testing.T) {
	boolFalse := types.BoolValue(false)
	boolTrue := types.BoolValue(true)
	cases := []struct {
		name   string
		input  string
		secret bool
		want   []BuildArgumentModel
	}{
		{"empty", "", false, nil},
		{"whitespace", "  \n  ", false, nil},
		{"single plain", "KEY=value", false, []BuildArgumentModel{{Name: strVal("KEY"), Value: strVal("value"), SecretEnabled: boolFalse}}},
		{"single secret", "KEY=value", true, []BuildArgumentModel{{Name: strVal("KEY"), Value: strVal("value"), SecretEnabled: boolTrue}}},
		{"trailing LF", "KEY=value\n", false, []BuildArgumentModel{{Name: strVal("KEY"), Value: strVal("value"), SecretEnabled: boolFalse}}},
		{"two entries sorted", "Z=last\nA=first", false, []BuildArgumentModel{
			{Name: strVal("A"), Value: strVal("first"), SecretEnabled: boolFalse},
			{Name: strVal("Z"), Value: strVal("last"), SecretEnabled: boolFalse},
		}},
		{"value with equals", "URL=http://x?a=1", false, []BuildArgumentModel{{Name: strVal("URL"), Value: strVal("http://x?a=1"), SecretEnabled: boolFalse}}},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := parseBuildArguments(tc.input, tc.secret)
			if len(got) != len(tc.want) {
				t.Fatalf("len=%d want %d: %v", len(got), len(tc.want), got)
			}
			for i := range got {
				if got[i].Name != tc.want[i].Name || got[i].Value != tc.want[i].Value || got[i].SecretEnabled != tc.want[i].SecretEnabled {
					t.Errorf("[%d] got {%s=%s secret=%v} want {%s=%s secret=%v}",
						i, got[i].Name, got[i].Value, got[i].SecretEnabled,
						tc.want[i].Name, tc.want[i].Value, tc.want[i].SecretEnabled)
				}
			}
		})
	}
}

func TestUnitBuildResource_buildArgsToString(t *testing.T) {
	boolFalse := types.BoolValue(false)
	cases := []struct {
		name  string
		input []BuildArgumentModel
		want  string
	}{
		{"empty", nil, ""},
		{"single", []BuildArgumentModel{{Name: strVal("KEY"), Value: strVal("v"), SecretEnabled: boolFalse}}, "KEY=v"},
		{"sorted", []BuildArgumentModel{
			{Name: strVal("Z"), Value: strVal("last"), SecretEnabled: boolFalse},
			{Name: strVal("A"), Value: strVal("first"), SecretEnabled: boolFalse},
		}, "A=first\nZ=last"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := buildArgsToString(tc.input)
			if got != tc.want {
				t.Errorf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestUnitBuildResource_argsRoundTrip(t *testing.T) {
	boolFalse := types.BoolValue(false)
	args := []BuildArgumentModel{
		{Name: strVal("LOG_LEVEL"), Value: strVal("info"), SecretEnabled: boolFalse},
		{Name: strVal("BUILD_ENV"), Value: strVal("production"), SecretEnabled: boolFalse},
	}
	serialised := buildArgsToString(args)
	parsed := parseBuildArguments(serialised, false)
	if len(parsed) != len(args) {
		t.Fatalf("got %d entries want %d", len(parsed), len(args))
	}
	// both are sorted by name
	expected := []BuildArgumentModel{
		{Name: strVal("BUILD_ENV"), Value: strVal("production"), SecretEnabled: boolFalse},
		{Name: strVal("LOG_LEVEL"), Value: strVal("info"), SecretEnabled: boolFalse},
	}
	for i := range parsed {
		if parsed[i].Name != expected[i].Name || parsed[i].Value != expected[i].Value || parsed[i].SecretEnabled != expected[i].SecretEnabled {
			t.Errorf("[%d] got {%s=%s secret=%v} want {%s=%s secret=%v}",
				i, parsed[i].Name, parsed[i].Value, parsed[i].SecretEnabled,
				expected[i].Name, expected[i].Value, expected[i].SecretEnabled)
		}
	}
}

// strVal is a test helper to create a types.String value.
func strVal(s string) types.String { return types.StringValue(s) }

// ---------------------------------------------------------------------------
// Acceptance tests – build.argument / labels / dockerfile
// round-trip without drift (regression for trailing-newline bug)
// ---------------------------------------------------------------------------

func TestAccBuildResource_argsNoDrift(t *testing.T) {
	const name = "tf-acc-build-args-drift"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with argument blocks - second plan must be empty (no drift).
			{
				Config: fmt.Sprintf(`
resource "komodo_build" "test" {
  name = %q
  build {
    argument {
      name  = "BUILD_ENV"
      value = "production"
    }
  }
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_build.test", "build.argument.0.name", "BUILD_ENV"),
					resource.TestCheckResourceAttr("komodo_build.test", "build.argument.0.value", "production"),
				),
			},
			// ExpectNonEmptyPlan would fail this step if there is drift.
			{
				Config: fmt.Sprintf(`
resource "komodo_build" "test" {
  name = %q
  build {
    argument {
      name  = "BUILD_ENV"
      value = "production"
    }
  }
}
`, name),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			// Multiple arguments.
			{
				Config: fmt.Sprintf(`
resource "komodo_build" "test" {
  name = %q
  build {
    argument {
      name  = "BUILD_ENV"
      value = "production"
    }
    argument {
      name  = "LOG_LEVEL"
      value = "info"
    }
  }
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_build.test", "build.argument.#", "2"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "komodo_build" "test" {
  name = %q
  build {
    argument {
      name  = "BUILD_ENV"
      value = "production"
    }
    argument {
      name  = "LOG_LEVEL"
      value = "info"
    }
  }
}
`, name),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccBuildResource_labelsNoDrift(t *testing.T) {
	const name = "tf-acc-build-labels-drift"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "komodo_build" "test" {
  name   = %q
  labels = "maintainer=team"
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_build.test", "labels", "maintainer=team"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "komodo_build" "test" {
  name   = %q
  labels = "maintainer=team"
}
`, name),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

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
				// webhook.secret and build.secret_argument are sensitive and may not round-trip
				ImportStateVerifyIgnore: []string{"webhook.secret", "build.secret_argument"},
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
    extra_arguments = ["--no-cache"]
    argument {
      name  = "ENV"
      value = "production"
    }
  }
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_build.test", "build.path", "./app"),
					resource.TestCheckResourceAttr("komodo_build.test", "build.extra_arguments.#", "1"),
					resource.TestCheckResourceAttr("komodo_build.test", "build.extra_arguments.0", "--no-cache"),
					resource.TestCheckResourceAttr("komodo_build.test", "build.argument.0.name", "ENV"),
					resource.TestCheckResourceAttr("komodo_build.test", "build.argument.0.value", "production"),
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
					resource.TestCheckResourceAttr("komodo_build.test", "build.extra_arguments.#", "0"),
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
