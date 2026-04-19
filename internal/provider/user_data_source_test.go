// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"testing"

	datasource "github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserDataSource_byUsername(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserDataSourceConfig_byUsername("tf-user-ds-name", "Password1!"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_user.test", "username", "tf-user-ds-name"),
					resource.TestCheckResourceAttrSet("data.komodo_user.test", "id"),
					resource.TestCheckResourceAttr("data.komodo_user.test", "enabled", "true"),
					resource.TestCheckResourceAttr("data.komodo_user.test", "admin_enabled", "false"),
					resource.TestCheckResourceAttr("data.komodo_user.test", "create_server_enabled", "false"),
					resource.TestCheckResourceAttr("data.komodo_user.test", "create_build_enabled", "false"),
				),
			},
		},
	})
}

func TestAccUserDataSource_byID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserDataSourceConfig_byID("tf-user-ds-id", "Password1!"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_user.test", "username", "tf-user-ds-id"),
					resource.TestCheckResourceAttrSet("data.komodo_user.test", "id"),
					resource.TestCheckResourceAttrPair(
						"data.komodo_user.test", "id",
						"komodo_user.test", "id",
					),
				),
			},
		},
	})
}

func TestAccUserDataSource_withPermissions(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserDataSourceConfig_withPermissions("tf-user-ds-perms", "Password1!"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_user.test", "username", "tf-user-ds-perms"),
					resource.TestCheckResourceAttr("data.komodo_user.test", "create_server_enabled", "true"),
					resource.TestCheckResourceAttr("data.komodo_user.test", "create_build_enabled", "true"),
				),
			},
		},
	})
}

// Config helpers

func testAccUserDataSourceConfig_byUsername(username, password string) string {
	return fmt.Sprintf(`
resource "komodo_user" "test" {
  username = %[1]q
  password = %[2]q
}

data "komodo_user" "test" {
  username = komodo_user.test.username
}
`, username, password)
}

func testAccUserDataSourceConfig_byID(username, password string) string {
	return fmt.Sprintf(`
resource "komodo_user" "test" {
  username = %[1]q
  password = %[2]q
}

data "komodo_user" "test" {
  id = komodo_user.test.id
}
`, username, password)
}

func testAccUserDataSourceConfig_withPermissions(username, password string) string {
	return fmt.Sprintf(`
resource "komodo_user" "test" {
  username       = %[1]q
  password       = %[2]q
  create_server_enabled = true
  create_build_enabled  = true
}

data "komodo_user" "test" {
  username = komodo_user.test.username
}
`, username, password)
}

func TestUnitUserDataSource_configure(t *testing.T) {
	d := &UserDataSource{}
	resp := &datasource.ConfigureResponse{}
	d.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: "wrong"}, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for wrong provider data type")
	}
}

func TestAccUserDataSource_bothSet_isError(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccUserDataSourceConfig_bothSet(),
				ExpectError: regexp.MustCompile(`Only one of`),
			},
		},
	})
}

func TestAccUserDataSource_neitherSet_isError(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccUserDataSourceConfig_neitherSet(),
				ExpectError: regexp.MustCompile(`One of`),
			},
		},
	})
}

func testAccUserDataSourceConfig_bothSet() string {
	return `
data "komodo_user" "test" {
  id       = "some-id"
  username = "some-username"
}
`
}

func testAccUserDataSourceConfig_neitherSet() string {
	return `
data "komodo_user" "test" {}
`
}

// ─── Unit tests for ValidateConfig / Read HasError ────────────────────────────

func TestUnitUserDataSource_validateConfigGetError(t *testing.T) {
	d := &UserDataSource{}
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
		t.Fatal("expected error from malformed config")
	}
}

func TestUnitUserDataSource_validateConfigUnknown(t *testing.T) {
	// username is Unknown → ValidateConfig should return early without error (unknown values deferred).
	d := &UserDataSource{}
	ctx := context.Background()
	schemaResp := &datasource.SchemaResponse{}
	d.Schema(ctx, datasource.SchemaRequest{}, schemaResp)
	objType := schemaResp.Schema.Type().TerraformType(ctx)
	configVal := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                    tftypes.NewValue(tftypes.String, nil), // null
		"username":              tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"enabled":               tftypes.NewValue(tftypes.Bool, nil),
		"admin_enabled":         tftypes.NewValue(tftypes.Bool, nil),
		"create_server_enabled": tftypes.NewValue(tftypes.Bool, nil),
		"create_build_enabled":  tftypes.NewValue(tftypes.Bool, nil),
	})
	req := datasource.ValidateConfigRequest{
		Config: tfsdk.Config{Raw: configVal, Schema: schemaResp.Schema},
	}
	resp := &datasource.ValidateConfigResponse{}
	d.ValidateConfig(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("expected no error for unknown username: %v", resp.Diagnostics)
	}
}

func TestUnitUserDataSource_readConfigGetError(t *testing.T) {
	d := &UserDataSource{}
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

const mockUserDataSourceConfig = `
data "komodo_user" "test" {
  username = "tf-mock-user"
}
`

func TestAccUserDataSource_findUserError(t *testing.T) {
	srv := newUserMockServer(t, map[string]mockUserRoute{
		"FindUser": {http.StatusInternalServerError, `"find error"`},
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockUserProviderConfig(srv.URL) + mockUserDataSourceConfig,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

func TestAccUserDataSource_findUserNil(t *testing.T) {
	srv := newUserMockServer(t, map[string]mockUserRoute{
		"FindUser": {http.StatusNotFound, `"not found"`},
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockUserProviderConfig(srv.URL) + mockUserDataSourceConfig,
				ExpectError: regexp.MustCompile(`(?i)not found`),
			},
		},
	})
}

// ─── Successful data source read test ────────────────────────────────────────

func TestAccUserDataSource_readSuccess(t *testing.T) {
	// All calls succeed → data source read maps user fields to state.
	srv := newUserMockServer(t, nil)
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockUserProviderConfig(srv.URL) + mockUserDataSourceConfig,
				Check:  resource.TestCheckResourceAttr("data.komodo_user.test", "username", "tf-mock-user"),
			},
		},
	})
}
