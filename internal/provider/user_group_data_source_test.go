// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"

	datasource "github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserGroupDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupDataSourceConfig_basic("tf-test-ds-group"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_user_group.test", "name", "tf-test-ds-group"),
					resource.TestCheckResourceAttr("data.komodo_user_group.test", "everyone_enabled", "false"),
					resource.TestCheckResourceAttrSet("data.komodo_user_group.test", "id"),
					resource.TestCheckResourceAttrSet("data.komodo_user_group.test", "updated_at"),
					resource.TestCheckResourceAttr("data.komodo_user_group.test", "users.#", "0"),
				),
			},
		},
	})
}

func TestAccUserGroupDataSource_withUsers(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupDataSourceConfig_withUser("tf-test-ds-group-users", "tf-acc-ugds-user", "Password1!"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_user_group.test", "name", "tf-test-ds-group-users"),
					resource.TestCheckResourceAttr("data.komodo_user_group.test", "users.#", "1"),
					resource.TestCheckResourceAttrPair("data.komodo_user_group.test", "users.0", "komodo_user.member", "id"),
				),
			},
		},
	})
}

func testAccUserGroupDataSourceConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "komodo_user_group" "test" {
  name = %q
}

data "komodo_user_group" "test" {
  name = komodo_user_group.test.name
}
`, name)
}

func testAccUserGroupDataSourceConfig_withUser(name, username, password string) string {
	return fmt.Sprintf(`
resource "komodo_user" "member" {
  username = %[2]q
  password = %[3]q
}

resource "komodo_user_group" "test" {
  name  = %[1]q
  users = [komodo_user.member.id]
}

data "komodo_user_group" "test" {
  name = komodo_user_group.test.name
}
`, name, username, password)
}

func TestUnitUserGroupDataSource_configure(t *testing.T) {
	d := &UserGroupDataSource{}
	resp := &datasource.ConfigureResponse{}
	d.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: "wrong"}, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for wrong provider data type")
	}
}

func TestAccUserGroupDataSource_bothSet_isError(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccUserGroupDataSourceConfig_bothSet(),
				ExpectError: regexp.MustCompile(`Only one of`),
			},
		},
	})
}

func TestAccUserGroupDataSource_neitherSet_isError(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccUserGroupDataSourceConfig_neitherSet(),
				ExpectError: regexp.MustCompile(`One of`),
			},
		},
	})
}

func testAccUserGroupDataSourceConfig_bothSet() string {
	return `
data "komodo_user_group" "test" {
  id   = "some-id"
  name = "some-name"
}
`
}

func testAccUserGroupDataSourceConfig_neitherSet() string {
	return `
