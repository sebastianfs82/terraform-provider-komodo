// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"testing"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	tftest "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"
)

// testAccTerminalServerID returns the ID of the first available server in the Komodo instance.
// Falls back to KOMODO_TEST_SERVER_ID if set. Skips the test if no servers are found.
func testAccTerminalServerID(t *testing.T) string {
	t.Helper()
	return testAccLookupServerID(t, "terminal acceptance tests")
}

// testAccTerminalNginxStackConfig returns a Terraform config that creates a
// komodo_stack hosting a single nginx service. Used for container and stack
// terminal tests that need a real target (without requiring env vars).
func testAccTerminalNginxStackConfig(name string) string {
	return fmt.Sprintf(`
data "komodo_servers" "all" {}

resource "komodo_stack" "nginx" {
  name      = %q
  server_id = data.komodo_servers.all.servers[0].id

  source {
    contents = <<-EOT
      services:
        web:
          image: nginx:latest
    EOT
  }

  lifecycle {
    action_trigger {
      events  = [after_create, after_update]
      actions = [action.komodo_stack.deploy]
    }
  }
}

action "komodo_stack" "deploy" {
  config {
    id     = komodo_stack.nginx.id
    action = "deploy"
  }
}

resource "time_sleep" "wait" {
  create_duration = "60s"
  depends_on      = [komodo_stack.nginx]
}
`, name)
}

// --- Unit tests (no live API required) ---

// TestTerminalResource_ValidateConfig_attachOnServer verifies that
// mode="attach" on target_type="Server" produces a diagnostic error.
func TestTerminalResource_ValidateConfig_attachOnServer(t *testing.T) {
	ctx := context.Background()
	r := &TerminalResource{}

	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)

	rawVal := tftypes.NewValue(
		schemaResp.Schema.Type().TerraformType(ctx),
		map[string]tftypes.Value{
			"id":             tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			"name":           tftypes.NewValue(tftypes.String, "test"),
			"target_type":    tftypes.NewValue(tftypes.String, "Server"),
			"target_id":      tftypes.NewValue(tftypes.String, "my-server"),
			"container":      tftypes.NewValue(tftypes.String, ""),
			"service":        tftypes.NewValue(tftypes.String, ""),
			"mode":           tftypes.NewValue(tftypes.String, "attach"),
			"command":        tftypes.NewValue(tftypes.String, ""),
			"created_at":     tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			"stored_size_kb": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
		},
	)

	req := fwresource.ValidateConfigRequest{
		Config: tfsdk.Config{
			Raw:    rawVal,
			Schema: schemaResp.Schema,
		},
	}
	resp := &fwresource.ValidateConfigResponse{}
	r.ValidateConfig(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Errorf("expected a diagnostic error for mode=attach on Server target, got none")
	}
}

// TestTerminalResource_ValidateConfig_attachOnContainerIsValid verifies that
// mode="attach" on target_type="Container" does NOT produce an error.
func TestTerminalResource_ValidateConfig_attachOnContainerIsValid(t *testing.T) {
	ctx := context.Background()
	r := &TerminalResource{}

	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)

	rawVal := tftypes.NewValue(
		schemaResp.Schema.Type().TerraformType(ctx),
		map[string]tftypes.Value{
			"id":             tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			"name":           tftypes.NewValue(tftypes.String, "test"),
			"target_type":    tftypes.NewValue(tftypes.String, "Container"),
			"target_id":      tftypes.NewValue(tftypes.String, "my-server"),
			"container":      tftypes.NewValue(tftypes.String, "my-container"),
			"service":        tftypes.NewValue(tftypes.String, ""),
			"mode":           tftypes.NewValue(tftypes.String, "attach"),
			"command":        tftypes.NewValue(tftypes.String, ""),
			"created_at":     tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			"stored_size_kb": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
		},
	)

	req := fwresource.ValidateConfigRequest{
		Config: tfsdk.Config{
			Raw:    rawVal,
			Schema: schemaResp.Schema,
		},
	}
	resp := &fwresource.ValidateConfigResponse{}
	r.ValidateConfig(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected diagnostic error for mode=attach on Container target: %s", resp.Diagnostics)
	}
}

// --- Resource acceptance tests ---

