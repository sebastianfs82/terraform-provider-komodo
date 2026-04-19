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

func TestAccUserGroupResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupResourceConfig_basic("tf-test-group"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user_group.test", "name", "tf-test-group"),
					resource.TestCheckResourceAttr("komodo_user_group.test", "everyone_enabled", "false"),
					resource.TestCheckResourceAttrSet("komodo_user_group.test", "id"),
					resource.TestCheckResourceAttrSet("komodo_user_group.test", "updated_at"),
				),
			},
		},
	})
}

func TestAccUserGroupResource_rename(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupResourceConfig_basic("tf-test-group-original"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user_group.test", "name", "tf-test-group-original"),
				),
			},
			{
				Config: testAccUserGroupResourceConfig_basic("tf-test-group-renamed"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user_group.test", "name", "tf-test-group-renamed"),
					resource.TestCheckResourceAttrSet("komodo_user_group.test", "id"),
				),
			},
		},
	})
}

func TestAccUserGroupResource_withUsers(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupResourceConfig_withNewUser("tf-test-group-users"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user_group.test", "name", "tf-test-group-users"),
					resource.TestCheckResourceAttr("komodo_user_group.test", "users.#", "1"),
					resource.TestCheckResourceAttrPair("komodo_user_group.test", "users.0", "komodo_user.test", "id"),
				),
			},
		},
	})
}

func TestAccUserGroupResource_addRemoveUser(t *testing.T) {
	var groupID string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupResourceConfig_withNewUser("tf-test-group-add-remove"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user_group.test", "users.#", "1"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["komodo_user_group.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						groupID = rs.Primary.ID
						return nil
					},
				),
			},
			// Remove the user (user resource still present, but removed from group)
			{
				Config: testAccUserGroupResourceConfig_withNewUserOnly("tf-test-group-add-remove"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("komodo_user_group.test", "users"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["komodo_user_group.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						if rs.Primary.ID != groupID {
							return fmt.Errorf("expected same group ID after user removal, got %s", rs.Primary.ID)
						}
						return nil
					},
				),
			},
		},
	})
}

func TestAccUserGroupResource_import(t *testing.T) {
	var groupID string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupResourceConfig_basic("tf-test-group-import"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_user_group.test", "id"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["komodo_user_group.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						groupID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config:            testAccUserGroupResourceConfig_basic("tf-test-group-import"),
				ResourceName:      "komodo_user_group.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(*terraform.State) (string, error) {
					return groupID, nil
				},
			},
		},
	})
}

func TestAccUserGroupResource_everyoneConflictsWithUsers(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccUserGroupResourceConfig_everyoneAndUsers("tf-test-group-invalid"),
				ExpectError: regexp.MustCompile("Conflicting configuration"),
			},
		},
	})
}

// TestAccUserGroupResource_unmanagedUsersNoDrift verifies that when users is
// not configured, externally-added users do not cause a non-empty plan (no drift).
func TestAccUserGroupResource_unmanagedUsersNoDrift(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create the group and a user; add user out-of-band
				Config: testAccUserGroupResourceConfig_withUserResOnly("tf-test-group-unmanaged"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_user_group.test", "id"),
					// Add the provisioned user out-of-band via the API
					testAccUserGroupAddUserFromState("komodo_user_group.test", "komodo_user.test"),
				),
				// No diff should be produced despite the external user addition
				ExpectNonEmptyPlan: false,
			},
			{
				// Re-apply the same config — must produce an empty plan
				Config:   testAccUserGroupResourceConfig_withUserResOnly("tf-test-group-unmanaged"),
				PlanOnly: true,
			},
		},
	})
}

// TestAccUserGroupResource_unmanagedUsersNotRemoved verifies that switching from
// a managed users list back to no users config removes the previously-managed users
// once, then stops tracking the list (future manual additions are not touched).
func TestAccUserGroupResource_unmanagedUsersNotRemoved(t *testing.T) {
	var savedUserID string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Start with managed user list
				Config: testAccUserGroupResourceConfig_withNewUser("tf-test-group-unmanage-transition"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user_group.test", "users.#", "1"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["komodo_user.test"]
						if !ok {
							return fmt.Errorf("komodo_user.test not found in state")
						}
						savedUserID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				// Remove users from config — Terraform should remove the previously-managed
				// user once, then stop managing the list.
				Config: testAccUserGroupResourceConfig_withNewUserOnly("tf-test-group-unmanage-transition"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// users attribute is now null in state (unmanaged)
					resource.TestCheckNoResourceAttr("komodo_user_group.test", "users"),
					// The previously-managed user has been removed from the group
					func(s *terraform.State) error {
						return testAccUserGroupNotHasMemberID(s, "komodo_user_group.test", savedUserID)
					},
				),
			},
			{
				// Add the user back out-of-band — should produce no plan diff (truly unmanaged now)
				Config: testAccUserGroupResourceConfig_withUserResOnly("tf-test-group-unmanage-transition"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccUserGroupAddUserFromState("komodo_user_group.test", "komodo_user.test"),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				// Re-apply the same config — must produce an empty plan (out-of-band user ignored)
				Config:   testAccUserGroupResourceConfig_withUserResOnly("tf-test-group-unmanage-transition"),
				PlanOnly: true,
			},
		},
	})
}

// TestAccUserGroupResource_managedUsersFullControl verifies that when users is
// specified, Terraform enforces the exact list and removes unlisted members.
func TestAccUserGroupResource_managedUsersFullControl(t *testing.T) {
	var savedUserID string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Start with managed user
				Config: testAccUserGroupResourceConfig_withNewUser("tf-test-group-full-control"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user_group.test", "users.#", "1"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["komodo_user.test"]
						if !ok {
							return fmt.Errorf("komodo_user.test not found in state")
						}
						savedUserID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				// Manage with empty list — user must be removed
				Config: testAccUserGroupResourceConfig_emptyUsersWithNewUser("tf-test-group-full-control"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user_group.test", "users.#", "0"),
					func(s *terraform.State) error {
						return testAccUserGroupNotHasMemberID(s, "komodo_user_group.test", savedUserID)
					},
				),
			},
		},
	})
}