data "komodo_user_group" "test" {}
`
}

// ─── ValidateConfig unit tests ────────────────────────────────────────────────

func TestUnitUserGroupDataSource_validateConfigGetError(t *testing.T) {
	d := &UserGroupDataSource{}
	ctx := context.Background()
	schemaResp := &datasource.SchemaResponse{}
	d.Schema(ctx, datasource.SchemaRequest{}, schemaResp)
	req := datasource.ValidateConfigRequest{
		Config: tfsdk.Config{
			Raw:    tftypes.NewValue(tftypes.String, "invalid"),
			Schema: schemaResp.Schema,
		},
	}
	resp := &datasource.ValidateConfigResponse{}
	d.ValidateConfig(ctx, req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error for malformed config in ValidateConfig")
	}
}

func TestUnitUserGroupDataSource_validateConfigUnknown(t *testing.T) {
	d := &UserGroupDataSource{}
	ctx := context.Background()
	schemaResp := &datasource.SchemaResponse{}
	d.Schema(ctx, datasource.SchemaRequest{}, schemaResp)
	objType := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	configVal := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"name":             tftypes.NewValue(tftypes.String, nil),
		"everyone_enabled": tftypes.NewValue(tftypes.Bool, nil),
		"users":            tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"all":              tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
		"updated_at":       tftypes.NewValue(tftypes.String, nil),
	})
	req := datasource.ValidateConfigRequest{
		Config: tfsdk.Config{Raw: configVal, Schema: schemaResp.Schema},
	}
	resp := &datasource.ValidateConfigResponse{}
	d.ValidateConfig(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("expected no error for unknown id: %v", resp.Diagnostics)
	}
}

func TestUnitUserGroupDataSource_readConfigGetError(t *testing.T) {
	d := &UserGroupDataSource{}
	ctx := context.Background()
	schemaResp := &datasource.SchemaResponse{}
	d.Schema(ctx, datasource.SchemaRequest{}, schemaResp)
	req := datasource.ReadRequest{
		Config: tfsdk.Config{
			Raw:    tftypes.NewValue(tftypes.String, "invalid"),
			Schema: schemaResp.Schema,
		},
	}
	resp := &datasource.ReadResponse{}
	d.Read(ctx, req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error from malformed config")
	}
}

// ─── Mock-server tests ────────────────────────────────────────────────────────

const mockUserGroupDataSourceConfig = `
data "komodo_user_group" "test" {
  name = "tf-mock-group"
}
`

func TestAccUserGroupDataSource_getGroupError(t *testing.T) {
	srv := newUserGroupMockServer(t, map[string]mockUserGroupRoute{
		"GetUserGroup": {http.StatusInternalServerError, `"get error"`},
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockUserGroupProviderConfig(srv.URL) + mockUserGroupDataSourceConfig,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

func TestAccUserGroupDataSource_readSuccess(t *testing.T) {
	srv := newUserGroupMockServer(t, nil)
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockUserGroupProviderConfig(srv.URL) + mockUserGroupDataSourceConfig,
				Check:  resource.TestCheckResourceAttr("data.komodo_user_group.test", "name", "tf-mock-group"),
			},
		},
	})
}

// TestUnitUserGroupDataSource_readByIDDirect tests the ID-fallback lookup path (name == "" → use ID).
func TestUnitUserGroupDataSource_readByIDDirect(t *testing.T) {
	srv := newUserGroupMockServer(t, map[string]mockUserGroupRoute{
		"GetUserGroup": {http.StatusOK, mockValidUserGroupJSON},
	})
	c := client.NewClient(srv.URL, "mock-user", "mock-pass")
	d := &UserGroupDataSource{client: c}
	ctx := context.Background()
	schemaResp := &datasource.SchemaResponse{}
	d.Schema(ctx, datasource.SchemaRequest{}, schemaResp)
	objType := schemaResp.Schema.Type().TerraformType(ctx)
	configVal := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, "aabbccddeeff001122334455"),
		"name":             tftypes.NewValue(tftypes.String, ""),
		"everyone_enabled": tftypes.NewValue(tftypes.Bool, nil),
		"users":            tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"all":              tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
		"updated_at":       tftypes.NewValue(tftypes.String, nil),
	})
	req := datasource.ReadRequest{Config: tfsdk.Config{Raw: configVal, Schema: schemaResp.Schema}}
	resp := &datasource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	d.Read(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %v", resp.Diagnostics)
	}
}

// TestUnitUserGroupDataSource_readWithNonEmptyAllDirect exercises the non-empty All loop body.
func TestUnitUserGroupDataSource_readWithNonEmptyAllDirect(t *testing.T) {
	srv := newUserGroupMockServer(t, map[string]mockUserGroupRoute{
		"GetUserGroup": {http.StatusOK, mockValidUserGroupWithUsersJSON},
	})
	c := client.NewClient(srv.URL, "mock-user", "mock-pass")
	// Verify client directly
	group, err := c.GetUserGroup(context.Background(), "tf-mock-group")
	if err != nil {
		t.Fatalf("GetUserGroup error: %v", err)
	}
	if group == nil {
		t.Fatal("expected non-nil group")
	}
	if len(group.All) == 0 {
		t.Fatalf("expected non-empty All, got: %v", group.All)
	}
	d := &UserGroupDataSource{client: c}
	ctx := context.Background()
	schemaResp := &datasource.SchemaResponse{}
	d.Schema(ctx, datasource.SchemaRequest{}, schemaResp)
	objType := schemaResp.Schema.Type().TerraformType(ctx)
	configVal := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, nil),
		"name":             tftypes.NewValue(tftypes.String, "tf-mock-group"),
		"everyone_enabled": tftypes.NewValue(tftypes.Bool, nil),
		"users":            tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"all":              tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
		"updated_at":       tftypes.NewValue(tftypes.String, nil),
	})
	req := datasource.ReadRequest{Config: tfsdk.Config{Raw: configVal, Schema: schemaResp.Schema}}
	resp := &datasource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	d.Read(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %v", resp.Diagnostics)
	}
	// Verify All map was populated (non-empty all loop body was exercised)
	var stateData UserGroupDataSourceModel
	resp.State.Get(ctx, &stateData)
	if stateData.All.IsNull() || stateData.All.IsUnknown() || len(stateData.All.Elements()) == 0 {
		t.Fatalf("expected non-empty All in state, got: %v", stateData.All)
	}
}

// TestUnitUserGroupDataSource_readNilUsersDirect covers the userIDs == nil branch.
func TestUnitUserGroupDataSource_readNilUsersDirect(t *testing.T) {
	srv := newUserGroupMockServer(t, map[string]mockUserGroupRoute{
		"GetUserGroup": {http.StatusOK, mockValidUserGroupNullUsersJSON},
	})
	c := client.NewClient(srv.URL, "mock-user", "mock-pass")
	d := &UserGroupDataSource{client: c}
	ctx := context.Background()
	schemaResp := &datasource.SchemaResponse{}
	d.Schema(ctx, datasource.SchemaRequest{}, schemaResp)
	objType := schemaResp.Schema.Type().TerraformType(ctx)
	configVal := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, nil),
		"name":             tftypes.NewValue(tftypes.String, "tf-mock-group"),
		"everyone_enabled": tftypes.NewValue(tftypes.Bool, nil),
		"users":            tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"all":              tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
		"updated_at":       tftypes.NewValue(tftypes.String, nil),
	})
	req := datasource.ReadRequest{Config: tfsdk.Config{Raw: configVal, Schema: schemaResp.Schema}}
	resp := &datasource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	d.Read(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %v", resp.Diagnostics)
	}
}
