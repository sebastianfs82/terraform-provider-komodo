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

func TestAccTagResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTagResourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_tag.example", "name"),
					resource.TestCheckResourceAttrSet("komodo_tag.example", "color"),
				),
			},
		},
	})
}

const testAccTagResourceConfig = `
resource "komodo_tag" "example" {
  name  = "tf_tag"
  color = "Green"
}
`

func TestAccTagResource_disappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTagResourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_tag.example", "id"),
					testAccTagDisappears("komodo_tag.example"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccTagResource_updateColor(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTagResourceConfigWithColor("tf-acc-tag-color-update", "Green"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_tag.test", "color", "Green"),
				),
			},
			{
				Config: testAccTagResourceConfigWithColor("tf-acc-tag-color-update", "Blue"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_tag.test", "color", "Blue"),
				),
			},
		},
	})
}

func TestAccTagResource_ownerStableOnUpdate(t *testing.T) {
	var savedOwner string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTagResourceConfigWithColor("tf-acc-tag-owner-stable", "Green"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_tag.test", "owner"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["komodo_tag.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						savedOwner = rs.Primary.Attributes["owner"]
						return nil
					},
				),
			},
			{
				Config: testAccTagResourceConfigWithColor("tf-acc-tag-owner-stable", "Blue"),
				Check: resource.ComposeAggregateTestCheckFunc(
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["komodo_tag.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						if rs.Primary.Attributes["owner"] != savedOwner {
							return fmt.Errorf("owner changed after color update: was %q, got %q", savedOwner, rs.Primary.Attributes["owner"])
						}
						return nil
					},
				),
			},
		},
	})
}

func TestAccTagResource_importState(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTagResourceConfigWithColor("tf-acc-tag-import", "Red"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_tag.test", "id"),
				),
			},
			{
				ResourceName:      "komodo_tag.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccTagResourceConfigWithColor(name, color string) string {
	return fmt.Sprintf(`
resource "komodo_tag" "test" {
  name  = %q
  color = %q
}
`, name, color)
}

func TestAccTagResource_rename(t *testing.T) {
	var savedID string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTagResourceConfigWithColor("tf-acc-tag-rename-orig", "Green"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_tag.test", "name", "tf-acc-tag-rename-orig"),
					resource.TestCheckResourceAttrSet("komodo_tag.test", "id"),
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["komodo_tag.test"]
						savedID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config: testAccTagResourceConfigWithColor("tf-acc-tag-rename-new", "Green"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_tag.test", "name", "tf-acc-tag-rename-new"),
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["komodo_tag.test"]
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

func testAccTagDisappears(resourceName string) resource.TestCheckFunc {
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
		return c.DeleteTag(context.Background(), rs.Primary.Attributes["name"])
	}
}

// ─── Unit tests ──────────────────────────────────────────────────────────────

func wrongRawTagPlan(t *testing.T, r *TagResource) tfsdk.Plan {
	t.Helper()
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schemaResp.Schema,
	}
}

func wrongRawTagState(t *testing.T, r *TagResource) tfsdk.State {
	t.Helper()
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	return tfsdk.State{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schemaResp.Schema,
	}
}

func TestUnitTagResource_configure(t *testing.T) {
	t.Run("wrong_type", func(t *testing.T) {
		r := &TagResource{}
		req := fwresource.ConfigureRequest{ProviderData: "not-a-client"}
		resp := &fwresource.ConfigureResponse{}
		r.Configure(context.Background(), req, resp)
		if !resp.Diagnostics.HasError() {
			t.Fatal("expected diagnostic error for wrong ProviderData type")
		}
	})
}

func TestUnitTagResource_createPlanGetError(t *testing.T) {
	r := &TagResource{client: &client.Client{}}
	req := fwresource.CreateRequest{Plan: wrongRawTagPlan(t, r)}
	resp := &fwresource.CreateResponse{}
	r.Create(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed plan")
	}
}

func TestUnitTagResource_readStateGetError(t *testing.T) {
	r := &TagResource{client: &client.Client{}}
	req := fwresource.ReadRequest{State: wrongRawTagState(t, r)}
	resp := &fwresource.ReadResponse{}
	r.Read(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed state")
	}
}

func TestUnitTagResource_updatePlanGetError(t *testing.T) {
	r := &TagResource{client: &client.Client{}}
	req := fwresource.UpdateRequest{Plan: wrongRawTagPlan(t, r)}
	resp := &fwresource.UpdateResponse{}
	r.Update(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed plan")
	}
}

func TestUnitTagResource_deleteStateGetError(t *testing.T) {
	r := &TagResource{client: &client.Client{}}
	req := fwresource.DeleteRequest{State: wrongRawTagState(t, r)}
	resp := &fwresource.DeleteResponse{}
	r.Delete(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed state")
	}
}

// ─── Mock-server helpers ──────────────────────────────────────────────────────

const (
	mockValidTagJSON    = `{"_id":{"$oid":"abc123"},"name":"tf-mock-tag","color":"Green","owner":"owner-id"}`
	mockTagListJSON     = `[{"_id":{"$oid":"abc123"},"name":"tf-mock-tag","color":"Green","owner":"owner-id"}]`
	mockTagListEmpty    = `[]`
	mockEmptyOIDTagJSON = `{"_id":{"$oid":""},"name":"tf-mock-tag","color":"Green","owner":"owner-id"}`

	// Use a valid 24-char hex OID so the client's DeleteTag takes the "by-ID" path.
	mockTagOID              = "aabbccddeeff001122334455"
	mockValidTagJSONLongOID = `{"_id":{"$oid":"aabbccddeeff001122334455"},"name":"tf-mock-tag","color":"Green","owner":"owner-id"}`
	mockTagListJSONLongOID  = `[{"_id":{"$oid":"aabbccddeeff001122334455"},"name":"tf-mock-tag","color":"Green","owner":"owner-id"}]`
)

// newTypedTagMockServer routes by the "type" field in JSON body. Login and GetVersion
// are handled automatically. Routes in the map return the given status+body;
// all other types return mockValidTagJSON or mockTagListJSON depending on the operation.
func newTypedTagMockServer(t *testing.T, routes map[string]mockTagRoute) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/auth/login/LoginLocalUser" {
			_, _ = w.Write([]byte(`{"type":"Jwt","data":{"jwt":"mock-token"}}`))
			return
		}
		body, _ := io.ReadAll(r.Body)
		var req struct {
			Type string `json:"type"`
		}
		_ = json.Unmarshal(body, &req)
		if route, ok := routes[req.Type]; ok {
			w.WriteHeader(route.statusCode)
			_, _ = w.Write([]byte(route.body))
			return
		}
		switch req.Type {
		case "GetVersion":
			_, _ = w.Write([]byte(`{"version":"2.0.0"}`))
		case "ListTags":
			_, _ = w.Write([]byte(mockTagListJSON))
		default:
			_, _ = w.Write([]byte(mockValidTagJSON))
		}
	}))
}

