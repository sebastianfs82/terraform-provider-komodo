// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"sync"
	"testing"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccStackResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStackResourceConfig_basic("tf-test-stack"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_stack.test", "name", "tf-test-stack"),
					resource.TestCheckResourceAttrSet("komodo_stack.test", "id"),
				),
			},
		},
	})
}

func TestAccStackResource_withFileContents(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStackResourceConfig_withFileContents("tf-test-stack-inline"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_stack.test", "name", "tf-test-stack-inline"),
					resource.TestCheckResourceAttrSet("komodo_stack.test", "id"),
					resource.TestCheckResourceAttrSet("komodo_stack.test", "source.contents"),
				),
			},
		},
	})
}

func TestAccStackResource_withGit(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStackResourceConfig_withGit(
					"tf-test-stack-git",
					"owner/my-stack-repo",
					"main",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_stack.test", "name", "tf-test-stack-git"),
					resource.TestCheckResourceAttr("komodo_stack.test", "source.url", "https://github.com"),
					resource.TestCheckResourceAttr("komodo_stack.test", "source.path", "owner/my-stack-repo"),
					resource.TestCheckResourceAttr("komodo_stack.test", "source.branch", "main"),
					resource.TestCheckResourceAttrSet("komodo_stack.test", "id"),
				),
			},
		},
	})
}

// TestAccStackResource_gitRecloneComputed verifies that omitting reclone from a
// git block does not cause a "provider produced inconsistent result" error.
// The API always returns the zero value (false) for reclone; without Computed:true
// on the schema attribute, the planned null vs actual false mismatch was fatal.
func TestAccStackResource_gitRecloneComputed(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Apply without reclone — must succeed without an inconsistency error.
			{
				Config: testAccStackResourceConfig_withGit(
					"tf-test-stack-reclone",
					"owner/my-stack-repo",
					"main",
				),
				Check: resource.TestCheckResourceAttr("komodo_stack.test", "source.reclone_enabled", "false"),
			},
			// Re-plan with the same config — must produce an empty diff.
			{
				Config:             testAccStackResourceConfig_withGit("tf-test-stack-reclone", "owner/my-stack-repo", "main"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccStackResource_withPreDeploy(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStackResourceConfig_withPreDeploy("tf-test-stack-pre"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_stack.test", "name", "tf-test-stack-pre"),
					resource.TestCheckResourceAttr("komodo_stack.test", "pre_deploy.command", "echo pre"),
					resource.TestCheckResourceAttrSet("komodo_stack.test", "id"),
				),
			},
		},
	})
}

func TestAccStackResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStackResourceConfig_withGit("tf-update-stack", "owner/repo", "main"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_stack.test", "source.branch", "main"),
				),
			},
			{
				Config: testAccStackResourceConfig_withGit("tf-update-stack", "owner/repo", "develop"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_stack.test", "source.branch", "develop"),
				),
			},
		},
	})
}

