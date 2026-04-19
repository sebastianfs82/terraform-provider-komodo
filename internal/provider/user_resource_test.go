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
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccUserResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserResourceConfig_basic("tf-user-basic", "Password1!"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user.test", "username", "tf-user-basic"),
					resource.TestCheckResourceAttrSet("komodo_user.test", "id"),
					resource.TestCheckResourceAttr("komodo_user.test", "enabled", "true"),
					resource.TestCheckResourceAttr("komodo_user.test", "admin_enabled", "false"),
					resource.TestCheckResourceAttr("komodo_user.test", "create_server_enabled", "false"),
					resource.TestCheckResourceAttr("komodo_user.test", "create_build_enabled", "false"),
				),
			},
		},
	})
}

func TestAccUserResource_withPermissions(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserResourceConfig_withPermissions("tf-user-perms", "Password1!", true, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user.test", "username", "tf-user-perms"),
					resource.TestCheckResourceAttr("komodo_user.test", "create_server_enabled", "true"),
					resource.TestCheckResourceAttr("komodo_user.test", "create_build_enabled", "true"),
				),
			},
			// Update: revoke permissions
			{
				Config: testAccUserResourceConfig_withPermissions("tf-user-perms", "Password1!", false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user.test", "create_server_enabled", "false"),
					resource.TestCheckResourceAttr("komodo_user.test", "create_build_enabled", "false"),
				),
			},
		},
	})
}

func TestAccUserResource_disableEnable(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserResourceConfig_withEnabled("tf-user-toggle", "Password1!", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user.test", "enabled", "true"),
				),
			},
			{
				Config: testAccUserResourceConfig_withEnabled("tf-user-toggle", "Password1!", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user.test", "enabled", "false"),
				),
			},
		},
	})
}

func TestAccUserResource_import(t *testing.T) {
	var userID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserResourceConfig_basic("tf-user-import", "Password1!"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user.test", "username", "tf-user-import"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["komodo_user.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						userID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config:            testAccUserResourceConfig_basic("tf-user-import", "Password1!"),
				ResourceName:      "komodo_user.test",
				ImportState:       true,
				ImportStateVerify: true,
				// password is not returned by the API on Read
				ImportStateVerifyIgnore: []string{"password"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return userID, nil
				},
			},
		},
	})
}

// Config helpers

func testAccUserResourceConfig_basic(username, password string) string {
	return fmt.Sprintf(`
resource "komodo_user" "test" {
  username = %[1]q
  password = %[2]q
}
`, username, password)
}

func testAccUserResourceConfig_withPermissions(username, password string, createServers, createBuilds bool) string {
	return fmt.Sprintf(`
resource "komodo_user" "test" {
  username       = %[1]q
  password       = %[2]q
  create_server_enabled = %[3]t
  create_build_enabled  = %[4]t
}
`, username, password, createServers, createBuilds)
}

func testAccUserResourceConfig_withEnabled(username, password string, enabled bool) string {
	return fmt.Sprintf(`
resource "komodo_user" "test" {
  username = %[1]q
  password = %[2]q
  enabled  = %[3]t
}
`, username, password, enabled)
}

func TestAccUserResource_adminEnabled(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with admin_enabled = true
			{
				Config: testAccUserResourceConfig_withAdmin("tf-user-admin", "Password1!", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user.test", "admin_enabled", "true"),
					resource.TestCheckResourceAttr("komodo_user.test", "enabled", "true"),
				),
			},
			// Update: revoke admin
			{
				Config: testAccUserResourceConfig_withAdmin("tf-user-admin", "Password1!", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user.test", "admin_enabled", "false"),
				),
			},
		},
	})
}

func TestAccUserResource_enabledDefault(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// enabled not set explicitly — should default to true
			{
				Config: testAccUserResourceConfig_basic("tf-user-enabled-default", "Password1!"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user.test", "enabled", "true"),
				),
			},
			// Explicitly disable
			{
				Config: testAccUserResourceConfig_withEnabled("tf-user-enabled-default", "Password1!", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user.test", "enabled", "false"),
				),
			},
			// Remove enabled from config — should plan a change back to true and apply it
			{
				Config: testAccUserResourceConfig_basic("tf-user-enabled-default", "Password1!"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user.test", "enabled", "true"),
				),
			},
		},
	})
}

