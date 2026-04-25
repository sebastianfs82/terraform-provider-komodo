// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// ── Resource tests ────────────────────────────────────────────────────────────

func TestAccResourceSyncResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSyncResourceConfig("tf-acc-rsync-basic", "# managed by terraform"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_resource_sync.test", "name", "tf-acc-rsync-basic"),
					resource.TestCheckResourceAttr("komodo_resource_sync.test", "source.contents", "# managed by terraform"),
					resource.TestCheckResourceAttrSet("komodo_resource_sync.test", "id"),
				),
			},
		},
	})
}

func TestAccResourceSyncResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSyncResourceConfig("tf-acc-rsync-update", "# v1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_resource_sync.test", "source.contents", "# v1"),
				),
			},
			{
				Config: testAccResourceSyncResourceConfig("tf-acc-rsync-update", "# v2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_resource_sync.test", "source.contents", "# v2"),
				),
			},
		},
	})
}

func TestAccResourceSyncResource_import(t *testing.T) {
	var syncID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSyncResourceConfig("tf-acc-rsync-import", "# import test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_resource_sync.test", "id"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["komodo_resource_sync.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						syncID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config:            testAccResourceSyncResourceConfig("tf-acc-rsync-import", "# import test"),
				ResourceName:      "komodo_resource_sync.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(_ *terraform.State) (string, error) { return syncID, nil },
			},
		},
	})
}

func TestAccResourceSyncResource_disappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSyncResourceConfig("tf-acc-rsync-disappears", "# disappears"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_resource_sync.test", "id"),
					testAccResourceSyncDisappears("komodo_resource_sync.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccResourceSyncResource_rename(t *testing.T) {
	var savedID string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSyncResourceConfig("tf-acc-rsync-rename-orig", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_resource_sync.test", "name", "tf-acc-rsync-rename-orig"),
					resource.TestCheckResourceAttrSet("komodo_resource_sync.test", "id"),
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["komodo_resource_sync.test"]
						savedID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config: testAccResourceSyncResourceConfig("tf-acc-rsync-rename-new", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_resource_sync.test", "name", "tf-acc-rsync-rename-new"),
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["komodo_resource_sync.test"]
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

func testAccResourceSyncDisappears(resourceName string) resource.TestCheckFunc {
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
		return c.DeleteResourceSync(context.Background(), rs.Primary.ID)
	}
}

func testAccResourceSyncResourceConfig(name, fileContents string) string {
	return fmt.Sprintf(`
resource "komodo_resource_sync" "test" {
  name = %q

  source {
    contents = %q
  }
}
`, name, fileContents)
}

// ── Data source test ──────────────────────────────────────────────────────────

func TestAccResourceSyncDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSyncDataSourceConfig("tf-acc-rsync-ds", "# ds test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.komodo_resource_sync.test", "id",
						"komodo_resource_sync.test", "id",
					),
					resource.TestCheckResourceAttr("data.komodo_resource_sync.test", "name", "tf-acc-rsync-ds"),
					resource.TestCheckResourceAttr("data.komodo_resource_sync.test", "contents", "# ds test"),
				),
			},
		},
	})
}

func testAccResourceSyncDataSourceConfig(name, fileContents string) string {
	return fmt.Sprintf(`
resource "komodo_resource_sync" "test" {
  name = %q

  source {
    contents = %q
  }
}

data "komodo_resource_sync" "test" {
  id = komodo_resource_sync.test.id
}
`, name, fileContents)
}

// TestAccResourceSyncDataSource_byName verifies that the data source can look
// up a resource sync by name instead of by ID.
func TestAccResourceSyncDataSource_byName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSyncDataSourceByNameConfig("tf-acc-rsync-ds-name", "# by name"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.komodo_resource_sync.test", "id",
						"komodo_resource_sync.test", "id",
					),
					resource.TestCheckResourceAttr("data.komodo_resource_sync.test", "name", "tf-acc-rsync-ds-name"),
					resource.TestCheckResourceAttr("data.komodo_resource_sync.test", "contents", "# by name"),
				),
			},
		},
	})
}

func testAccResourceSyncDataSourceByNameConfig(name, fileContents string) string {
	return fmt.Sprintf(`
resource "komodo_resource_sync" "test" {
  name = %q

  source {
    contents = %q
  }
}

data "komodo_resource_sync" "test" {
  name = komodo_resource_sync.test.name
}
`, name, fileContents)
}

