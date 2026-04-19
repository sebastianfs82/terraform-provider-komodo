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

func TestAccUserGroupMembershipResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupMembershipResourceConfig_basic("tf-test-membership-group", "tf-test-membership-svc"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user_group_membership.test", "user_group_id", "tf-test-membership-group"),
					resource.TestCheckResourceAttr("komodo_user_group_membership.test", "user_id", "tf-test-membership-svc"),
					resource.TestCheckResourceAttr("komodo_user_group_membership.test", "id", "tf-test-membership-group/tf-test-membership-svc"),
				),
			},
		},
	})
}

func TestAccUserGroupMembershipResource_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupMembershipResourceConfig_basic("tf-test-membership-import-group", "tf-test-membership-import-svc"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_user_group_membership.test", "id"),
				),
			},
			{
				Config:            testAccUserGroupMembershipResourceConfig_basic("tf-test-membership-import-group", "tf-test-membership-import-svc"),
				ResourceName:      "komodo_user_group_membership.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "tf-test-membership-import-group/tf-test-membership-import-svc",
			},
		},
	})
}

func TestAccUserGroupMembershipResource_disappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupMembershipResourceConfig_basic("tf-test-membership-disappears-group", "tf-test-membership-disappears-svc"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user_group_membership.test", "user_group_id", "tf-test-membership-disappears-group"),
					testAccUserGroupMembershipDisappears("komodo_user_group_membership.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccUserGroupMembershipDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		userGroup := rs.Primary.Attributes["user_group_id"]
		user := rs.Primary.Attributes["user_id"]

		c := client.NewClient(
			os.Getenv("KOMODO_ENDPOINT"),
			os.Getenv("KOMODO_USERNAME"),
			os.Getenv("KOMODO_PASSWORD"),
		)

		_, err := c.RemoveUserFromUserGroup(context.Background(), client.RemoveUserFromUserGroupRequest{
			UserGroup: userGroup,
			User:      user,
		})
		return err
	}
}

// Config helpers

func testAccUserGroupMembershipResourceConfig_basic(groupName, svcUserName string) string {
	return fmt.Sprintf(`
resource "komodo_user_group" "test" {
  name = %q
}

resource "komodo_service_user" "test" {
  username    = %q
  description = "test service user for membership"
}

resource "komodo_user_group_membership" "test" {
  user_group_id = komodo_user_group.test.name
  user_id       = komodo_service_user.test.username

  depends_on = [komodo_user_group.test, komodo_service_user.test]
}
`, groupName, svcUserName)
}

func TestAccUserGroupMembershipResource_everyoneEnabledBlocked(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccUserGroupMembershipResourceConfig_everyoneEnabled("tf-test-membership-everyone-group", "tf-test-membership-everyone-svc"),
				ExpectError: regexp.MustCompile(`everyone_enabled.*is true`),
			},
		},
	})
}

func testAccUserGroupMembershipResourceConfig_everyoneEnabled(groupName, svcUserName string) string {
	return fmt.Sprintf(`
resource "komodo_user_group" "test" {
  name             = %q
  everyone_enabled = true
}

resource "komodo_service_user" "test" {
  username    = %q
  description = "test service user for membership everyone check"
}

resource "komodo_user_group_membership" "test" {
  user_group_id = komodo_user_group.test.name
  user_id       = komodo_service_user.test.username

  depends_on = [komodo_user_group.test, komodo_service_user.test]
}
`, groupName, svcUserName)
}

// ─── Unit tests ──────────────────────────────────────────────────────────────

func wrongRawUserGroupMembershipPlan(t *testing.T, r *UserGroupMembershipResource) tfsdk.Plan {
	t.Helper()
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schemaResp.Schema,
	}
}

func wrongRawUserGroupMembershipState(t *testing.T, r *UserGroupMembershipResource) tfsdk.State {
	t.Helper()
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	return tfsdk.State{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schemaResp.Schema,
	}
}

func TestUnitUserGroupMembershipResource_configure(t *testing.T) {
	t.Run("wrong_type", func(t *testing.T) {
		r := &UserGroupMembershipResource{}
		req := fwresource.ConfigureRequest{ProviderData: "not-a-client"}
		resp := &fwresource.ConfigureResponse{}
		r.Configure(context.Background(), req, resp)
		if !resp.Diagnostics.HasError() {
			t.Fatal("expected diagnostic error for wrong ProviderData type")
		}
	})
}

func TestUnitUserGroupMembershipResource_createPlanGetError(t *testing.T) {
	r := &UserGroupMembershipResource{client: &client.Client{}}
	req := fwresource.CreateRequest{Plan: wrongRawUserGroupMembershipPlan(t, r)}
	resp := &fwresource.CreateResponse{}
	r.Create(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed plan")
	}
}