// testAccUserGroupAddUserFromState adds the user identified by userResourceName to the group
// identified by groupResourceName, using IDs looked up from Terraform state.
func testAccUserGroupAddUserFromState(groupResourceName, userResourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		groupRS, ok := s.RootModule().Resources[groupResourceName]
		if !ok {
			return fmt.Errorf("resource not found in state: %s", groupResourceName)
		}
		userRS, ok := s.RootModule().Resources[userResourceName]
		if !ok {
			return fmt.Errorf("resource not found in state: %s", userResourceName)
		}
		c := client.NewClient(
			os.Getenv("KOMODO_ENDPOINT"),
			os.Getenv("KOMODO_USERNAME"),
			os.Getenv("KOMODO_PASSWORD"),
		)
		_, err := c.AddUserToUserGroup(context.Background(), client.AddUserToUserGroupRequest{
			UserGroup: groupRS.Primary.ID,
			User:      userRS.Primary.ID,
		})
		return err
	}
}

// testAccUserGroupNotHasMemberID checks directly without wrapping in TestCheckFunc.
func testAccUserGroupNotHasMemberID(s *terraform.State, groupResourceName, userID string) error {
	rs, ok := s.RootModule().Resources[groupResourceName]
	if !ok {
		return fmt.Errorf("resource not found in state: %s", groupResourceName)
	}
	c := client.NewClient(
		os.Getenv("KOMODO_ENDPOINT"),
		os.Getenv("KOMODO_USERNAME"),
		os.Getenv("KOMODO_PASSWORD"),
	)
	group, err := c.GetUserGroup(context.Background(), rs.Primary.ID)
	if err != nil {
		return fmt.Errorf("unable to fetch group: %s", err)
	}
	for _, u := range group.Users {
		if u == userID {
			return fmt.Errorf("expected user %s to NOT be a member of group %s, but was", userID, rs.Primary.ID)
		}
	}
	return nil
}

// Config helpers

func testAccUserGroupResourceConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "komodo_user_group" "test" {
  name = %q
}
`, name)
}

// testAccUserGroupResourceConfig_withNewUser creates a komodo_user and a group that includes it.
func testAccUserGroupResourceConfig_withNewUser(name string) string {
	return fmt.Sprintf(`
resource "komodo_user" "test" {
  username = "tf-test-grp-res-user"
  password = "Password1!"
}

resource "komodo_user_group" "test" {
  name  = %q
  users = [komodo_user.test.id]
}
`, name)
}

// testAccUserGroupResourceConfig_withNewUserOnly creates the same komodo_user but
// without adding it to the group (group has no managed users list).
func testAccUserGroupResourceConfig_withNewUserOnly(name string) string {
	return fmt.Sprintf(`
resource "komodo_user" "test" {
  username = "tf-test-grp-res-user"
  password = "Password1!"
}

resource "komodo_user_group" "test" {
  name = %q
}
`, name)
}

// testAccUserGroupResourceConfig_withUserResOnly creates the user and the group
// (without the user in the group's managed list) — for out-of-band membership tests.
func testAccUserGroupResourceConfig_withUserResOnly(name string) string {
	return testAccUserGroupResourceConfig_withNewUserOnly(name)
}

// testAccUserGroupResourceConfig_emptyUsersWithNewUser creates the user and a group
// with an explicit empty users list.
func testAccUserGroupResourceConfig_emptyUsersWithNewUser(name string) string {
	return fmt.Sprintf(`
resource "komodo_user" "test" {
  username = "tf-test-grp-res-user"
  password = "Password1!"
}

resource "komodo_user_group" "test" {
  name  = %q
  users = []
}
`, name)
}

func testAccUserGroupResourceConfig_everyoneEnabled(name string, enabled bool) string {
	return fmt.Sprintf(`
resource "komodo_user_group" "test" {
  name             = %q
  everyone_enabled = %t
}
`, name, enabled)
}

func testAccUserGroupResourceConfig_everyoneAndUsers(name string) string {
	return fmt.Sprintf(`