func TestAccTerminalResource_basic(t *testing.T) {
	serverID := testAccTerminalServerID(t)
	tftest.Test(t, tftest.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tftest.TestStep{
			{
				Config: testAccTerminalResourceServerConfig(serverID, "tf-test-terminal", ""),
				Check: tftest.ComposeAggregateTestCheckFunc(
					tftest.TestCheckResourceAttr("komodo_terminal.test", "target_type", "Server"),
					tftest.TestCheckResourceAttr("komodo_terminal.test", "target_id", serverID),
					tftest.TestCheckResourceAttr("komodo_terminal.test", "name", "tf-test-terminal"),
					tftest.TestCheckResourceAttrSet("komodo_terminal.test", "id"),
					tftest.TestCheckResourceAttrSet("komodo_terminal.test", "created_at"),
					tftest.TestCheckResourceAttrSet("komodo_terminal.test", "stored_size_kb"),
				),
			},
			// Verify state is stable on a second plan (no-diff).
			{
				Config:   testAccTerminalResourceServerConfig(serverID, "tf-test-terminal", ""),
				PlanOnly: true,
			},
		},
	})
}

func TestAccTerminalResource_withCommand(t *testing.T) {
	serverID := testAccTerminalServerID(t)
	tftest.Test(t, tftest.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tftest.TestStep{
			{
				Config: testAccTerminalResourceServerConfig(serverID, "tf-test-terminal-bash", "bash"),
				Check: tftest.ComposeAggregateTestCheckFunc(
					tftest.TestCheckResourceAttr("komodo_terminal.test", "target_type", "Server"),
					tftest.TestCheckResourceAttr("komodo_terminal.test", "target_id", serverID),
					tftest.TestCheckResourceAttr("komodo_terminal.test", "name", "tf-test-terminal-bash"),
					tftest.TestCheckResourceAttr("komodo_terminal.test", "command", "bash"),
					tftest.TestCheckResourceAttrSet("komodo_terminal.test", "id"),
				),
			},
		},
	})
}

// TestAccTerminalResource_importState_twopart tests the 2-part import ("target_id:name"),
// which defaults target_type to "Server".
func TestAccTerminalResource_importState_twopart(t *testing.T) {
	serverID := testAccTerminalServerID(t)
	tftest.Test(t, tftest.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tftest.TestStep{
			{
				Config: testAccTerminalResourceServerConfig(serverID, "tf-import-terminal", ""),
				Check: tftest.ComposeAggregateTestCheckFunc(
					tftest.TestCheckResourceAttrSet("komodo_terminal.test", "id"),
				),
			},
			{
				ResourceName:            "komodo_terminal.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"mode", "command", "stored_size_kb"},
			},
		},
	})
}

// TestAccTerminalResource_importState_threepart tests the 3-part import
// ("target_type:target_id:name") for non-Server targets.
func TestAccTerminalResource_importState_threepart(t *testing.T) {
	serverID := testAccTerminalServerID(t)
	tftest.Test(t, tftest.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tftest.TestStep{
			{
				Config: testAccTerminalResourceServerConfig(serverID, "tf-import3-terminal", ""),
				Check: tftest.ComposeAggregateTestCheckFunc(
					tftest.TestCheckResourceAttrSet("komodo_terminal.test", "id"),
				),
			},
			{
				ResourceName:            "komodo_terminal.test",
				ImportState:             true,
				ImportStateIdFunc:       testAccTerminalImportID3Part("komodo_terminal.test"),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"mode", "command", "stored_size_kb"},
			},
		},
	})
}

// TestAccTerminalResource_container tests creating a terminal against a container target.
// Uses a self-contained komodo_stack with an nginx compose service so no env vars are needed.
func TestAccTerminalResource_container(t *testing.T) {
	tftest.Test(t, tftest.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		ExternalProviders: map[string]tftest.ExternalProvider{
			"time": {Source: "hashicorp/time"},
		},
		Steps: []tftest.TestStep{
			{
				Config: testAccTerminalNginxStackConfig("tf-nginx-container-stack") + `
resource "komodo_terminal" "test" {
  target_type = "Container"
  target_id   = data.komodo_servers.all.servers[0].id
  container   = "tf-nginx-container-stack-web-1"
  name        = "tf-container-terminal"
  command     = "/bin/sh"
  depends_on  = [time_sleep.wait]
}
`,
				Check: tftest.ComposeAggregateTestCheckFunc(
					tftest.TestCheckResourceAttr("komodo_terminal.test", "target_type", "Container"),
					tftest.TestCheckResourceAttr("komodo_terminal.test", "container", "tf-nginx-container-stack-web-1"),
					tftest.TestCheckResourceAttr("komodo_terminal.test", "name", "tf-container-terminal"),
					tftest.TestCheckResourceAttrSet("komodo_terminal.test", "id"),
					tftest.TestCheckResourceAttrSet("komodo_terminal.test", "created_at"),
				),
			},
		},
	})
}

