// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	tftest "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// testAccTerminalServerID returns the server ID/name to use for terminal acceptance tests.
// Set KOMODO_TEST_SERVER_ID to the name or ID of an available server in your Komodo instance.
func testAccTerminalServerID(t *testing.T) string {
	t.Helper()
	v := os.Getenv("KOMODO_TEST_SERVER_ID")
	if v == "" {
		t.Skip("KOMODO_TEST_SERVER_ID must be set for terminal acceptance tests")
	}
	return v
}

// testAccTerminalContainerServer returns the server ID that hosts the container for container-terminal tests.
// Set KOMODO_TEST_CONTAINER_SERVER_ID and KOMODO_TEST_CONTAINER_NAME for container acceptance tests.
func testAccTerminalContainerEnv(t *testing.T) (serverID, containerName string) {
	t.Helper()
	serverID = os.Getenv("KOMODO_TEST_CONTAINER_SERVER_ID")
	containerName = os.Getenv("KOMODO_TEST_CONTAINER_NAME")
	if serverID == "" || containerName == "" {
		t.Skip("KOMODO_TEST_CONTAINER_SERVER_ID and KOMODO_TEST_CONTAINER_NAME must be set for container terminal tests")
	}
	return
}

// testAccTerminalStackEnv returns the stack ID and service name for stack-terminal tests.
// Set KOMODO_TEST_STACK_ID and KOMODO_TEST_STACK_SERVICE for stack acceptance tests.
func testAccTerminalStackEnv(t *testing.T) (stackID, service string) {
	t.Helper()
	stackID = os.Getenv("KOMODO_TEST_STACK_ID")
	service = os.Getenv("KOMODO_TEST_STACK_SERVICE")
	if stackID == "" || service == "" {
		t.Skip("KOMODO_TEST_STACK_ID and KOMODO_TEST_STACK_SERVICE must be set for stack terminal tests")
	}
	return
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
				ImportStateVerifyIgnore: []string{"mode", "command"},
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
				ImportStateVerifyIgnore: []string{"mode", "command"},
			},
		},
	})
}

// TestAccTerminalResource_container tests creating a terminal against a container target.
func TestAccTerminalResource_container(t *testing.T) {
	serverID, containerName := testAccTerminalContainerEnv(t)
	tftest.Test(t, tftest.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tftest.TestStep{
			{
				Config: testAccTerminalResourceContainerConfig(serverID, containerName, "tf-container-terminal"),
				Check: tftest.ComposeAggregateTestCheckFunc(
					tftest.TestCheckResourceAttr("komodo_terminal.test", "target_type", "Container"),
					tftest.TestCheckResourceAttr("komodo_terminal.test", "target_id", serverID),
					tftest.TestCheckResourceAttr("komodo_terminal.test", "container", containerName),
					tftest.TestCheckResourceAttr("komodo_terminal.test", "name", "tf-container-terminal"),
					tftest.TestCheckResourceAttrSet("komodo_terminal.test", "id"),
					tftest.TestCheckResourceAttrSet("komodo_terminal.test", "created_at"),
				),
			},
		},
	})
}

// TestAccTerminalResource_stack tests creating a terminal against a stack service target.
func TestAccTerminalResource_stack(t *testing.T) {
	stackID, service := testAccTerminalStackEnv(t)
	tftest.Test(t, tftest.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tftest.TestStep{
			{
				Config: testAccTerminalResourceStackConfig(stackID, service, "tf-stack-terminal"),
				Check: tftest.ComposeAggregateTestCheckFunc(
					tftest.TestCheckResourceAttr("komodo_terminal.test", "target_type", "Stack"),
					tftest.TestCheckResourceAttr("komodo_terminal.test", "target_id", stackID),
					tftest.TestCheckResourceAttr("komodo_terminal.test", "service", service),
					tftest.TestCheckResourceAttr("komodo_terminal.test", "name", "tf-stack-terminal"),
					tftest.TestCheckResourceAttrSet("komodo_terminal.test", "id"),
				),
			},
		},
	})
}

// TestAccTerminalResource_attachMode tests that mode="attach" is accepted for a container target.
func TestAccTerminalResource_attachMode(t *testing.T) {
	serverID, containerName := testAccTerminalContainerEnv(t)
	tftest.Test(t, tftest.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tftest.TestStep{
			{
				Config: testAccTerminalResourceContainerModeConfig(serverID, containerName, "tf-attach-terminal", "attach"),
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

func testAccTerminalResourceContainerConfig(serverID, containerName, name string) string {
	return fmt.Sprintf(`
resource "komodo_terminal" "test" {
  target_type = "Container"
  target_id   = %q
  container   = %q
  name        = %q
}
`, serverID, containerName, name)
}

func testAccTerminalResourceContainerModeConfig(serverID, containerName, name, mode string) string {
	return fmt.Sprintf(`
resource "komodo_terminal" "test" {
  target_type = "Container"
  target_id   = %q
  container   = %q
  name        = %q
  mode        = %q
}
`, serverID, containerName, name, mode)
}

func testAccTerminalResourceStackConfig(stackID, service, name string) string {
	return fmt.Sprintf(`
resource "komodo_terminal" "test" {
  target_type = "Stack"
  target_id   = %q
  service     = %q
  name        = %q
}
`, stackID, service, name)
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