resource "komodo_user_group" "test" {
  name     = %q
  everyone_enabled = true
  users    = ["some-user-id"]
}
`, name)
}

// TestAccUserGroupResource_everyoneEnabledDefault verifies that omitting
// everyone_enabled results in false in state (not unknown after apply).
func TestAccUserGroupResource_everyoneEnabledDefault(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupResourceConfig_basic("tf-test-group-everyone-default"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user_group.test", "everyone_enabled", "false"),
				),
			},
		},
	})
}

// TestAccUserGroupResource_everyoneEnabledUpdate verifies that toggling
// everyone_enabled is applied and reflected correctly in state.
func TestAccUserGroupResource_everyoneEnabledUpdate(t *testing.T) {
	const name = "tf-test-group-everyone-update"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupResourceConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user_group.test", "everyone_enabled", "false"),
				),
			},
			{
				Config: testAccUserGroupResourceConfig_everyoneEnabled(name, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user_group.test", "everyone_enabled", "true"),
				),
			},
			{
				Config: testAccUserGroupResourceConfig_everyoneEnabled(name, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user_group.test", "everyone_enabled", "false"),
				),
			},
		},
	})
}

// TestAccUserGroupResource_everyoneEnabledDrift verifies that an external change
// to everyone_enabled is detected as drift (non-empty plan).
func TestAccUserGroupResource_everyoneEnabledDrift(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupResourceConfig_basic("tf-test-group-everyone-drift"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_user_group.test", "everyone_enabled", "false"),
					// Simulate external change: set everyone_enabled = true out-of-band
					testAccUserGroupSetEveryoneEnabled("komodo_user_group.test", true),
				),
				// After the external change the plan must show a diff
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccUserGroupSetEveryoneEnabled(resourceName string, enabled bool) resource.TestCheckFunc {
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
		_, err := c.SetEveryoneUserGroup(context.Background(), client.SetEveryoneUserGroupRequest{
			UserGroup: rs.Primary.ID,
			Everyone:  enabled,
		})
		return err
	}
}

func TestAccUserGroupResource_disappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupResourceConfig_basic("tf-disappear-group"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_user_group.test", "id"),
					testAccUserGroupDisappears("komodo_user_group.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccUserGroupDisappears(resourceName string) resource.TestCheckFunc {
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
		return c.DeleteUserGroup(context.Background(), rs.Primary.ID)
	}
}

// ─── Unit tests ──────────────────────────────────────────────────────────────

func wrongRawUserGroupPlan(t *testing.T, r *UserGroupResource) tfsdk.Plan {
	t.Helper()
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schemaResp.Schema,
	}
}

func wrongRawUserGroupState(t *testing.T, r *UserGroupResource) tfsdk.State {
	t.Helper()
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	return tfsdk.State{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schemaResp.Schema,
	}
}

func TestUnitUserGroupResource_configure(t *testing.T) {
	t.Run("wrong_type", func(t *testing.T) {
		r := &UserGroupResource{}
		req := fwresource.ConfigureRequest{ProviderData: "not-a-client"}
		resp := &fwresource.ConfigureResponse{}
		r.Configure(context.Background(), req, resp)
		if !resp.Diagnostics.HasError() {
			t.Fatal("expected diagnostic error for wrong ProviderData type")
		}
	})
}

func TestUnitUserGroupResource_createPlanGetError(t *testing.T) {
	r := &UserGroupResource{client: &client.Client{}}
	req := fwresource.CreateRequest{Plan: wrongRawUserGroupPlan(t, r)}
	resp := &fwresource.CreateResponse{}
	r.Create(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed plan")
	}
}

func TestUnitUserGroupResource_readStateGetError(t *testing.T) {
	r := &UserGroupResource{client: &client.Client{}}
	req := fwresource.ReadRequest{State: wrongRawUserGroupState(t, r)}
	resp := &fwresource.ReadResponse{}
	r.Read(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed state")
	}
}

func TestUnitUserGroupResource_updatePlanGetError(t *testing.T) {
	r := &UserGroupResource{client: &client.Client{}}
	req := fwresource.UpdateRequest{Plan: wrongRawUserGroupPlan(t, r)}
	resp := &fwresource.UpdateResponse{}
	r.Update(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed plan")
	}
}

func TestUnitUserGroupResource_deleteStateGetError(t *testing.T) {
	r := &UserGroupResource{client: &client.Client{}}
	req := fwresource.DeleteRequest{State: wrongRawUserGroupState(t, r)}
	resp := &fwresource.DeleteResponse{}
	r.Delete(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed state")
	}
}

// ─── Mock server infrastructure ──────────────────────────────────────────────

const mockValidUserGroupJSON = `{
  "_id": {"$oid": "aabbccddeeff001122334455"},
  "name": "tf-mock-group",
  "everyone": false,
  "users": [],
  "all": {},
  "updated_at": 0
}`

type mockUserGroupRoute struct {
	statusCode int
	body       string
}

func newUserGroupMockServerWithDefault(t *testing.T, routes map[string]mockUserGroupRoute, defaultJSON string) *httptest.Server {
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

func newUserGroupMockServer(t *testing.T, routes map[string]mockUserGroupRoute) *httptest.Server {
	return newUserGroupMockServerWithDefault(t, routes, mockValidUserGroupJSON)
}

func newStatefulUserGroupMockServer(t *testing.T, handler func(typ string, callCount int) (int, string)) *httptest.Server {
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

func mockUserGroupProviderConfig(srvURL string) string {
	return fmt.Sprintf(`