func TestAccResourceSyncResource_tags(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSyncWithTagConfig("tf-acc-rsync-tags"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_resource_sync.test", "tags.#", "1"),
					resource.TestCheckResourceAttrPair("komodo_resource_sync.test", "tags.0", "komodo_tag.test", "id"),
				),
			},
			{
				Config: testAccResourceSyncClearTagsConfig("tf-acc-rsync-tags"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_resource_sync.test", "tags.#", "0"),
				),
			},
		},
	})
}

func testAccResourceSyncWithTagConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_tag" "test" {
  name  = "tf-acc-tag-resource-sync"
  color = "Green"
}

resource "komodo_resource_sync" "test" {
  name = %q
  tags = [komodo_tag.test.id]

  source {
    contents = ""
  }
}
`, name)
}

func testAccResourceSyncClearTagsConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_resource_sync" "test" {
  name = %q
  tags = []

  source {
    contents = ""
  }
}
`, name)
}

// ─── Unit tests ──────────────────────────────────────────────────────────────

func wrongRawResourceSyncPlan(t *testing.T, r *ResourceSyncResource) tfsdk.Plan {
	t.Helper()
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schemaResp.Schema,
	}
}

func wrongRawResourceSyncState(t *testing.T, r *ResourceSyncResource) tfsdk.State {
	t.Helper()
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	return tfsdk.State{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schemaResp.Schema,
	}
}

func TestUnitResourceSyncResource_configure(t *testing.T) {
	t.Run("wrong_type", func(t *testing.T) {
		r := &ResourceSyncResource{}
		req := fwresource.ConfigureRequest{ProviderData: "not-a-client"}
		resp := &fwresource.ConfigureResponse{}
		r.Configure(context.Background(), req, resp)
		if !resp.Diagnostics.HasError() {
			t.Fatal("expected diagnostic error for wrong ProviderData type")
		}
	})
}

func TestUnitResourceSyncResource_createPlanGetError(t *testing.T) {
	r := &ResourceSyncResource{client: &client.Client{}}
	req := fwresource.CreateRequest{Plan: wrongRawResourceSyncPlan(t, r)}
	resp := &fwresource.CreateResponse{}
	r.Create(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed plan")
	}
}

func TestUnitResourceSyncResource_readStateGetError(t *testing.T) {
	r := &ResourceSyncResource{client: &client.Client{}}
	req := fwresource.ReadRequest{State: wrongRawResourceSyncState(t, r)}
	resp := &fwresource.ReadResponse{}
	r.Read(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed state")
	}
}

func TestUnitResourceSyncResource_updatePlanGetError(t *testing.T) {
	r := &ResourceSyncResource{client: &client.Client{}}
	req := fwresource.UpdateRequest{Plan: wrongRawResourceSyncPlan(t, r)}
	resp := &fwresource.UpdateResponse{}
	r.Update(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed plan")
	}
}

func TestUnitResourceSyncResource_deleteStateGetError(t *testing.T) {
	r := &ResourceSyncResource{client: &client.Client{}}
	req := fwresource.DeleteRequest{State: wrongRawResourceSyncState(t, r)}
	resp := &fwresource.DeleteResponse{}
	r.Delete(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed state")
	}
}

