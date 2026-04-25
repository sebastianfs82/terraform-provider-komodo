// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"
)

// testAccNetworkServerID returns the ID of the first available server in the Komodo instance.
// Falls back to KOMODO_TEST_SERVER_ID if set. Skips the test if no servers are found.
func testAccNetworkServerID(t *testing.T) string {
	t.Helper()
	return testAccLookupServerID(t, "network acceptance tests")
}

// --- Resource tests ---

func TestAccNetworkResource_basic(t *testing.T) {
	serverID := testAccNetworkServerID(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkResourceConfig(serverID, "tf-test-network"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_network.test", "server_id", serverID),
					resource.TestCheckResourceAttr("komodo_network.test", "name", "tf-test-network"),
					resource.TestCheckResourceAttrSet("komodo_network.test", "id"),
				),
			},
			// Verify state is stable on a second plan/apply (no-diff).
			{
				Config:   testAccNetworkResourceConfig(serverID, "tf-test-network"),
				PlanOnly: true,
			},
		},
	})
}

func TestAccNetworkResource_importState(t *testing.T) {
	serverID := testAccNetworkServerID(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkResourceConfig(serverID, "tf-import-network"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_network.test", "id"),
				),
			},
			{
				ResourceName:      "komodo_network.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNetworkResourceConfig(serverID, name string) string {
	return fmt.Sprintf(`
resource "komodo_network" "test" {
  server_id = %q
  name      = %q
}
`, serverID, name)
}

// ─── Unit tests ──────────────────────────────────────────────────────────────

func wrongRawNetworkPlan(t *testing.T, r *NetworkResource) tfsdk.Plan {
	t.Helper()
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schemaResp.Schema,
	}
}

func wrongRawNetworkState(t *testing.T, r *NetworkResource) tfsdk.State {
	t.Helper()
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	return tfsdk.State{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schemaResp.Schema,
	}
}

func TestUnitNetworkResource_configure(t *testing.T) {
	t.Run("wrong_type", func(t *testing.T) {
		r := &NetworkResource{}
		req := fwresource.ConfigureRequest{ProviderData: "not-a-client"}
		resp := &fwresource.ConfigureResponse{}
		r.Configure(context.Background(), req, resp)
		if !resp.Diagnostics.HasError() {
			t.Fatal("expected diagnostic error for wrong ProviderData type")
		}
	})
}

func TestUnitNetworkResource_createPlanGetError(t *testing.T) {
	r := &NetworkResource{client: &client.Client{}}
	req := fwresource.CreateRequest{Plan: wrongRawNetworkPlan(t, r)}
	resp := &fwresource.CreateResponse{}
	r.Create(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed plan")
	}
}

func TestUnitNetworkResource_readStateGetError(t *testing.T) {
	r := &NetworkResource{client: &client.Client{}}
	req := fwresource.ReadRequest{State: wrongRawNetworkState(t, r)}
	resp := &fwresource.ReadResponse{}
	r.Read(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed state")
	}
}

func TestUnitNetworkResource_deleteStateGetError(t *testing.T) {
	r := &NetworkResource{client: &client.Client{}}
	req := fwresource.DeleteRequest{State: wrongRawNetworkState(t, r)}
	resp := &fwresource.DeleteResponse{}
	r.Delete(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed state")
	}
}

func TestUnitNetworkResource_updateIsNoop(t *testing.T) {
	r := &NetworkResource{}
	req := fwresource.UpdateRequest{}
	resp := &fwresource.UpdateResponse{}
	r.Update(context.Background(), req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error in no-op Update: %v", resp.Diagnostics)
	}
}

func TestUnitNetworkResource_importState_invalidFormat(t *testing.T) {
	r := &NetworkResource{}
	req := fwresource.ImportStateRequest{ID: "no-colon-separator"}
	resp := &fwresource.ImportStateResponse{}
	r.ImportState(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for invalid import ID format")
	}
}