func TestUnitUserGroupMembershipResource_readStateGetError(t *testing.T) {
	r := &UserGroupMembershipResource{client: &client.Client{}}
	req := fwresource.ReadRequest{State: wrongRawUserGroupMembershipState(t, r)}
	resp := &fwresource.ReadResponse{}
	r.Read(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed state")
	}
}

func TestUnitUserGroupMembershipResource_deleteStateGetError(t *testing.T) {
	r := &UserGroupMembershipResource{client: &client.Client{}}
	req := fwresource.DeleteRequest{State: wrongRawUserGroupMembershipState(t, r)}
	resp := &fwresource.DeleteResponse{}
	r.Delete(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed state")
	}
}

func TestUnitUserGroupMembershipResource_updateIsNoop(t *testing.T) {
	r := &UserGroupMembershipResource{}
	req := fwresource.UpdateRequest{}
	resp := &fwresource.UpdateResponse{}
	r.Update(context.Background(), req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error in no-op Update: %v", resp.Diagnostics)
	}
}

func TestUnitUserGroupMembershipResource_importState_invalidFormat(t *testing.T) {
	r := &UserGroupMembershipResource{}
	req := fwresource.ImportStateRequest{ID: "missing-slash"}
	resp := &fwresource.ImportStateResponse{}
	r.ImportState(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for invalid import ID format")
	}
}

// ─── Mock server infrastructure ──────────────────────────────────────────────

// Membership tests reuse the userGroup mock server infrastructure defined in
// user_group_resource_test.go (same package). The constants/helpers there are
// available here without re-declaration.

const mockValidMembershipUserJSON = `{
  "_id": {"$oid": "bbccddeeff001122334455aa"},
  "username": "tf-mock-member",
  "enabled": true,
  "admin": false,
  "create_server_permissions": false,
  "create_build_permissions": false
}`

const mockValidMembershipGroupJSON = `{
  "_id": {"$oid": "aabbccddeeff001122334455"},
  "name": "tf-mock-group",
  "everyone": false,
  "users": ["bbccddeeff001122334455aa"],
  "all": {},
  "updated_at": 0
}`

const mockEmptyGroupJSON = `{
  "_id": {"$oid": "aabbccddeeff001122334455"},
  "name": "tf-mock-group",
  "everyone": false,
  "users": [],
  "all": {},
  "updated_at": 0
}`

const mockEveryoneGroupJSON = `{
  "_id": {"$oid": "aabbccddeeff001122334455"},
  "name": "tf-mock-group",
  "everyone": true,
  "users": [],
  "all": {},
  "updated_at": 0
}`

func mockMembershipProviderConfig(srvURL string) string {
	return fmt.Sprintf(`
provider "komodo" {
  endpoint = %[1]q
  username = "mock-user"
  password = "mock-pass"
}
`, srvURL)
}

const mockMembershipResourceConfig = `
resource "komodo_user_group_membership" "test" {
  user_group_id = "tf-mock-group"
  user_id       = "tf-mock-member"
}
`

func newMembershipMockServer(t *testing.T, getGroupBody string, extraRoutes map[string]mockUserGroupRoute) *httptest.Server {
	t.Helper()
	defaults := map[string]mockUserGroupRoute{
		"GetUserGroup": {http.StatusOK, getGroupBody},
		"FindUser":     {http.StatusOK, mockValidMembershipUserJSON},
	}
	for k, v := range extraRoutes {
		defaults[k] = v
	}
	return newUserGroupMockServerWithDefault(t, defaults, mockValidMembershipGroupJSON)
}

// ─── Create error path tests ──────────────────────────────────────────────────

func TestAccUserGroupMembershipResource_createGetGroupError(t *testing.T) {
	srv := newUserGroupMockServer(t, map[string]mockUserGroupRoute{
		"GetUserGroup": {http.StatusInternalServerError, `"get error"`},
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockMembershipProviderConfig(srv.URL) + mockMembershipResourceConfig,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

func TestAccUserGroupMembershipResource_createGroupNil(t *testing.T) {
	srv := newUserGroupMockServer(t, map[string]mockUserGroupRoute{
		"GetUserGroup": {http.StatusNotFound, `"not found"`},
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockMembershipProviderConfig(srv.URL) + mockMembershipResourceConfig,
				ExpectError: regexp.MustCompile(`(?i)not found`),
			},
		},
	})
}

func TestAccUserGroupMembershipResource_createEveryoneBlocked(t *testing.T) {
	srv := newUserGroupMockServer(t, map[string]mockUserGroupRoute{
		"GetUserGroup": {http.StatusOK, mockEveryoneGroupJSON},
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockMembershipProviderConfig(srv.URL) + mockMembershipResourceConfig,
				ExpectError: regexp.MustCompile(`(?i)everyone`),
			},
		},
	})
}