func TestUnitResourceSyncResource_partialConfigFromModel(t *testing.T) {
	ctx := context.Background()
	c := &client.Client{}

	t.Run("https_url", func(t *testing.T) {
		m := &ResourceSyncResourceModel{
			Source: &ResourceSyncSourceModel{
				RepoID: types.StringNull(),
				URL:    types.StringValue("https://github.com/owner/repo"),
			},
			Webhook: &WebhookModel{Enabled: types.BoolValue(false), Secret: types.StringValue("")},
		}
		cfg := partialResourceSyncConfigFromModel(ctx, c, m)
		if cfg.GitProvider == nil || *cfg.GitProvider != "github.com/owner/repo" {
			t.Fatalf("expected GitProvider %q, got %v", "github.com/owner/repo", cfg.GitProvider)
		}
		if cfg.GitHttps == nil || !*cfg.GitHttps {
			t.Fatal("expected GitHttps=true for https URL")
		}
	})

	t.Run("http_url", func(t *testing.T) {
		m := &ResourceSyncResourceModel{
			Source: &ResourceSyncSourceModel{
				RepoID: types.StringNull(),
				URL:    types.StringValue("http://gitea.local/owner/repo"),
			},
			Webhook: &WebhookModel{Enabled: types.BoolValue(false), Secret: types.StringValue("")},
		}
		cfg := partialResourceSyncConfigFromModel(ctx, c, m)
		if cfg.GitProvider == nil || *cfg.GitProvider != "gitea.local/owner/repo" {
			t.Fatalf("expected GitProvider %q, got %v", "gitea.local/owner/repo", cfg.GitProvider)
		}
		if cfg.GitHttps == nil || *cfg.GitHttps {
			t.Fatal("expected GitHttps=false for http URL")
		}
	})

	t.Run("no_prefix_url", func(t *testing.T) {
		m := &ResourceSyncResourceModel{
			Source: &ResourceSyncSourceModel{
				RepoID: types.StringNull(),
				URL:    types.StringValue("github.com/owner/repo"),
			},
			Webhook: &WebhookModel{Enabled: types.BoolValue(false), Secret: types.StringValue("")},
		}
		cfg := partialResourceSyncConfigFromModel(ctx, c, m)
		if cfg.GitProvider == nil || *cfg.GitProvider != "github.com/owner/repo" {
			t.Fatalf("expected GitProvider %q, got %v", "github.com/owner/repo", cfg.GitProvider)
		}
		if cfg.GitHttps == nil || !*cfg.GitHttps {
			t.Fatal("expected GitHttps=true for no-prefix URL")
		}
	})

	t.Run("linked_repo_id", func(t *testing.T) {
		m := &ResourceSyncResourceModel{
			Source: &ResourceSyncSourceModel{
				RepoID: types.StringValue("my-repo-id"),
				URL:    types.StringNull(),
			},
			Webhook: &WebhookModel{Enabled: types.BoolValue(false), Secret: types.StringValue("")},
		}
		cfg := partialResourceSyncConfigFromModel(ctx, c, m)
		if cfg.LinkedRepo == nil || *cfg.LinkedRepo != "my-repo-id" {
			t.Fatal("expected LinkedRepo to be set")
		}
	})

	t.Run("nil_webhook", func(t *testing.T) {
		m := &ResourceSyncResourceModel{
			Source: &ResourceSyncSourceModel{
				RepoID: types.StringNull(),
				URL:    types.StringValue("https://github.com/owner/repo"),
			},
			Webhook: nil,
		}
		cfg := partialResourceSyncConfigFromModel(ctx, c, m)
		if cfg.WebhookEnabled == nil || *cfg.WebhookEnabled {
			t.Fatal("expected WebhookEnabled=false when webhook block is nil")
		}
	})

	t.Run("scope_list", func(t *testing.T) {
		m := &ResourceSyncResourceModel{
			Scope: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("resources"),
				types.StringValue("variables"),
				types.StringValue("user_groups"),
			}),
			Webhook: nil,
		}
		cfg := partialResourceSyncConfigFromModel(ctx, c, m)
		if cfg.IncludeResources == nil || !*cfg.IncludeResources {
			t.Fatal("expected IncludeResources=true")
		}
		if cfg.IncludeVariables == nil || !*cfg.IncludeVariables {
			t.Fatal("expected IncludeVariables=true")
		}
		if cfg.IncludeUserGroups == nil || !*cfg.IncludeUserGroups {
			t.Fatal("expected IncludeUserGroups=true")
		}
	})

	t.Run("managed_mode", func(t *testing.T) {
		m := &ResourceSyncResourceModel{
			Webhook: nil,
			ManagedMode: &ResourceSyncManagedModeModel{
				Enabled:   types.BoolValue(true),
				TagFilter: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("env:prod")}),
			},
		}
		cfg := partialResourceSyncConfigFromModel(ctx, c, m)
		if cfg.Managed == nil || !*cfg.Managed {
			t.Fatal("expected Managed=true")
		}
		if len(cfg.MatchTags) != 1 || cfg.MatchTags[0] != "env:prod" {
			t.Fatal("expected MatchTags=[env:prod]")
		}
	})
}

