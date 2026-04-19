// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccServiceUserResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceUserResourceConfig_basic("tf-svc-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_service_user.test", "username", "tf-svc-basic"),
					resource.TestCheckResourceAttrSet("komodo_service_user.test", "id"),
					resource.TestCheckResourceAttr("komodo_service_user.test", "enabled", "true"),
					resource.TestCheckResourceAttr("komodo_service_user.test", "admin_enabled", "false"),
					resource.TestCheckResourceAttr("komodo_service_user.test", "create_server_enabled", "false"),
					resource.TestCheckResourceAttr("komodo_service_user.test", "create_build_enabled", "false"),
				),
			},
		},
	})
}

func TestAccServiceUserResource_withDescription(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceUserResourceConfig_withDescription("tf-svc-desc", "initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_service_user.test", "username", "tf-svc-desc"),
					resource.TestCheckResourceAttr("komodo_service_user.test", "description", "initial description"),
				),
			},
			// Update description
			{
				Config: testAccServiceUserResourceConfig_withDescription("tf-svc-desc", "updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_service_user.test", "description", "updated description"),
				),
			},
		},
	})
}

func TestAccServiceUserResource_withPermissions(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceUserResourceConfig_withPermissions("tf-svc-perms", true, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_service_user.test", "username", "tf-svc-perms"),
					resource.TestCheckResourceAttr("komodo_service_user.test", "create_server_enabled", "true"),
					resource.TestCheckResourceAttr("komodo_service_user.test", "create_build_enabled", "true"),
				),
			},
			// Update permissions
			{
				Config: testAccServiceUserResourceConfig_withPermissions("tf-svc-perms", false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_service_user.test", "create_server_enabled", "false"),
					resource.TestCheckResourceAttr("komodo_service_user.test", "create_build_enabled", "false"),
				),
			},
		},
	})
}

func TestAccServiceUserResource_import(t *testing.T) {
	var userID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceUserResourceConfig_basic("tf-svc-import"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_service_user.test", "username", "tf-svc-import"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["komodo_service_user.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						userID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config:            testAccServiceUserResourceConfig_basic("tf-svc-import"),
				ResourceName:      "komodo_service_user.test",
				ImportState:       true,
				ImportStateVerify: true,
				// description is not returned by the API on Read
				ImportStateVerifyIgnore: []string{"description"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return userID, nil
				},
			},
		},
	})
}

func TestAccServiceUserResource_withApiKey(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceUserResourceConfig_withApiKey("tf-svc-apikey", "svc-user-key"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_service_user.test", "username", "tf-svc-apikey"),
					resource.TestCheckResourceAttr("komodo_api_key.svc_key", "name", "svc-user-key"),
					resource.TestCheckResourceAttrSet("komodo_api_key.svc_key", "key"),
					resource.TestCheckResourceAttrSet("komodo_api_key.svc_key", "secret"),
					resource.TestCheckResourceAttr("komodo_api_key.svc_key", "expires_at", ""),
					resource.TestCheckResourceAttrPair(
						"komodo_api_key.svc_key", "service_user_id",
						"komodo_service_user.test", "id",
					),
				),
			},
		},
	})
}

// Config helpers

func testAccServiceUserResourceConfig_basic(username string) string {
	return fmt.Sprintf(`
resource "komodo_service_user" "test" {
  username = %[1]q
}
`, username)
}

func testAccServiceUserResourceConfig_withDescription(username, description string) string {
	return fmt.Sprintf(`
resource "komodo_service_user" "test" {
  username    = %[1]q
  description = %[2]q
}
`, username, description)
}

func testAccServiceUserResourceConfig_withPermissions(username string, createServers, createBuilds bool) string {
	return fmt.Sprintf(`
resource "komodo_service_user" "test" {
  username               = %[1]q
  create_server_enabled  = %[2]t
  create_build_enabled   = %[3]t
}
`, username, createServers, createBuilds)
}

func testAccServiceUserResourceConfig_withApiKey(username, keyName string) string {
	return fmt.Sprintf(`
resource "komodo_service_user" "test" {
  username = %[1]q
}

resource "komodo_api_key" "svc_key" {
  name            = %[2]q
  service_user_id = komodo_service_user.test.id
}
`, username, keyName)
}