// ─── Mock-HTTP-server unit tests ──────────────────────────────────────────────

// networkTestClient returns a *client.Client backed by a one-shot mock server
// that responds with the given HTTP status and body for every request.
func networkTestClient(t *testing.T, status int, body string) (*client.Client, func()) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
		fmt.Fprint(w, body)
	}))
	return client.NewClientWithApiKey(srv.URL, "key", "secret"), srv.Close
}

// networkReadState builds a valid tfsdk.State for NetworkResource Read/Delete tests.
func networkReadState(t *testing.T, r *NetworkResource, serverID string) tfsdk.State {
	t.Helper()
	const name = "test-net"
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	stateVal := tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), map[string]tftypes.Value{
		"id":        tftypes.NewValue(tftypes.String, serverID+":"+name),
		"server_id": tftypes.NewValue(tftypes.String, serverID),
		"name":      tftypes.NewValue(tftypes.String, name),
	})
	return tfsdk.State{Schema: schemaResp.Schema, Raw: stateVal}
}

func TestUnitNetworkResource_createClientError(t *testing.T) {
	c, closeSrv := networkTestClient(t, http.StatusInternalServerError, `"create error"`)
	defer closeSrv()

	r := &NetworkResource{client: c}
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	planVal := tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), map[string]tftypes.Value{
		"id":        tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"server_id": tftypes.NewValue(tftypes.String, "server-1"),
		"name":      tftypes.NewValue(tftypes.String, "test-net"),
	})
	req := fwresource.CreateRequest{Plan: tfsdk.Plan{Schema: schemaResp.Schema, Raw: planVal}}
	resp := &fwresource.CreateResponse{}
	r.Create(ctx, req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error when CreateNetwork fails")
	}
}

func TestUnitNetworkResource_readServerIDEmpty(t *testing.T) {
	r := &NetworkResource{client: &client.Client{}}
	ctx := context.Background()
	state := networkReadState(t, r, "")
	req := fwresource.ReadRequest{State: state}
	resp := &fwresource.ReadResponse{State: tfsdk.State{Schema: state.Schema}}
	r.Read(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error on empty server_id path: %v", resp.Diagnostics)
	}
}

func TestUnitNetworkResource_readListNetworksNotFound(t *testing.T) {
	// body contains "did not find" → Read must call RemoveResource without an error diagnostic.
	c, closeSrv := networkTestClient(t, http.StatusInternalServerError, `"did not find server"`)
	defer closeSrv()

	r := &NetworkResource{client: c}
	ctx := context.Background()
	state := networkReadState(t, r, "server-1")
	req := fwresource.ReadRequest{State: state}
	resp := &fwresource.ReadResponse{State: tfsdk.State{Schema: state.Schema}}
	r.Read(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostic on not-found path: %v", resp.Diagnostics)
	}
}

func TestUnitNetworkResource_readListNetworksClientError(t *testing.T) {
	// body does NOT contain a not-found phrase → Read must call AddError.
	c, closeSrv := networkTestClient(t, http.StatusInternalServerError, `"internal error"`)
	defer closeSrv()

	r := &NetworkResource{client: c}
	ctx := context.Background()
	req := fwresource.ReadRequest{State: networkReadState(t, r, "server-1")}
	resp := &fwresource.ReadResponse{}
	r.Read(ctx, req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error when ListDockerNetworks fails with non-not-found error")
	}
}

func TestUnitNetworkResource_readNetworkNotInList(t *testing.T) {
	// Returns an empty network list → network not found → RemoveResource, no error.
	c, closeSrv := networkTestClient(t, http.StatusOK, `[]`)
	defer closeSrv()

	r := &NetworkResource{client: c}
	ctx := context.Background()
	state := networkReadState(t, r, "server-1")
	req := fwresource.ReadRequest{State: state}
	resp := &fwresource.ReadResponse{State: tfsdk.State{Schema: state.Schema}}
	r.Read(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error when network not in list: %v", resp.Diagnostics)
	}
}
