// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccRepoResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRepoResourceConfig_basic("tf-test-repo"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_repo.test", "name", "tf-test-repo"),
					resource.TestCheckResourceAttrSet("komodo_repo.test", "id"),
				),
			},
		},
	})
}

func TestAccRepoResource_withConfig(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRepoResourceConfig_withConfig(
					"tf-test-repo-full",
					"owner/my-repo",
					"main",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_repo.test", "name", "tf-test-repo-full"),
					resource.TestCheckResourceAttr("komodo_repo.test", "source.domain", "github.com"),
					resource.TestCheckResourceAttr("komodo_repo.test", "source.https_enabled", "true"),
					resource.TestCheckResourceAttr("komodo_repo.test", "source.path", "owner/my-repo"),
					resource.TestCheckResourceAttr("komodo_repo.test", "source.branch", "main"),
					resource.TestCheckResourceAttrSet("komodo_repo.test", "id"),
				),
			},
		},
	})
}

func TestAccRepoResource_withOnClone(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRepoResourceConfig_withOnClone("tf-test-repo-clone"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_repo.test", "name", "tf-test-repo-clone"),
					resource.TestCheckResourceAttr("komodo_repo.test", "on_clone.command", "echo cloned"),
					resource.TestCheckResourceAttrSet("komodo_repo.test", "id"),
				),
			},
		},
	})
}

func TestAccRepoResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRepoResourceConfig_withConfig("tf-update-repo", "owner/repo", "main"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_repo.test", "source.branch", "main"),
				),
			},
			{
				Config: testAccRepoResourceConfig_withConfig("tf-update-repo", "owner/repo", "develop"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_repo.test", "source.branch", "develop"),
				),
			},
		},
	})
}