func TestAccStackResource_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStackResourceConfig_basic("tf-import-stack"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_stack.test", "name", "tf-import-stack"),
					resource.TestCheckResourceAttrSet("komodo_stack.test", "id"),
				),
			},
			{
				ResourceName:      "komodo_stack.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccStackResource_disappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStackResourceConfig_basic("tf-disappear-stack"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_stack.test", "id"),
					testAccStackDisappears("komodo_stack.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccStackResource_rename(t *testing.T) {
	var savedID string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStackResourceConfig_basic("tf-acc-stack-rename-orig"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_stack.test", "name", "tf-acc-stack-rename-orig"),
					resource.TestCheckResourceAttrSet("komodo_stack.test", "id"),
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["komodo_stack.test"]
						savedID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config: testAccStackResourceConfig_basic("tf-acc-stack-rename-new"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_stack.test", "name", "tf-acc-stack-rename-new"),
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["komodo_stack.test"]
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

func testAccStackDisappears(resourceName string) resource.TestCheckFunc {
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
		return c.DeleteStack(context.Background(), rs.Primary.ID)
	}
}

// TestAccStackResource_alertsEnabledDefault verifies that omitting alerts_enabled
// after it was explicitly set to false causes Terraform to plan a change back to
// the default value of true.
func TestAccStackResource_alertsEnabledDefault(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: create with alerts_enabled explicitly false.
			{
				Config: testAccStackResourceConfig_withAlertsEnabled("tf-test-stack-alerts-default", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_stack.test", "alerts_enabled", "false"),
				),
			},
			// Step 2: remove alerts_enabled from config → default kicks in, must plan true.
			{
				Config: testAccStackResourceConfig_basic("tf-test-stack-alerts-default"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_stack.test", "alerts_enabled", "true"),
				),
			},
			// Step 3: re-plan with same config → no further changes.
			{
				Config:             testAccStackResourceConfig_basic("tf-test-stack-alerts-default"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccStackResourceConfig_withAlertsEnabled(name string, alertsEnabled bool) string {
	return fmt.Sprintf(`
resource "komodo_stack" "test" {
  name           = "%s"
  alerts_enabled = %t
}
`, name, alertsEnabled)
}

// Test configuration functions

func testAccStackResourceConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "komodo_stack" "test" {
  name = "%s"
}
`, name)
}

func testAccStackResourceConfig_withFileContents(name string) string {
	return fmt.Sprintf(`
resource "komodo_stack" "test" {
  name = "%s"

  source {
    contents = <<-EOT
      services:
        web:
          image: nginx:latest
    EOT
  }
}
`, name)
}

func testAccStackResourceConfig_withGit(name, repo, branch string) string {
	return fmt.Sprintf(`
resource "komodo_stack" "test" {
  name = "%s"

  source {
    url    = "https://github.com"
    path   = "%s"
    branch = "%s"
  }
}
`, name, repo, branch)
}

func testAccStackResourceConfig_withPreDeploy(name string) string {
	return fmt.Sprintf(`
resource "komodo_stack" "test" {
  name = "%s"

  pre_deploy {
    command = "echo pre"
  }
}
`, name)
}

// TestAccStackResource_preDeployPathWithoutCommand verifies that setting path
// without command is rejected at plan time.
func TestAccStackResource_preDeployPathWithoutCommand(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccStackResourceConfig_preDeployPathOnly("tf-test-stack-pre-invalid"),
				ExpectError: regexp.MustCompile(`path requires command`),
			},
		},
	})
}

// TestAccStackResource_postDeployPathWithoutCommand verifies that setting path
// without command is rejected at plan time for post_deploy.
func TestAccStackResource_postDeployPathWithoutCommand(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccStackResourceConfig_postDeployPathOnly("tf-test-stack-post-invalid"),
				ExpectError: regexp.MustCompile(`path requires command`),
			},
		},
	})
}

// TestAccStackResource_gitRepoConflicts verifies that setting source.repo_id
// alongside any direct-clone field is rejected at plan time.
func TestAccStackResource_gitRepoConflicts(t *testing.T) {
	cases := []struct {
		name   string
		config string
	}{
		{
			name: "repo_and_url",
			config: `
resource "komodo_stack" "test" {
  name = "tf-test-conflict"
  source {
    repo_id = "my-git-repo"
    url  = "https://github.com"
  }
}`,
		},
		{
			name: "repo_and_path",
			config: `
resource "komodo_stack" "test" {
  name = "tf-test-conflict"
  source {
    repo_id = "my-git-repo"
    path = "owner/repo"
  }
}`,
		},
		{
			name: "repo_and_branch",
			config: `
resource "komodo_stack" "test" {
  name = "tf-test-conflict"
  source {
    repo_id = "my-git-repo"
    branch = "main"
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

func testAccStackResourceConfig_preDeployPathOnly(name string) string {
	return fmt.Sprintf(`
resource "komodo_stack" "test" {
  name = "%s"

  pre_deploy {
    path = "/opt/app"
  }
}
`, name)
}

func testAccStackResourceConfig_postDeployPathOnly(name string) string {
	return fmt.Sprintf(`
resource "komodo_stack" "test" {
  name = "%s"

  post_deploy {
    path = "/opt/app"
  }
}
`, name)
}

// TestAccStackResource_preDeployPartialFieldRemoval verifies that removing only
// one field (path or command) from a pre_deploy block is applied correctly —
// i.e. the plan is non-empty and the removed field is cleared on the API.
func TestAccStackResource_preDeployPartialFieldRemoval(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: create with both path and command set.
			{
				Config: testAccStackResourceConfig_preDeployBothFields("tf-test-stack-pre-partial"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_stack.test", "pre_deploy.path", "/opt/app"),
					resource.TestCheckResourceAttr("komodo_stack.test", "pre_deploy.command", "echo pre"),
				),
			},
			// Step 2: remove path only — must produce a diff and clear path.
			{
				Config: testAccStackResourceConfig_preDeployCommandOnly("tf-test-stack-pre-partial"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_stack.test", "pre_deploy.command", "echo pre"),
					resource.TestCheckNoResourceAttr("komodo_stack.test", "pre_deploy.path"),
				),
			},
			// Step 3: remove command only (path already gone) — clears the whole block.
			{
				Config: testAccStackResourceConfig_basic("tf-test-stack-pre-partial"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("komodo_stack.test", "pre_deploy.command"),
					resource.TestCheckNoResourceAttr("komodo_stack.test", "pre_deploy.path"),
				),
			},
		},
	})
}

// TestAccStackResource_postDeployPartialFieldRemoval mirrors the pre_deploy
// test for post_deploy.
func TestAccStackResource_postDeployPartialFieldRemoval(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: create with both path and command set.
			{
				Config: testAccStackResourceConfig_postDeployBothFields("tf-test-stack-post-partial"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_stack.test", "post_deploy.path", "/opt/app"),
					resource.TestCheckResourceAttr("komodo_stack.test", "post_deploy.command", "echo post"),
				),
			},
			// Step 2: remove path only — must produce a diff and clear path.
			{
				Config: testAccStackResourceConfig_postDeployCommandOnly("tf-test-stack-post-partial"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_stack.test", "post_deploy.command", "echo post"),
					resource.TestCheckNoResourceAttr("komodo_stack.test", "post_deploy.path"),
				),
			},
			// Step 3: remove the whole post_deploy block.
			{
				Config: testAccStackResourceConfig_basic("tf-test-stack-post-partial"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("komodo_stack.test", "post_deploy.command"),
					resource.TestCheckNoResourceAttr("komodo_stack.test", "post_deploy.path"),
				),
			},
		},
	})
}

func testAccStackResourceConfig_preDeployBothFields(name string) string {
	return fmt.Sprintf(`
resource "komodo_stack" "test" {
  name = "%s"

  pre_deploy {
    path    = "/opt/app"
    command = "echo pre"
  }
}
`, name)
}

func testAccStackResourceConfig_preDeployCommandOnly(name string) string {
	return fmt.Sprintf(`
resource "komodo_stack" "test" {
  name = "%s"

  pre_deploy {
    command = "echo pre"
  }
}
`, name)
}

func testAccStackResourceConfig_postDeployBothFields(name string) string {
	return fmt.Sprintf(`
resource "komodo_stack" "test" {
  name = "%s"

  post_deploy {
    path    = "/opt/app"
    command = "echo post"
  }
}
`, name)
}

func testAccStackResourceConfig_postDeployCommandOnly(name string) string {
	return fmt.Sprintf(`
resource "komodo_stack" "test" {
  name = "%s"

  post_deploy {
    command = "echo post"
  }
}
`, name)
}

func TestAccStackResource_tags(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStackWithTagConfig("tf-acc-stack-tags"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_stack.test", "tags.#", "1"),
					resource.TestCheckResourceAttrPair("komodo_stack.test", "tags.0", "komodo_tag.test", "id"),
				),
			},
			{
				Config: testAccStackClearTagsConfig("tf-acc-stack-tags"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_stack.test", "tags.#", "0"),
				),
			},
		},
	})
}

func testAccStackWithTagConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_tag" "test" {
  name  = "tf-acc-tag-stack"
  color = "Green"
}

resource "komodo_stack" "test" {
  name = %q
  tags = [komodo_tag.test.id]
}
`, name)
}

func testAccStackClearTagsConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_stack" "test" {
  name = %q
  tags = []
}
`, name)
}

// ---------------------------------------------------------------------------
// Test constants and helpers for mock-server tests
// ---------------------------------------------------------------------------

const (
	mockValidStackJSON    = `{"_id":{"$oid":"abc123"},"name":"tf-mock-stack","tags":[],"config":{"send_alerts":true,"links":[],"extra_args":[],"ignore_services":[],"file_paths":[],"build_extra_args":[],"compose_cmd_wrapper_include":[]}}`
	mockEmptyOIDStackJSON = `{"_id":{"$oid":""},"name":"tf-mock-stack","tags":[],"config":{}}`
)

func mockStackProviderConfig(srvURL string) string {
	return fmt.Sprintf(`
provider "komodo" {
  endpoint = %q
  username = "mock"
  password = "mock"
}`, srvURL)
}

const mockStackResourceConfig = `
resource "komodo_stack" "test" {
  name = "tf-mock-stack"
}`

func wrongRawStackPlan(t *testing.T, r *StackResource) tfsdk.Plan {
	t.Helper()
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schemaResp.Schema,
	}
}