type mockTagRoute struct {
	statusCode int
	body       string
}

func mockTagProviderConfig(srvURL string) string {
	return fmt.Sprintf(`
provider "komodo" {
  endpoint = %q
  username = "mock"
  password = "mock"
}`, srvURL)
}

const mockTagResourceConfig = `
resource "komodo_tag" "test" {
  name  = "tf-mock-tag"
  color = "Green"
}`

// ─── Create error paths ───────────────────────────────────────────────────────

// TestAccTagResource_createClientError covers Create → client error (non-2xx from CreateTag).
func TestAccTagResource_createClientError(t *testing.T) {
	srv := newTypedTagMockServer(t, map[string]mockTagRoute{
		"CreateTag": {http.StatusInternalServerError, `"create failed"`},
	})
	defer srv.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockTagProviderConfig(srv.URL) + mockTagResourceConfig,
				ExpectError: regexp.MustCompile("(?i)error"),
			},
		},
	})
}

// TestAccTagResource_createAmbiguousError covers Create → "multiple tags found" branch.
func TestAccTagResource_createAmbiguousError(t *testing.T) {
	srv := newTypedTagMockServer(t, map[string]mockTagRoute{
		"CreateTag": {http.StatusInternalServerError, `"multiple tags found matching"`},
	})
	defer srv.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockTagProviderConfig(srv.URL) + mockTagResourceConfig,
				ExpectError: regexp.MustCompile("(?i)error"),
			},
		},
	})
}

// TestAccTagResource_createEmptyID covers Create → empty OID branch.
func TestAccTagResource_createEmptyID(t *testing.T) {
	srv := newTypedTagMockServer(t, map[string]mockTagRoute{
		"CreateTag": {http.StatusOK, mockEmptyOIDTagJSON},
	})
	defer srv.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockTagProviderConfig(srv.URL) + mockTagResourceConfig,
				ExpectError: regexp.MustCompile("(?i)missing id"),
			},
		},
	})
}

// ─── Read error paths ─────────────────────────────────────────────────────────

// TestAccTagResource_readListTagsError covers Read → ListTags failure.
func TestAccTagResource_readListTagsError(t *testing.T) {
	var mu sync.Mutex
	listCount := 0

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/auth/login/LoginLocalUser" {
			_, _ = w.Write([]byte(`{"type":"Jwt","data":{"jwt":"mock-token"}}`))
			return
		}
		body, _ := io.ReadAll(r.Body)
		var reqBody struct {
			Type string `json:"type"`
		}
		_ = json.Unmarshal(body, &reqBody)
		switch reqBody.Type {
		case "GetVersion":
			_, _ = w.Write([]byte(`{"version":"2.0.0"}`))
		case "CreateTag":
			_, _ = w.Write([]byte(mockValidTagJSON))
		case "ListTags":
			mu.Lock()
			listCount++
			n := listCount
			mu.Unlock()
			// Fail only on the second call (first is post-create read,
			// second is the refresh in step 2). All others (cleanup) succeed.
			if n == 2 {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`"list failed"`))
			} else {
				_, _ = w.Write([]byte(mockTagListJSON))
			}
		default:
			_, _ = w.Write([]byte(mockValidTagJSON))
		}
	}))
	defer srv.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockTagProviderConfig(srv.URL) + mockTagResourceConfig,
			},
			{
				RefreshState: true,
				ExpectError:  regexp.MustCompile("(?i)error"),
			},
		},
	})
}