func TestAccUserGroupMembershipResource_createAddUserError(t *testing.T) {
	srv := newMembershipMockServer(t, mockEmptyGroupJSON, map[string]mockUserGroupRoute{
		"AddUserToUserGroup": {http.StatusInternalServerError, `"add error"`},
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockMembershipProviderConfig(srv.URL) + mockMembershipResourceConfig,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

// ─── Create + Read success test ───────────────────────────────────────────────

func TestAccUserGroupMembershipResource_createSuccess(t *testing.T) {
	srv := newMembershipMockServer(t, mockValidMembershipGroupJSON, nil)
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockMembershipProviderConfig(srv.URL) + mockMembershipResourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user_group_membership.test", "user_group_id", "tf-mock-group"),
					resource.TestCheckResourceAttr("komodo_user_group_membership.test", "user_id", "tf-mock-member"),
				),
			},
		},
	})
}

// ─── Read: user no longer in group (disappears) ───────────────────────────────

func TestAccUserGroupMembershipResource_readDisappears_unit(t *testing.T) {
	// Create sees user in group; subsequent read sees empty users → RemoveResource
	var createCount int
	srv := newStatefulUserGroupMockServer(t, func(typ string, n int) (int, string) {
		switch typ {
		case "GetUserGroup":
			createCount++
			if createCount > 1 {
				return http.StatusOK, mockEmptyGroupJSON
			}
			return http.StatusOK, mockValidMembershipGroupJSON
		case "FindUser":
			return http.StatusOK, mockValidMembershipUserJSON
		case "AddUserToUserGroup":
			return http.StatusOK, mockValidMembershipGroupJSON
		}
		return http.StatusOK, mockValidMembershipGroupJSON
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             mockMembershipProviderConfig(srv.URL) + mockMembershipResourceConfig,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// ─── Read: GetUserGroup error ─────────────────────────────────────────────────

func TestAccUserGroupMembershipResource_readGetGroupError(t *testing.T) {
	var createDone bool
	srv := newStatefulUserGroupMockServer(t, func(typ string, n int) (int, string) {
		switch typ {
		case "GetUserGroup":
			if createDone {
				return http.StatusInternalServerError, `"read error"`
			}
			return http.StatusOK, mockValidMembershipGroupJSON
		case "AddUserToUserGroup":
			createDone = true
			return http.StatusOK, mockValidMembershipGroupJSON
		case "FindUser":
			return http.StatusOK, mockValidMembershipUserJSON
		}
		return http.StatusOK, mockValidMembershipGroupJSON
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockMembershipProviderConfig(srv.URL) + mockMembershipResourceConfig,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

// ─── Delete error test ────────────────────────────────────────────────────────

func TestAccUserGroupMembershipResource_deleteError(t *testing.T) {
	srv := newStatefulUserGroupMockServer(t, func(typ string, n int) (int, string) {
		switch typ {
		case "GetUserGroup":
			return http.StatusOK, mockValidMembershipGroupJSON
		case "AddUserToUserGroup":
			return http.StatusOK, mockValidMembershipGroupJSON
		case "FindUser":
			return http.StatusOK, mockValidMembershipUserJSON
		case "RemoveUserFromUserGroup":
			if n == 1 {
				return http.StatusInternalServerError, `"remove error"`
			}
			return http.StatusOK, mockEmptyGroupJSON
		}
		return http.StatusOK, mockValidMembershipGroupJSON
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockMembershipProviderConfig(srv.URL) + mockMembershipResourceConfig,
			},
			{
				Config:      mockMembershipProviderConfig(srv.URL),
				Destroy:     true,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

// ─── ImportState: valid format direct unit test ───────────────────────────────

func TestUnitUserGroupMembershipResource_importStateDirect(t *testing.T) {
	r := &UserGroupMembershipResource{}
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	emptyState := tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), map[string]tftypes.Value{
		"id":            tftypes.NewValue(tftypes.String, nil),
		"user_group_id": tftypes.NewValue(tftypes.String, nil),
		"user_id":       tftypes.NewValue(tftypes.String, nil),
	})
	req := fwresource.ImportStateRequest{ID: "my-group/my-user"}
	resp := &fwresource.ImportStateResponse{State: tfsdk.State{Schema: schemaResp.Schema, Raw: emptyState}}
	r.ImportState(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error in ImportState: %v", resp.Diagnostics)
	}
}