func TestAccUserResource_adminConflictWithCreateServer(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccUserResourceConfig_adminWithCreateServer("tf-user-conflict-srv", "Password1!"),
				ExpectError: regexp.MustCompile(`create_server_enabled cannot be set to true alongside admin_enabled = true`),
			},
		},
	})
}

func TestAccUserResource_adminConflictWithCreateBuild(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccUserResourceConfig_adminWithCreateBuild("tf-user-conflict-bld", "Password1!"),
				ExpectError: regexp.MustCompile(`create_build_enabled cannot be set to true alongside admin_enabled = true`),
			},
		},
	})
}

func testAccUserResourceConfig_withAdmin(username, password string, admin bool) string {
	return fmt.Sprintf(`
resource "komodo_user" "test" {
  username      = %[1]q
  password      = %[2]q
  admin_enabled = %[3]t
}
`, username, password, admin)
}

func testAccUserResourceConfig_adminWithCreateServer(username, password string) string {
	return fmt.Sprintf(`
resource "komodo_user" "test" {
  username              = %[1]q
  password              = %[2]q
  admin_enabled         = true
  create_server_enabled = true
}
`, username, password)
}

func testAccUserResourceConfig_adminWithCreateBuild(username, password string) string {
	return fmt.Sprintf(`
resource "komodo_user" "test" {
  username             = %[1]q
  password             = %[2]q
  admin_enabled        = true
  create_build_enabled = true
}
`, username, password)
}

func TestAccUserResource_disappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserResourceConfig_basic("disappear-user", "Password123!"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_user.test", "id"),
					testAccUserDisappears("komodo_user.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccUserDisappears(resourceName string) resource.TestCheckFunc {
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
		return c.DeleteUser(context.Background(), rs.Primary.ID)
	}
}

func TestUnitUserResource_readNon404ErrorDirect(t *testing.T) {
	// Direct call to r.Read() with a mock returning non-404 error.
	srv := newUserMockServer(t, map[string]mockUserRoute{
		"FindUser": {http.StatusInternalServerError, `"server error"`},
	})
	c := client.NewClient(srv.URL, "mock-user", "mock-pass")
	r := &UserResource{client: c}
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)

	stateVal := tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), map[string]tftypes.Value{
		"id":                    tftypes.NewValue(tftypes.String, "aabbccddeeff001122334455"),
		"username":              tftypes.NewValue(tftypes.String, "tf-mock-user"),
		"password":              tftypes.NewValue(tftypes.String, "Password1!"),
		"enabled":               tftypes.NewValue(tftypes.Bool, true),
		"admin_enabled":         tftypes.NewValue(tftypes.Bool, false),
		"create_server_enabled": tftypes.NewValue(tftypes.Bool, false),
		"create_build_enabled":  tftypes.NewValue(tftypes.Bool, false),
	})
	state := tfsdk.State{Schema: schemaResp.Schema, Raw: stateVal}
	req := fwresource.ReadRequest{State: state}
	resp := &fwresource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	r.Read(ctx, req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for non-404 FindUser error")
	}
}

func TestUnitUserResource_readNilUserDirect(t *testing.T) {
	// Direct call to r.Read() with FindUser returning nil (404).
	srv := newUserMockServer(t, map[string]mockUserRoute{
		"FindUser": {http.StatusNotFound, `"not found"`},
	})
	c := client.NewClient(srv.URL, "mock-user", "mock-pass")
	r := &UserResource{client: c}
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)

	stateVal := tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), map[string]tftypes.Value{
		"id":                    tftypes.NewValue(tftypes.String, "aabbccddeeff001122334455"),
		"username":              tftypes.NewValue(tftypes.String, "tf-mock-user"),
		"password":              tftypes.NewValue(tftypes.String, "Password1!"),
		"enabled":               tftypes.NewValue(tftypes.Bool, true),
		"admin_enabled":         tftypes.NewValue(tftypes.Bool, false),
		"create_server_enabled": tftypes.NewValue(tftypes.Bool, false),
		"create_build_enabled":  tftypes.NewValue(tftypes.Bool, false),
	})
	state := tfsdk.State{Schema: schemaResp.Schema, Raw: stateVal}
	req := fwresource.ReadRequest{State: state}
	resp := &fwresource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	r.Read(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("expected no error for nil user (resource removed from state): %v", resp.Diagnostics)
	}
}

