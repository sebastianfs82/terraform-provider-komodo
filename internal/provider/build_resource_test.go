// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
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

// ---------------------------------------------------------------------------
// build.buildx_enabled
// ---------------------------------------------------------------------------

func TestAccBuildResource_buildxEnabled(t *testing.T) {
	const name = "tf-acc-build-buildx"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Enable buildx inside the build block
			{
				Config: fmt.Sprintf(`
resource "komodo_build" "test" {
  name = %q
  build {
    buildx_enabled = true
  }
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_build.test", "build.buildx_enabled", "true"),
					resource.TestCheckResourceAttrSet("komodo_build.test", "id"),
				),
			},
			// Second plan must be empty (no drift)
			{
				Config: fmt.Sprintf(`
resource "komodo_build" "test" {
  name = %q
  build {
    buildx_enabled = true
  }
}
`, name),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			// Disable buildx
			{
				Config: fmt.Sprintf(`
resource "komodo_build" "test" {
  name = %q
  build {
    buildx_enabled = false
  }
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_build.test", "build.buildx_enabled", "false"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// image.dockerfile block
// ---------------------------------------------------------------------------

func TestAccBuildResource_imageDockerfile(t *testing.T) {
	const name = "tf-acc-build-dockerfile"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Set a custom Dockerfile path inside image block
			{
				Config: fmt.Sprintf(`
resource "komodo_build" "test" {
  name = %q
  image {
    name = "myorg/my-service"
    dockerfile {
      path = "docker/Dockerfile.prod"
    }
  }
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_build.test", "image.dockerfile.path", "docker/Dockerfile.prod"),
					resource.TestCheckResourceAttrSet("komodo_build.test", "id"),
				),
			},
			// Second plan must be empty (no drift)
			{
				Config: fmt.Sprintf(`
resource "komodo_build" "test" {
  name = %q
  image {
    name = "myorg/my-service"
    dockerfile {
      path = "docker/Dockerfile.prod"
    }
  }
}
`, name),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			// Update dockerfile path
			{
				Config: fmt.Sprintf(`
resource "komodo_build" "test" {
  name = %q
  image {
    name = "myorg/my-service"
    dockerfile {
      path = "Dockerfile"
    }
  }
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_build.test", "image.dockerfile.path", "Dockerfile"),
				),
			},
			// Remove dockerfile block
			{
				Config: fmt.Sprintf(`
resource "komodo_build" "test" {
  name = %q
  image {
    name = "myorg/my-service"
  }
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("komodo_build.test", "image.dockerfile.path"),
				),
			},
		},
	})
}

// ─── Unit tests ──────────────────────────────────────────────────────────────

func wrongRawBuildPlan(t *testing.T, r *BuildResource) tfsdk.Plan {
	t.Helper()
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schemaResp.Schema,
	}
}

func wrongRawBuildState(t *testing.T, r *BuildResource) tfsdk.State {
	t.Helper()
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	return tfsdk.State{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schemaResp.Schema,
	}
}

func TestUnitBuildResource_configure(t *testing.T) {
	t.Run("wrong_type", func(t *testing.T) {
		r := &BuildResource{}
		req := fwresource.ConfigureRequest{ProviderData: "not-a-client"}
		resp := &fwresource.ConfigureResponse{}
		r.Configure(context.Background(), req, resp)
		if !resp.Diagnostics.HasError() {
			t.Fatal("expected diagnostic error for wrong ProviderData type")
		}
	})
}

func TestUnitBuildResource_createPlanGetError(t *testing.T) {
	r := &BuildResource{client: &client.Client{}}
	req := fwresource.CreateRequest{Plan: wrongRawBuildPlan(t, r)}
	resp := &fwresource.CreateResponse{}
	r.Create(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed plan")
	}
}

func TestUnitBuildResource_readStateGetError(t *testing.T) {
	r := &BuildResource{client: &client.Client{}}
	req := fwresource.ReadRequest{State: wrongRawBuildState(t, r)}
	resp := &fwresource.ReadResponse{}
	r.Read(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed state")
	}
}

func TestUnitBuildResource_updatePlanGetError(t *testing.T) {
	r := &BuildResource{client: &client.Client{}}
	req := fwresource.UpdateRequest{Plan: wrongRawBuildPlan(t, r)}
	resp := &fwresource.UpdateResponse{}
	r.Update(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed plan")
	}
}

func TestUnitBuildResource_deleteStateGetError(t *testing.T) {
	r := &BuildResource{client: &client.Client{}}
	req := fwresource.DeleteRequest{State: wrongRawBuildState(t, r)}
	resp := &fwresource.DeleteResponse{}
	r.Delete(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed state")
	}
}

func TestUnitBuildResource_matchPriorOrder(t *testing.T) {
	makeArgs := func(names ...string) []BuildArgumentModel {
		out := make([]BuildArgumentModel, len(names))
		for i, n := range names {
			out[i] = BuildArgumentModel{
				Name:          types.StringValue(n),
				Value:         types.StringValue(n + "_val"),
				SecretEnabled: types.BoolValue(false),
			}
		}
		return out
	}

	t.Run("empty_prior_returns_api_order", func(t *testing.T) {
		api := makeArgs("B", "A")
		result := matchPriorOrder(nil, api)
		if len(result) != 2 || result[0].Name.ValueString() != "B" || result[1].Name.ValueString() != "A" {
			t.Fatal("expected original B,A order preserved when prior is nil")
		}
	})

	t.Run("prior_order_applied", func(t *testing.T) {
		prior := makeArgs("A", "B", "C")
		api := makeArgs("C", "B", "A")
		result := matchPriorOrder(prior, api)
		if result[0].Name.ValueString() != "A" || result[1].Name.ValueString() != "B" || result[2].Name.ValueString() != "C" {
			t.Fatalf("expected prior order A,B,C but got %s,%s,%s",
				result[0].Name.ValueString(), result[1].Name.ValueString(), result[2].Name.ValueString())
		}
	})

	t.Run("new_args_appended_after_prior", func(t *testing.T) {
		prior := makeArgs("A")
		api := makeArgs("B", "A")
		result := matchPriorOrder(prior, api)
		if len(result) != 2 || result[0].Name.ValueString() != "A" || result[1].Name.ValueString() != "B" {
			t.Fatalf("expected A first then B, got %v", result)
		}
	})

	t.Run("prior_entries_absent_in_api_are_skipped", func(t *testing.T) {
		prior := makeArgs("A", "REMOVED")
		api := makeArgs("A")
		result := matchPriorOrder(prior, api)
		if len(result) != 1 || result[0].Name.ValueString() != "A" {
			t.Fatalf("expected only A in result, got %v", result)
		}
	})
}

func TestUnitBuildResource_partialBuildConfigFromModel(t *testing.T) {
	ctx := context.Background()
	c := &client.Client{}

	t.Run("no_version_block_resets_to_zero_with_auto_increment", func(t *testing.T) {
		data := &BuildResourceModel{
			BuilderID:        types.StringNull(),
			Version:          nil,
			Image:            nil,
			Links:            types.ListNull(types.StringType),
			Source:           nil,
			Webhook:          nil,
			Build:            nil,
			PreBuild:         nil,
			Labels:           NewTrimmedStringNull(),
			SkipSecretInterp: types.BoolNull(),
		}
		cfg := partialBuildConfigFromModel(ctx, c, data)
		if cfg.Version == nil || cfg.Version.Major != 0 || cfg.Version.Minor != 0 || cfg.Version.Patch != 0 {
			t.Fatal("expected version reset to 0.0.0 when no version block")
		}
		if cfg.AutoIncrementVersion == nil || !*cfg.AutoIncrementVersion {
			t.Fatal("expected AutoIncrementVersion=true when no version block")
		}
	})

	t.Run("with_version_string_parsed", func(t *testing.T) {
		data := &BuildResourceModel{
			BuilderID: types.StringNull(),
			Version: &BuildVersionModel{
				Value:                types.StringValue("2.3.4"),
				AutoIncrementEnabled: types.BoolValue(false),
			},
			Image:            nil,
			Links:            types.ListNull(types.StringType),
			Source:           nil,
			Webhook:          nil,
			Build:            nil,
			PreBuild:         nil,
			Labels:           NewTrimmedStringNull(),
			SkipSecretInterp: types.BoolNull(),
		}
		cfg := partialBuildConfigFromModel(ctx, c, data)
		if cfg.Version == nil || cfg.Version.Major != 2 || cfg.Version.Minor != 3 || cfg.Version.Patch != 4 {
			t.Fatalf("expected version 2.3.4, got %v", cfg.Version)
		}
		if cfg.AutoIncrementVersion == nil || *cfg.AutoIncrementVersion {
			t.Fatal("expected AutoIncrementVersion=false")
		}
	})

	t.Run("no_source_block_clears_git_fields", func(t *testing.T) {
		data := &BuildResourceModel{
			BuilderID:        types.StringNull(),
			Version:          nil,
			Image:            nil,
			Links:            types.ListNull(types.StringType),
			Source:           nil,
			Webhook:          nil,
			Build:            nil,
			PreBuild:         nil,
			Labels:           NewTrimmedStringNull(),
			SkipSecretInterp: types.BoolNull(),
		}
		cfg := partialBuildConfigFromModel(ctx, c, data)
		if cfg.LinkedRepo == nil || *cfg.LinkedRepo != "" {
			t.Fatalf("expected empty LinkedRepo, got %v", cfg.LinkedRepo)
		}
		if cfg.GitProvider == nil || *cfg.GitProvider != "" {
			t.Fatalf("expected empty GitProvider, got %v", cfg.GitProvider)
		}
		if cfg.Repo == nil || *cfg.Repo != "" {
			t.Fatalf("expected empty Repo, got %v", cfg.Repo)
		}
	})

	t.Run("with_https_source_url", func(t *testing.T) {
		data := &BuildResourceModel{
			BuilderID: types.StringNull(),
			Version:   nil,
			Image:     nil,
			Links:     types.ListNull(types.StringType),
			Source: &BuildSourceModel{
				RepoID:      types.StringNull(),
				URL:         types.StringValue("https://github.com"),
				AccountID:   types.StringNull(),
				Path:        types.StringValue("owner/repo"),
				Branch:      types.StringValue("main"),
				Commit:      types.StringNull(),
				FilesOnHost: types.BoolValue(false),
			},
			Webhook:          nil,
			Build:            nil,
			PreBuild:         nil,
			Labels:           NewTrimmedStringNull(),
			SkipSecretInterp: types.BoolNull(),
		}
		cfg := partialBuildConfigFromModel(ctx, c, data)
		if cfg.GitProvider == nil || *cfg.GitProvider != "github.com" {
			t.Fatalf("expected GitProvider=github.com, got %v", cfg.GitProvider)
		}
		if cfg.GitHttps == nil || !*cfg.GitHttps {
			t.Fatal("expected GitHttps=true for https URL")
		}
		if cfg.Repo == nil || *cfg.Repo != "owner/repo" {
			t.Fatalf("expected Repo=owner/repo, got %v", cfg.Repo)
		}
		if cfg.Branch == nil || *cfg.Branch != "main" {
			t.Fatalf("expected Branch=main, got %v", cfg.Branch)
		}
	})

	t.Run("with_build_args_split_plain_and_secret", func(t *testing.T) {
		data := &BuildResourceModel{
			BuilderID: types.StringNull(),
			Version:   nil,
			Image:     nil,
			Links:     types.ListNull(types.StringType),
			Source:    nil,
			Webhook:   nil,
			Build: &DockerBuildModel{
				Path:           types.StringValue("."),
				ExtraArguments: types.ListNull(types.StringType),
				Arguments: []BuildArgumentModel{
					{Name: types.StringValue("FOO"), Value: types.StringValue("bar"), SecretEnabled: types.BoolValue(false)},
					{Name: types.StringValue("SEC"), Value: types.StringValue("s3cr3t"), SecretEnabled: types.BoolValue(true)},
				},
				UseBuildx: types.BoolNull(),
			},
			PreBuild:         nil,
			Labels:           NewTrimmedStringNull(),
			SkipSecretInterp: types.BoolNull(),
		}
		cfg := partialBuildConfigFromModel(ctx, c, data)
		if cfg.BuildArgs == nil || !strings.Contains(*cfg.BuildArgs, "FOO=bar") {
			t.Fatalf("expected FOO=bar in build_args, got %v", cfg.BuildArgs)
		}
		if cfg.SecretArgs == nil || !strings.Contains(*cfg.SecretArgs, "SEC=s3cr3t") {
			t.Fatalf("expected SEC=s3cr3t in secret_args, got %v", cfg.SecretArgs)
		}
	})

	t.Run("with_pre_build_command", func(t *testing.T) {
		data := &BuildResourceModel{
			BuilderID: types.StringNull(),
			Version:   nil,
			Image:     nil,
			Links:     types.ListNull(types.StringType),
			Source:    nil,
			Webhook:   nil,
			Build:     nil,
			PreBuild: &SystemCommandModel{
				Path:    types.StringValue("/scripts"),
				Command: NewTrimmedStringValue("./build.sh"),
			},
			Labels:           NewTrimmedStringNull(),
			SkipSecretInterp: types.BoolNull(),
		}
		cfg := partialBuildConfigFromModel(ctx, c, data)
		if cfg.PreBuild == nil || cfg.PreBuild.Path != "/scripts" || cfg.PreBuild.Command != "./build.sh" {
			t.Fatalf("unexpected pre_build: %v", cfg.PreBuild)
		}
	})

	t.Run("with_labels", func(t *testing.T) {
		data := &BuildResourceModel{
			BuilderID:        types.StringNull(),
			Version:          nil,
			Image:            nil,
			Links:            types.ListNull(types.StringType),
			Source:           nil,
			Webhook:          nil,
			Build:            nil,
			PreBuild:         nil,
			Labels:           NewTrimmedStringValue("env=prod\nteam=infra"),
			SkipSecretInterp: types.BoolNull(),
		}
		cfg := partialBuildConfigFromModel(ctx, c, data)
		if cfg.Labels == nil || *cfg.Labels != "env=prod\nteam=infra" {
			t.Fatalf("unexpected labels: %v", cfg.Labels)
		}
	})
}

func TestUnitBuildResource_buildToModel(t *testing.T) {
	ctx := context.Background()
	c := &client.Client{}

	t.Run("basic_fields", func(t *testing.T) {
		b := &client.Build{
			ID:   client.OID{OID: "build001"},
			Name: "my-build",
			Tags: []string{"t1"},
			Config: client.BuildConfig{
				BuilderID: "builder-xyz",
				Labels:    "env=prod",
			},
		}
		data := &BuildResourceModel{Tags: types.ListValueMust(types.StringType, nil)}
		buildToModel(ctx, c, b, data)
		if data.ID.ValueString() != "build001" {
			t.Fatalf("unexpected ID: %s", data.ID.ValueString())
		}
		if data.Name.ValueString() != "my-build" {
			t.Fatalf("unexpected Name: %s", data.Name.ValueString())
		}
		if data.BuilderID.ValueString() != "builder-xyz" {
			t.Fatalf("unexpected BuilderID: %s", data.BuilderID.ValueString())
		}
		if data.Labels.ValueString() != "env=prod" {
			t.Fatalf("unexpected Labels: %s", data.Labels.ValueString())
		}
		if len(data.Tags.Elements()) != 1 {
			t.Fatalf("expected 1 tag, got %d", len(data.Tags.Elements()))
		}
	})

	t.Run("version_block_populated_when_set", func(t *testing.T) {
		b := &client.Build{
			ID:   client.OID{OID: "build002"},
			Name: "versioned",
			Tags: []string{},
			Config: client.BuildConfig{
				Version:              client.BuildVersion{Major: 2, Minor: 3, Patch: 4},
				AutoIncrementVersion: false,
			},
		}
		data := &BuildResourceModel{
			Tags:    types.ListValueMust(types.StringType, nil),
			Version: &BuildVersionModel{},
		}
		buildToModel(ctx, c, b, data)
		if data.Version == nil {
			t.Fatal("expected non-nil version block")
		}
		if data.Version.Value.ValueString() != "2.3.4" {
			t.Fatalf("expected version=2.3.4, got %s", data.Version.Value.ValueString())
		}
		if data.Version.AutoIncrementEnabled.ValueBool() {
			t.Fatal("expected auto_increment_enabled=false")
		}
	})

	t.Run("version_block_stays_nil_when_not_set", func(t *testing.T) {
		b := &client.Build{
			ID:   client.OID{OID: "build003"},
			Name: "no-version",
			Tags: []string{},
			Config: client.BuildConfig{
				Version: client.BuildVersion{Major: 1, Minor: 0, Patch: 0},
			},
		}
		data := &BuildResourceModel{
			Tags:    types.ListValueMust(types.StringType, nil),
			Version: nil,
		}
		buildToModel(ctx, c, b, data)
		if data.Version != nil {
			t.Fatal("expected nil version block when not previously set in model")
		}
	})

	t.Run("source_set_when_repo_non_empty", func(t *testing.T) {
		b := &client.Build{
			ID:   client.OID{OID: "build004"},
			Name: "repo-build",
			Tags: []string{},
			Config: client.BuildConfig{
				Repo:   "owner/myrepo",
				Branch: "main",
			},
		}
		data := &BuildResourceModel{
			Tags:   types.ListValueMust(types.StringType, nil),
			Source: nil,
		}
		buildToModel(ctx, c, b, data)
		if data.Source == nil {
			t.Fatal("expected source block set when Repo is non-empty")
		}
		if data.Source.Path.ValueString() != "owner/myrepo" {
			t.Fatalf("expected source.path=owner/myrepo, got %s", data.Source.Path.ValueString())
		}
		if data.Source.Branch.ValueString() != "main" {
			t.Fatalf("expected source.branch=main, got %s", data.Source.Branch.ValueString())
		}
	})

	t.Run("source_nil_when_git_empty_and_prior_nil", func(t *testing.T) {
		b := &client.Build{
			ID:     client.OID{OID: "build005"},
			Name:   "no-source",
			Tags:   []string{},
			Config: client.BuildConfig{},
		}
		data := &BuildResourceModel{
			Tags:   types.ListValueMust(types.StringType, nil),
			Source: nil,
		}
		buildToModel(ctx, c, b, data)
		if data.Source != nil {
			t.Fatal("expected nil source when all git fields empty and prior Source nil")
		}
	})

	t.Run("pre_build_set_when_non_empty", func(t *testing.T) {
		b := &client.Build{
			ID:   client.OID{OID: "build006"},
			Name: "pre-build",
			Tags: []string{},
			Config: client.BuildConfig{
				PreBuild: client.SystemCommand{Path: "/src", Command: "make build\n"},
			},
		}
		data := &BuildResourceModel{Tags: types.ListValueMust(types.StringType, nil)}
		buildToModel(ctx, c, b, data)
		if data.PreBuild == nil {
			t.Fatal("expected non-nil pre_build block")
		}
		if data.PreBuild.Path.ValueString() != "/src" {
			t.Fatalf("expected pre_build.path=/src, got %s", data.PreBuild.Path.ValueString())
		}
		// Trailing newline stripped from Command.
		if data.PreBuild.Command.ValueString() != "make build" {
			t.Fatalf("expected trailing newline stripped, got %q", data.PreBuild.Command.ValueString())
		}
	})

	t.Run("build_block_args_round_trip", func(t *testing.T) {
		b := &client.Build{
			ID:   client.OID{OID: "build007"},
			Name: "with-args",
			Tags: []string{},
			Config: client.BuildConfig{
				BuildPath: ".",
				BuildArgs: "BAZ=qux\nFOO=bar",
			},
		}
		data := &BuildResourceModel{
			Tags:  types.ListValueMust(types.StringType, nil),
			Build: &DockerBuildModel{},
		}
		buildToModel(ctx, c, b, data)
		if data.Build == nil {
			t.Fatal("expected non-nil build block")
		}
		if len(data.Build.Arguments) != 2 {
			t.Fatalf("expected 2 arguments, got %d", len(data.Build.Arguments))
		}
		names := map[string]bool{}
		for _, a := range data.Build.Arguments {
			names[a.Name.ValueString()] = true
		}
		if !names["FOO"] || !names["BAZ"] {
			t.Fatalf("expected FOO and BAZ in arguments, got %v", names)
		}
	})

	t.Run("linked_repo_clears_direct_fields", func(t *testing.T) {
		b := &client.Build{
			ID:   client.OID{OID: "build-linked"},
			Name: "linked-build",
			Tags: []string{},
			Config: client.BuildConfig{
				LinkedRepo: "my-repo-id",
			},
		}
		data := &BuildResourceModel{
			Tags:   types.ListValueMust(types.StringType, nil),
			Source: &BuildSourceModel{},
		}
		buildToModel(ctx, c, b, data)
		if data.Source == nil {
			t.Fatal("expected source block when LinkedRepo is set")
		}
		if data.Source.RepoID.ValueString() != "my-repo-id" {
			t.Fatalf("expected repo_id=my-repo-id, got %s", data.Source.RepoID.ValueString())
		}
		// Direct fields should be null when linked repo is used.
		if !data.Source.URL.IsNull() {
			t.Fatalf("expected url null for linked repo, got %s", data.Source.URL.ValueString())
		}
	})

	t.Run("files_on_host_set", func(t *testing.T) {
		b := &client.Build{
			ID:   client.OID{OID: "build-foh"},
			Name: "foh-build",
			Tags: []string{},
			Config: client.BuildConfig{
				FilesOnHost: true,
			},
		}
		data := &BuildResourceModel{
			Tags:   types.ListValueMust(types.StringType, nil),
			Source: &BuildSourceModel{},
		}
		buildToModel(ctx, c, b, data)
		if data.Source == nil {
			t.Fatal("expected source block when FilesOnHost is true")
		}
		if !data.Source.FilesOnHost.ValueBool() {
			t.Fatal("expected files_on_host=true")
		}
	})
}

func TestUnitBuildResource_partialBuildConfigFromModel_nilBranches(t *testing.T) {
	ctx := context.Background()
	c := &client.Client{}

	t.Run("version_nil_resets_to_zero", func(t *testing.T) {
		data := &BuildResourceModel{
			BuilderID: types.StringNull(),
			Version:   nil, // no version block → should reset to 0.0.0
			Links:     types.ListNull(types.StringType),
		}
		cfg := partialBuildConfigFromModel(ctx, c, data)
		if cfg.Version == nil {
			t.Fatal("expected non-nil version when Version block is nil")
		}
		if cfg.Version.Major != 0 || cfg.Version.Minor != 0 || cfg.Version.Patch != 0 {
			t.Fatalf("expected 0.0.0, got %d.%d.%d", cfg.Version.Major, cfg.Version.Minor, cfg.Version.Patch)
		}
		if cfg.AutoIncrementVersion == nil || !*cfg.AutoIncrementVersion {
			t.Fatal("expected AutoIncrementVersion=true when version block is nil")
		}
	})

	t.Run("source_nil_clears_all_git_fields", func(t *testing.T) {
		data := &BuildResourceModel{
			BuilderID: types.StringNull(),
			Source:    nil,
			Links:     types.ListNull(types.StringType),
		}
		cfg := partialBuildConfigFromModel(ctx, c, data)
		if cfg.LinkedRepo == nil || *cfg.LinkedRepo != "" {
			t.Fatal("expected empty LinkedRepo when Source is nil")
		}
		if cfg.Repo == nil || *cfg.Repo != "" {
			t.Fatal("expected empty Repo when Source is nil")
		}
		if cfg.Branch == nil || *cfg.Branch != "" {
			t.Fatal("expected empty Branch when Source is nil")
		}
	})

	t.Run("webhook_nil_clears_webhook_fields", func(t *testing.T) {
		data := &BuildResourceModel{
			BuilderID: types.StringNull(),
			Webhook:   nil,
			Links:     types.ListNull(types.StringType),
		}
		cfg := partialBuildConfigFromModel(ctx, c, data)
		if cfg.WebhookEnabled == nil || *cfg.WebhookEnabled {
			t.Fatal("expected WebhookEnabled=false when Webhook is nil")
		}
		if cfg.WebhookSecret == nil || *cfg.WebhookSecret != "" {
			t.Fatal("expected empty WebhookSecret when Webhook is nil")
		}
	})
}

func TestUnitBuildResource_gitRepoConflictsValidator(t *testing.T) {
	cases := []struct {
		name   string
		config string
	}{
		{
			name: "repo_id_with_url",
			config: `
resource "komodo_build" "test" {
  name = "tf-test-build-conflict"
  source {
    repo_id = "my-git-repo"
    url     = "https://github.com"
  }
}`,
		},
		{
			name: "repo_id_with_path",
			config: `
resource "komodo_build" "test" {
  name = "tf-test-build-conflict"
  source {
    repo_id = "my-git-repo"
    path    = "owner/repo"
  }
}`,
		},
		{
			name: "repo_id_with_branch",
			config: `
resource "komodo_build" "test" {
  name = "tf-test-build-conflict"
  source {
    repo_id = "my-git-repo"
    branch  = "main"
  }
}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config:      tc.config,
						ExpectError: regexp.MustCompile(`source\.repo_id conflicts with other source fields`),
					},
				},
			})
		})
	}
}