provider "komodo" {
  endpoint = %[1]q
  username = "mock-user"
  password = "mock-pass"
}
`, srvURL)
}

const mockUserGroupResourceConfig = `
resource "komodo_user_group" "test" {
  name = "tf-mock-group"
}
`

const mockUserGroupResourceConfigEveryone = `
resource "komodo_user_group" "test" {
  name             = "tf-mock-group"
  everyone_enabled = true
}
`

const mockUserGroupResourceConfigUpdated = `
resource "komodo_user_group" "test" {
  name = "tf-mock-group-renamed"
}
`

// ─── Create error path tests ──────────────────────────────────────────────────

func TestAccUserGroupResource_createClientError(t *testing.T) {
	srv := newUserGroupMockServer(t, map[string]mockUserGroupRoute{
		"CreateUserGroup": {http.StatusInternalServerError, `"create error"`},
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockUserGroupProviderConfig(srv.URL) + mockUserGroupResourceConfig,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

func TestAccUserGroupResource_createSetEveryoneError(t *testing.T) {
	srv := newUserGroupMockServer(t, map[string]mockUserGroupRoute{
		"SetEveryoneUserGroup": {http.StatusInternalServerError, `"set everyone error"`},
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockUserGroupProviderConfig(srv.URL) + mockUserGroupResourceConfigEveryone,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

func TestAccUserGroupResource_createFetchError(t *testing.T) {
	// CreateUserGroup succeeds but the final GetUserGroup call fails
	var createDone bool
	srv := newStatefulUserGroupMockServer(t, func(typ string, n int) (int, string) {
		switch typ {
		case "CreateUserGroup":
			createDone = true
			return http.StatusOK, mockValidUserGroupJSON
		case "GetUserGroup":
			if createDone {
				return http.StatusInternalServerError, `"fetch error"`
			}
			return http.StatusOK, mockValidUserGroupJSON
		}
		return http.StatusOK, mockValidUserGroupJSON
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockUserGroupProviderConfig(srv.URL) + mockUserGroupResourceConfig,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

// ─── Read error/disappears tests ─────────────────────────────────────────────

func TestAccUserGroupResource_readDisappears_unit(t *testing.T) {
	// First GetUserGroup (create refresh) succeeds; second (pre-plan read) returns nil → resource removed.
	var createCount int
	srv := newStatefulUserGroupMockServer(t, func(typ string, n int) (int, string) {
		if typ == "GetUserGroup" {
			createCount++
			if createCount > 1 {
				return http.StatusNotFound, `"not found"`
			}
		}
		return http.StatusOK, mockValidUserGroupJSON
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             mockUserGroupProviderConfig(srv.URL) + mockUserGroupResourceConfig,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccUserGroupResource_readNon404Error(t *testing.T) {
	var count int
	srv := newStatefulUserGroupMockServer(t, func(typ string, n int) (int, string) {
		if typ == "GetUserGroup" {
			count++
			if count > 1 {
				return http.StatusInternalServerError, `"server error"`
			}
		}
		return http.StatusOK, mockValidUserGroupJSON
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockUserGroupProviderConfig(srv.URL) + mockUserGroupResourceConfig,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

// ─── Delete error test ────────────────────────────────────────────────────────

func TestAccUserGroupResource_deleteClientError(t *testing.T) {
	srv := newStatefulUserGroupMockServer(t, func(typ string, n int) (int, string) {
		if typ == "DeleteUserGroup" && n == 1 {
			return http.StatusInternalServerError, `"delete error"`
		}
		return http.StatusOK, mockValidUserGroupJSON
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockUserGroupProviderConfig(srv.URL) + mockUserGroupResourceConfig,
			},
			{
				Config:      mockUserGroupProviderConfig(srv.URL),
				Destroy:     true,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

// ─── Update tests ─────────────────────────────────────────────────────────────

func TestAccUserGroupResource_updateRenameError(t *testing.T) {
	srv := newUserGroupMockServer(t, map[string]mockUserGroupRoute{
		"RenameUserGroup": {http.StatusInternalServerError, `"rename error"`},
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockUserGroupProviderConfig(srv.URL) + mockUserGroupResourceConfig,
			},
			{
				Config:      mockUserGroupProviderConfig(srv.URL) + mockUserGroupResourceConfigUpdated,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

func TestAccUserGroupResource_updateSetEveryoneError(t *testing.T) {
	// Create succeeds, then Update tries to change everyone_enabled → fails
	var createDone bool
	srv := newStatefulUserGroupMockServer(t, func(typ string, n int) (int, string) {
		switch typ {
		case "CreateUserGroup":
			createDone = true
			return http.StatusOK, mockValidUserGroupJSON
		case "SetEveryoneUserGroup":
			if createDone {
				return http.StatusInternalServerError, `"set everyone error"`
			}
		}
		return http.StatusOK, mockValidUserGroupJSON
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockUserGroupProviderConfig(srv.URL) + mockUserGroupResourceConfig,
			},
			{
				Config:      mockUserGroupProviderConfig(srv.URL) + mockUserGroupResourceConfigEveryone,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

func TestAccUserGroupResource_updateFetchError(t *testing.T) {
	// Create+Read succeed, then Update's final GetUserGroup fails
	var updateStarted bool
	srv := newStatefulUserGroupMockServer(t, func(typ string, n int) (int, string) {
		switch typ {
		case "RenameUserGroup":
			updateStarted = true
			return http.StatusOK, mockValidUserGroupJSON
		case "GetUserGroup":
			if updateStarted {
				return http.StatusInternalServerError, `"fetch error"`
			}
		}
		return http.StatusOK, mockValidUserGroupJSON
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockUserGroupProviderConfig(srv.URL) + mockUserGroupResourceConfig,
			},
			{
				Config:      mockUserGroupProviderConfig(srv.URL) + mockUserGroupResourceConfigUpdated,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

func TestAccUserGroupResource_updateSuccess(t *testing.T) {
	const updatedJSON = `{
  "_id": {"$oid": "aabbccddeeff001122334455"},
  "name": "tf-mock-group-renamed",
  "everyone": false,
  "users": [],
  "all": {},
  "updated_at": 0
}`
	var renameCount int
	srv := newStatefulUserGroupMockServer(t, func(typ string, n int) (int, string) {
		if typ == "RenameUserGroup" {
			renameCount++
		}
		if renameCount > 0 {
			return http.StatusOK, updatedJSON
		}
		return http.StatusOK, mockValidUserGroupJSON
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockUserGroupProviderConfig(srv.URL) + mockUserGroupResourceConfig,
				Check:  resource.TestCheckResourceAttr("komodo_user_group.test", "name", "tf-mock-group"),
			},
			{
				Config: mockUserGroupProviderConfig(srv.URL) + mockUserGroupResourceConfigUpdated,
				Check:  resource.TestCheckResourceAttr("komodo_user_group.test", "name", "tf-mock-group-renamed"),
			},
		},
	})
}

// ─── ValidateConfig unit test ─────────────────────────────────────────────────

func TestUnitUserGroupResource_validateConfigGetError(t *testing.T) {
	r := &UserGroupResource{}
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
		t.Fatal("expected error for malformed config in ValidateConfig")
	}
}

func TestAccUserGroupResource_validateConfigConflict_unit(t *testing.T) {
	// everyone_enabled=true + users set → conflict error
	srv := newUserGroupMockServer(t, nil)
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockUserGroupProviderConfig(srv.URL) + `
resource "komodo_user_group" "test" {
  name             = "tf-mock-conflict"
  everyone_enabled = true
  users            = ["some-user"]
}
`,
				ExpectError: regexp.MustCompile(`(?i)conflict`),
			},
		},
	})
}

// ─── ImportState direct unit test ────────────────────────────────────────────

func TestUnitUserGroupResource_importStateDirect(t *testing.T) {
	r := &UserGroupResource{}
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	emptyState := tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, nil),
		"name":             tftypes.NewValue(tftypes.String, nil),
		"everyone_enabled": tftypes.NewValue(tftypes.Bool, nil),
		"users":            tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"all":              tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
		"updated_at":       tftypes.NewValue(tftypes.String, nil),
	})
	req := fwresource.ImportStateRequest{ID: "aabbccddeeff001122334455"}
	resp := &fwresource.ImportStateResponse{State: tfsdk.State{Schema: schemaResp.Schema, Raw: emptyState}}
	r.ImportState(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error in ImportState: %v", resp.Diagnostics)
	}
}

// ─── Direct Read unit tests ───────────────────────────────────────────────────

func TestUnitUserGroupResource_readNilGroupDirect(t *testing.T) {
	srv := newUserGroupMockServer(t, map[string]mockUserGroupRoute{
		"GetUserGroup": {http.StatusNotFound, `"not found"`},
	})
	c := client.NewClient(srv.URL, "mock-user", "mock-pass")
	r := &UserGroupResource{client: c}
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	stateVal := tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, "aabbccddeeff001122334455"),
		"name":             tftypes.NewValue(tftypes.String, "tf-mock-group"),
		"everyone_enabled": tftypes.NewValue(tftypes.Bool, false),
		"users":            tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"all":              tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
		"updated_at":       tftypes.NewValue(tftypes.String, "1970-01-01T00:00:00Z"),
	})
	state := tfsdk.State{Schema: schemaResp.Schema, Raw: stateVal}
	req := fwresource.ReadRequest{State: state}
	resp := &fwresource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	r.Read(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("expected no error for nil group (removed from state): %v", resp.Diagnostics)
	}
}

func TestUnitUserGroupResource_readErrorContainsNotFoundDirect(t *testing.T) {
	// Error message contains "not found" → inner if-branch → RemoveResource
	srv := newUserGroupMockServer(t, map[string]mockUserGroupRoute{
		"GetUserGroup": {http.StatusInternalServerError, `"not found error"`},
	})
	c := client.NewClient(srv.URL, "mock-user", "mock-pass")
	r := &UserGroupResource{client: c}
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	stateVal := tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, "aabbccddeeff001122334455"),
		"name":             tftypes.NewValue(tftypes.String, "tf-mock-group"),
		"everyone_enabled": tftypes.NewValue(tftypes.Bool, false),
		"users":            tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"all":              tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
		"updated_at":       tftypes.NewValue(tftypes.String, "1970-01-01T00:00:00Z"),
	})
	state := tfsdk.State{Schema: schemaResp.Schema, Raw: stateVal}
	req := fwresource.ReadRequest{State: state}
	resp := &fwresource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	r.Read(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("expected no error (resource removed from state for not-found error): %v", resp.Diagnostics)
	}
}

func TestUnitUserGroupResource_readNon404ErrorDirect(t *testing.T) {
	// Error message does NOT contain "not found" → AddError branch
	srv := newUserGroupMockServer(t, map[string]mockUserGroupRoute{
		"GetUserGroup": {http.StatusInternalServerError, `"server error"`},
	})
	c := client.NewClient(srv.URL, "mock-user", "mock-pass")
	r := &UserGroupResource{client: c}
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	stateVal := tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, "aabbccddeeff001122334455"),
		"name":             tftypes.NewValue(tftypes.String, "tf-mock-group"),
		"everyone_enabled": tftypes.NewValue(tftypes.Bool, false),
		"users":            tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"all":              tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
		"updated_at":       tftypes.NewValue(tftypes.String, "1970-01-01T00:00:00Z"),
	})
	state := tfsdk.State{Schema: schemaResp.Schema, Raw: stateVal}
	req := fwresource.ReadRequest{State: state}
	resp := &fwresource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	r.Read(ctx, req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for non-404 GetUserGroup error")
	}
}

// ─── Users-managed and non-empty All branch tests ────────────────────────────

// JSON that includes users and a non-empty all map.
const mockValidUserGroupWithUsersJSON = `{
  "_id": {"$oid": "aabbccddeeff001122334455"},
  "name": "tf-mock-group",
  "everyone": false,
  "users": ["bbccddeeff001122334455aa"],
  "all": {"key": "val"},
  "updated_at": 0
}`

// JSON with users but empty all (for plan-consistency in create/update tests).
const mockValidUserGroupWithUsersNoAllJSON = `{
  "_id": {"$oid": "aabbccddeeff001122334455"},
  "name": "tf-mock-group",
  "everyone": false,
  "users": ["bbccddeeff001122334455aa"],
  "all": {},
  "updated_at": 0
}`

// JSON with null users (for nil-users branch coverage).
const mockValidUserGroupNullUsersJSON = `{
  "_id": {"$oid": "aabbccddeeff001122334455"},
  "name": "tf-mock-group",
  "everyone": false,
  "users": null,
  "all": {},
  "updated_at": 0
}`

// JSON after user is removed: empty users, non-empty all still present.
const mockValidUserGroupNoUsersJSON = `{
  "_id": {"$oid": "aabbccddeeff001122334455"},
  "name": "tf-mock-group",
  "everyone": false,
  "users": [],
  "all": {"key": "val"},
  "updated_at": 0
}`

const mockUserGroupResourceConfigWithUsers = `
resource "komodo_user_group" "test" {
  name  = "tf-mock-group"
  users = ["bbccddeeff001122334455aa"]
}
`

const mockUserGroupResourceConfigNoUsers = `
resource "komodo_user_group" "test" {
  name  = "tf-mock-group"
  users = []
}
`

// TestAccUserGroupResource_createAddUserError tests the AddUserToUserGroup error path in Create.
func TestAccUserGroupResource_createAddUserError(t *testing.T) {
	srv := newUserGroupMockServer(t, map[string]mockUserGroupRoute{
		"AddUserToUserGroup": {http.StatusInternalServerError, `"add user error"`},
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockUserGroupProviderConfig(srv.URL) + mockUserGroupResourceConfigWithUsers,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

// TestAccUserGroupResource_createWithUsersSuccess exercises the non-null users and non-empty All branches in Create.
func TestAccUserGroupResource_createWithUsersSuccess(t *testing.T) {
	srv := newUserGroupMockServer(t, map[string]mockUserGroupRoute{
		"GetUserGroup": {http.StatusOK, mockValidUserGroupWithUsersNoAllJSON},
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockUserGroupProviderConfig(srv.URL) + mockUserGroupResourceConfigWithUsers,
				Check:  resource.TestCheckResourceAttr("komodo_user_group.test", "users.#", "1"),
			},
		},
	})
}

// TestUnitUserGroupResource_readWithNonNullUsersAndAll exercises the non-null users + non-empty All branch in Read.
func TestUnitUserGroupResource_readWithNonNullUsersAndAll(t *testing.T) {
	srv := newUserGroupMockServer(t, map[string]mockUserGroupRoute{
		"GetUserGroup": {http.StatusOK, mockValidUserGroupWithUsersJSON},
	})
	c := client.NewClient(srv.URL, "mock-user", "mock-pass")
	r := &UserGroupResource{client: c}
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	// Use non-null users list in state to trigger the `!data.Users.IsNull()` branch
	stateVal := tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, "aabbccddeeff001122334455"),
		"name":             tftypes.NewValue(tftypes.String, "tf-mock-group"),
		"everyone_enabled": tftypes.NewValue(tftypes.Bool, false),
		"users": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
			tftypes.NewValue(tftypes.String, "bbccddeeff001122334455aa"),
		}),
		"all": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{
			"key": tftypes.NewValue(tftypes.String, "val"),
		}),
		"updated_at": tftypes.NewValue(tftypes.String, "1970-01-01T00:00:00Z"),
	})
	state := tfsdk.State{Schema: schemaResp.Schema, Raw: stateVal}
	readReq := fwresource.ReadRequest{State: state}
	readResp := &fwresource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema, Raw: stateVal}}
	r.Read(ctx, readReq, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %v", readResp.Diagnostics)
	}
}

// TestAccUserGroupResource_updateAddUserError tests the AddUserToUserGroup error in Update (Case 1).
func TestAccUserGroupResource_updateAddUserError(t *testing.T) {
	var createDone bool
	srv := newStatefulUserGroupMockServer(t, func(typ string, n int) (int, string) {
		switch typ {
		case "CreateUserGroup":
			createDone = true
			return http.StatusOK, mockValidUserGroupJSON
		case "AddUserToUserGroup":
			if createDone {
				return http.StatusInternalServerError, `"add user error"`
			}
		}
		return http.StatusOK, mockValidUserGroupJSON
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockUserGroupProviderConfig(srv.URL) + mockUserGroupResourceConfig,
			},
			{
				Config:      mockUserGroupProviderConfig(srv.URL) + mockUserGroupResourceConfigWithUsers,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

// TestAccUserGroupResource_updateRemoveUserError tests RemoveUserFromUserGroup error in Update (Case 1 remove path).
func TestAccUserGroupResource_updateRemoveUserError(t *testing.T) {
	var createDone bool
	srv := newStatefulUserGroupMockServer(t, func(typ string, n int) (int, string) {
		switch typ {
		case "CreateUserGroup":
			createDone = true
			return http.StatusOK, mockValidUserGroupWithUsersNoAllJSON
		case "GetUserGroup":
			if createDone {
				return http.StatusOK, mockValidUserGroupWithUsersNoAllJSON
			}
		case "RemoveUserFromUserGroup":
			return http.StatusInternalServerError, `"remove user error"`
		}
		return http.StatusOK, mockValidUserGroupWithUsersNoAllJSON
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockUserGroupProviderConfig(srv.URL) + mockUserGroupResourceConfigWithUsers,
			},
			{
				Config:      mockUserGroupProviderConfig(srv.URL) + mockUserGroupResourceConfigNoUsers,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

// TestAccUserGroupResource_updateUsersToNullSuccess tests Case 2: users transitions from non-null to null in plan.
// This removes previously-managed users and then stops tracking them.
func TestAccUserGroupResource_updateUsersToNullSuccess(t *testing.T) {
	var step2 bool
	srv := newStatefulUserGroupMockServer(t, func(typ string, n int) (int, string) {
		switch typ {
		case "CreateUserGroup":
			return http.StatusOK, mockValidUserGroupWithUsersNoAllJSON
		case "RemoveUserFromUserGroup":
			step2 = true
		case "GetUserGroup":
			if step2 {
				return http.StatusOK, mockValidUserGroupJSON
			}
			return http.StatusOK, mockValidUserGroupWithUsersNoAllJSON
		}
		return http.StatusOK, mockValidUserGroupWithUsersNoAllJSON
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockUserGroupProviderConfig(srv.URL) + mockUserGroupResourceConfigWithUsers,
				Check:  resource.TestCheckResourceAttr("komodo_user_group.test", "users.#", "1"),
			},
			{
				// No users attribute → null in plan → triggers Case 2
				Config: mockUserGroupProviderConfig(srv.URL) + mockUserGroupResourceConfig,
				Check:  resource.TestCheckNoResourceAttr("komodo_user_group.test", "users.0"),
			},
		},
	})
}

// TestAccUserGroupResource_updateUsersSuccessWithAdd tests adding a user in update (Case 1 add path).
func TestAccUserGroupResource_updateUsersSuccessWithAdd(t *testing.T) {
	var addDone bool
	srv := newStatefulUserGroupMockServer(t, func(typ string, n int) (int, string) {
		switch typ {
		case "AddUserToUserGroup":
			addDone = true
		case "GetUserGroup":
			if addDone {
				return http.StatusOK, mockValidUserGroupWithUsersNoAllJSON
			}
		}
		return http.StatusOK, mockValidUserGroupJSON
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockUserGroupProviderConfig(srv.URL) + mockUserGroupResourceConfigNoUsers,
				Check:  resource.TestCheckResourceAttr("komodo_user_group.test", "users.#", "0"),
			},
			{
				Config: mockUserGroupProviderConfig(srv.URL) + mockUserGroupResourceConfigWithUsers,
				Check:  resource.TestCheckResourceAttr("komodo_user_group.test", "users.#", "1"),
			},
		},
	})
}

// TestAccUserGroupResource_updateRemoveUserSuccess tests removing a user in update (Case 1 remove path succeeds).
func TestAccUserGroupResource_updateRemoveUserSuccess(t *testing.T) {
	var removeDone bool
	srv := newStatefulUserGroupMockServer(t, func(typ string, n int) (int, string) {
		switch typ {
		case "RemoveUserFromUserGroup":
			removeDone = true
		case "GetUserGroup":
			if removeDone {
				return http.StatusOK, mockValidUserGroupJSON
			}
			return http.StatusOK, mockValidUserGroupWithUsersNoAllJSON
		}
		return http.StatusOK, mockValidUserGroupWithUsersNoAllJSON
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockUserGroupProviderConfig(srv.URL) + mockUserGroupResourceConfigWithUsers,
				Check:  resource.TestCheckResourceAttr("komodo_user_group.test", "users.#", "1"),
			},
			{
				Config: mockUserGroupProviderConfig(srv.URL) + mockUserGroupResourceConfigNoUsers,
				Check:  resource.TestCheckResourceAttr("komodo_user_group.test", "users.#", "0"),
			},
		},
	})
}

// TestAccUserGroupDataSource_readSuccessWithAll exercises the non-empty All branch in the data source.
func TestAccUserGroupDataSource_readSuccessWithAll(t *testing.T) {
	srv := newUserGroupMockServer(t, map[string]mockUserGroupRoute{
		"GetUserGroup": {http.StatusOK, mockValidUserGroupWithUsersJSON},
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockUserGroupProviderConfig(srv.URL) + mockUserGroupDataSourceConfig,
				Check:  resource.TestCheckResourceAttr("data.komodo_user_group.test", "users.#", "1"),
			},
		},
	})
}

// TestAccUserGroupsDataSource_readSuccessWithAll exercises non-empty All branch in the list data source.
func TestAccUserGroupsDataSource_readSuccessWithAll(t *testing.T) {
	const groupWithAll = `[` + mockValidUserGroupWithUsersJSON + `]`
	srv := newUserGroupMockServer(t, map[string]mockUserGroupRoute{
		"ListUserGroups": {http.StatusOK, groupWithAll},
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockUserGroupProviderConfig(srv.URL) + mockUserGroupsDataSourceConfig,
				Check:  resource.TestCheckResourceAttr("data.komodo_user_groups.test", "groups.#", "1"),
			},
		},
	})
}

// ─── Additional coverage for uncovered branches ───────────────────────────────

// mockUserGroupResourceConfigWithAll has an explicit all map in config.
const mockUserGroupResourceConfigWithAll = `
resource "komodo_user_group" "test" {
  name = "tf-mock-group"
  all  = {"key": "val"}
}
`

// TestAccUserGroupResource_createWithAllSuccess covers:
// - data.All.ElementsAs in Create (line ~93)
// - non-empty fetched.All in Create (line ~147)
func TestAccUserGroupResource_createWithAllSuccess(t *testing.T) {
	const jsonWithAll = `{"_id":{"$oid":"aabbccddeeff001122334455"},"name":"tf-mock-group","everyone":false,"users":[],"all":{"key":"val"},"updated_at":0}`
	srv := newUserGroupMockServer(t, map[string]mockUserGroupRoute{
		"GetUserGroup": {http.StatusOK, jsonWithAll},
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockUserGroupProviderConfig(srv.URL) + mockUserGroupResourceConfigWithAll,
				Check:  resource.TestCheckResourceAttr("komodo_user_group.test", "all.key", "val"),
			},
		},
	})
}

// TestUnitUserGroupResource_readWithEmptyIDFallback covers the ID=="" → name fallback in Read.
func TestUnitUserGroupResource_readWithEmptyIDFallback(t *testing.T) {
	srv := newUserGroupMockServer(t, map[string]mockUserGroupRoute{
		"GetUserGroup": {http.StatusOK, mockValidUserGroupJSON},
	})
	c := client.NewClient(srv.URL, "mock-user", "mock-pass")
	r := &UserGroupResource{client: c}
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	stateVal := tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, ""),
		"name":             tftypes.NewValue(tftypes.String, "tf-mock-group"),
		"everyone_enabled": tftypes.NewValue(tftypes.Bool, false),
		"users":            tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"all":              tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
		"updated_at":       tftypes.NewValue(tftypes.String, "1970-01-01T00:00:00Z"),
	})
	state := tfsdk.State{Schema: schemaResp.Schema, Raw: stateVal}
	req := fwresource.ReadRequest{State: state}
	resp := &fwresource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema, Raw: stateVal}}
	r.Read(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %v", resp.Diagnostics)
	}
}

// TestAccUserGroupResource_updateWithAllSuccess covers the non-empty All branch in Update's final state set.
func TestAccUserGroupResource_updateWithAllSuccess(t *testing.T) {
	const jsonWithAll = `{"_id":{"$oid":"aabbccddeeff001122334455"},"name":"tf-mock-group-renamed","everyone":false,"users":[],"all":{"key":"val"},"updated_at":0}`
	var renameCount int
	srv := newStatefulUserGroupMockServer(t, func(typ string, n int) (int, string) {
		if typ == "RenameUserGroup" {
			renameCount++
		}
		if renameCount > 0 {
			return http.StatusOK, jsonWithAll
		}
		const jsonNoAll = `{"_id":{"$oid":"aabbccddeeff001122334455"},"name":"tf-mock-group","everyone":false,"users":[],"all":{"key":"val"},"updated_at":0}`
		return http.StatusOK, jsonNoAll
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockUserGroupProviderConfig(srv.URL) + mockUserGroupResourceConfigWithAll,
				Check:  resource.TestCheckResourceAttr("komodo_user_group.test", "all.key", "val"),
			},
			{
				Config: mockUserGroupProviderConfig(srv.URL) + `