func wrongRawStackState(t *testing.T, r *StackResource) tfsdk.State {
	t.Helper()
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	return tfsdk.State{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schemaResp.Schema,
	}
}

// newStackAPIClientMock creates an httptest server and a *client.Client
// (using API key auth, so no login needed) routing by JSON "type" field.
func newStackAPIClientMock(t *testing.T, routes map[string]string) *client.Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		body, _ := io.ReadAll(r.Body)
		var req struct {
			Type string `json:"type"`
		}
		_ = json.Unmarshal(body, &req)
		if resp, ok := routes[req.Type]; ok {
			_, _ = w.Write([]byte(resp))
			return
		}
		_, _ = w.Write([]byte(`null`))
	}))
	t.Cleanup(srv.Close)
	return client.NewClientWithApiKey(srv.URL, "key", "secret")
}

// ---------------------------------------------------------------------------
// Unit tests – Configure
// ---------------------------------------------------------------------------

func TestUnitStackResource_configure(t *testing.T) {
	t.Run("wrong_provider_data_type_adds_error", func(t *testing.T) {
		r := &StackResource{}
		req := fwresource.ConfigureRequest{ProviderData: "not-a-client"}
		resp := &fwresource.ConfigureResponse{}
		r.Configure(context.Background(), req, resp)
		if !resp.Diagnostics.HasError() {
			t.Fatal("expected diagnostic error for wrong ProviderData type")
		}
	})

	t.Run("nil_provider_data_is_noop", func(t *testing.T) {
		r := &StackResource{}
		req := fwresource.ConfigureRequest{ProviderData: nil}
		resp := &fwresource.ConfigureResponse{}
		r.Configure(context.Background(), req, resp)
		if resp.Diagnostics.HasError() {
			t.Fatalf("expected no error for nil ProviderData, got: %v", resp.Diagnostics)
		}
	})
}

// ---------------------------------------------------------------------------
// Unit tests – Validator MarkdownDescription methods
// ---------------------------------------------------------------------------

func TestUnitStackResource_validatorDescriptions(t *testing.T) {
	ctx := context.Background()

	v1 := systemCommandPathRequiresCommandValidator{blockName: "pre_deploy"}
	if got := v1.MarkdownDescription(ctx); got != v1.Description(ctx) {
		t.Fatalf("systemCommandPathRequiresCommandValidator: MarkdownDescription=%q, want %q", got, v1.Description(ctx))
	}

	v2 := autoUpdateRequiresPollUpdatesValidator{}
	if got := v2.MarkdownDescription(ctx); got != v2.Description(ctx) {
		t.Fatalf("autoUpdateRequiresPollUpdatesValidator: MarkdownDescription=%q, want %q", got, v2.Description(ctx))
	}

	v3 := autoUpdateScopeValidator{}
	if got := v3.MarkdownDescription(ctx); got != v3.Description(ctx) {
		t.Fatalf("autoUpdateScopeValidator: MarkdownDescription=%q, want %q", got, v3.Description(ctx))
	}

	v4 := gitRepoConflictsValidator{}
	if got := v4.MarkdownDescription(ctx); got != v4.Description(ctx) {
		t.Fatalf("gitRepoConflictsValidator: MarkdownDescription=%q, want %q", got, v4.Description(ctx))
	}
}

// ---------------------------------------------------------------------------
// Unit tests – Validator config.Get errors (cover the HasError early-return)
// ---------------------------------------------------------------------------

func TestUnitStackResource_validatorConfigGetError(t *testing.T) {
	ctx := context.Background()
	r := &StackResource{}
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	// A config whose Raw type is String instead of Object causes Config.Get to fail.
	badConfig := tfsdk.Config{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schemaResp.Schema,
	}
	req := fwresource.ValidateConfigRequest{Config: badConfig}

	t.Run("systemCommandPathRequiresCommand", func(t *testing.T) {
		resp := &fwresource.ValidateConfigResponse{}
		v := systemCommandPathRequiresCommandValidator{blockName: "pre_deploy"}
		v.ValidateResource(ctx, req, resp)
		if !resp.Diagnostics.HasError() {
			t.Fatal("expected error for bad config")
		}
	})

	t.Run("autoUpdateRequiresPollUpdates", func(t *testing.T) {
		resp := &fwresource.ValidateConfigResponse{}
		v := autoUpdateRequiresPollUpdatesValidator{}
		v.ValidateResource(ctx, req, resp)
		if !resp.Diagnostics.HasError() {
			t.Fatal("expected error for bad config")
		}
	})

	t.Run("autoUpdateScope", func(t *testing.T) {
		resp := &fwresource.ValidateConfigResponse{}
		v := autoUpdateScopeValidator{}
		v.ValidateResource(ctx, req, resp)
		if !resp.Diagnostics.HasError() {
			t.Fatal("expected error for bad config")
		}
	})

	t.Run("gitRepoConflicts", func(t *testing.T) {
		resp := &fwresource.ValidateConfigResponse{}
		v := gitRepoConflictsValidator{}
		v.ValidateResource(ctx, req, resp)
		if !resp.Diagnostics.HasError() {
			t.Fatal("expected error for bad config")
		}
	})
}

// ---------------------------------------------------------------------------
// UnitTest cases – auto_update validators (no API required)
// ---------------------------------------------------------------------------

// TestAccStackResource_autoUpdateRequiresPollUpdates verifies that enabling
// auto_update without poll_updates_enabled is rejected at plan time.
func TestAccStackResource_autoUpdateRequiresPollUpdates(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "komodo_stack" "test" {
  name = "tf-test-stack-autoupdate"
  auto_update {
    enabled = true
  }
}`,
				ExpectError: regexp.MustCompile(`(?i)poll_updates`),
			},
		},
	})
}

// TestAccStackResource_autoUpdateScopeInvalid verifies that an unknown scope
// value is rejected at plan time.
func TestAccStackResource_autoUpdateScopeInvalid(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "komodo_stack" "test" {
  name = "tf-test-stack-scope"
  auto_update {
    scope = "bad_scope"
  }
}`,
				ExpectError: regexp.MustCompile(`(?i)scope`),
			},
		},
	})
}