func TestUnitUserResource_readErrorContains404Direct(t *testing.T) {
	// Direct call to r.Read() with FindUser returning error that contains "404" → RemoveResource.
	// Body "error 404" causes err.Error() to contain "404", triggering inner if-branch.
	srv := newUserMockServer(t, map[string]mockUserRoute{
		"FindUser": {http.StatusInternalServerError, `"error 404"`},
	})
	c := client.NewClient(srv.URL, "mock-user", "mock-pass")
	r := &UserResource{client: c}
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)

	stateVal := tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), map[string]tftypes.Value{
		"id":                    tftypes.NewValue(tftypes.String, "aabbccddeeff001122334455"),
		"username":              tftypes.NewValue(tftypes.String, "tf-mock-user"),
		"password":              tftypes.NewValue(tftypes.String, "Password1!"),
		"enabled":               tftypes.NewValue(tftypes.Bool, true),
		"admin_enabled":         tftypes.NewValue(tftypes.Bool, false),
		"create_server_enabled": tftypes.NewValue(tftypes.Bool, false),
		"create_build_enabled":  tftypes.NewValue(tftypes.Bool, false),
	})
	state := tfsdk.State{Schema: schemaResp.Schema, Raw: stateVal}
	req := fwresource.ReadRequest{State: state}
	resp := &fwresource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	r.Read(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("expected no error (resource removed from state for 404-like error): %v", resp.Diagnostics)
	}
}

// ─── Unit tests ──────────────────────────────────────────────────────────────

func TestUnitUserResource_importStateDirect(t *testing.T) {
	r := &UserResource{}
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	emptyState := tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), map[string]tftypes.Value{
		"id":                    tftypes.NewValue(tftypes.String, nil),
		"username":              tftypes.NewValue(tftypes.String, nil),
		"password":              tftypes.NewValue(tftypes.String, nil),
		"enabled":               tftypes.NewValue(tftypes.Bool, nil),
		"admin_enabled":         tftypes.NewValue(tftypes.Bool, nil),
		"create_server_enabled": tftypes.NewValue(tftypes.Bool, nil),
		"create_build_enabled":  tftypes.NewValue(tftypes.Bool, nil),
	})
	req := fwresource.ImportStateRequest{ID: "aabbccddeeff001122334455"}
	resp := &fwresource.ImportStateResponse{State: tfsdk.State{Schema: schemaResp.Schema, Raw: emptyState}}
	r.ImportState(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error in ImportState: %v", resp.Diagnostics)
	}
}

func TestUnitUserResource_updateFindUserErrorDirect(t *testing.T) {
	// Direct call to r.Update() with FindUser failing → error path at Update's end.
	srv := newUserMockServer(t, map[string]mockUserRoute{
		"FindUser": {http.StatusInternalServerError, `"server error"`},
	})
	c := client.NewClient(srv.URL, "mock-user", "mock-pass")
	r := &UserResource{client: c}
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)

	modelVal := tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), map[string]tftypes.Value{
		"id":                    tftypes.NewValue(tftypes.String, "aabbccddeeff001122334455"),
		"username":              tftypes.NewValue(tftypes.String, "tf-mock-user"),
		"password":              tftypes.NewValue(tftypes.String, "Password1!"),
		"enabled":               tftypes.NewValue(tftypes.Bool, true),
		"admin_enabled":         tftypes.NewValue(tftypes.Bool, false),
		"create_server_enabled": tftypes.NewValue(tftypes.Bool, false),
		"create_build_enabled":  tftypes.NewValue(tftypes.Bool, false),
	})
	state := tfsdk.State{Schema: schemaResp.Schema, Raw: modelVal}
	plan := tfsdk.Plan{Schema: schemaResp.Schema, Raw: modelVal}
	req := fwresource.UpdateRequest{Plan: plan, State: state}
	resp := &fwresource.UpdateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	r.Update(ctx, req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for FindUser error in Update")
	}
}

