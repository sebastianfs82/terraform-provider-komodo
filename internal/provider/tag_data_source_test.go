// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	datasource "github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTagDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTagDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_tag.example", "name"),
					resource.TestCheckResourceAttrSet("data.komodo_tag.example", "color"),
				),
			},
		},
	})
}

const testAccTagDataSourceConfig = `
resource "komodo_tag" "example" {
  name  = "tf_tag_ds"
  color = "Blue"
}

data "komodo_tag" "example" {
  name = komodo_tag.example.name
}
`

func TestAccTagDataSource_fields(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTagDataSourceConfig_fields,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_tag.example", "name", "tf-acc-tag-ds-fields"),
					resource.TestCheckResourceAttr("data.komodo_tag.example", "color", "Purple"),
					resource.TestCheckResourceAttrSet("data.komodo_tag.example", "owner"),
				),
			},
		},
	})
}

func TestAccTagDataSource_byID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTagDataSourceConfig_byID,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_tag.byid", "id"),
					resource.TestCheckResourceAttrSet("data.komodo_tag.byid", "name"),
					resource.TestCheckResourceAttrSet("data.komodo_tag.byid", "color"),
				),
			},
		},
	})
}

const testAccTagDataSourceConfig_fields = `
resource "komodo_tag" "src" {
  name  = "tf-acc-tag-ds-fields"
  color = "Purple"
}

data "komodo_tag" "example" {
  name = komodo_tag.src.name
}
`

const testAccTagDataSourceConfig_byID = `
resource "komodo_tag" "src" {
  name  = "tf-acc-tag-ds-byid"
  color = "Red"
}

data "komodo_tag" "byid" {
  id = komodo_tag.src.id
}
`

func TestUnitTagDataSource_configure(t *testing.T) {
	d := &TagDataSource{}
	resp := &datasource.ConfigureResponse{}
	d.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: "wrong"}, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for wrong provider data type")
	}
}

func TestAccTagDataSource_bothSet_isError(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccTagDataSourceConfig_bothSet(),
				ExpectError: regexp.MustCompile(`Only one of`),
			},
		},
	})
}

func TestAccTagDataSource_neitherSet_isError(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccTagDataSourceConfig_neitherSet(),
				ExpectError: regexp.MustCompile(`One of`),
			},
		},
	})
}

func testAccTagDataSourceConfig_bothSet() string {
	return `
data "komodo_tag" "test" {
  id   = "some-id"
  name = "some-name"
}
`
}

func testAccTagDataSourceConfig_neitherSet() string {
	return `
data "komodo_tag" "test" {}
`
}

// ─── Unit tests – ValidateConfig early-return on HasError ────────────────────

// TestUnitTagDataSource_validateConfigGetError covers the ValidateConfig "HasError" early return.
func TestUnitTagDataSource_validateConfigGetError(t *testing.T) {
	d := &TagDataSource{}
	ctx := context.Background()
	schemaResp := &datasource.SchemaResponse{}
	d.Schema(ctx, datasource.SchemaRequest{}, schemaResp)
	badRaw := tfsdk.Config{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schemaResp.Schema,
	}
	resp := &datasource.ValidateConfigResponse{}
	d.ValidateConfig(ctx, datasource.ValidateConfigRequest{Config: badRaw}, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error from ValidateConfig for malformed config")
	}
}

// TestUnitTagDataSource_readConfigGetError covers the Read "HasError" early return.
func TestUnitTagDataSource_readConfigGetError(t *testing.T) {
	d := &TagDataSource{}
	ctx := context.Background()
	schemaResp := &datasource.SchemaResponse{}
	d.Schema(ctx, datasource.SchemaRequest{}, schemaResp)
	badRaw := tfsdk.Config{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schemaResp.Schema,
	}
	resp := &datasource.ReadResponse{}
	d.Read(ctx, datasource.ReadRequest{Config: badRaw}, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error from Read for malformed config")
	}
}

// ─── Mock-server tests – Read error paths ────────────────────────────────────

// newTagDSMockServer creates an httptest server for data source tests.
// Login and GetVersion are handled automatically; entries in routes override by type.
func newTagDSMockServer(t *testing.T, routes map[string]mockTagRoute) *httptest.Server {
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
		case "CreateTag":
			_, _ = w.Write([]byte(mockValidTagJSON))
		case "ListTags":
			_, _ = w.Write([]byte(mockTagListJSON))
		default:
			_, _ = w.Write([]byte(mockValidTagJSON))
		}
	}))
}

// TestAccTagDataSource_readByIDListError covers Read → ListTags failure when lookup by ID.
func TestAccTagDataSource_readByIDListError(t *testing.T) {
	srv := newTagDSMockServer(t, map[string]mockTagRoute{
		"ListTags": {http.StatusInternalServerError, `"list failed"`},
	})
	defer srv.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockTagProviderConfig(srv.URL) + `
data "komodo_tag" "test" {
  id = "abc123"
}`,
				ExpectError: regexp.MustCompile("(?i)error"),
			},
		},
	})
}

// TestAccTagDataSource_readByIDNotFound covers Read → tag not found by ID.
func TestAccTagDataSource_readByIDNotFound(t *testing.T) {
	srv := newTagDSMockServer(t, map[string]mockTagRoute{
		"ListTags": {http.StatusOK, mockTagListEmpty},
	})
	defer srv.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockTagProviderConfig(srv.URL) + `
data "komodo_tag" "test" {
  id = "abc123"
}`,
				ExpectError: regexp.MustCompile("(?i)not found"),
			},
		},
	})
}

// TestAccTagDataSource_readByNameError covers Read → GetTag error (name lookup fails).
func TestAccTagDataSource_readByNameError(t *testing.T) {
	srv := newTagDSMockServer(t, map[string]mockTagRoute{
		"ListTags": {http.StatusInternalServerError, `"list failed"`},
	})
	defer srv.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockTagProviderConfig(srv.URL) + `
data "komodo_tag" "test" {
  name = "some-tag"
}`,
				ExpectError: regexp.MustCompile("(?i)error"),
			},
		},
	})
}