// ─── Acceptance tests – line endings idempotency ─────────────────────────────

// TestAccBuildResource_labelsLineEndingsNoDrift verifies that a labels value
// with a trailing newline does not produce plan drift on re-plan. State mirrors
// the config value; SemanticEquals prevents unnecessary diffs.
func TestAccBuildResource_labelsLineEndingsNoDrift(t *testing.T) {
	const name = "tf-acc-build-labels-le"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "komodo_build" "test" {
  name   = %q
  labels = "maintainer=team\n"
}
`, name),
				Check: resource.TestCheckResourceAttr("komodo_build.test", "labels", "maintainer=team\n"),
			},
			{
				Config: fmt.Sprintf(`
resource "komodo_build" "test" {
  name   = %q
  labels = "maintainer=team\n"
}
`, name),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			// CRLF variant: multiline labels with CRLF endings.
			{
				Config: fmt.Sprintf(`
resource "komodo_build" "test" {
  name   = %q
  labels = "maintainer=team\r\nenv=prod\r\n"
}
`, name),
				Check: resource.TestCheckResourceAttr("komodo_build.test", "labels", "maintainer=team\r\nenv=prod\r\n"),
			},
			{
				Config: fmt.Sprintf(`
resource "komodo_build" "test" {
  name   = %q
  labels = "maintainer=team\r\nenv=prod\r\n"
}
`, name),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// TestAccBuildResource_preBuildCommandLineEndingsNoDrift verifies that
// pre_build.command with trailing LF or CRLF is stored trimmed and does not
// drift on re-plan.
func TestAccBuildResource_preBuildCommandLineEndingsNoDrift(t *testing.T) {
	const name = "tf-acc-build-prebuild-le"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "komodo_build" "test" {
  name = %q
  pre_build {
    command = "make test\n"
  }
}
`, name),
				Check: resource.TestCheckResourceAttr("komodo_build.test", "pre_build.command", "make test\n"),
			},
			{
				Config: fmt.Sprintf(`
resource "komodo_build" "test" {
  name = %q
  pre_build {
    command = "make test\n"
  }
}
`, name),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			// CRLF variant.
			{
				Config: fmt.Sprintf(`
resource "komodo_build" "test" {
  name = %q
  pre_build {
    command = "step1\r\nstep2\r\n"
  }
}
`, name),
				Check: resource.TestCheckResourceAttr("komodo_build.test", "pre_build.command", "step1\r\nstep2\r\n"),
			},
			{
				Config: fmt.Sprintf(`
resource "komodo_build" "test" {
  name = %q
  pre_build {
    command = "step1\r\nstep2\r\n"
  }
}
`, name),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// TestAccBuildResource_dockerfileContentsLineEndingsNoDrift verifies that
// image.dockerfile.contents with trailing LF or CRLF is stored trimmed and
// does not drift on re-plan.
func TestAccBuildResource_dockerfileContentsLineEndingsNoDrift(t *testing.T) {
	const name = "tf-acc-build-dockerfile-le"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "komodo_build" "test" {
  name = %q
  image {
    dockerfile {
      contents = "FROM ubuntu:22.04\n"
    }
  }
}
`, name),
				Check: resource.TestCheckResourceAttr("komodo_build.test", "image.dockerfile.contents", "FROM ubuntu:22.04\n"),
			},
			{
				Config: fmt.Sprintf(`
resource "komodo_build" "test" {
  name = %q
  image {
    dockerfile {
      contents = "FROM ubuntu:22.04\n"
    }
  }
}
`, name),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			// CRLF variant.
			{
				Config: fmt.Sprintf(`
resource "komodo_build" "test" {
  name = %q
  image {
    dockerfile {
      contents = "FROM ubuntu:22.04\r\nRUN apt-get update\r\n"
    }
  }
}
`, name),
				Check: resource.TestCheckResourceAttr("komodo_build.test", "image.dockerfile.contents", "FROM ubuntu:22.04\r\nRUN apt-get update\r\n"),
			},
			{
				Config: fmt.Sprintf(`
resource "komodo_build" "test" {
  name = %q
  image {
    dockerfile {
      contents = "FROM ubuntu:22.04\r\nRUN apt-get update\r\n"
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