func TestUnitUserResource_updateFindUserNilDirect(t *testing.T) {
	// Direct call to r.Update() with FindUser returning nil → nil user error path.
	srv := newUserMockServer(t, map[string]mockUserRoute{
		"FindUser": {http.StatusNotFound, `"not found"`},
	})
	c := client.NewClient(srv.URL, "mock-user", "mock-pass")
	r := &UserResource{client: c}
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)

	modelVal := tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), map[string]tftypes.Value{
		"id":                    tftypes.NewValue(tftypes.String, "aabbccddeeff001122334455"),
		"username":              tftypes.NewValue(tftypes.String, "tf-mock-user"),
		"password":              tftypes.NewValue(tftypes.String, "Password1!"),
		"enabled":               tftypes.NewValue(tftypes.Bool, true),
		"admin_enabled":         tftypes.NewValue(tftypes.Bool, false),
		"create_server_enabled": tftypes.NewValue(tftypes.Bool, false),
		"create_build_enabled":  tftypes.NewValue(tftypes.Bool, false),
	})
	state := tfsdk.State{Schema: schemaResp.Schema, Raw: modelVal}
	plan := tfsdk.Plan{Schema: schemaResp.Schema, Raw: modelVal}
	req := fwresource.UpdateRequest{Plan: plan, State: state}
	resp := &fwresource.UpdateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	r.Update(ctx, req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for nil user in Update")
	}
}

func wrongRawUserPlan(t *testing.T, r *UserResource) tfsdk.Plan {
	t.Helper()
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schemaResp.Schema,
	}
}

func wrongRawUserState(t *testing.T, r *UserResource) tfsdk.State {
	t.Helper()
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	return tfsdk.State{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schemaResp.Schema,
	}
}

func TestUnitUserResource_configure(t *testing.T) {
	t.Run("wrong_type", func(t *testing.T) {
		r := &UserResource{}
		req := fwresource.ConfigureRequest{ProviderData: "not-a-client"}
		resp := &fwresource.ConfigureResponse{}
		r.Configure(context.Background(), req, resp)
		if !resp.Diagnostics.HasError() {
			t.Fatal("expected diagnostic error for wrong ProviderData type")
		}
	})
}

func TestUnitUserResource_createPlanGetError(t *testing.T) {
	r := &UserResource{client: &client.Client{}}
	req := fwresource.CreateRequest{Plan: wrongRawUserPlan(t, r)}
	resp := &fwresource.CreateResponse{}
	r.Create(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed plan")
	}
}

func TestUnitUserResource_readStateGetError(t *testing.T) {
	r := &UserResource{client: &client.Client{}}
	req := fwresource.ReadRequest{State: wrongRawUserState(t, r)}
	resp := &fwresource.ReadResponse{}
	r.Read(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed state")
	}
}

func TestUnitUserResource_updatePlanGetError(t *testing.T) {
	r := &UserResource{client: &client.Client{}}
	req := fwresource.UpdateRequest{
		Plan:  wrongRawUserPlan(t, r),
		State: wrongRawUserState(t, r),
	}
	resp := &fwresource.UpdateResponse{}
	r.Update(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed plan")
	}
}

func TestUnitUserResource_deleteStateGetError(t *testing.T) {
	r := &UserResource{client: &client.Client{}}
	req := fwresource.DeleteRequest{State: wrongRawUserState(t, r)}
	resp := &fwresource.DeleteResponse{}
	r.Delete(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed state")
	}
}

// ─── Mock-server helpers (user) ───────────────────────────────────────────────

const mockValidUserJSON = `{"_id":{"$oid":"aabbccddeeff001122334455"},"username":"tf-mock-user","enabled":true,"admin":false,"create_server_permissions":false,"create_build_permissions":false,"config":{"type":"Local"}}`

type mockUserRoute struct {
	statusCode int
	body       string
}