func TestAccRepoResource_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRepoResourceConfig_basic("tf-import-repo"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_repo.test", "name", "tf-import-repo"),
					resource.TestCheckResourceAttrSet("komodo_repo.test", "id"),
				),
			},
			{
				ResourceName:      "komodo_repo.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccRepoResource_disappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRepoResourceConfig_basic("tf-disappear-repo"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_repo.test", "id"),
					testAccRepoDisappears("komodo_repo.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRepoResource_rename(t *testing.T) {
	var savedID string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRepoResourceConfig_basic("tf-acc-repo-rename-orig"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_repo.test", "name", "tf-acc-repo-rename-orig"),
					resource.TestCheckResourceAttrSet("komodo_repo.test", "id"),
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["komodo_repo.test"]
						savedID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config: testAccRepoResourceConfig_basic("tf-acc-repo-rename-new"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_repo.test", "name", "tf-acc-repo-rename-new"),
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["komodo_repo.test"]
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

func testAccRepoDisappears(resourceName string) resource.TestCheckFunc {
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
		return c.DeleteGitRepository(context.Background(), rs.Primary.ID)
	}
}

// TestAccRepoResource_sourceHttpsEnabledDefault verifies that omitting
// source.https_enabled after it was explicitly set to false causes Terraform to
// plan a change back to the default value of true.
func TestAccRepoResource_sourceHttpsEnabledDefault(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: create with https_enabled explicitly false.
			{
				Config: testAccRepoResourceConfig_withHttps("tf-test-repo-https-default", "github.com", "owner/repo", "main", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_repo.test", "source.https_enabled", "false"),
				),
			},
			// Step 2: remove https_enabled from config → default kicks in, must plan true.
			{
				Config: testAccRepoResourceConfig_withConfig("tf-test-repo-https-default", "owner/repo", "main"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_repo.test", "source.https_enabled", "true"),
				),
			},
			// Step 3: re-plan with same config → no further changes.
			{
				Config:             testAccRepoResourceConfig_withConfig("tf-test-repo-https-default", "owner/repo", "main"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccRepoResourceConfig_withHttps(name, domain, repo, branch string, https bool) string {
	return fmt.Sprintf(`
resource "komodo_repo" "test" {
  name = "%s"
  source {
    domain        = "%s"
    https_enabled = %t
    path          = "%s"
    branch        = "%s"
  }
}
`, name, domain, https, repo, branch)
}

// Test configuration functions

func testAccRepoResourceConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "komodo_repo" "test" {
  name = "%s"
}
`, name)
}

func testAccRepoResourceConfig_withConfig(name, repo, branch string) string {
	return fmt.Sprintf(`
resource "komodo_repo" "test" {
  name = "%s"
  source {
    domain        = "github.com"
    https_enabled = true
    path          = "%s"
    branch        = "%s"
  }
}
`, name, repo, branch)
}

func testAccRepoResourceConfig_withOnClone(name string) string {
	return fmt.Sprintf(`
resource "komodo_repo" "test" {
  name = "%s"

  on_clone {
    command = "echo cloned"
  }
}
`, name)
}

func testAccRepoResourceConfig_withServerID(name, serverID string) string {
	return fmt.Sprintf(`
resource "komodo_repo" "test" {
  name      = "%s"
  server_id = "%s"
}
`, name, serverID)
}

// TestAccRepoResource_serverIDDrift verifies that if server_id is changed
// out-of-band (e.g. via the Komodo portal) Terraform detects the drift and corrects it.
func TestAccRepoResource_serverIDDrift(t *testing.T) {
	serverID := testAccLookupServerID(t, "server_id drift tests")
	const repoName = "tf-drift-repo"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: create without server_id
			{
				Config: testAccRepoResourceConfig_basic(repoName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_repo.test", "id"),
					resource.TestCheckNoResourceAttr("komodo_repo.test", "server_id"),
				),
			},
			// Step 2: set server_id out-of-band via the API, then re-apply the same
			// config — Terraform must detect the drift and clear server_id back to "".
			{
				Config: testAccRepoResourceConfig_basic(repoName),
				// Inject the out-of-band change before Terraform's plan/apply.
				PreConfig: func() {
					c := client.NewClient(
						os.Getenv("KOMODO_ENDPOINT"),
						os.Getenv("KOMODO_USERNAME"),
						os.Getenv("KOMODO_PASSWORD"),
					)
					repo, err := c.GetGitRepository(context.Background(), repoName)
					if err != nil || repo == nil {
						t.Fatalf("pre-config: failed to fetch repo %q: %v", repoName, err)
					}
					cfg := repo.Config
					cfg.ServerID = serverID
					_, err = c.UpdateGitRepository(context.Background(), client.UpdateGitRepositoryRequest{
						ID:     repo.ID.OID,
						Config: cfg,
					})
					if err != nil {
						t.Fatalf("pre-config: failed to inject out-of-band server_id: %v", err)
					}
				},
				// After apply Terraform must have cleared the server_id back to "".
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("komodo_repo.test", "server_id"),
				),
				// The plan must be non-empty because drift was detected.
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// TestAccRepoResource_serverIDRemove verifies that removing server_id from
// config sends an empty string to the API and clears the value.
func TestAccRepoResource_serverIDRemove(t *testing.T) {
	serverID := testAccLookupServerID(t, "server_id remove tests")
	const repoName = "tf-remove-serverid-repo"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: create with an explicit (fake) server_id
			{
				Config: testAccRepoResourceConfig_withServerID(repoName, serverID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_repo.test", "server_id", serverID),
				),
			},
			// Step 2: remove server_id from config — must be cleared in the API
			{
				Config: testAccRepoResourceConfig_basic(repoName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("komodo_repo.test", "server_id"),
				),
			},
		},
	})
}

func testAccRepoResourceConfig_withGitAccount(name, gitAccount string) string {
	return fmt.Sprintf(`
resource "komodo_repo" "test" {
  name = "%s"
  source {
    account_id = "%s"
  }
}
`, name, gitAccount)
}

// TestAccRepoResource_gitAccountDrift verifies that if git_account is changed
// out-of-band (e.g. via the Komodo portal) Terraform detects the drift and corrects it.
func TestAccRepoResource_gitAccountDrift(t *testing.T) {
	const repoName = "tf-drift-gitaccount-repo"
	const fakeAccount = "some-git-account"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: create without git_account
			{
				Config: testAccRepoResourceConfig_basic(repoName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_repo.test", "id"),
					resource.TestCheckNoResourceAttr("komodo_repo.test", "source.account_id"),
				),
			},
			// Step 2: set git_account out-of-band via the API, then re-apply the same
			// config — Terraform must detect the drift and clear git_account back to "".
			{
				Config: testAccRepoResourceConfig_basic(repoName),
				PreConfig: func() {
					c := client.NewClient(
						os.Getenv("KOMODO_ENDPOINT"),
						os.Getenv("KOMODO_USERNAME"),
						os.Getenv("KOMODO_PASSWORD"),
					)
					repo, err := c.GetGitRepository(context.Background(), repoName)
					if err != nil || repo == nil {
						t.Fatalf("pre-config: failed to fetch repo %q: %v", repoName, err)
					}
					cfg := repo.Config
					cfg.GitAccount = fakeAccount
					_, err = c.UpdateGitRepository(context.Background(), client.UpdateGitRepositoryRequest{
						ID:     repo.ID.OID,
						Config: cfg,
					})
					if err != nil {
						t.Fatalf("pre-config: failed to inject out-of-band git_account: %v", err)
					}
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("komodo_repo.test", "source.account_id"),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// TestAccRepoResource_gitAccountRemove verifies that removing git_account from
// config sends an empty string to the API and clears the value.
func TestAccRepoResource_gitAccountRemove(t *testing.T) {
	const repoName = "tf-remove-gitaccount-repo"
	const account = "my-git-account"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: create with an explicit git_account
			{
				Config: testAccRepoResourceConfig_withGitAccount(repoName, account),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_repo.test", "source.account_id", account),
				),
			},
			// Step 2: remove git_account from config — must be cleared in the API
			{
				Config: testAccRepoResourceConfig_basic(repoName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("komodo_repo.test", "source.account_id"),
				),
			},
		},
	})
}

func TestAccRepoResource_tags(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRepoWithTagConfig("tf-acc-repo-tags"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_repo.test", "tags.#", "1"),
					resource.TestCheckResourceAttrPair("komodo_repo.test", "tags.0", "komodo_tag.test", "id"),
				),
			},
			{
				Config: testAccRepoClearTagsConfig("tf-acc-repo-tags"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_repo.test", "tags.#", "0"),
				),
			},
		},
	})
}

func testAccRepoWithTagConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_tag" "test" {
  name  = "tf-acc-tag-repo"
  color = "Green"
}

resource "komodo_repo" "test" {
  name = %q
  tags = [komodo_tag.test.id]
}
`, name)
}

func testAccRepoClearTagsConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_repo" "test" {
  name = %q
  tags = []
}
`, name)
}

// ─── Unit tests ──────────────────────────────────────────────────────────────

func wrongRawRepoPlan(t *testing.T, r *RepoResource) tfsdk.Plan {
	t.Helper()
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schemaResp.Schema,
	}
}

func wrongRawRepoState(t *testing.T, r *RepoResource) tfsdk.State {
	t.Helper()
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	return tfsdk.State{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schemaResp.Schema,
	}
}

func TestUnitRepoResource_configure(t *testing.T) {
	t.Run("wrong_type", func(t *testing.T) {
		r := &RepoResource{}
		req := fwresource.ConfigureRequest{ProviderData: "not-a-client"}
		resp := &fwresource.ConfigureResponse{}
		r.Configure(context.Background(), req, resp)
		if !resp.Diagnostics.HasError() {
			t.Fatal("expected diagnostic error for wrong ProviderData type")
		}
	})
}

func TestUnitRepoResource_createPlanGetError(t *testing.T) {
	r := &RepoResource{client: &client.Client{}}
	req := fwresource.CreateRequest{Plan: wrongRawRepoPlan(t, r)}
	resp := &fwresource.CreateResponse{}
	r.Create(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed plan")
	}
}

func TestUnitRepoResource_readStateGetError(t *testing.T) {
	r := &RepoResource{client: &client.Client{}}
	req := fwresource.ReadRequest{State: wrongRawRepoState(t, r)}
	resp := &fwresource.ReadResponse{}
	r.Read(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed state")
	}
}

func TestUnitRepoResource_updatePlanGetError(t *testing.T) {
	r := &RepoResource{client: &client.Client{}}
	req := fwresource.UpdateRequest{Plan: wrongRawRepoPlan(t, r)}
	resp := &fwresource.UpdateResponse{}
	r.Update(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed plan")
	}
}

func TestUnitRepoResource_deleteStateGetError(t *testing.T) {
	r := &RepoResource{client: &client.Client{}}
	req := fwresource.DeleteRequest{State: wrongRawRepoState(t, r)}
	resp := &fwresource.DeleteResponse{}
	r.Delete(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed state")
	}
}

func TestUnitRepoResource_repoConfigFromModel(t *testing.T) {
	ctx := context.Background()
	c := &client.Client{}

	t.Run("minimal_no_source", func(t *testing.T) {
		data := &RepoResourceModel{
			ServerID:    types.StringValue("server-1"),
			BuilderID:   types.StringNull(),
			Path:        types.StringNull(),
			Webhook:     nil,
			OnClone:     nil,
			OnPull:      nil,
			Links:       types.ListNull(types.StringType),
			Environment: nil,
			Source:      nil,
			Tags:        types.ListValueMust(types.StringType, nil),
		}
		cfg := repoConfigFromModel(ctx, c, data)
		if cfg.ServerID != "server-1" {
			t.Fatalf("expected server_id=server-1, got %s", cfg.ServerID)
		}
		if cfg.WebhookEnabled {
			t.Fatal("expected webhook disabled")
		}
	})

	t.Run("with_webhook_enabled", func(t *testing.T) {
		data := &RepoResourceModel{
			ServerID:    types.StringNull(),
			BuilderID:   types.StringNull(),
			Path:        types.StringNull(),
			Links:       types.ListNull(types.StringType),
			Source:      nil,
			Environment: nil,
			Tags:        types.ListValueMust(types.StringType, nil),
			Webhook: &WebhookModel{
				Enabled: types.BoolValue(true),
				Secret:  types.StringValue("mysecret"),
			},
		}
		cfg := repoConfigFromModel(ctx, c, data)
		if !cfg.WebhookEnabled {
			t.Fatal("expected webhook enabled=true")
		}
		if cfg.WebhookSecret != "mysecret" {
			t.Fatalf("expected WebhookSecret=mysecret, got %s", cfg.WebhookSecret)
		}
	})

	t.Run("with_on_clone_command", func(t *testing.T) {
		data := &RepoResourceModel{
			ServerID:    types.StringNull(),
			BuilderID:   types.StringNull(),
			Path:        types.StringNull(),
			Links:       types.ListNull(types.StringType),
			Source:      nil,
			Environment: nil,
			Webhook:     nil,
			Tags:        types.ListValueMust(types.StringType, nil),
			OnClone: &SystemCommandModel{
				Path:    types.StringValue("/app"),
				Command: NewTrimmedStringValue("./setup.sh"),
			},
		}
		cfg := repoConfigFromModel(ctx, c, data)
		if cfg.OnClone.Path != "/app" || cfg.OnClone.Command != "./setup.sh" {
			t.Fatalf("unexpected on_clone: %+v", cfg.OnClone)
		}
	})

	t.Run("with_source_domain_and_path", func(t *testing.T) {
		data := &RepoResourceModel{
			ServerID:    types.StringNull(),
			BuilderID:   types.StringNull(),
			Path:        types.StringNull(),
			Links:       types.ListNull(types.StringType),
			Environment: nil,
			Webhook:     nil,
			Tags:        types.ListValueMust(types.StringType, nil),
			Source: &RepositoryProviderModel{
				Domain:       types.StringValue("github.com"),
				HttpsEnabled: types.BoolValue(true),
				AccountID:    types.StringNull(),
				Path:         types.StringValue("owner/repo"),
				Branch:       types.StringValue("main"),
				Commit:       types.StringNull(),
			},
		}
		cfg := repoConfigFromModel(ctx, c, data)
		if cfg.GitProvider != "github.com" {
			t.Fatalf("expected GitProvider=github.com, got %s", cfg.GitProvider)
		}
		if !cfg.GitHttps {
			t.Fatal("expected GitHttps=true")
		}
		if cfg.Repo != "owner/repo" {
			t.Fatalf("expected Repo=owner/repo, got %s", cfg.Repo)
		}
		if cfg.Branch != "main" {
			t.Fatalf("expected Branch=main, got %s", cfg.Branch)
		}
	})

	t.Run("with_links", func(t *testing.T) {
		data := &RepoResourceModel{
			ServerID:    types.StringNull(),
			BuilderID:   types.StringNull(),
			Path:        types.StringNull(),
			Links:       types.ListValueMust(types.StringType, nil),
			Source:      nil,
			Environment: nil,
			Webhook:     nil,
			Tags:        types.ListValueMust(types.StringType, nil),
		}
		cfg := repoConfigFromModel(ctx, c, data)
		if cfg.Links == nil {
			t.Fatal("expected non-nil links slice from empty list")
		}
	})
}

func TestUnitRepoResource_repoToModel(t *testing.T) {
	ctx := context.Background()
	c := &client.Client{}

	t.Run("basic_fields", func(t *testing.T) {
		repo := &client.GitRepository{
			ID:   client.OID{OID: "repo123"},
			Name: "my-repo",
			Tags: []string{"prod"},
			Config: client.GitRepositoryConfig{
				ServerID: "srv-1",
			},
		}
		data := &RepoResourceModel{Tags: types.ListValueMust(types.StringType, nil)}
		repoToModel(ctx, c, repo, data)
		if data.ID.ValueString() != "repo123" {
			t.Fatalf("unexpected ID: %s", data.ID.ValueString())
		}
		if data.Name.ValueString() != "my-repo" {
			t.Fatalf("unexpected Name: %s", data.Name.ValueString())
		}
		if data.ServerID.ValueString() != "srv-1" {
			t.Fatalf("unexpected server_id: %s", data.ServerID.ValueString())
		}
		if len(data.Tags.Elements()) != 1 {
			t.Fatalf("expected 1 tag, got %d", len(data.Tags.Elements()))
		}
	})

	t.Run("source_set_when_repo_non_empty", func(t *testing.T) {
		repo := &client.GitRepository{
			ID:   client.OID{OID: "repo456"},
			Name: "git-repo",
			Config: client.GitRepositoryConfig{
				Repo:     "owner/myrepo",
				Branch:   "develop",
				GitHttps: true,
			},
		}
		data := &RepoResourceModel{
			Tags:   types.ListValueMust(types.StringType, nil),
			Source: nil,
		}
		repoToModel(ctx, c, repo, data)
		if data.Source == nil {
			t.Fatal("expected source block set when Repo is non-empty")
		}
		if data.Source.Path.ValueString() != "owner/myrepo" {
			t.Fatalf("expected source.path=owner/myrepo, got %s", data.Source.Path.ValueString())
		}
		if data.Source.Branch.ValueString() != "develop" {
			t.Fatalf("expected source.branch=develop, got %s", data.Source.Branch.ValueString())
		}
		if !data.Source.HttpsEnabled.ValueBool() {
			t.Fatal("expected https_enabled=true")
		}
	})

	t.Run("source_nil_when_all_git_empty_and_prior_nil", func(t *testing.T) {
		repo := &client.GitRepository{
			ID:   client.OID{OID: "repo789"},
			Name: "no-git",
			Config: client.GitRepositoryConfig{
				ServerID: "srv-1",
			},
		}
		data := &RepoResourceModel{
			Tags:   types.ListValueMust(types.StringType, nil),
			Source: nil,
		}
		repoToModel(ctx, c, repo, data)
		if data.Source != nil {
			t.Fatal("expected nil source when no git fields set and prior Source nil")
		}
	})

	t.Run("webhook_set_when_enabled", func(t *testing.T) {
		repo := &client.GitRepository{
			ID:   client.OID{OID: "repo000"},
			Name: "webhook-repo",
			Config: client.GitRepositoryConfig{
				WebhookEnabled: true,
				WebhookSecret:  "s3cr3t",
			},
		}
		data := &RepoResourceModel{Tags: types.ListValueMust(types.StringType, nil)}
		repoToModel(ctx, c, repo, data)
		if data.Webhook == nil {
			t.Fatal("expected non-nil webhook block")
		}
		if !data.Webhook.Enabled.ValueBool() {
			t.Fatal("expected webhook enabled=true")
		}
		if data.Webhook.Secret.ValueString() != "s3cr3t" {
			t.Fatalf("expected webhook secret=s3cr3t, got %s", data.Webhook.Secret.ValueString())
		}
	})

	t.Run("on_clone_set_when_non_empty", func(t *testing.T) {
		repo := &client.GitRepository{
			ID:   client.OID{OID: "repo111"},
			Name: "clone-hooks",
			Config: client.GitRepositoryConfig{
				OnClone: client.SystemCommand{Path: "/scripts", Command: "install.sh"},
			},
		}
		data := &RepoResourceModel{Tags: types.ListValueMust(types.StringType, nil)}
		repoToModel(ctx, c, repo, data)
		if data.OnClone == nil {
			t.Fatal("expected non-nil on_clone")
		}
		if data.OnClone.Path.ValueString() != "/scripts" {
			t.Fatalf("unexpected on_clone.path: %s", data.OnClone.Path.ValueString())
		}
	})

	t.Run("on_clone_nil_when_empty", func(t *testing.T) {
		repo := &client.GitRepository{
			ID:     client.OID{OID: "repo222"},
			Name:   "no-hooks",
			Config: client.GitRepositoryConfig{},
		}
		data := &RepoResourceModel{Tags: types.ListValueMust(types.StringType, nil)}
		repoToModel(ctx, c, repo, data)
		if data.OnClone != nil {
			t.Fatal("expected nil on_clone when API returns empty command")
		}
	})

	t.Run("env_file_path_populates_environment_block", func(t *testing.T) {
		repo := &client.GitRepository{
			ID:   client.OID{OID: "repo333"},
			Name: "env-repo",
			Config: client.GitRepositoryConfig{
				EnvFilePath: "/app/.env",
			},
		}
		data := &RepoResourceModel{Tags: types.ListValueMust(types.StringType, nil)}
		repoToModel(ctx, c, repo, data)
		if data.Environment == nil {
			t.Fatal("expected non-nil environment block when env_file_path set")
		}
		if data.Environment.FilePath.ValueString() != "/app/.env" {
			t.Fatalf("expected env_file_path=/app/.env, got %s", data.Environment.FilePath.ValueString())
		}
	})
}

func TestUnitRepoResource_envStringToMap(t *testing.T) {
	t.Run("empty_string_returns_null_map", func(t *testing.T) {
		m := envStringToMap("")
		if !m.IsNull() {
			t.Fatal("expected null map for empty input")
		}
	})

	t.Run("blank_only_lines_return_null_map", func(t *testing.T) {
		m := envStringToMap("   \n   \n")
		if !m.IsNull() {
			t.Fatal("expected null map when all lines are blank")
		}
	})

	t.Run("simple_key_value_pair", func(t *testing.T) {
		m := envStringToMap("FOO=bar")
		if m.IsNull() {
			t.Fatal("expected non-null map")
		}
		if len(m.Elements()) != 1 {
			t.Fatalf("expected 1 element, got %d", len(m.Elements()))
		}
		val := m.Elements()["FOO"]
		strVal, ok := val.(types.String)
		if !ok {
			t.Fatal("expected val to be types.String")
		}
		if strVal.ValueString() != "bar" {
			t.Fatalf("expected FOO=bar, got %s", strVal.ValueString())
		}
	})

	t.Run("lowercase_key_uppercased", func(t *testing.T) {
		m := envStringToMap("database_url=postgres://localhost/db")
		if _, ok := m.Elements()["DATABASE_URL"]; !ok {
			t.Fatal("expected lowercase key to be uppercased to DATABASE_URL")
		}
	})

	t.Run("line_without_equals_gets_empty_value", func(t *testing.T) {
		m := envStringToMap("MYVAR")
		if m.IsNull() {
			t.Fatal("expected non-null map")
		}
		val := m.Elements()["MYVAR"]
		strVal2, ok := val.(types.String)
		if !ok {
			t.Fatal("expected val to be types.String")
		}
		if strVal2.ValueString() != "" {
			t.Fatalf("expected empty value for key without =, got %q", strVal2.ValueString())
		}
	})

	t.Run("value_with_equals_sign", func(t *testing.T) {
		m := envStringToMap("DSN=host=localhost port=5432")
		val := m.Elements()["DSN"]
		strVal3, ok := val.(types.String)
		if !ok {
			t.Fatal("expected val to be types.String")
		}
		if strVal3.ValueString() != "host=localhost port=5432" {
			t.Fatalf("expected full value after first =, got %q", strVal3.ValueString())
		}
	})

	t.Run("multiple_lines", func(t *testing.T) {
		m := envStringToMap("FOO=1\nBAR=2\nBAZ=3")
		if len(m.Elements()) != 3 {
			t.Fatalf("expected 3 elements, got %d", len(m.Elements()))
		}
	})
}

// ─── Acceptance tests – line endings idempotency ─────────────────────────────

// TestAccRepoResource_onCloneCommandLineEndingsNoDrift verifies that
// on_clone.command with trailing LF or CRLF does not produce plan drift.
// State mirrors the config value; SemanticEquals prevents unnecessary diffs.
func TestAccRepoResource_onCloneCommandLineEndingsNoDrift(t *testing.T) {
	const name = "tf-acc-repo-clone-le"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "komodo_repo" "test" {
  name = %q
  on_clone {
    command = "echo cloned\n"
  }
}
`, name),
				Check: resource.TestCheckResourceAttr("komodo_repo.test", "on_clone.command", "echo cloned\n"),
			},
			{
				Config: fmt.Sprintf(`
resource "komodo_repo" "test" {
  name = %q
  on_clone {
    command = "echo cloned\n"
  }
}
`, name),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			// CRLF variant.
			{
				Config: fmt.Sprintf(`
resource "komodo_repo" "test" {
  name = %q
  on_clone {
    command = "step1\r\nstep2\r\n"
  }
}
`, name),
				Check: resource.TestCheckResourceAttr("komodo_repo.test", "on_clone.command", "step1\r\nstep2\r\n"),
			},
			{
				Config: fmt.Sprintf(`
resource "komodo_repo" "test" {
  name = %q
  on_clone {
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

// TestAccRepoResource_onPullCommandLineEndingsNoDrift verifies that
// on_pull.command with trailing LF or CRLF does not produce plan drift.
// State mirrors the config value; SemanticEquals prevents unnecessary diffs.
func TestAccRepoResource_onPullCommandLineEndingsNoDrift(t *testing.T) {
	const name = "tf-acc-repo-pull-le"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "komodo_repo" "test" {
  name = %q
  on_pull {
    command = "echo pulled\n"
  }
}
`, name),
				Check: resource.TestCheckResourceAttr("komodo_repo.test", "on_pull.command", "echo pulled\n"),
			},
			{
				Config: fmt.Sprintf(`
resource "komodo_repo" "test" {
  name = %q
  on_pull {
    command = "echo pulled\n"
  }
}
`, name),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			// CRLF variant.
			{
				Config: fmt.Sprintf(`
resource "komodo_repo" "test" {
  name = %q
  on_pull {
    command = "step1\r\nstep2\r\n"
  }
}
`, name),
				Check: resource.TestCheckResourceAttr("komodo_repo.test", "on_pull.command", "step1\r\nstep2\r\n"),
			},
			{
				Config: fmt.Sprintf(`
resource "komodo_repo" "test" {
  name = %q
  on_pull {
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