resource "komodo_user_group" "test" {
  name = "tf-mock-group-renamed"
  all  = {"key": "val"}
}
`,
				Check: resource.TestCheckResourceAttr("komodo_user_group.test", "all.key", "val"),
			},
		},
	})
}

// TestUnitUserGroupResource_updateStateGetErrorDirect covers the req.State.Get HasError branch in Update.
func TestUnitUserGroupResource_updateStateGetErrorDirect(t *testing.T) {
	r := &UserGroupResource{}
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	// Valid plan value
	validVal := tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, "aabbccddeeff001122334455"),
		"name":             tftypes.NewValue(tftypes.String, "tf-mock-group"),
		"everyone_enabled": tftypes.NewValue(tftypes.Bool, false),
		"users":            tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"all":              tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
		"updated_at":       tftypes.NewValue(tftypes.String, "1970-01-01T00:00:00Z"),
	})
	// Invalid state (wrong raw type)
	invalidStateVal := tftypes.NewValue(tftypes.String, "invalid")
	req := fwresource.UpdateRequest{
		Plan:  tfsdk.Plan{Schema: schemaResp.Schema, Raw: validVal},
		State: tfsdk.State{Schema: schemaResp.Schema, Raw: invalidStateVal},
	}
	resp := &fwresource.UpdateResponse{State: tfsdk.State{Schema: schemaResp.Schema, Raw: validVal}}
	r.Update(ctx, req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error from malformed state in Update")
	}
}

// TestUnitUserGroupResource_updateAddUserDirectDirect directly tests the Update path
// where a new user is added (covers line 251: AddUserToUserGroup call body).
func TestUnitUserGroupResource_updateAddUserDirect(t *testing.T) {
	srv := newUserGroupMockServer(t, map[string]mockUserGroupRoute{
		"GetUserGroup": {http.StatusOK, mockValidUserGroupWithUsersNoAllJSON},
	})
	c := client.NewClient(srv.URL, "mock-user", "mock-pass")
	r := &UserGroupResource{client: c}
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	schemaType := schemaResp.Schema.Type().TerraformType(ctx)

	// state: users = [] (non-null empty list)
	stateVal := tftypes.NewValue(schemaType, map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, "aabbccddeeff001122334455"),
		"name":             tftypes.NewValue(tftypes.String, "tf-mock-group"),
		"everyone_enabled": tftypes.NewValue(tftypes.Bool, false),
		"users":            tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{}),
		"all":              tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
		"updated_at":       tftypes.NewValue(tftypes.String, "1970-01-01T00:00:00Z"),
	})
	// plan: users = ["bbccddeeff001122334455aa"]
	planVal := tftypes.NewValue(schemaType, map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, "aabbccddeeff001122334455"),
		"name":             tftypes.NewValue(tftypes.String, "tf-mock-group"),
		"everyone_enabled": tftypes.NewValue(tftypes.Bool, false),
		"users": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
			tftypes.NewValue(tftypes.String, "bbccddeeff001122334455aa"),
		}),
		"all":        tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
		"updated_at": tftypes.NewValue(tftypes.String, "1970-01-01T00:00:00Z"),
	})
	req := fwresource.UpdateRequest{
		Plan:  tfsdk.Plan{Schema: schemaResp.Schema, Raw: planVal},
		State: tfsdk.State{Schema: schemaResp.Schema, Raw: stateVal},
	}
	resp := &fwresource.UpdateResponse{State: tfsdk.State{Schema: schemaResp.Schema, Raw: stateVal}}
	r.Update(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %v", resp.Diagnostics)
	}
}

// TestUnitUserGroupResource_updateRemoveUserDirect directly tests the Update path
// where a user is removed (covers line 287: RemoveUserFromUserGroup call body).
func TestUnitUserGroupResource_updateRemoveUserDirect(t *testing.T) {
	srv := newUserGroupMockServer(t, map[string]mockUserGroupRoute{
		"GetUserGroup": {http.StatusOK, mockValidUserGroupJSON},
	})
	c := client.NewClient(srv.URL, "mock-user", "mock-pass")
	r := &UserGroupResource{client: c}
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	schemaType := schemaResp.Schema.Type().TerraformType(ctx)

	// state: users = ["bbccddeeff001122334455aa"]
	stateVal := tftypes.NewValue(schemaType, map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, "aabbccddeeff001122334455"),
		"name":             tftypes.NewValue(tftypes.String, "tf-mock-group"),
		"everyone_enabled": tftypes.NewValue(tftypes.Bool, false),
		"users": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
			tftypes.NewValue(tftypes.String, "bbccddeeff001122334455aa"),
		}),
		"all":        tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
		"updated_at": tftypes.NewValue(tftypes.String, "1970-01-01T00:00:00Z"),
	})
	// plan: users = [] (remove the user)
	planVal := tftypes.NewValue(schemaType, map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, "aabbccddeeff001122334455"),
		"name":             tftypes.NewValue(tftypes.String, "tf-mock-group"),
		"everyone_enabled": tftypes.NewValue(tftypes.Bool, false),
		"users":            tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{}),
		"all":              tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
		"updated_at":       tftypes.NewValue(tftypes.String, "1970-01-01T00:00:00Z"),
	})
	req := fwresource.UpdateRequest{
		Plan:  tfsdk.Plan{Schema: schemaResp.Schema, Raw: planVal},
		State: tfsdk.State{Schema: schemaResp.Schema, Raw: stateVal},
	}
	resp := &fwresource.UpdateResponse{State: tfsdk.State{Schema: schemaResp.Schema, Raw: stateVal}}
	r.Update(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %v", resp.Diagnostics)
	}
}

// TestUnitUserGroupResource_updateCase2UsersToNullDirect tests Case 2 in Update
// (users non-null in state, null in plan → remove all previously-managed users).
func TestUnitUserGroupResource_updateCase2UsersToNullDirect(t *testing.T) {
	srv := newUserGroupMockServer(t, map[string]mockUserGroupRoute{
		"GetUserGroup": {http.StatusOK, mockValidUserGroupJSON},
	})
	c := client.NewClient(srv.URL, "mock-user", "mock-pass")
	r := &UserGroupResource{client: c}
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	schemaType := schemaResp.Schema.Type().TerraformType(ctx)

	// state: users = ["bbccddeeff001122334455aa"]
	stateVal := tftypes.NewValue(schemaType, map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, "aabbccddeeff001122334455"),
		"name":             tftypes.NewValue(tftypes.String, "tf-mock-group"),
		"everyone_enabled": tftypes.NewValue(tftypes.Bool, false),
		"users": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
			tftypes.NewValue(tftypes.String, "bbccddeeff001122334455aa"),
		}),
		"all":        tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
		"updated_at": tftypes.NewValue(tftypes.String, "1970-01-01T00:00:00Z"),
	})
	// plan: users = null (unmanaged)
	planVal := tftypes.NewValue(schemaType, map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, "aabbccddeeff001122334455"),
		"name":             tftypes.NewValue(tftypes.String, "tf-mock-group"),
		"everyone_enabled": tftypes.NewValue(tftypes.Bool, false),
		"users":            tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"all":              tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
		"updated_at":       tftypes.NewValue(tftypes.String, "1970-01-01T00:00:00Z"),
	})
	req := fwresource.UpdateRequest{
		Plan:  tfsdk.Plan{Schema: schemaResp.Schema, Raw: planVal},
		State: tfsdk.State{Schema: schemaResp.Schema, Raw: stateVal},
	}
	resp := &fwresource.UpdateResponse{State: tfsdk.State{Schema: schemaResp.Schema, Raw: stateVal}}
	r.Update(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %v", resp.Diagnostics)
	}
}

// TestUnitUserGroupResource_updateCase2RemoveUserErrorDirect tests Case 2 where
// RemoveUserFromUserGroup fails, covering lines 292-295.
func TestUnitUserGroupResource_updateCase2RemoveUserErrorDirect(t *testing.T) {
	srv := newUserGroupMockServer(t, map[string]mockUserGroupRoute{
		"RemoveUserFromUserGroup": {http.StatusInternalServerError, `"remove error"`},
	})
	c := client.NewClient(srv.URL, "mock-user", "mock-pass")
	r := &UserGroupResource{client: c}
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	schemaType := schemaResp.Schema.Type().TerraformType(ctx)

	stateVal := tftypes.NewValue(schemaType, map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, "aabbccddeeff001122334455"),
		"name":             tftypes.NewValue(tftypes.String, "tf-mock-group"),
		"everyone_enabled": tftypes.NewValue(tftypes.Bool, false),
		"users": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
			tftypes.NewValue(tftypes.String, "bbccddeeff001122334455aa"),
		}),
		"all":        tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
		"updated_at": tftypes.NewValue(tftypes.String, "1970-01-01T00:00:00Z"),
	})
	// plan: users = null → Case 2
	planVal := tftypes.NewValue(schemaType, map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, "aabbccddeeff001122334455"),
		"name":             tftypes.NewValue(tftypes.String, "tf-mock-group"),
		"everyone_enabled": tftypes.NewValue(tftypes.Bool, false),
		"users":            tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"all":              tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
		"updated_at":       tftypes.NewValue(tftypes.String, "1970-01-01T00:00:00Z"),
	})
	req := fwresource.UpdateRequest{
		Plan:  tfsdk.Plan{Schema: schemaResp.Schema, Raw: planVal},
		State: tfsdk.State{Schema: schemaResp.Schema, Raw: stateVal},
	}
	resp := &fwresource.UpdateResponse{State: tfsdk.State{Schema: schemaResp.Schema, Raw: stateVal}}
	r.Update(ctx, req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error from RemoveUserFromUserGroup failure")
	}
}