// newUserMockServerWithDefault builds an httptest.Server; unmatched types return defaultJSON/200.
func newUserMockServerWithDefault(t *testing.T, routes map[string]mockUserRoute, defaultJSON string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/auth/login/LoginLocalUser" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"type":"Jwt","data":{"jwt":"mock-token"}}`)
			return
		}
		body, _ := io.ReadAll(r.Body)
		var req map[string]interface{}
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		typ, _ := req["type"].(string)
		if typ == "GetVersion" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"version":"2.0.0"}`)
			return
		}
		if route, ok := routes[typ]; ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(route.statusCode)
			fmt.Fprint(w, route.body)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, defaultJSON)
	}))
	t.Cleanup(srv.Close)
	return srv
}

// newUserMockServer builds an httptest.Server routing /write and /read
// by the "type" field in the request body. Routes not in the map return
// mockValidUserJSON with status 200.
func newUserMockServer(t *testing.T, routes map[string]mockUserRoute) *httptest.Server {
	return newUserMockServerWithDefault(t, routes, mockValidUserJSON)
}

// newStatefulUserMockServer builds a mock server with a custom handler per
// request type. Each handler is called with the raw request body and returns
// (statusCode, body). A nil handler entry means "return mockValidUserJSON 200".
func newStatefulUserMockServer(t *testing.T, handler func(typ string, callCount int) (int, string)) *httptest.Server {
	t.Helper()
	var mu sync.Mutex
	counts := map[string]int{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/auth/login/LoginLocalUser" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"type":"Jwt","data":{"jwt":"mock-token"}}`)
			return
		}
		body, _ := io.ReadAll(r.Body)
		var req map[string]interface{}
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		typ, _ := req["type"].(string)
		if typ == "GetVersion" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"version":"2.0.0"}`)
			return
		}
		mu.Lock()
		counts[typ]++
		n := counts[typ]
		mu.Unlock()
		code, respBody := handler(typ, n)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		fmt.Fprint(w, respBody)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func mockUserProviderConfig(srvURL string) string {
	return fmt.Sprintf(`