func TestAccServiceUserResource_disappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceUserResourceConfig_basic("disappear-svc-user"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_service_user.test", "id"),
					testAccServiceUserDisappears("komodo_service_user.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccServiceUserDisappears(resourceName string) resource.TestCheckFunc {
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

func TestAccServiceUserResource_adminEnabled(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceUserResourceConfig_withAdmin("tf-svc-admin", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_service_user.test", "admin_enabled", "true"),
					resource.TestCheckResourceAttr("komodo_service_user.test", "enabled", "true"),
				),
			},
			// Revoke admin
			{
				Config: testAccServiceUserResourceConfig_withAdmin("tf-svc-admin", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_service_user.test", "admin_enabled", "false"),
				),
			},
		},
	})
}

func TestAccServiceUserResource_enabledDefault(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// enabled not set explicitly — should default to true
			{
				Config: testAccServiceUserResourceConfig_basic("tf-svc-enabled-default"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_service_user.test", "enabled", "true"),
				),
			},
			// Explicitly disable
			{
				Config: testAccServiceUserResourceConfig_withEnabled("tf-svc-enabled-default", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_service_user.test", "enabled", "false"),
				),
			},
			// Remove enabled from config — default re-enables
			{
				Config: testAccServiceUserResourceConfig_basic("tf-svc-enabled-default"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_service_user.test", "enabled", "true"),
				),
			},
		},
	})
}

func TestAccServiceUserResource_adminConflictWithCreateServer(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccServiceUserResourceConfig_adminWithCreateServer("tf-svc-conflict-srv"),
				ExpectError: regexp.MustCompile(`create_server_enabled cannot be set to true alongside admin_enabled = true`),
			},
		},
	})
}

func TestAccServiceUserResource_adminConflictWithCreateBuild(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccServiceUserResourceConfig_adminWithCreateBuild("tf-svc-conflict-bld"),
				ExpectError: regexp.MustCompile(`create_build_enabled cannot be set to true alongside admin_enabled = true`),
			},
		},
	})
}

func testAccServiceUserResourceConfig_withAdmin(username string, admin bool) string {
	return fmt.Sprintf(`
resource "komodo_service_user" "test" {
  username      = %[1]q
  admin_enabled = %[2]t
}
`, username, admin)
}

func testAccServiceUserResourceConfig_withEnabled(username string, enabled bool) string {
	return fmt.Sprintf(`
resource "komodo_service_user" "test" {
  username = %[1]q
  enabled  = %[2]t
}
`, username, enabled)
}

func testAccServiceUserResourceConfig_adminWithCreateServer(username string) string {
	return fmt.Sprintf(`
resource "komodo_service_user" "test" {
  username               = %[1]q
  admin_enabled          = true
  create_server_enabled  = true
}
`, username)
}

func testAccServiceUserResourceConfig_adminWithCreateBuild(username string) string {
	return fmt.Sprintf(`
resource "komodo_service_user" "test" {
  username              = %[1]q
  admin_enabled         = true
  create_build_enabled  = true
}
`, username)
}

// ─── Unit tests ──────────────────────────────────────────────────────────────