func TestUnitResourceSyncResource_resourceSyncToModel(t *testing.T) {
	ctx := context.Background()
	c := &client.Client{}

	t.Run("basic_fields", func(t *testing.T) {
		rs := &client.ResourceSync{
			ID:   client.OID{OID: "sync001"},
			Name: "my-sync",
			Tags: []string{"prod"},
			Config: client.ResourceSyncConfig{
				Delete:       true,
				PendingAlert: true,
			},
		}
		m := &ResourceSyncResourceModel{}
		resourceSyncToModel(ctx, c, rs, m)
		if m.ID.ValueString() != "sync001" {
			t.Fatalf("unexpected ID: %s", m.ID.ValueString())
		}
		if m.Name.ValueString() != "my-sync" {
			t.Fatalf("unexpected Name: %s", m.Name.ValueString())
		}
		if !m.Delete.ValueBool() {
			t.Fatal("expected delete=true")
		}
		if !m.AlertsEnabled.ValueBool() {
			t.Fatal("expected alerts_enabled=true")
		}
		if m.Source != nil {
			t.Fatal("expected nil source when no git fields set")
		}
		if m.Webhook != nil {
			t.Fatal("expected nil webhook when disabled")
		}
		if len(m.Tags.Elements()) != 1 {
			t.Fatalf("expected 1 tag, got %d", len(m.Tags.Elements()))
		}
	})

	t.Run("source_with_linked_repo", func(t *testing.T) {
		rs := &client.ResourceSync{
			ID:   client.OID{OID: "sync002"},
			Name: "linked-sync",
			Config: client.ResourceSyncConfig{
				LinkedRepo: "repo-abc",
			},
		}
		m := &ResourceSyncResourceModel{}
		resourceSyncToModel(ctx, c, rs, m)
		if m.Source == nil {
			t.Fatal("expected non-nil source for linked_repo")
		}
		if m.Source.RepoID.ValueString() != "repo-abc" {
			t.Fatalf("expected RepoID=repo-abc, got %s", m.Source.RepoID.ValueString())
		}
	})

	t.Run("source_file_contents_trailing_newline_stripped", func(t *testing.T) {
		rs := &client.ResourceSync{
			ID:   client.OID{OID: "sync003"},
			Name: "file-sync",
			Config: client.ResourceSyncConfig{
				FileContents: "resource: foo\n",
			},
		}
		m := &ResourceSyncResourceModel{
			Source: &ResourceSyncSourceModel{FileContents: types.StringNull()},
		}
		resourceSyncToModel(ctx, c, rs, m)
		if m.Source == nil {
			t.Fatal("expected non-nil source")
		}
		if m.Source.FileContents.ValueString() != "resource: foo" {
			t.Fatalf("expected trailing newline stripped, got %q", m.Source.FileContents.ValueString())
		}
	})

	t.Run("webhook_set_when_enabled", func(t *testing.T) {
		rs := &client.ResourceSync{
			ID:   client.OID{OID: "sync004"},
			Name: "webhook-sync",
			Config: client.ResourceSyncConfig{
				WebhookEnabled: true,
				WebhookSecret:  "tok",
			},
		}
		m := &ResourceSyncResourceModel{}
		resourceSyncToModel(ctx, c, rs, m)
		if m.Webhook == nil {
			t.Fatal("expected non-nil webhook block")
		}
		if !m.Webhook.Enabled.ValueBool() {
			t.Fatal("expected webhook enabled=true")
		}
		if m.Webhook.Secret.ValueString() != "tok" {
			t.Fatalf("expected webhook secret=tok, got %s", m.Webhook.Secret.ValueString())
		}
	})

	t.Run("scope_from_include_booleans", func(t *testing.T) {
		rs := &client.ResourceSync{
			ID:   client.OID{OID: "sync005"},
			Name: "scoped-sync",
			Config: client.ResourceSyncConfig{
				IncludeResources:  true,
				IncludeVariables:  true,
				IncludeUserGroups: false,
			},
		}
		m := &ResourceSyncResourceModel{}
		resourceSyncToModel(ctx, c, rs, m)
		var scopeItems []string
		_ = m.Scope.ElementsAs(ctx, &scopeItems, false)
		found := map[string]bool{}
		for _, s := range scopeItems {
			found[s] = true
		}
		if !found["resources"] {
			t.Fatal("expected 'resources' in scope list")
		}
		if !found["variables"] {
			t.Fatal("expected 'variables' in scope list")
		}
		if found["user_groups"] {
			t.Fatal("expected 'user_groups' not in scope list")
		}
	})

	t.Run("managed_mode_block_set", func(t *testing.T) {
		rs := &client.ResourceSync{
			ID:   client.OID{OID: "sync006"},
			Name: "managed-sync",
			Config: client.ResourceSyncConfig{
				Managed:   true,
				MatchTags: []string{"env:prod"},
			},
		}
		m := &ResourceSyncResourceModel{}
		resourceSyncToModel(ctx, c, rs, m)
		if m.ManagedMode == nil {
			t.Fatal("expected non-nil managed_mode block")
		}
		if !m.ManagedMode.Enabled.ValueBool() {
			t.Fatal("expected managed_mode.enabled=true")
		}
		var tagItems []string
		_ = m.ManagedMode.TagFilter.ElementsAs(ctx, &tagItems, false)
		if len(tagItems) != 1 || tagItems[0] != "env:prod" {
			t.Fatalf("expected tag_filter=[env:prod], got %v", tagItems)
		}
	})
}