// TestAccTerminalResource_stack tests creating a terminal against a stack service target.
// Uses a self-contained komodo_stack with an nginx compose service so no env vars are needed.
func TestAccTerminalResource_stack(t *testing.T) {
	tftest.Test(t, tftest.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		ExternalProviders: map[string]tftest.ExternalProvider{
			"time": {Source: "hashicorp/time"},
		},
		Steps: []tftest.TestStep{
			{
				Config: testAccTerminalNginxStackConfig("tf-nginx-stack-terminal") + `
resource "komodo_terminal" "test" {
  target_type = "Stack"
  target_id   = komodo_stack.nginx.id
  service     = "web"
  name        = "tf-stack-terminal"
  depends_on  = [time_sleep.wait]
}
`,
				Check: tftest.ComposeAggregateTestCheckFunc(
					tftest.TestCheckResourceAttr("komodo_terminal.test", "target_type", "Stack"),
					tftest.TestCheckResourceAttrPair("komodo_terminal.test", "target_id", "komodo_stack.nginx", "id"),
					tftest.TestCheckResourceAttr("komodo_terminal.test", "service", "web"),
					tftest.TestCheckResourceAttr("komodo_terminal.test", "name", "tf-stack-terminal"),
					tftest.TestCheckResourceAttrSet("komodo_terminal.test", "id"),
				),
			},
		},
	})
}

// TestAccTerminalResource_attachMode tests that mode="attach" is accepted for a container target.
// Uses a self-contained komodo_stack with an nginx compose service so no env vars are needed.
func TestAccTerminalResource_attachMode(t *testing.T) {
	tftest.Test(t, tftest.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		ExternalProviders: map[string]tftest.ExternalProvider{
			"time": {Source: "hashicorp/time"},
		},
		Steps: []tftest.TestStep{
			{
				Config: testAccTerminalNginxStackConfig("tf-nginx-attach-stack") + `
resource "komodo_terminal" "test" {
  target_type = "Container"
  target_id   = data.komodo_servers.all.servers[0].id
  container   = "tf-nginx-attach-stack-web-1"
  name        = "tf-attach-terminal"
  mode        = "attach"
  depends_on  = [time_sleep.wait]
}
`,
				Check: tftest.ComposeAggregateTestCheckFunc(
					tftest.TestCheckResourceAttr("komodo_terminal.test", "mode", "attach"),
					tftest.TestCheckResourceAttr("komodo_terminal.test", "target_type", "Container"),
					tftest.TestCheckResourceAttrSet("komodo_terminal.test", "id"),
				),
			},
		},
	})
}

// --- Data source acceptance tests ---

func TestAccTerminalDataSource_basic(t *testing.T) {
	serverID := testAccTerminalServerID(t)
	tftest.Test(t, tftest.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tftest.TestStep{
			{
				Config: testAccTerminalDataSourceConfig(serverID, "tf-ds-terminal"),
				Check: tftest.ComposeAggregateTestCheckFunc(
					tftest.TestCheckResourceAttr("data.komodo_terminal.test", "target_type", "Server"),
					tftest.TestCheckResourceAttr("data.komodo_terminal.test", "target_id", serverID),
					tftest.TestCheckResourceAttr("data.komodo_terminal.test", "name", "tf-ds-terminal"),
					tftest.TestCheckResourceAttrSet("data.komodo_terminal.test", "created_at"),
					tftest.TestCheckResourceAttrSet("data.komodo_terminal.test", "stored_size_kb"),
				),
			},
		},
	})
}