provider "komodo" {
  endpoint = %[1]q
  username = "mock-user"
  password = "mock-pass"
}
`, srvURL)
}

const mockUserResourceConfig = `
resource "komodo_user" "test" {
  username = "tf-mock-user"
  password = "Password1!"
}
`

const mockUserResourceConfigAdmin = `
resource "komodo_user" "test" {
  username      = "tf-mock-user"
  password      = "Password1!"
  admin_enabled = true
}
`

// ─── Create error tests ───────────────────────────────────────────────────────

func TestAccUserResource_createClientError(t *testing.T) {
	srv := newUserMockServer(t, map[string]mockUserRoute{
		"CreateLocalUser": {http.StatusInternalServerError, `"internal error"`},
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockUserProviderConfig(srv.URL) + mockUserResourceConfig,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

func TestAccUserResource_createFindUserError(t *testing.T) {
	srv := newUserMockServer(t, map[string]mockUserRoute{
		"FindUser": {http.StatusInternalServerError, `"find error"`},
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockUserProviderConfig(srv.URL) + mockUserResourceConfig,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

func TestAccUserResource_createFindUserNil(t *testing.T) {
	srv := newUserMockServer(t, map[string]mockUserRoute{
		"FindUser": {http.StatusNotFound, `"not found"`},
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockUserProviderConfig(srv.URL) + mockUserResourceConfig,
				ExpectError: regexp.MustCompile(`(?i)not found`),
			},
		},
	})
}

func TestAccUserResource_createUpdateAdminError(t *testing.T) {
	srv := newUserMockServer(t, map[string]mockUserRoute{
		"UpdateUserAdmin": {http.StatusInternalServerError, `"admin error"`},
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockUserProviderConfig(srv.URL) + mockUserResourceConfigAdmin,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

func TestAccUserResource_createUpdatePermsError(t *testing.T) {
	srv := newUserMockServer(t, map[string]mockUserRoute{
		"UpdateUserBasePermissions": {http.StatusInternalServerError, `"perms error"`},
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockUserProviderConfig(srv.URL) + mockUserResourceConfig,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

func TestAccUserResource_createSecondFindUserError(t *testing.T) {
	// First FindUser (by username) succeeds; second (by ID, after perms) fails.
	srv := newStatefulUserMockServer(t, func(typ string, n int) (int, string) {
		if typ == "FindUser" && n >= 2 {
			return http.StatusInternalServerError, `"second find error"`
		}
		return http.StatusOK, mockValidUserJSON
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockUserProviderConfig(srv.URL) + mockUserResourceConfig,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

func TestAccUserResource_createSecondFindUserNil(t *testing.T) {
	// First FindUser succeeds; second returns 404 (nil).
	srv := newStatefulUserMockServer(t, func(typ string, n int) (int, string) {
		if typ == "FindUser" && n >= 2 {
			return http.StatusNotFound, `"not found"`
		}
		return http.StatusOK, mockValidUserJSON
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockUserProviderConfig(srv.URL) + mockUserResourceConfig,
				ExpectError: regexp.MustCompile(`(?i)not found`),
			},
		},
	})
}

// ─── Update error tests ───────────────────────────────────────────────────────

func TestAccUserResource_updateAdminError(t *testing.T) {
	// Step 1: create (admin_enabled=false → no UpdateUserAdmin call).
	// Step 2: change to admin_enabled=true → UpdateUserAdmin → 500.
	srv := newUserMockServer(t, map[string]mockUserRoute{
		"UpdateUserAdmin": {http.StatusInternalServerError, `"admin update error"`},
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockUserProviderConfig(srv.URL) + mockUserResourceConfig,
			},
			{
				Config:      mockUserProviderConfig(srv.URL) + mockUserResourceConfigAdmin,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

func TestAccUserResource_updatePermsError(t *testing.T) {
	// Call 1 of UpdateUserBasePermissions (in Create) succeeds.
	// Call 2+ (in Update when enabled changes) fails.
	srv := newStatefulUserMockServer(t, func(typ string, n int) (int, string) {
		if typ == "UpdateUserBasePermissions" && n >= 2 {
			return http.StatusInternalServerError, `"perms update error"`
		}
		return http.StatusOK, mockValidUserJSON
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockUserProviderConfig(srv.URL) + mockUserResourceConfig,
			},
			{
				Config: mockUserProviderConfig(srv.URL) + `
resource "komodo_user" "test" {
  username = "tf-mock-user"
  password = "Password1!"
  enabled  = false
}
`,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

// ─── Read disappears test ─────────────────────────────────────────────────────

func TestAccUserResource_readDisappears_unit(t *testing.T) {
	// Create succeeds; post-apply refresh (FindUser call 3) returns 404 → nil → RemoveResource.
	srv := newStatefulUserMockServer(t, func(typ string, n int) (int, string) {
		if typ == "FindUser" && n >= 3 {
			return http.StatusNotFound, `"not found"`
		}
		return http.StatusOK, mockValidUserJSON
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             mockUserProviderConfig(srv.URL) + mockUserResourceConfig,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// ─── Successful update test ───────────────────────────────────────────────────

func TestAccUserResource_updateSuccess(t *testing.T) {
	// All calls succeed; Update path + final FindUser + set-state.
	// FindUser calls 1-4 return enabled=true; calls 5+ return enabled=false to match updated plan.
	const mockUserUpdatedJSON = `{"_id":{"$oid":"aabbccddeeff001122334455"},"username":"tf-mock-user","enabled":false,"admin":false,"create_server_permissions":false,"create_build_permissions":false,"config":{"type":"Local"}}`
	srv := newStatefulUserMockServer(t, func(typ string, n int) (int, string) {
		if typ == "FindUser" && n >= 5 {
			return http.StatusOK, mockUserUpdatedJSON
		}
		return http.StatusOK, mockValidUserJSON
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockUserProviderConfig(srv.URL) + mockUserResourceConfig,
			},
			{
				Config: mockUserProviderConfig(srv.URL) + `