func TestUnitServiceUserResource_readNon404ErrorDirect(t *testing.T) {
	// Direct call to r.Read() with a mock returning non-404 error.
	srv := newSvcUserMockServer(t, map[string]mockUserRoute{
		"FindUser": {http.StatusInternalServerError, `"server error"`},
	})
	c := client.NewClient(srv.URL, "mock-user", "mock-pass")
	r := &ServiceUserResource{client: c}
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)

	stateVal := tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), map[string]tftypes.Value{
		"id":                    tftypes.NewValue(tftypes.String, "aabbccddeeff001122334455"),
		"username":              tftypes.NewValue(tftypes.String, "tf-mock-svc-user"),
		"description":           tftypes.NewValue(tftypes.String, ""),
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

func TestUnitServiceUserResource_readNilUserDirect(t *testing.T) {
	// Direct call to r.Read() with FindUser returning nil (404).
	srv := newSvcUserMockServer(t, map[string]mockUserRoute{
		"FindUser": {http.StatusNotFound, `"not found"`},
	})
	c := client.NewClient(srv.URL, "mock-user", "mock-pass")
	r := &ServiceUserResource{client: c}
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)

	stateVal := tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), map[string]tftypes.Value{
		"id":                    tftypes.NewValue(tftypes.String, "aabbccddeeff001122334455"),
		"username":              tftypes.NewValue(tftypes.String, "tf-mock-svc-user"),
		"description":           tftypes.NewValue(tftypes.String, ""),
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

func TestUnitServiceUserResource_readErrorContains404Direct(t *testing.T) {
	// Direct call to r.Read() with FindUser returning error that contains "404" → RemoveResource.
	srv := newSvcUserMockServer(t, map[string]mockUserRoute{
		"FindUser": {http.StatusInternalServerError, `"error 404"`},
	})
	c := client.NewClient(srv.URL, "mock-user", "mock-pass")
	r := &ServiceUserResource{client: c}
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)

	stateVal := tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), map[string]tftypes.Value{
		"id":                    tftypes.NewValue(tftypes.String, "aabbccddeeff001122334455"),
		"username":              tftypes.NewValue(tftypes.String, "tf-mock-svc-user"),
		"description":           tftypes.NewValue(tftypes.String, ""),
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

func TestUnitServiceUserResource_updateFindUserErrorDirect(t *testing.T) {
	// Direct call to r.Update() with FindUser failing → error path at Update's end.
	srv := newSvcUserMockServer(t, map[string]mockUserRoute{
		"FindUser": {http.StatusInternalServerError, `"server error"`},
	})
	c := client.NewClient(srv.URL, "mock-user", "mock-pass")
	r := &ServiceUserResource{client: c}
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)

	modelVal := tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), map[string]tftypes.Value{
		"id":                    tftypes.NewValue(tftypes.String, "aabbccddeeff001122334455"),
		"username":              tftypes.NewValue(tftypes.String, "tf-mock-svc-user"),
		"description":           tftypes.NewValue(tftypes.String, ""),
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

func TestUnitServiceUserResource_updateFindUserNilDirect(t *testing.T) {
	// Direct call to r.Update() with FindUser returning nil → nil user error path.
	srv := newSvcUserMockServer(t, map[string]mockUserRoute{
		"FindUser": {http.StatusNotFound, `"not found"`},
	})
	c := client.NewClient(srv.URL, "mock-user", "mock-pass")
	r := &ServiceUserResource{client: c}
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)

	modelVal := tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), map[string]tftypes.Value{
		"id":                    tftypes.NewValue(tftypes.String, "aabbccddeeff001122334455"),
		"username":              tftypes.NewValue(tftypes.String, "tf-mock-svc-user"),
		"description":           tftypes.NewValue(tftypes.String, ""),
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

func TestUnitServiceUserResource_importStateDirect(t *testing.T) {
	r := &ServiceUserResource{}
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	emptyState := tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), map[string]tftypes.Value{
		"id":                    tftypes.NewValue(tftypes.String, nil),
		"username":              tftypes.NewValue(tftypes.String, nil),
		"description":           tftypes.NewValue(tftypes.String, nil),
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

func wrongRawServiceUserPlan(t *testing.T, r *ServiceUserResource) tfsdk.Plan {
	t.Helper()
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schemaResp.Schema,
	}
}

func wrongRawServiceUserState(t *testing.T, r *ServiceUserResource) tfsdk.State {
	t.Helper()
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	return tfsdk.State{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schemaResp.Schema,
	}
}

func TestUnitServiceUserResource_configure(t *testing.T) {
	t.Run("wrong_type", func(t *testing.T) {
		r := &ServiceUserResource{}
		req := fwresource.ConfigureRequest{ProviderData: "not-a-client"}
		resp := &fwresource.ConfigureResponse{}
		r.Configure(context.Background(), req, resp)
		if !resp.Diagnostics.HasError() {
			t.Fatal("expected diagnostic error for wrong ProviderData type")
		}
	})
}

func TestUnitServiceUserResource_createPlanGetError(t *testing.T) {
	r := &ServiceUserResource{client: &client.Client{}}
	req := fwresource.CreateRequest{Plan: wrongRawServiceUserPlan(t, r)}
	resp := &fwresource.CreateResponse{}
	r.Create(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed plan")
	}
}

func TestUnitServiceUserResource_readStateGetError(t *testing.T) {
	r := &ServiceUserResource{client: &client.Client{}}
	req := fwresource.ReadRequest{State: wrongRawServiceUserState(t, r)}
	resp := &fwresource.ReadResponse{}
	r.Read(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed state")
	}
}

func TestUnitServiceUserResource_updatePlanGetError(t *testing.T) {
	r := &ServiceUserResource{client: &client.Client{}}
	req := fwresource.UpdateRequest{
		Plan:  wrongRawServiceUserPlan(t, r),
		State: wrongRawServiceUserState(t, r),
	}
	resp := &fwresource.UpdateResponse{}
	r.Update(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed plan")
	}
}

func TestUnitServiceUserResource_deleteStateGetError(t *testing.T) {
	r := &ServiceUserResource{client: &client.Client{}}
	req := fwresource.DeleteRequest{State: wrongRawServiceUserState(t, r)}
	resp := &fwresource.DeleteResponse{}
	r.Delete(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed state")
	}
}

// ─── Mock-server helpers (service_user) ──────────────────────────────────────

const mockValidServiceUserJSON = `{"_id":{"$oid":"aabbccddeeff001122334455"},"username":"tf-mock-svc-user","enabled":true,"admin":false,"create_server_permissions":false,"create_build_permissions":false,"config":{"type":"Service"}}`

// newSvcUserMockServer builds a mock server routing by "type" with a static
// route map; unmatched types return mockValidServiceUserJSON / 200.
func newSvcUserMockServer(t *testing.T, routes map[string]mockUserRoute) *httptest.Server {
	t.Helper()
	return newUserMockServerWithDefault(t, routes, mockValidServiceUserJSON)
}

// newSvcUserStatefulMockServer is the stateful variant for service_user.
func newSvcUserStatefulMockServer(t *testing.T, handler func(typ string, n int) (int, string)) *httptest.Server {
	t.Helper()
	return newStatefulUserMockServer(t, handler)
}

func mockSvcUserProviderConfig(srvURL string) string {
	return fmt.Sprintf(`
provider "komodo" {
  endpoint = %[1]q
  username = "mock-user"
  password = "mock-pass"
}
`, srvURL)
}

const mockSvcUserResourceConfig = `
resource "komodo_service_user" "test" {
  username = "tf-mock-svc-user"
}
`

const mockSvcUserResourceConfigAdmin = `
resource "komodo_service_user" "test" {
  username      = "tf-mock-svc-user"
  admin_enabled = true
}
`

const mockSvcUserResourceConfigDesc = `
resource "komodo_service_user" "test" {
  username    = "tf-mock-svc-user"
  description = "initial desc"
}
`

const mockSvcUserResourceConfigDesc2 = `
resource "komodo_service_user" "test" {
  username    = "tf-mock-svc-user"
  description = "updated desc"
}
`

// ─── Create error tests ───────────────────────────────────────────────────────

func TestAccServiceUserResource_createClientError(t *testing.T) {
	srv := newSvcUserMockServer(t, map[string]mockUserRoute{
		"CreateServiceUser": {http.StatusInternalServerError, `"internal error"`},
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockSvcUserProviderConfig(srv.URL) + mockSvcUserResourceConfig,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

func TestAccServiceUserResource_createFindUserError(t *testing.T) {
	srv := newSvcUserMockServer(t, map[string]mockUserRoute{
		"FindUser": {http.StatusInternalServerError, `"find error"`},
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockSvcUserProviderConfig(srv.URL) + mockSvcUserResourceConfig,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

func TestAccServiceUserResource_createFindUserNil(t *testing.T) {
	srv := newSvcUserMockServer(t, map[string]mockUserRoute{
		"FindUser": {http.StatusNotFound, `"not found"`},
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockSvcUserProviderConfig(srv.URL) + mockSvcUserResourceConfig,
				ExpectError: regexp.MustCompile(`(?i)not found`),
			},
		},
	})
}

func TestAccServiceUserResource_createUpdateAdminError(t *testing.T) {
	srv := newSvcUserMockServer(t, map[string]mockUserRoute{
		"UpdateUserAdmin": {http.StatusInternalServerError, `"admin error"`},
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockSvcUserProviderConfig(srv.URL) + mockSvcUserResourceConfigAdmin,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

func TestAccServiceUserResource_createUpdatePermsError(t *testing.T) {
	srv := newSvcUserMockServer(t, map[string]mockUserRoute{
		"UpdateUserBasePermissions": {http.StatusInternalServerError, `"perms error"`},
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockSvcUserProviderConfig(srv.URL) + mockSvcUserResourceConfig,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

func TestAccServiceUserResource_createSecondFindUserError(t *testing.T) {
	srv := newSvcUserStatefulMockServer(t, func(typ string, n int) (int, string) {
		if typ == "FindUser" && n >= 2 {
			return http.StatusInternalServerError, `"second find error"`
		}
		return http.StatusOK, mockValidServiceUserJSON
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockSvcUserProviderConfig(srv.URL) + mockSvcUserResourceConfig,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

func TestAccServiceUserResource_createSecondFindUserNil(t *testing.T) {
	srv := newSvcUserStatefulMockServer(t, func(typ string, n int) (int, string) {
		if typ == "FindUser" && n >= 2 {
			return http.StatusNotFound, `"not found"`
		}
		return http.StatusOK, mockValidServiceUserJSON
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockSvcUserProviderConfig(srv.URL) + mockSvcUserResourceConfig,
				ExpectError: regexp.MustCompile(`(?i)not found`),
			},
		},
	})
}

// ─── Read error test ──────────────────────────────────────────────────────────

func TestAccServiceUserResource_readFindUserError(t *testing.T) {
	// Step 1 Create: FindUser calls 1 & 2; post-apply refresh: call 3 (all succeed).
	// Step 2 pre-plan refresh (Read): FindUser call 4 → 500 → non-404 AddError path.
	srv := newSvcUserStatefulMockServer(t, func(typ string, n int) (int, string) {
		if typ == "FindUser" && n >= 4 {
			return http.StatusInternalServerError, `"read refresh error"`
		}
		return http.StatusOK, mockValidServiceUserJSON
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockSvcUserProviderConfig(srv.URL) + mockSvcUserResourceConfig,
			},
			{
				Config:      mockSvcUserProviderConfig(srv.URL) + mockSvcUserResourceConfigDesc,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

// ─── Update error tests ───────────────────────────────────────────────────────

func TestAccServiceUserResource_updateDescriptionError(t *testing.T) {
	srv := newSvcUserMockServer(t, map[string]mockUserRoute{
		"UpdateServiceUserDescription": {http.StatusInternalServerError, `"desc update error"`},
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockSvcUserProviderConfig(srv.URL) + mockSvcUserResourceConfigDesc,
			},
			{
				Config:      mockSvcUserProviderConfig(srv.URL) + mockSvcUserResourceConfigDesc2,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

func TestAccServiceUserResource_updateAdminError(t *testing.T) {
	srv := newSvcUserMockServer(t, map[string]mockUserRoute{
		"UpdateUserAdmin": {http.StatusInternalServerError, `"admin update error"`},
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockSvcUserProviderConfig(srv.URL) + mockSvcUserResourceConfig,
			},
			{
				Config:      mockSvcUserProviderConfig(srv.URL) + mockSvcUserResourceConfigAdmin,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

func TestAccServiceUserResource_updatePermsError(t *testing.T) {
	// UpdateUserBasePermissions call 1 (Create) succeeds; call 2+ (Update) fails.
	srv := newSvcUserStatefulMockServer(t, func(typ string, n int) (int, string) {
		if typ == "UpdateUserBasePermissions" && n >= 2 {
			return http.StatusInternalServerError, `"perms update error"`
		}
		return http.StatusOK, mockValidServiceUserJSON
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockSvcUserProviderConfig(srv.URL) + mockSvcUserResourceConfig,
			},
			{
				Config: mockSvcUserProviderConfig(srv.URL) + `
resource "komodo_service_user" "test" {
  username = "tf-mock-svc-user"
  enabled  = false
}
`,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

func TestAccServiceUserResource_updateFindUserError(t *testing.T) {
	// Step 1 Create: FindUser calls 1 & 2 succeed.
	// Step 2 Read refresh: FindUser call 3 succeeds.
	// Step 2 Update end: FindUser call 4 fails.
	srv := newSvcUserStatefulMockServer(t, func(typ string, n int) (int, string) {
		if typ == "FindUser" && n >= 4 {
			return http.StatusInternalServerError, `"find after update error"`
		}
		return http.StatusOK, mockValidServiceUserJSON
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockSvcUserProviderConfig(srv.URL) + mockSvcUserResourceConfig,
			},
			{
				Config: mockSvcUserProviderConfig(srv.URL) + `
resource "komodo_service_user" "test" {
  username = "tf-mock-svc-user"
  enabled  = false
}
`,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

func TestAccServiceUserResource_updateFindUserNil(t *testing.T) {
	srv := newSvcUserStatefulMockServer(t, func(typ string, n int) (int, string) {
		if typ == "FindUser" && n >= 4 {
			return http.StatusNotFound, `"not found after update"`
		}
		return http.StatusOK, mockValidServiceUserJSON
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockSvcUserProviderConfig(srv.URL) + mockSvcUserResourceConfig,
			},
			{
				Config: mockSvcUserProviderConfig(srv.URL) + `
resource "komodo_service_user" "test" {
  username = "tf-mock-svc-user"
  enabled  = false
}
`,
				ExpectError: regexp.MustCompile(`(?i)not found`),
			},
		},
	})
}

// ─── Delete error test ────────────────────────────────────────────────────────

func TestAccServiceUserResource_deleteClientError(t *testing.T) {
	// First DeleteUser call (the Destroy step) → 500; subsequent calls succeed for cleanup.
	srv := newSvcUserStatefulMockServer(t, func(typ string, n int) (int, string) {
		if typ == "DeleteUser" && n == 1 {
			return http.StatusInternalServerError, `"delete error"`
		}
		return http.StatusOK, mockValidServiceUserJSON
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockSvcUserProviderConfig(srv.URL) + mockSvcUserResourceConfig,
			},
			{
				Config:      mockSvcUserProviderConfig(srv.URL),
				Destroy:     true,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

// ─── Read disappears test ─────────────────────────────────────────────────────

func TestAccServiceUserResource_readDisappears_unit(t *testing.T) {
	// Create succeeds; post-apply refresh (FindUser call 3) returns 404 → nil → RemoveResource.
	srv := newSvcUserStatefulMockServer(t, func(typ string, n int) (int, string) {
		if typ == "FindUser" && n >= 3 {
			return http.StatusNotFound, `"not found"`
		}
		return http.StatusOK, mockValidServiceUserJSON
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             mockSvcUserProviderConfig(srv.URL) + mockSvcUserResourceConfig,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// ─── Successful update test ───────────────────────────────────────────────────

func TestAccServiceUserResource_updateSuccess(t *testing.T) {
	// All calls succeed; Update path + final FindUser + set-state.
	srv := newSvcUserMockServer(t, nil)
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockSvcUserProviderConfig(srv.URL) + mockSvcUserResourceConfig,
			},
			{
				Config: mockSvcUserProviderConfig(srv.URL) + `
resource "komodo_service_user" "test" {
  username    = "tf-mock-svc-user"
  description = "updated desc"
}
`,
				Check: resource.TestCheckResourceAttr("komodo_service_user.test", "username", "tf-mock-svc-user"),
			},
		},
	})
}

// ─── Unit tests for ValidateConfig ───────────────────────────────────────────

func TestUnitServiceUserResource_validateConfigGetError(t *testing.T) {
	r := &ServiceUserResource{}
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

func TestAccServiceUserResource_adminConflictServer_unit(t *testing.T) {
	srv := newSvcUserMockServer(t, nil)
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockSvcUserProviderConfig(srv.URL) + testAccServiceUserResourceConfig_adminWithCreateServer("tf-mock-svc-user"),
				ExpectError: regexp.MustCompile(`create_server_enabled cannot be set to true alongside admin_enabled = true`),
			},
		},
	})
}

func TestAccServiceUserResource_adminConflictBuild_unit(t *testing.T) {
	srv := newSvcUserMockServer(t, nil)
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockSvcUserProviderConfig(srv.URL) + testAccServiceUserResourceConfig_adminWithCreateBuild("tf-mock-svc-user"),
				ExpectError: regexp.MustCompile(`create_build_enabled cannot be set to true alongside admin_enabled = true`),
			},
		},
	})
}

// ─── Unit test for Read non-404 error path ────────────────────────────────────

func TestAccServiceUserResource_readNon404Error(t *testing.T) {
	// Step 1 Create: calls 1 & 2 succeed; post-apply refresh (call 3) → 500.
	srv := newSvcUserStatefulMockServer(t, func(typ string, n int) (int, string) {
		if typ == "FindUser" && n >= 3 {
			return http.StatusInternalServerError, `"read error"`
		}
		return http.StatusOK, mockValidServiceUserJSON
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockSvcUserProviderConfig(srv.URL) + mockSvcUserResourceConfig,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}