// TestAccTerminalsDataSource_basic lists all terminals (no target filter).
func TestAccTerminalsDataSource_basic(t *testing.T) {
	serverID := testAccTerminalServerID(t)
	tftest.Test(t, tftest.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tftest.TestStep{
			{
				// Create a terminal first so the list is non-empty, then verify the data source.
				Config: testAccTerminalsDataSourceConfig(serverID),
				Check: tftest.ComposeAggregateTestCheckFunc(
					tftest.TestCheckResourceAttrSet("data.komodo_terminals.test", "terminals.#"),
				),
			},
		},
	})
}

// --- Config helpers ---

func testAccTerminalResourceServerConfig(serverID, name, command string) string {
	if command != "" {
		return fmt.Sprintf(`
resource "komodo_terminal" "test" {
  target_type = "Server"
  target_id   = %q
  name        = %q
  command     = %q
}
`, serverID, name, command)
	}
	return fmt.Sprintf(`
resource "komodo_terminal" "test" {
  target_type = "Server"
  target_id   = %q
  name        = %q
}
`, serverID, name)
}

func testAccTerminalDataSourceConfig(serverID, name string) string {
	return fmt.Sprintf(`
resource "komodo_terminal" "test" {
  target_type = "Server"
  target_id   = %q
  name        = %q
}

data "komodo_terminal" "test" {
  target_type = komodo_terminal.test.target_type
  target_id   = komodo_terminal.test.target_id
  name        = komodo_terminal.test.name
}
`, serverID, name)
}

func testAccTerminalsDataSourceConfig(serverID string) string {
	return fmt.Sprintf(`
resource "komodo_terminal" "seed" {
  target_type = "Server"
  target_id   = %q
  name        = "tf-terminals-ds-seed"
}

data "komodo_terminals" "test" {
  depends_on = [komodo_terminal.seed]
}
`, serverID)
}

// testAccTerminalImportID3Part returns an ImportStateIdFunc that builds a
// 3-part import ID ("target_type:target_id:name") from the resource's state.
func testAccTerminalImportID3Part(resourceAddress string) tftest.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceAddress]
		if !ok {
			return "", fmt.Errorf("resource %q not found in state", resourceAddress)
		}
		targetType := rs.Primary.Attributes["target_type"]
		targetID := rs.Primary.Attributes["target_id"]
		name := rs.Primary.Attributes["name"]
		return targetType + ":" + targetID + ":" + name, nil
	}
}

// ─── Unit tests ──────────────────────────────────────────────────────────────

func wrongRawTerminalPlan(t *testing.T, r *TerminalResource) tfsdk.Plan {
	t.Helper()
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schemaResp.Schema,
	}
}

func wrongRawTerminalState(t *testing.T, r *TerminalResource) tfsdk.State {
	t.Helper()
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	return tfsdk.State{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schemaResp.Schema,
	}
}

func TestUnitTerminalResource_configure(t *testing.T) {
	t.Run("wrong_type", func(t *testing.T) {
		r := &TerminalResource{}
		req := fwresource.ConfigureRequest{ProviderData: "not-a-client"}
		resp := &fwresource.ConfigureResponse{}
		r.Configure(context.Background(), req, resp)
		if !resp.Diagnostics.HasError() {
			t.Fatal("expected diagnostic error for wrong ProviderData type")
		}
	})
}

func TestUnitTerminalResource_createPlanGetError(t *testing.T) {
	r := &TerminalResource{client: &client.Client{}}
	req := fwresource.CreateRequest{Plan: wrongRawTerminalPlan(t, r)}
	resp := &fwresource.CreateResponse{}
	r.Create(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed plan")
	}
}

func TestUnitTerminalResource_readStateGetError(t *testing.T) {
	r := &TerminalResource{client: &client.Client{}}
	req := fwresource.ReadRequest{State: wrongRawTerminalState(t, r)}
	resp := &fwresource.ReadResponse{}
	r.Read(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed state")
	}
}

func TestUnitTerminalResource_deleteStateGetError(t *testing.T) {
	r := &TerminalResource{client: &client.Client{}}
	req := fwresource.DeleteRequest{State: wrongRawTerminalState(t, r)}
	resp := &fwresource.DeleteResponse{}
	r.Delete(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed state")
	}
}

func TestUnitTerminalResource_updateIsNoop(t *testing.T) {
	r := &TerminalResource{}
	req := fwresource.UpdateRequest{}
	resp := &fwresource.UpdateResponse{}
	r.Update(context.Background(), req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error in no-op Update: %v", resp.Diagnostics)
	}
}