resource "komodo_user" "test" {
  username = "tf-mock-user"
  password = "Password1!"
  enabled  = false
}
`,
				Check: resource.TestCheckResourceAttr("komodo_user.test", "username", "tf-mock-user"),
			},
		},
	})
}

func TestUnitUserResource_validateConfigGetError(t *testing.T) {
	r := &UserResource{}
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	req := fwresource.ValidateConfigRequest{
		Config: tfsdk.Config{
			Raw:    tftypes.NewValue(tftypes.String, "invalid"),
			Schema: schemaResp.Schema,
		},
	}
	resp := &fwresource.ValidateConfigResponse{}
	r.ValidateConfig(ctx, req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error from malformed config")
	}
}

func TestAccUserResource_adminConflictServer_unit(t *testing.T) {
	srv := newUserMockServer(t, nil)
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockUserProviderConfig(srv.URL) + testAccUserResourceConfig_adminWithCreateServer("tf-mock-user", "Password1!"),
				ExpectError: regexp.MustCompile(`create_server_enabled cannot be set to true alongside admin_enabled = true`),
			},
		},
	})
}

func TestAccUserResource_adminConflictBuild_unit(t *testing.T) {
	srv := newUserMockServer(t, nil)
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockUserProviderConfig(srv.URL) + testAccUserResourceConfig_adminWithCreateBuild("tf-mock-user", "Password1!"),
				ExpectError: regexp.MustCompile(`create_build_enabled cannot be set to true alongside admin_enabled = true`),
			},
		},
	})
}

// ─── Unit test for Read non-404 error path ────────────────────────────────────

func TestAccUserResource_readNon404Error(t *testing.T) {
	// Step 1 Create: calls 1 & 2 succeed; post-apply refresh (call 3) → 500 (non-404).
	srv := newStatefulUserMockServer(t, func(typ string, n int) (int, string) {
		if typ == "FindUser" && n >= 3 {
			return http.StatusInternalServerError, `"read error"`
		}
		return http.StatusOK, mockValidUserJSON
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockUserProviderConfig(srv.URL) + mockUserResourceConfig,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

func TestAccUserResource_updateFindUserError(t *testing.T) {
	// Step 1 Create: FindUser calls 1 (by username) & 2 (by ID).
	// Post-apply refresh (step 1): FindUser call 3.
	// Step 2 pre-plan refresh: FindUser call 4.
	// Step 2 Update apply end: FindUser call 5 → fails.
	srv := newStatefulUserMockServer(t, func(typ string, n int) (int, string) {
		if typ == "FindUser" && n >= 5 {
			return http.StatusInternalServerError, `"find after update error"`
		}
		return http.StatusOK, mockValidUserJSON
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockUserProviderConfig(srv.URL) + mockUserResourceConfig,
			},
			{
				Config: mockUserProviderConfig(srv.URL) + `
resource "komodo_user" "test" {
  username = "tf-mock-user"
  password = "Password1!"
  enabled  = false
}
`,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

func TestAccUserResource_updateFindUserNil(t *testing.T) {
	// Same as above but FindUser call 5 returns 404 (nil).
	srv := newStatefulUserMockServer(t, func(typ string, n int) (int, string) {
		if typ == "FindUser" && n >= 5 {
			return http.StatusNotFound, `"not found after update"`
		}
		return http.StatusOK, mockValidUserJSON
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockUserProviderConfig(srv.URL) + mockUserResourceConfig,
			},
			{
				Config: mockUserProviderConfig(srv.URL) + `
resource "komodo_user" "test" {
  username = "tf-mock-user"
  password = "Password1!"
  enabled  = false
}
`,
				ExpectError: regexp.MustCompile(`(?i)not found`),
			},
		},
	})
}

// ─── Delete error test ────────────────────────────────────────────────────────

func TestAccUserResource_deleteClientError(t *testing.T) {
	// First DeleteUser call (the Destroy step) → 500; subsequent calls succeed for cleanup.
	srv := newStatefulUserMockServer(t, func(typ string, n int) (int, string) {
		if typ == "DeleteUser" && n == 1 {
			return http.StatusInternalServerError, `"delete error"`
		}
		return http.StatusOK, mockValidUserJSON
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockUserProviderConfig(srv.URL) + mockUserResourceConfig,
			},
			{
				Config:      mockUserProviderConfig(srv.URL),
				Destroy:     true,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}