// TestAccStackResource_autoUpdateScopeStackWithoutEnabled verifies that
// scope = "stack" without enabled = true is rejected at plan time.
func TestAccStackResource_autoUpdateScopeStackWithoutEnabled(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "komodo_stack" "test" {
  name = "tf-test-stack-scope-stack"
  auto_update {
    scope   = "stack"
    enabled = false
  }
}`,
				ExpectError: regexp.MustCompile(`(?i)scope`),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// Unit tests – CRUD Plan/State.Get error branches
// ---------------------------------------------------------------------------

func TestUnitStackResource_createPlanGetError(t *testing.T) {
	r := &StackResource{client: &client.Client{}}
	req := fwresource.CreateRequest{Plan: wrongRawStackPlan(t, r)}
	resp := &fwresource.CreateResponse{}
	r.Create(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed plan")
	}
}

func TestUnitStackResource_readStateGetError(t *testing.T) {
	r := &StackResource{client: &client.Client{}}
	req := fwresource.ReadRequest{State: wrongRawStackState(t, r)}
	resp := &fwresource.ReadResponse{}
	r.Read(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed state")
	}
}

func TestUnitStackResource_updatePlanGetError(t *testing.T) {
	r := &StackResource{client: &client.Client{}}
	req := fwresource.UpdateRequest{Plan: wrongRawStackPlan(t, r)}
	resp := &fwresource.UpdateResponse{}
	r.Update(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed plan")
	}
}

func TestUnitStackResource_updateStateGetError(t *testing.T) {
	r := &StackResource{client: &client.Client{}}
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)

	// Build a null-per-attribute plan so Plan.Get succeeds and we reach State.Get.
	schemaType := schemaResp.Schema.Type().TerraformType(ctx)
	objType, ok := schemaType.(tftypes.Object)
	if !ok {
		t.Fatal("schema type is not an object")
	}
	attrVals := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name, attrType := range objType.AttributeTypes {
		attrVals[name] = tftypes.NewValue(attrType, nil)
	}
	validRaw := tftypes.NewValue(schemaType, attrVals)
	req := fwresource.UpdateRequest{
		Plan:  tfsdk.Plan{Raw: validRaw, Schema: schemaResp.Schema},
		State: wrongRawStackState(t, r),
	}
	resp := &fwresource.UpdateResponse{}
	r.Update(ctx, req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed state")
	}
}

func TestUnitStackResource_deleteStateGetError(t *testing.T) {
	r := &StackResource{client: &client.Client{}}
	req := fwresource.DeleteRequest{State: wrongRawStackState(t, r)}
	resp := &fwresource.DeleteResponse{}
	r.Delete(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed state")
	}
}

// ---------------------------------------------------------------------------
// Mock-server acceptance tests – CRUD client error paths
// ---------------------------------------------------------------------------

func TestAccStackResource_createApiError(t *testing.T) {
	srv := newTypedMockServer(t, map[string]mockRoute{
		"CreateStack": {http.StatusInternalServerError, `"create failed"`},
	})
	defer srv.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockStackProviderConfig(srv.URL) + mockStackResourceConfig,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

func TestAccStackResource_createEmptyID(t *testing.T) {
	srv := newTypedMockServer(t, map[string]mockRoute{
		"CreateStack": {http.StatusOK, mockEmptyOIDStackJSON},
	})
	defer srv.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockStackProviderConfig(srv.URL) + mockStackResourceConfig,
				ExpectError: regexp.MustCompile(`(?i)missing id`),
			},
		},
	})
}

func TestAccStackResource_createUpdateMetaError(t *testing.T) {
	srv := newTypedMockServer(t, map[string]mockRoute{
		"CreateStack":        {http.StatusOK, mockValidStackJSON},
		"UpdateResourceMeta": {http.StatusInternalServerError, `"meta error"`},
	})
	defer srv.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockStackProviderConfig(srv.URL) + `
resource "komodo_stack" "test" {
  name = "tf-mock-stack"
  tags = ["tag1"]
}`,
				ExpectError: regexp.MustCompile(`(?i)tags`),
			},
		},
	})
}

func TestAccStackResource_readClientError(t *testing.T) {
	var mu sync.Mutex
	getStackCount := 0

	srv := newStatefulMockServer(t, func(w http.ResponseWriter, reqType string) {
		if reqType == "GetStack" {
			mu.Lock()
			n := getStackCount
			getStackCount++
			mu.Unlock()
			if n == 0 {
				_, _ = w.Write([]byte(mockValidStackJSON))
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`"get failed"`))
			}
			return
		}
		_, _ = w.Write([]byte(mockValidStackJSON))
	})
	defer srv.Close()

	provCfg := mockStackProviderConfig(srv.URL)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: provCfg + mockStackResourceConfig},
			{RefreshState: true, ExpectError: regexp.MustCompile(`(?i)error`)},
		},
	})
}

func TestAccStackResource_readNilToNil(t *testing.T) {
	var mu sync.Mutex
	getStackCount := 0

	srv := newStatefulMockServer(t, func(w http.ResponseWriter, reqType string) {
		if reqType == "GetStack" {
			mu.Lock()
			n := getStackCount
			getStackCount++
			mu.Unlock()
			if n == 0 {
				_, _ = w.Write([]byte(mockValidStackJSON))
			} else {
				// Both ID-based and name-based lookups return not-found → RemoveResource.
				w.WriteHeader(http.StatusNotFound)
			}
			return
		}
		_, _ = w.Write([]byte(mockValidStackJSON))
	})
	defer srv.Close()

	provCfg := mockStackProviderConfig(srv.URL)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: provCfg + mockStackResourceConfig},
			{RefreshState: true, ExpectNonEmptyPlan: true},
		},
	})
}

func TestAccStackResource_readNilNameLookupError(t *testing.T) {
	var mu sync.Mutex
	getStackCount := 0

	srv := newStatefulMockServer(t, func(w http.ResponseWriter, reqType string) {
		if reqType == "GetStack" {
			mu.Lock()
			n := getStackCount
			getStackCount++
			mu.Unlock()
			switch n {
			case 0:
				// Post-create refresh – success.
				_, _ = w.Write([]byte(mockValidStackJSON))
			case 1:
				// Refresh step: ID-based lookup → nil.
				w.WriteHeader(http.StatusNotFound)
			default:
				// Refresh step: name-based lookup → error.
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`"name lookup failed"`))
			}
			return
		}
		_, _ = w.Write([]byte(mockValidStackJSON))
	})
	defer srv.Close()

	provCfg := mockStackProviderConfig(srv.URL)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: provCfg + mockStackResourceConfig},
			{RefreshState: true, ExpectError: regexp.MustCompile(`(?i)error`)},
		},
	})
}

// TestAccStackResource_readNameAdoption covers the "adopted by name" log path:
// the first GetStack (by ID) returns nil, but the second (by name) returns the stack.
func TestAccStackResource_readNameAdoption(t *testing.T) {
	var mu sync.Mutex
	getStackCount := 0

	srv := newStatefulMockServer(t, func(w http.ResponseWriter, reqType string) {
		if reqType == "GetStack" {
			mu.Lock()
			n := getStackCount
			getStackCount++
			mu.Unlock()
			switch n {
			case 0:
				// Post-create refresh – success by ID.
				_, _ = w.Write([]byte(mockValidStackJSON))
			case 1:
				// Refresh step: ID-based lookup → nil (externally recreated).
				w.WriteHeader(http.StatusNotFound)
			default:
				// Refresh step: name-based lookup → success with new ID.
				_, _ = w.Write([]byte(mockValidStackJSON))
			}
			return
		}
		_, _ = w.Write([]byte(mockValidStackJSON))
	})
	defer srv.Close()

	provCfg := mockStackProviderConfig(srv.URL)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: provCfg + mockStackResourceConfig},
			// Refresh: ID lookup nil → name lookup succeeds → adopted.
			{Config: provCfg + mockStackResourceConfig},
		},
	})
}

func TestAccStackResource_updateRenameError(t *testing.T) {
	srv := newTypedMockServer(t, map[string]mockRoute{
		"CreateStack": {http.StatusOK, mockValidStackJSON},
		"GetStack":    {http.StatusOK, mockValidStackJSON},
		"RenameStack": {http.StatusInternalServerError, `"rename failed"`},
	})
	defer srv.Close()

	provCfg := mockStackProviderConfig(srv.URL)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: provCfg + mockStackResourceConfig},
			{
				Config: provCfg + `
resource "komodo_stack" "test" {
  name = "tf-mock-stack-renamed"
}`,
				ExpectError: regexp.MustCompile(`(?i)rename`),
			},
		},
	})
}

func TestAccStackResource_updateStackError(t *testing.T) {
	srv := newTypedMockServer(t, map[string]mockRoute{
		"CreateStack": {http.StatusOK, mockValidStackJSON},
		"GetStack":    {http.StatusOK, mockValidStackJSON},
		"UpdateStack": {http.StatusInternalServerError, `"update failed"`},
	})
	defer srv.Close()

	provCfg := mockStackProviderConfig(srv.URL)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: provCfg + mockStackResourceConfig},
			{
				Config: provCfg + `
resource "komodo_stack" "test" {
  name            = "tf-mock-stack"
  auto_pull_enabled = true
}`,
				ExpectError: regexp.MustCompile(`(?i)update`),
			},
		},
	})
}

func TestAccStackResource_updateMetaError(t *testing.T) {
	srv := newTypedMockServer(t, map[string]mockRoute{
		"CreateStack":        {http.StatusOK, mockValidStackJSON},
		"GetStack":           {http.StatusOK, mockValidStackJSON},
		"UpdateStack":        {http.StatusOK, mockValidStackJSON},
		"UpdateResourceMeta": {http.StatusInternalServerError, `"meta error"`},
	})
	defer srv.Close()

	provCfg := mockStackProviderConfig(srv.URL)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: provCfg + mockStackResourceConfig},
			{
				Config: provCfg + `
resource "komodo_stack" "test" {
  name = "tf-mock-stack"
  tags = ["tag1"]
}`,
				ExpectError: regexp.MustCompile(`(?i)tags`),
			},
		},
	})
}

func TestAccStackResource_deleteClientError(t *testing.T) {
	var mu sync.Mutex
	deleteCount := 0

	srv := newStatefulMockServer(t, func(w http.ResponseWriter, reqType string) {
		if reqType == "DeleteStack" {
			mu.Lock()
			n := deleteCount
			deleteCount++
			mu.Unlock()
			if n == 0 {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`"delete failed"`))
			} else {
				// Cleanup calls succeed so the test framework doesn't flag dangling resources.
				_, _ = w.Write([]byte(`null`))
			}
			return
		}
		_, _ = w.Write([]byte(mockValidStackJSON))
	})
	defer srv.Close()

	provCfg := mockStackProviderConfig(srv.URL)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: provCfg + mockStackResourceConfig},
			// Remove resource from config → triggers Delete → error.
			{Config: provCfg, ExpectError: regexp.MustCompile(`(?i)delete`)},
		},
	})
}

// ---------------------------------------------------------------------------
// Unit tests – stackToModel
// ---------------------------------------------------------------------------

func TestUnitStackResource_stackToModel(t *testing.T) {
	ctx := context.Background()

	t.Run("nil_tags_converted_to_empty", func(t *testing.T) {
		c := newStackAPIClientMock(t, nil)
		stack := &client.Stack{
			ID:   client.OID{OID: "id1"},
			Name: "test",
			Tags: nil, // nil → empty list
		}
		var data StackResourceModel
		stackToModel(ctx, c, stack, &data)
		if data.Tags.IsNull() || data.Tags.IsUnknown() {
			t.Fatal("expected Tags to be an empty list, not null/unknown")
		}
		if len(data.Tags.Elements()) != 0 {
			t.Fatalf("expected 0 tags, got %d", len(data.Tags.Elements()))
		}
	})

	t.Run("auto_update_all_services", func(t *testing.T) {
		c := newStackAPIClientMock(t, nil)
		stack := &client.Stack{
			ID:   client.OID{OID: "id1"},
			Name: "test",
			Config: client.StackConfig{
				AutoUpdate:            true,
				AutoUpdateAllServices: true,
			},
		}
		var data StackResourceModel
		stackToModel(ctx, c, stack, &data)
		if data.AutoUpdate == nil {
			t.Fatal("expected AutoUpdate to be set")
		}
		if data.AutoUpdate.Scope.ValueString() != "stack" {
			t.Fatalf("expected scope=stack, got %q", data.AutoUpdate.Scope.ValueString())
		}
		if !data.AutoUpdate.Enabled.ValueBool() {
			t.Fatal("expected Enabled=true")
		}
	})

	t.Run("registry_set", func(t *testing.T) {
		c := newStackAPIClientMock(t, map[string]string{
			"ListDockerRegistryAccounts": `[{"_id":{"$oid":"reg1"},"domain":"docker.io","username":"myuser","token":""}]`,
		})
		stack := &client.Stack{
			ID:   client.OID{OID: "id1"},
			Name: "test",
			Config: client.StackConfig{
				RegistryProvider: "docker.io",
				RegistryAccount:  "myuser",
			},
		}
		var data StackResourceModel
		stackToModel(ctx, c, stack, &data)
		if data.Registry == nil {
			t.Fatal("expected Registry to be set")
		}
		if data.Registry.AccountID.ValueString() != "reg1" {
			t.Fatalf("expected account_id=reg1, got %q", data.Registry.AccountID.ValueString())
		}
	})

	t.Run("source_with_linked_repo", func(t *testing.T) {
		c := newStackAPIClientMock(t, nil)
		stack := &client.Stack{
			ID:   client.OID{OID: "id1"},
			Name: "test",
			Config: client.StackConfig{
				LinkedRepo: "repo-id-123",
			},
		}
		var data StackResourceModel
		stackToModel(ctx, c, stack, &data)
		if data.Source == nil {
			t.Fatal("expected Source to be set")
		}
		if data.Source.RepoID.ValueString() != "repo-id-123" {
			t.Fatalf("expected RepoID=repo-id-123, got %q", data.Source.RepoID.ValueString())
		}
		// URL should be null when LinkedRepo is set.
		if !data.Source.URL.IsNull() {
			t.Fatalf("expected URL=null when LinkedRepo is set, got %q", data.Source.URL.ValueString())
		}
	})

	t.Run("source_git_https_false", func(t *testing.T) {
		c := newStackAPIClientMock(t, nil)
		stack := &client.Stack{
			ID:   client.OID{OID: "id1"},
			Name: "test",
			Config: client.StackConfig{
				GitProvider: "gitlab.com",
				GitHttps:    false,
				GitAccount:  "",
			},
		}
		// Prior source has URL set → priorURLSet = true.
		data := StackResourceModel{
			Source: &StackSourceModel{
				URL: types.StringValue("http://gitlab.com"),
			},
		}
		stackToModel(ctx, c, stack, &data)
		if data.Source == nil {
			t.Fatal("expected Source to be set")
		}
		if data.Source.URL.ValueString() != "http://gitlab.com" {
			t.Fatalf("expected URL=http://gitlab.com, got %q", data.Source.URL.ValueString())
		}
	})

	t.Run("source_git_account_resolution", func(t *testing.T) {
		c := newStackAPIClientMock(t, map[string]string{
			"ListGitProviderAccounts": `[{"_id":{"$oid":"acc1"},"domain":"github.com","username":"myuser","https":true,"token":""}]`,
		})
		stack := &client.Stack{
			ID:   client.OID{OID: "id1"},
			Name: "test",
			Config: client.StackConfig{
				GitProvider: "github.com",
				GitHttps:    true,
				GitAccount:  "myuser",
			},
		}
		// Prior source has URL set → priorURLSet = true.
		data := StackResourceModel{
			Source: &StackSourceModel{
				URL: types.StringValue("https://github.com"),
			},
		}
		stackToModel(ctx, c, stack, &data)
		if data.Source == nil {
			t.Fatal("expected Source to be set")
		}
		if data.Source.AccountID.ValueString() != "acc1" {
			t.Fatalf("expected AccountID=acc1, got %q", data.Source.AccountID.ValueString())
		}
	})

	t.Run("webhook_set", func(t *testing.T) {
		c := newStackAPIClientMock(t, nil)
		stack := &client.Stack{
			ID:   client.OID{OID: "id1"},
			Name: "test",
			Config: client.StackConfig{
				WebhookEnabled:     true,
				WebhookSecret:      "mysecret",
				WebhookForceDeploy: true,
			},
		}
		var data StackResourceModel
		stackToModel(ctx, c, stack, &data)
		if data.Webhook == nil {
			t.Fatal("expected Webhook to be set")
		}
		if !data.Webhook.Enabled.ValueBool() {
			t.Fatal("expected Webhook.Enabled=true")
		}
		if data.Webhook.Secret.ValueString() != "mysecret" {
			t.Fatalf("expected Webhook.Secret=mysecret, got %q", data.Webhook.Secret.ValueString())
		}
		if !data.Webhook.ForceDeploy.ValueBool() {
			t.Fatal("expected Webhook.ForceDeploy=true")
		}
	})

	t.Run("post_deploy_set", func(t *testing.T) {
		c := newStackAPIClientMock(t, nil)
		stack := &client.Stack{
			ID:   client.OID{OID: "id1"},
			Name: "test",
			Config: client.StackConfig{
				PostDeploy: client.SystemCommand{
					Path:    "/opt/app",
					Command: "echo post",
				},
			},
		}
		var data StackResourceModel
		stackToModel(ctx, c, stack, &data)
		if data.PostDeploy == nil {
			t.Fatal("expected PostDeploy to be set")
		}
		if data.PostDeploy.Command.ValueString() != "echo post" {
			t.Fatalf("expected Command=echo post, got %q", data.PostDeploy.Command.ValueString())
		}
	})

	t.Run("environment_with_file_path", func(t *testing.T) {
		c := newStackAPIClientMock(t, nil)
		stack := &client.Stack{
			ID:   client.OID{OID: "id1"},
			Name: "test",
			Config: client.StackConfig{
				EnvFilePath: "/etc/env",
			},
		}
		var data StackResourceModel
		stackToModel(ctx, c, stack, &data)
		if data.Environment == nil {
			t.Fatal("expected Environment to be set")
		}
		if data.Environment.FilePath.ValueString() != "/etc/env" {
			t.Fatalf("expected FilePath=/etc/env, got %q", data.Environment.FilePath.ValueString())
		}
	})

	t.Run("environment_cleared", func(t *testing.T) {
		// Environment block was set in prior state but API now returns nothing.
		c := newStackAPIClientMock(t, nil)
		stack := &client.Stack{
			ID:   client.OID{OID: "id1"},
			Name: "test",
			// no Environment fields
		}
		emptyVars, _ := types.MapValueFrom(ctx, types.StringType, map[string]string{})
		data := StackResourceModel{
			Environment: &EnvironmentModel{
				FilePath:  types.StringNull(),
				Variables: emptyVars,
			},
		}
		stackToModel(ctx, c, stack, &data)
		// data.Environment should be set to an empty model (not nil), because it was non-nil before.
		if data.Environment == nil {
			t.Fatal("expected Environment to be retained as empty model")
		}
	})

	t.Run("build_extra_args_present", func(t *testing.T) {
		c := newStackAPIClientMock(t, nil)
		stack := &client.Stack{
			ID:   client.OID{OID: "id1"},
			Name: "test",
			Config: client.StackConfig{
				RunBuild:       true,
				BuildExtraArgs: []string{"--no-cache"},
			},
		}
		var data StackResourceModel
		stackToModel(ctx, c, stack, &data)
		if data.Build == nil {
			t.Fatal("expected Build to be set")
		}
		var args []string
		data.Build.ExtraArguments.ElementsAs(ctx, &args, false)
		if len(args) != 1 || args[0] != "--no-cache" {
			t.Fatalf("expected BuildExtraArgs=[--no-cache], got %v", args)
		}
	})

	t.Run("build_already_in_state", func(t *testing.T) {
		// data.Build is non-nil but API returns RunBuild=false and no BuildExtraArgs
		// → Build block should remain (nil-guarded by `data.Build != nil`).
		c := newStackAPIClientMock(t, nil)
		stack := &client.Stack{
			ID:   client.OID{OID: "id1"},
			Name: "test",
			Config: client.StackConfig{
				RunBuild: false,
			},
		}
		data := StackResourceModel{
			Build: &BuildConfigModel{
				Enabled:        types.BoolValue(false),
				ExtraArguments: types.ListNull(types.StringType),
			},
		}
		stackToModel(ctx, c, stack, &data)
		if data.Build == nil {
			t.Fatal("expected Build block to be retained")
		}
	})

	t.Run("wrapper_set", func(t *testing.T) {
		c := newStackAPIClientMock(t, nil)
		stack := &client.Stack{
			ID:   client.OID{OID: "id1"},
			Name: "test",
			Config: client.StackConfig{
				ComposeCmdWrapper:        "sudo",
				ComposeCmdWrapperInclude: []string{"web"},
			},
		}
		var data StackResourceModel
		stackToModel(ctx, c, stack, &data)
		if data.Wrapper == nil {
			t.Fatal("expected Wrapper to be set")
		}
		if data.Wrapper.Command.ValueString() != "sudo" {
			t.Fatalf("expected Wrapper.Command=sudo, got %q", data.Wrapper.Command.ValueString())
		}
		var inc []string
		data.Wrapper.Include.ElementsAs(ctx, &inc, false)
		if len(inc) != 1 || inc[0] != "web" {
			t.Fatalf("expected Wrapper.Include=[web], got %v", inc)
		}
	})
}

// ---------------------------------------------------------------------------
// Unit tests – stackConfigFromModel
// ---------------------------------------------------------------------------

func TestUnitStackResource_stackConfigFromModel(t *testing.T) {
	ctx := context.Background()

	t.Run("registry_with_account", func(t *testing.T) {
		c := newStackAPIClientMock(t, map[string]string{
			"GetDockerRegistryAccount": `{"_id":{"$oid":"reg1"},"domain":"docker.io","username":"myuser","token":""}`,
		})
		data := &StackResourceModel{
			Registry: &RegistryConfigModel{
				AccountID: types.StringValue("reg1"),
			},
		}
		cfg := stackConfigFromModel(ctx, c, data)
		if cfg.RegistryProvider != "docker.io" {
			t.Fatalf("expected RegistryProvider=docker.io, got %q", cfg.RegistryProvider)
		}
		if cfg.RegistryAccount != "myuser" {
			t.Fatalf("expected RegistryAccount=myuser, got %q", cfg.RegistryAccount)
		}
	})

	t.Run("wrapper_command", func(t *testing.T) {
		c := newStackAPIClientMock(t, nil)
		inc, _ := types.ListValueFrom(ctx, types.StringType, []string{})
		data := &StackResourceModel{
			Wrapper: &StackCmdWrapperModel{
				Command: types.StringValue("sudo"),
				Include: inc,
			},
		}
		cfg := stackConfigFromModel(ctx, c, data)
		if cfg.ComposeCmdWrapper != "sudo" {
			t.Fatalf("expected ComposeCmdWrapper=sudo, got %q", cfg.ComposeCmdWrapper)
		}
	})

	t.Run("webhook_fields", func(t *testing.T) {
		c := newStackAPIClientMock(t, nil)
		data := &StackResourceModel{
			Webhook: &StackWebhookModel{
				Enabled:     types.BoolValue(true),
				Secret:      types.StringValue("s3cr3t"),
				ForceDeploy: types.BoolValue(true),
			},
		}
		cfg := stackConfigFromModel(ctx, c, data)
		if !cfg.WebhookEnabled {
			t.Fatal("expected WebhookEnabled=true")
		}
		if cfg.WebhookSecret != "s3cr3t" {
			t.Fatalf("expected WebhookSecret=s3cr3t, got %q", cfg.WebhookSecret)
		}
		if !cfg.WebhookForceDeploy {
			t.Fatal("expected WebhookForceDeploy=true")
		}
	})

	t.Run("environment_fields", func(t *testing.T) {
		c := newStackAPIClientMock(t, nil)
		vars, _ := types.MapValueFrom(ctx, types.StringType, map[string]string{"KEY": "val"})
		data := &StackResourceModel{
			Environment: &EnvironmentModel{
				FilePath:  types.StringValue("/etc/env"),
				Variables: vars,
			},
		}
		cfg := stackConfigFromModel(ctx, c, data)
		if cfg.EnvFilePath != "/etc/env" {
			t.Fatalf("expected EnvFilePath=/etc/env, got %q", cfg.EnvFilePath)
		}
		if cfg.Environment == "" {
			t.Fatal("expected Environment to be non-empty")
		}
	})

	t.Run("source_http_url", func(t *testing.T) {
		c := newStackAPIClientMock(t, nil)
		data := &StackResourceModel{
			Source: &StackSourceModel{
				URL:       types.StringValue("http://gitlab.com"),
				FilePaths: types.ListNull(types.StringType),
			},
		}
		cfg := stackConfigFromModel(ctx, c, data)
		if cfg.GitProvider != "gitlab.com" {
			t.Fatalf("expected GitProvider=gitlab.com, got %q", cfg.GitProvider)
		}
		if cfg.GitHttps {
			t.Fatal("expected GitHttps=false for http:// URL")
		}
	})

	t.Run("source_plain_url", func(t *testing.T) {
		// URL without http/https scheme → stored as-is, GitHttps=true.
		c := newStackAPIClientMock(t, nil)
		data := &StackResourceModel{
			Source: &StackSourceModel{
				URL:       types.StringValue("gitlab.com"),
				FilePaths: types.ListNull(types.StringType),
			},
		}
		cfg := stackConfigFromModel(ctx, c, data)
		if cfg.GitProvider != "gitlab.com" {
			t.Fatalf("expected GitProvider=gitlab.com, got %q", cfg.GitProvider)
		}
		if !cfg.GitHttps {
			t.Fatal("expected GitHttps=true for plain URL")
		}
	})

	t.Run("source_file_paths", func(t *testing.T) {
		c := newStackAPIClientMock(t, nil)
		fps, _ := types.ListValueFrom(ctx, types.StringType, []string{"docker-compose.yml", "override.yml"})
		data := &StackResourceModel{
			Source: &StackSourceModel{
				URL:       types.StringNull(),
				FilePaths: fps,
			},
		}
		cfg := stackConfigFromModel(ctx, c, data)
		if len(cfg.FilePaths) != 2 {
			t.Fatalf("expected 2 FilePaths, got %d", len(cfg.FilePaths))
		}
	})

	t.Run("post_deploy", func(t *testing.T) {
		c := newStackAPIClientMock(t, nil)
		data := &StackResourceModel{
			PostDeploy: &SystemCommandModel{
				Path:    types.StringValue("/opt"),
				Command: types.StringValue("echo post"),
			},
		}
		cfg := stackConfigFromModel(ctx, c, data)
		if cfg.PostDeploy.Command != "echo post" {
			t.Fatalf("expected PostDeploy.Command=echo post, got %q", cfg.PostDeploy.Command)
		}
	})

	t.Run("git_account_resolve_error_falls_back_to_id", func(t *testing.T) {
		// When ResolveGitAccountUsername fails (server returns 500), it falls back to the raw ID.
		srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			body, _ := io.ReadAll(r.Body)
			var req struct {
				Type string `json:"type"`
			}
			_ = json.Unmarshal(body, &req)
			if req.Type == "GetGitProviderAccount" {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`"error"`))
			} else {
				_, _ = w.Write([]byte(`null`))
			}
		}))
		defer srv2.Close()
		c2 := client.NewClientWithApiKey(srv2.URL, "key", "secret")

		hexID := "aabbccddee0011223344aabb" // 24-char hex
		data := &StackResourceModel{
			Source: &StackSourceModel{
				URL:       types.StringValue("https://github.com"),
				AccountID: types.StringValue(hexID),
				FilePaths: types.ListNull(types.StringType),
			},
		}
		cfg := stackConfigFromModel(ctx, c2, data)
		// Falls back to the raw account ID when resolution fails.
		if cfg.GitAccount != hexID {
			t.Fatalf("expected GitAccount=%q (fallback), got %q", hexID, cfg.GitAccount)
		}
	})

	t.Run("source_account_resolves_domain", func(t *testing.T) {
		// source.url is empty but source.account_id is a valid ObjectID that resolves.
		// This covers the `else if acc := c.ResolveGitAccountFull(...)` branch.
		srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			body, _ := io.ReadAll(r.Body)
			var req struct {
				Type string `json:"type"`
			}
			_ = json.Unmarshal(body, &req)
			if req.Type == "GetGitProviderAccount" {
				_, _ = w.Write([]byte(`{"_id":{"$oid":"acc1"},"domain":"gitlab.com","https":true,"username":"myuser","token":""}`))
			} else {
				_, _ = w.Write([]byte(`null`))
			}
		}))
		defer srv2.Close()
		c2 := client.NewClientWithApiKey(srv2.URL, "key", "secret")
		hexID := "aabbccddee0011223344aabb"
		data := &StackResourceModel{
			Source: &StackSourceModel{
				URL:       types.StringNull(), // no URL → fall through to account lookup
				AccountID: types.StringValue(hexID),
				FilePaths: types.ListNull(types.StringType),
			},
		}
		cfg := stackConfigFromModel(ctx, c2, data)
		if cfg.GitProvider != "gitlab.com" {
			t.Fatalf("expected GitProvider=gitlab.com from account domain, got %q", cfg.GitProvider)
		}
		if !cfg.GitHttps {
			t.Fatal("expected GitHttps=true")
		}
	})

	t.Run("build_extra_arguments_set", func(t *testing.T) {
		// Covers the `Build.ExtraArguments not null` branch in stackConfigFromModel.
		c := newStackAPIClientMock(t, nil)
		args, _ := types.ListValueFrom(ctx, types.StringType, []string{"--no-cache"})
		data := &StackResourceModel{
			Build: &BuildConfigModel{
				Enabled:        types.BoolValue(true),
				ExtraArguments: args,
			},
		}
		cfg := stackConfigFromModel(ctx, c, data)
		if len(cfg.BuildExtraArgs) != 1 || cfg.BuildExtraArgs[0] != "--no-cache" {
			t.Fatalf("expected BuildExtraArgs=[--no-cache], got %v", cfg.BuildExtraArgs)
		}
	})

	t.Run("build_extra_arguments_null_clears", func(t *testing.T) {
		// Covers the else branch: Build != nil but ExtraArguments is null → BuildExtraArgs = []string{}.
		c := newStackAPIClientMock(t, nil)
		data := &StackResourceModel{
			Build: &BuildConfigModel{
				Enabled:        types.BoolValue(true),
				ExtraArguments: types.ListNull(types.StringType),
			},
		}
		cfg := stackConfigFromModel(ctx, c, data)
		if cfg.BuildExtraArgs == nil {
			t.Fatal("expected BuildExtraArgs to be empty slice, not nil")
		}
		if len(cfg.BuildExtraArgs) != 0 {
			t.Fatalf("expected empty BuildExtraArgs, got %v", cfg.BuildExtraArgs)
		}
	})

	t.Run("wrapper_include_set", func(t *testing.T) {
		// Covers the `Wrapper.Include not null` branch in stackConfigFromModel.
		c := newStackAPIClientMock(t, nil)
		inc, _ := types.ListValueFrom(ctx, types.StringType, []string{"web", "db"})
		data := &StackResourceModel{
			Wrapper: &StackCmdWrapperModel{
				Command: types.StringValue("sudo"),
				Include: inc,
			},
		}
		cfg := stackConfigFromModel(ctx, c, data)
		if len(cfg.ComposeCmdWrapperInclude) != 2 {
			t.Fatalf("expected 2 ComposeCmdWrapperInclude, got %v", cfg.ComposeCmdWrapperInclude)
		}
	})
}

// ---------------------------------------------------------------------------
// Unit tests – trimTrailingNewlinePlanModifier
// ---------------------------------------------------------------------------

func TestUnitStackResource_trimTrailingNewline(t *testing.T) {
	ctx := context.Background()
	m := trimTrailingNewlinePlanModifier{}

	t.Run("markdown_description_equals_description", func(t *testing.T) {
		if m.MarkdownDescription(ctx) != m.Description(ctx) {
			t.Fatal("expected MarkdownDescription == Description")
		}
	})

	t.Run("null_value_noop", func(t *testing.T) {
		req := planmodifier.StringRequest{PlanValue: types.StringNull()}
		resp := &planmodifier.StringResponse{PlanValue: types.StringNull()}
		m.PlanModifyString(ctx, req, resp)
		if !resp.PlanValue.IsNull() {
			t.Fatal("expected PlanValue to remain null")
		}
	})

	t.Run("unknown_value_noop", func(t *testing.T) {
		req := planmodifier.StringRequest{PlanValue: types.StringUnknown()}
		resp := &planmodifier.StringResponse{PlanValue: types.StringUnknown()}
		m.PlanModifyString(ctx, req, resp)
		if !resp.PlanValue.IsUnknown() {
			t.Fatal("expected PlanValue to remain unknown")
		}
	})

	t.Run("strips_trailing_newline", func(t *testing.T) {
		req := planmodifier.StringRequest{PlanValue: types.StringValue("echo hello\n")}
		resp := &planmodifier.StringResponse{PlanValue: types.StringValue("echo hello\n")}
		m.PlanModifyString(ctx, req, resp)
		if resp.PlanValue.ValueString() != "echo hello" {
			t.Fatalf("expected trimmed value, got %q", resp.PlanValue.ValueString())
		}
	})
}