// TestAccTagResource_readTagDisappears covers Read → tag not found by ID → removed from state.
func TestAccTagResource_readTagDisappears(t *testing.T) {
	var mu sync.Mutex
	listCount := 0

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/auth/login/LoginLocalUser" {
			_, _ = w.Write([]byte(`{"type":"Jwt","data":{"jwt":"mock-token"}}`))
			return
		}
		body, _ := io.ReadAll(r.Body)
		var reqBody struct {
			Type string `json:"type"`
		}
		_ = json.Unmarshal(body, &reqBody)
		switch reqBody.Type {
		case "GetVersion":
			_, _ = w.Write([]byte(`{"version":"2.0.0"}`))
		case "CreateTag":
			_, _ = w.Write([]byte(mockValidTagJSON))
		case "ListTags":
			mu.Lock()
			listCount++
			n := listCount
			mu.Unlock()
			if n > 1 {
				// Return empty list → tag not found by ID → state removal
				_, _ = w.Write([]byte(mockTagListEmpty))
			} else {
				_, _ = w.Write([]byte(mockTagListJSON))
			}
		default:
			_, _ = w.Write([]byte(mockValidTagJSON))
		}
	}))
	defer srv.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockTagProviderConfig(srv.URL) + mockTagResourceConfig,
			},
			{
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// ─── Update error paths ───────────────────────────────────────────────────────

// TestAccTagResource_updateClientError covers Update → client error from UpdateTag.
func TestAccTagResource_updateClientError(t *testing.T) {
	srv := newTypedTagMockServer(t, map[string]mockTagRoute{
		"UpdateTagColor": {http.StatusInternalServerError, `"update failed"`},
	})
	defer srv.Close()

	provCfg := mockTagProviderConfig(srv.URL)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: provCfg + mockTagResourceConfig,
			},
			{
				Config: provCfg + `
resource "komodo_tag" "test" {
  name  = "tf-mock-tag"
  color = "Blue"
}`,
				ExpectError: regexp.MustCompile("(?i)error"),
			},
		},
	})
}

// TestAccTagResource_updateAmbiguousError covers Update → "multiple tags found" branch.
func TestAccTagResource_updateAmbiguousError(t *testing.T) {
	srv := newTypedTagMockServer(t, map[string]mockTagRoute{
		"UpdateTagColor": {http.StatusInternalServerError, `"multiple tags found matching"`},
	})
	defer srv.Close()

	provCfg := mockTagProviderConfig(srv.URL)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: provCfg + mockTagResourceConfig,
			},
			{
				Config: provCfg + `
resource "komodo_tag" "test" {
  name  = "tf-mock-tag"
  color = "Blue"
}`,
				ExpectError: regexp.MustCompile("(?i)error"),
			},
		},
	})
}

// ─── Delete error paths ───────────────────────────────────────────────────────

// TestAccTagResource_deleteClientError covers Delete → client error from DeleteTag.
func TestAccTagResource_deleteClientError(t *testing.T) {
	var mu sync.Mutex
	deleteCount := 0

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/auth/login/LoginLocalUser" {
			_, _ = w.Write([]byte(`{"type":"Jwt","data":{"jwt":"mock-token"}}`))
			return
		}
		body, _ := io.ReadAll(r.Body)
		var reqBody struct {
			Type string `json:"type"`
		}
		_ = json.Unmarshal(body, &reqBody)
		switch reqBody.Type {
		case "GetVersion":
			_, _ = w.Write([]byte(`{"version":"2.0.0"}`))
		case "ListTags":
			_, _ = w.Write([]byte(mockTagListJSONLongOID))
		case "DeleteTag":
			mu.Lock()
			deleteCount++
			n := deleteCount
			mu.Unlock()
			if n == 1 {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`"delete failed"`))
			} else {
				_, _ = w.Write([]byte(`null`))
			}
		default:
			_, _ = w.Write([]byte(mockValidTagJSONLongOID))
		}
	}))
	defer srv.Close()

	provCfg := mockTagProviderConfig(srv.URL)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: provCfg + mockTagResourceConfig,
			},
			// Remove resource from config → triggers Delete → first call returns 500.
			{Config: provCfg, ExpectError: regexp.MustCompile("(?i)error")},
		},
	})
}
