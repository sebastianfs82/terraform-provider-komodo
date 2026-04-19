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
	"strings"
	"sync"
	"testing"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"
)

func TestAccBuilderResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBuilderResourceUrlConfig("tf-acc-builder-basic", "http://localhost:8120"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_builder.test", "name", "tf-acc-builder-basic"),
					resource.TestCheckResourceAttr("komodo_builder.test", "type", "Url"),
					resource.TestCheckResourceAttr("komodo_builder.test", "url_config.address", "http://localhost:8120"),
					resource.TestCheckResourceAttrSet("komodo_builder.test", "id"),
				),
			},
		},
	})
}

func TestAccBuilderResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBuilderResourceUrlConfig("tf-acc-builder-update", "http://localhost:8120"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_builder.test", "url_config.address", "http://localhost:8120"),
				),
			},
			{
				Config: testAccBuilderResourceUrlConfig("tf-acc-builder-update", "http://localhost:9000"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_builder.test", "url_config.address", "http://localhost:9000"),
				),
			},
		},
	})
}

func TestAccBuilderResource_import(t *testing.T) {
	var builderID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBuilderResourceUrlConfig("tf-acc-builder-import", "http://localhost:8120"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_builder.test", "id"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["komodo_builder.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						builderID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config:            testAccBuilderResourceUrlConfig("tf-acc-builder-import", "http://localhost:8120"),
				ResourceName:      "komodo_builder.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(_ *terraform.State) (string, error) { return builderID, nil },
			},
		},
	})
}

func TestAccBuilderResource_disappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBuilderResourceUrlConfig("tf-acc-builder-disappears", "http://localhost:8120"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_builder.test", "id"),
					testAccBuilderDisappears("komodo_builder.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccBuilderResource_rename(t *testing.T) {
	var savedID string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBuilderResourceUrlConfig("tf-acc-builder-rename-orig", "http://localhost:8120"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_builder.test", "name", "tf-acc-builder-rename-orig"),
					resource.TestCheckResourceAttrSet("komodo_builder.test", "id"),
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["komodo_builder.test"]
						savedID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config: testAccBuilderResourceUrlConfig("tf-acc-builder-rename-new", "http://localhost:8120"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_builder.test", "name", "tf-acc-builder-rename-new"),
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["komodo_builder.test"]
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

func testAccBuilderDisappears(resourceName string) resource.TestCheckFunc {
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
		return c.DeleteBuilder(context.Background(), rs.Primary.ID)
	}
}

func testAccBuilderResourceUrlConfig(name, address string) string {
	return fmt.Sprintf(`
resource "komodo_builder" "test" {
  name = %q
  type = "Url"
  url_config {
    address = %q
  }
}
`, name, address)
}

func TestAccBuilderResource_tags(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBuilderWithTagConfig("tf-acc-builder-tags"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_builder.test", "tags.#", "1"),
					resource.TestCheckResourceAttrPair("komodo_builder.test", "tags.0", "komodo_tag.test", "id"),
				),
			},
			{
				Config: testAccBuilderClearTagsConfig("tf-acc-builder-tags"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_builder.test", "tags.#", "0"),
				),
			},
		},
	})
}

func testAccBuilderWithTagConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_tag" "test" {
  name  = "tf-acc-tag-builder"
  color = "Green"
}

resource "komodo_builder" "test" {
  name = %q
  type = "Url"
  url_config {
    address = "http://localhost:8120"
  }
  tags = [komodo_tag.test.id]
}
`, name)
}

func testAccBuilderClearTagsConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_builder" "test" {
  name = %q
  type = "Url"
  url_config {
    address = "http://localhost:8120"
  }
  tags = []
}
`, name)
}

// ---------------------------------------------------------------------------
// Server-type builder acceptance tests
// ---------------------------------------------------------------------------

func TestAccBuilderResource_server(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create a Server-type builder; exercises builderConfigInputFromModel "Server" branch
				// and builderToModel "Server" branch.
				Config: testAccBuilderResourceServerConfig("tf-acc-builder-server-a"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_builder.test", "name", "tf-acc-builder-server-a"),
					resource.TestCheckResourceAttr("komodo_builder.test", "type", "Server"),
					resource.TestCheckResourceAttrSet("komodo_builder.test", "server_config.server_id"),
				),
			},
			{
				// Rename the builder — exercises the Update rename + builderConfigInputFromModel "Server" path.
				Config: testAccBuilderResourceServerConfig("tf-acc-builder-server-b"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_builder.test", "name", "tf-acc-builder-server-b"),
					resource.TestCheckResourceAttr("komodo_builder.test", "type", "Server"),
				),
			},
		},
	})
}

func testAccBuilderResourceServerConfig(name string) string {
	return fmt.Sprintf(`
data "komodo_servers" "main" {}

resource "komodo_builder" "test" {
  name = %q
  type = "Server"
  server_config {
    server_id = data.komodo_servers.main.servers[0].id
  }
}
`, name)
}

// ---------------------------------------------------------------------------
// Aws-type builder acceptance tests
// ---------------------------------------------------------------------------

func TestAccBuilderResource_aws(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create an Aws-type builder; exercises builderConfigInputFromModel "Aws" branch
				// and builderToModel "Aws" branch (config is stored by Komodo without spawning EC2).
				Config: testAccBuilderResourceAwsConfig("tf-acc-builder-aws"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_builder.test", "name", "tf-acc-builder-aws"),
					resource.TestCheckResourceAttr("komodo_builder.test", "type", "Aws"),
					resource.TestCheckResourceAttr("komodo_builder.test", "aws_config.region", "us-east-1"),
					resource.TestCheckResourceAttr("komodo_builder.test", "aws_config.instance_type", "t3.micro"),
				),
			},
		},
	})
}

func testAccBuilderResourceAwsConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_builder" "test" {
  name = %q
  type = "Aws"
  aws_config {
    region        = "us-east-1"
    instance_type = "t3.micro"
  }
}
`, name)
}

// ---------------------------------------------------------------------------
// Missing-config error acceptance tests
// Each exercises the nil-config error branch in builderConfigInputFromModel
// and the resulting "Config Error" diagnostic in Create.
// ---------------------------------------------------------------------------

func TestAccBuilderResource_urlMissingConfig(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "komodo_builder" "test" {
  name = "tf-acc-builder-url-missing"
  type = "Url"
}`,
				ExpectError: regexp.MustCompile("(?i)config error"),
			},
		},
	})
}

func TestAccBuilderResource_serverMissingConfig(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "komodo_builder" "test" {
  name = "tf-acc-builder-server-missing"
  type = "Server"
}`,
				ExpectError: regexp.MustCompile("(?i)config error"),
			},
		},
	})
}

func TestAccBuilderResource_awsMissingConfig(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "komodo_builder" "test" {
  name = "tf-acc-builder-aws-missing"
  type = "Aws"
}`,
				ExpectError: regexp.MustCompile("(?i)config error"),
			},
		},
	})
}

func TestAccBuilderResource_unknownType(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Exercises the default branch in builderConfigInputFromModel.
				Config: `
resource "komodo_builder" "test" {
  name = "tf-acc-builder-unknown-type"
  type = "Invalid"
}`,
				ExpectError: regexp.MustCompile("(?i)config error"),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// Unit tests – builderConfigInputFromModel (no API required)
// ---------------------------------------------------------------------------

func TestUnitBuilderResource_configInputFromModel(t *testing.T) {
	ctx := context.Background()

	t.Run("url_nil_config_error", func(t *testing.T) {
		data := &BuilderResourceModel{BuilderType: types.StringValue("Url")}
		_, err := builderConfigInputFromModel(ctx, data)
		if err == nil || !strings.Contains(err.Error(), "url_config is required") {
			t.Fatalf("expected url_config error, got: %v", err)
		}
	})

	t.Run("server_nil_config_error", func(t *testing.T) {
		data := &BuilderResourceModel{BuilderType: types.StringValue("Server")}
		_, err := builderConfigInputFromModel(ctx, data)
		if err == nil || !strings.Contains(err.Error(), "server_config is required") {
			t.Fatalf("expected server_config error, got: %v", err)
		}
	})

	t.Run("aws_nil_config_error", func(t *testing.T) {
		data := &BuilderResourceModel{BuilderType: types.StringValue("Aws")}
		_, err := builderConfigInputFromModel(ctx, data)
		if err == nil || !strings.Contains(err.Error(), "aws_config is required") {
			t.Fatalf("expected aws_config error, got: %v", err)
		}
	})

	t.Run("unknown_type_error", func(t *testing.T) {
		data := &BuilderResourceModel{BuilderType: types.StringValue("Bogus")}
		_, err := builderConfigInputFromModel(ctx, data)
		if err == nil || !strings.Contains(err.Error(), "unknown type") {
			t.Fatalf("expected unknown type error, got: %v", err)
		}
	})

	t.Run("server_valid", func(t *testing.T) {
		data := &BuilderResourceModel{
			BuilderType: types.StringValue("Server"),
			ServerConfig: &ServerConfigModel{
				ServerID: types.StringValue("server-id-123"),
			},
		}
		cfg, err := builderConfigInputFromModel(ctx, data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Type != "Server" {
			t.Fatalf("expected type Server, got %q", cfg.Type)
		}
		sc, ok := cfg.Params.(client.ServerBuilderConfig)
		if !ok || sc.ServerID != "server-id-123" {
			t.Fatalf("unexpected ServerBuilderConfig: %+v", cfg.Params)
		}
	})

	t.Run("aws_valid_with_lists", func(t *testing.T) {
		sgIDs, _ := types.ListValueFrom(ctx, types.StringType, []string{"sg-abc"})
		secrets, _ := types.ListValueFrom(ctx, types.StringType, []string{"my-secret"})
		data := &BuilderResourceModel{
			BuilderType: types.StringValue("Aws"),
			AwsConfig: &AwsConfigModel{
				Region:             types.StringValue("eu-west-1"),
				InstanceType:       types.StringValue("t3.small"),
				VolumeGb:           types.Int64Value(50),
				AmiID:              types.StringValue("ami-12345678"),
				SubnetID:           types.StringValue("subnet-abc"),
				KeyPairName:        types.StringValue("my-key"),
				AssignPublicIP:     types.BoolValue(true),
				UsePublicIP:        types.BoolValue(false),
				SecurityGroupIDs:   sgIDs,
				UserData:           types.StringValue("#!/bin/bash\necho hello"),
				Port:               types.Int64Value(8120),
				UseHttps:           types.BoolValue(false),
				PeripheryPublicKey: types.StringValue(""),
				InsecureTls:        types.BoolValue(false),
				Secrets:            secrets,
			},
		}
		cfg, err := builderConfigInputFromModel(ctx, data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Type != "Aws" {
			t.Fatalf("expected type Aws, got %q", cfg.Type)
		}
		ac, ok := cfg.Params.(client.AwsBuilderConfig)
		if !ok || ac.Region != "eu-west-1" {
			t.Fatalf("unexpected AwsBuilderConfig: %+v", cfg.Params)
		}
		if len(ac.SecurityGroupIDs) != 1 || ac.SecurityGroupIDs[0] != "sg-abc" {
			t.Fatalf("unexpected SecurityGroupIDs: %v", ac.SecurityGroupIDs)
		}
		if len(ac.Secrets) != 1 || ac.Secrets[0] != "my-secret" {
			t.Fatalf("unexpected Secrets: %v", ac.Secrets)
		}
	})
}

// ---------------------------------------------------------------------------
// Unit tests – builderToModel (no API required)
// ---------------------------------------------------------------------------

func TestUnitBuilderResource_builderToModel(t *testing.T) {
	ctx := context.Background()

	t.Run("server_type", func(t *testing.T) {
		serverParams, _ := json.Marshal(client.ServerBuilderConfig{ServerID: "server-id-456"})
		b := &client.Builder{
			ID:   client.OID{OID: "builder-id-1"},
			Name: "test-server-builder",
			Tags: []string{},
			Config: client.BuilderConfig{
				Type:   "Server",
				Params: json.RawMessage(serverParams),
			},
		}
		var data BuilderResourceModel
		diags := builderToModel(ctx, b, &data)
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		if data.BuilderType.ValueString() != "Server" {
			t.Fatalf("expected type Server, got %q", data.BuilderType.ValueString())
		}
		if data.ServerConfig == nil {
			t.Fatal("expected ServerConfig to be set")
		}
		if data.ServerConfig.ServerID.ValueString() != "server-id-456" {
			t.Fatalf("expected ServerID = server-id-456, got %q", data.ServerConfig.ServerID.ValueString())
		}
		if data.UrlConfig != nil {
			t.Fatal("expected UrlConfig to be nil for Server type")
		}
		if data.AwsConfig != nil {
			t.Fatal("expected AwsConfig to be nil for Server type")
		}
	})

	t.Run("aws_type", func(t *testing.T) {
		awsParams, _ := json.Marshal(client.AwsBuilderConfig{
			Region:           "ap-southeast-1",
			InstanceType:     "t3.medium",
			SecurityGroupIDs: []string{"sg-xyz"},
			Secrets:          []string{"db-password"},
		})
		b := &client.Builder{
			ID:   client.OID{OID: "builder-id-2"},
			Name: "test-aws-builder",
			Tags: []string{},
			Config: client.BuilderConfig{
				Type:   "Aws",
				Params: json.RawMessage(awsParams),
			},
		}
		var data BuilderResourceModel
		diags := builderToModel(ctx, b, &data)
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		if data.BuilderType.ValueString() != "Aws" {
			t.Fatalf("expected type Aws, got %q", data.BuilderType.ValueString())
		}
		if data.AwsConfig == nil {
			t.Fatal("expected AwsConfig to be set")
		}
		if data.AwsConfig.Region.ValueString() != "ap-southeast-1" {
			t.Fatalf("expected Region = ap-southeast-1, got %q", data.AwsConfig.Region.ValueString())
		}
		var sgIDs []string
		if d := data.AwsConfig.SecurityGroupIDs.ElementsAs(ctx, &sgIDs, false); d.HasError() {
			t.Fatalf("unexpected diags converting SecurityGroupIDs: %v", d)
		}
		if len(sgIDs) != 1 || sgIDs[0] != "sg-xyz" {
			t.Fatalf("unexpected SecurityGroupIDs: %v", sgIDs)
		}
		var secrets []string
		if d := data.AwsConfig.Secrets.ElementsAs(ctx, &secrets, false); d.HasError() {
			t.Fatalf("unexpected diags converting Secrets: %v", d)
		}
		if len(secrets) != 1 || secrets[0] != "db-password" {
			t.Fatalf("unexpected Secrets: %v", secrets)
		}
		if data.UrlConfig != nil {
			t.Fatal("expected UrlConfig to be nil for Aws type")
		}
		if data.ServerConfig != nil {
			t.Fatal("expected ServerConfig to be nil for Aws type")
		}
	})

	t.Run("null_tags_converted_to_empty", func(t *testing.T) {
		serverParams, _ := json.Marshal(client.ServerBuilderConfig{ServerID: "s-1"})
		b := &client.Builder{
			ID:   client.OID{OID: "id-nil-tags"},
			Name: "nil-tags-builder",
			Tags: nil, // nil tags should become an empty list
			Config: client.BuilderConfig{
				Type:   "Server",
				Params: json.RawMessage(serverParams),
			},
		}
		var data BuilderResourceModel
		diags := builderToModel(ctx, b, &data)
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		if data.Tags.IsNull() || data.Tags.IsUnknown() {
			t.Fatal("expected Tags to be an empty list, not null/unknown")
		}
		if len(data.Tags.Elements()) != 0 {
			t.Fatalf("expected 0 tags, got %d", len(data.Tags.Elements()))
		}
	})

	t.Run("server_bad_json_error", func(t *testing.T) {
		b := &client.Builder{
			ID:   client.OID{OID: "bad-server"},
			Name: "bad",
			Tags: []string{},
			Config: client.BuilderConfig{
				Type:   "Server",
				Params: json.RawMessage(`not-valid-json`),
			},
		}
		var data BuilderResourceModel
		diags := builderToModel(ctx, b, &data)
		if !diags.HasError() {
			t.Fatal("expected a parse error for bad Server JSON, got none")
		}
	})

	t.Run("aws_bad_json_error", func(t *testing.T) {
		b := &client.Builder{
			ID:   client.OID{OID: "bad-aws"},
			Name: "bad",
			Tags: []string{},
			Config: client.BuilderConfig{
				Type:   "Aws",
				Params: json.RawMessage(`not-valid-json`),
			},
		}
		var data BuilderResourceModel
		diags := builderToModel(ctx, b, &data)
		if !diags.HasError() {
			t.Fatal("expected a parse error for bad Aws JSON, got none")
		}
	})

	t.Run("url_bad_json_error", func(t *testing.T) {
		b := &client.Builder{
			ID:   client.OID{OID: "bad-url"},
			Name: "bad",
			Tags: []string{},
			Config: client.BuilderConfig{
				Type:   "Url",
				Params: json.RawMessage(`not-valid-json`),
			},
		}
		var data BuilderResourceModel
		diags := builderToModel(ctx, b, &data)
		if !diags.HasError() {
			t.Fatal("expected a parse error for bad Url JSON, got none")
		}
	})

	t.Run("unknown_type_error", func(t *testing.T) {
		b := &client.Builder{
			ID:   client.OID{OID: "unknown"},
			Name: "unknown",
			Tags: []string{},
			Config: client.BuilderConfig{
				Type:   "Unknown",
				Params: json.RawMessage(`{}`),
			},
		}
		var data BuilderResourceModel
		diags := builderToModel(ctx, b, &data)
		if !diags.HasError() {
			t.Fatal("expected an error for unknown builder type, got none")
		}
	})
}

// ---------------------------------------------------------------------------
// Unit tests – Configure method (no API required)
// ---------------------------------------------------------------------------

func TestUnitBuilderResource_configure(t *testing.T) {
	t.Run("wrong_provider_data_type_adds_error", func(t *testing.T) {
		r := &BuilderResource{}
		req := fwresource.ConfigureRequest{ProviderData: "not-a-client"}
		resp := &fwresource.ConfigureResponse{}
		r.Configure(context.Background(), req, resp)
		if !resp.Diagnostics.HasError() {
			t.Fatal("expected diagnostic error for wrong ProviderData type")
		}
	})

	t.Run("nil_provider_data_is_noop", func(t *testing.T) {
		r := &BuilderResource{}
		req := fwresource.ConfigureRequest{ProviderData: nil}
		resp := &fwresource.ConfigureResponse{}
		r.Configure(context.Background(), req, resp)
		if resp.Diagnostics.HasError() {
			t.Fatalf("expected no error for nil ProviderData, got: %v", resp.Diagnostics)
		}
	})
}

// ---------------------------------------------------------------------------
// Mock-server acceptance tests – CRUD client error paths
// ---------------------------------------------------------------------------

// mockRoute holds the status code and response body for a single mock API route.
type mockRoute struct {
	statusCode int
	body       string
}

const (
	// mockValidBuilderJSON is a valid Komodo builder JSON response used by typed mock servers.
	mockValidBuilderJSON = `{"_id":{"$oid":"abc123"},"name":"tf-mock-builder","tags":[],"config":{"type":"Url","params":{"address":"http://localhost:8120","periphery_public_key":"","insecure_tls":false,"passkey":""}}}`
	// mockBadConfigBuilderJSON is a builder response where config.params is an invalid JSON type.
	mockBadConfigBuilderJSON = `{"_id":{"$oid":"abc123"},"name":"tf-mock-builder","tags":[],"config":{"type":"Url","params":"invalid"}}`
	// mockEmptyOIDBuilderJSON is a builder response with an empty ObjectId.
	mockEmptyOIDBuilderJSON = `{"_id":{"$oid":""},"name":"tf-mock-builder","tags":[],"config":{"type":"Url","params":{"address":"http://localhost:8120","periphery_public_key":"","insecure_tls":false,"passkey":""}}}`
)

// newTypedMockServer creates an httptest server that routes requests by the "type" field in the
// JSON body. Login (/auth/login/LoginLocalUser) always returns a mock JWT. GetVersion always
// returns "2.0.0". Any type present in routes returns the given statusCode+body. All other types
// return mockValidBuilderJSON with 200 OK.
func newTypedMockServer(t *testing.T, routes map[string]mockRoute) *httptest.Server {
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
		default:
			_, _ = w.Write([]byte(mockValidBuilderJSON))
		}
	}))
}

// newStatefulMockServer creates an httptest server where Login and GetVersion are handled
// automatically and all other requests are dispatched to handle(w, requestType).
func newStatefulMockServer(t *testing.T, handle func(w http.ResponseWriter, reqType string)) *httptest.Server {
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
		if req.Type == "GetVersion" {
			_, _ = w.Write([]byte(`{"version":"2.0.0"}`))
			return
		}
		handle(w, req.Type)
	}))
}

// newMockBuilderServer creates an httptest server that returns the given status+body for all
// requests after the initial login. The first call is always handled as authentication.
func newMockBuilderServer(t *testing.T, statusCode int, body string) *httptest.Server {
	t.Helper()
	call := 0
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call++
		w.Header().Set("Content-Type", "application/json")
		if call == 1 && r.URL.Path == "/auth/login/LoginLocalUser" {
			_, _ = w.Write([]byte(`{"type":"Jwt","data":{"jwt":"mock-token"}}`))
			return
		}
		w.WriteHeader(statusCode)
		_, _ = w.Write([]byte(body))
	}))
}

func TestAccBuilderResource_createClientError(t *testing.T) {
	srv := newMockBuilderServer(t, http.StatusInternalServerError, `"internal server error"`)
	defer srv.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "komodo" {
  endpoint = %q
  username = "mock"
  password = "mock"
}
resource "komodo_builder" "test" {
  name = "tf-mock-builder"
  type = "Url"
  url_config { address = "http://localhost:8120" }
}`, srv.URL),
				ExpectError: regexp.MustCompile("(?i)error"),
			},
		},
	})
}

// mockBuilderProviderConfig formats a provider block pointing at a mock server URL.
func mockBuilderProviderConfig(srvURL string) string {
	return fmt.Sprintf(`
provider "komodo" {
  endpoint = %q
  username = "mock"
  password = "mock"
}`, srvURL)
}

const mockBuilderResourceConfig = `
resource "komodo_builder" "test" {
  name = "tf-mock-builder"
  type = "Url"
  url_config { address = "http://localhost:8120" }
}`

// TestAccBuilderResource_createApiError covers the Create "client error" branch
// (CreateBuilder API returns a non-2xx response).
func TestAccBuilderResource_createApiError(t *testing.T) {
	srv := newTypedMockServer(t, map[string]mockRoute{
		"CreateBuilder": {http.StatusInternalServerError, `"create failed"`},
	})
	defer srv.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockBuilderProviderConfig(srv.URL) + mockBuilderResourceConfig,
				ExpectError: regexp.MustCompile("(?i)error"),
			},
		},
	})
}

// TestAccBuilderResource_createEmptyID covers the Create "missing ID" branch
// (CreateBuilder succeeds but the returned builder has an empty ObjectId).
func TestAccBuilderResource_createEmptyID(t *testing.T) {
	srv := newTypedMockServer(t, map[string]mockRoute{
		"CreateBuilder": {http.StatusOK, mockEmptyOIDBuilderJSON},
	})
	defer srv.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockBuilderProviderConfig(srv.URL) + mockBuilderResourceConfig,
				ExpectError: regexp.MustCompile("(?i)missing id"),
			},
		},
	})
}

// TestAccBuilderResource_createBuilderToModelError covers the Create
// "builderToModel returns error" branch (CreateBuilder returns a builder
// whose config.params cannot be decoded).
func TestAccBuilderResource_createBuilderToModelError(t *testing.T) {
	srv := newTypedMockServer(t, map[string]mockRoute{
		"CreateBuilder": {http.StatusOK, mockBadConfigBuilderJSON},
	})
	defer srv.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockBuilderProviderConfig(srv.URL) + mockBuilderResourceConfig,
				ExpectError: regexp.MustCompile("(?i)parse|decode"),
			},
		},
	})
}

// TestAccBuilderResource_createUpdateMetaError covers the Create
// "UpdateResourceMeta error" branch (tags are set but the meta-update call fails).
func TestAccBuilderResource_createUpdateMetaError(t *testing.T) {
	srv := newTypedMockServer(t, map[string]mockRoute{
		"UpdateResourceMeta": {http.StatusInternalServerError, `"meta error"`},
	})
	defer srv.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockBuilderProviderConfig(srv.URL) + `
resource "komodo_builder" "test" {
  name = "tf-mock-builder"
  type = "Url"
  url_config { address = "http://localhost:8120" }
  tags = ["tag1"]
}`,
				ExpectError: regexp.MustCompile("(?i)tags"),
			},
		},
	})
}

// TestAccBuilderResource_updateRenameError covers the Update "RenameBuilder error" branch.
func TestAccBuilderResource_updateRenameError(t *testing.T) {
	srv := newTypedMockServer(t, map[string]mockRoute{
		"RenameBuilder": {http.StatusInternalServerError, `"rename failed"`},
	})
	defer srv.Close()

	provCfg := mockBuilderProviderConfig(srv.URL)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: provCfg + mockBuilderResourceConfig,
			},
			{
				Config: provCfg + `
resource "komodo_builder" "test" {
  name = "tf-mock-builder-renamed"
  type = "Url"
  url_config { address = "http://localhost:8120" }
}`,
				ExpectError: regexp.MustCompile("(?i)rename"),
			},
		},
	})
}

// TestAccBuilderResource_updateBuilderError covers the Update "UpdateBuilder error" branch.
func TestAccBuilderResource_updateBuilderError(t *testing.T) {
	srv := newTypedMockServer(t, map[string]mockRoute{
		"UpdateBuilder": {http.StatusInternalServerError, `"update failed"`},
	})
	defer srv.Close()

	provCfg := mockBuilderProviderConfig(srv.URL)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: provCfg + mockBuilderResourceConfig,
			},
			{
				Config: provCfg + `
resource "komodo_builder" "test" {
  name = "tf-mock-builder"
  type = "Url"
  url_config { address = "http://localhost:9000" }
}`,
				ExpectError: regexp.MustCompile("(?i)update"),
			},
		},
	})
}

// TestAccBuilderResource_updateBuilderToModelError covers the Update
// "builderToModel returns error" branch (UpdateBuilder succeeds but returns
// a builder whose config.params cannot be decoded).
func TestAccBuilderResource_updateBuilderToModelError(t *testing.T) {
	srv := newTypedMockServer(t, map[string]mockRoute{
		"UpdateBuilder": {http.StatusOK, mockBadConfigBuilderJSON},
	})
	defer srv.Close()

	provCfg := mockBuilderProviderConfig(srv.URL)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: provCfg + mockBuilderResourceConfig,
			},
			{
				Config: provCfg + `
resource "komodo_builder" "test" {
  name = "tf-mock-builder"
  type = "Url"
  url_config { address = "http://localhost:9000" }
}`,
				ExpectError: regexp.MustCompile("(?i)parse|decode"),
			},
		},
	})
}

// TestAccBuilderResource_updateMetaError covers the Update "UpdateResourceMeta error"
// branch (tags change triggers a meta update that fails).
func TestAccBuilderResource_updateMetaError(t *testing.T) {
	srv := newTypedMockServer(t, map[string]mockRoute{
		"UpdateResourceMeta": {http.StatusInternalServerError, `"meta error"`},
	})
	defer srv.Close()

	provCfg := mockBuilderProviderConfig(srv.URL)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: provCfg + mockBuilderResourceConfig,
			},
			{
				Config: provCfg + `
resource "komodo_builder" "test" {
  name = "tf-mock-builder"
  type = "Url"
  url_config { address = "http://localhost:8120" }
  tags = ["tag1"]
}`,
				ExpectError: regexp.MustCompile("(?i)tags"),
			},
		},
	})
}

// TestAccBuilderResource_deleteClientError covers the Delete "client error" branch.
// Changing the type from Url to Server triggers RequiresReplace, so Terraform destroys
// the existing builder before creating a new one. The first DeleteBuilder call returns
// an error; subsequent calls (cleanup) succeed.
func TestAccBuilderResource_deleteClientError(t *testing.T) {
	var mu sync.Mutex
	deleteCount := 0

	srv := newStatefulMockServer(t, func(w http.ResponseWriter, reqType string) {
		switch reqType {
		case "DeleteBuilder":
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
			_, _ = w.Write([]byte(mockValidBuilderJSON))
		}
	})
	defer srv.Close()

	provCfg := mockBuilderProviderConfig(srv.URL)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: provCfg + mockBuilderResourceConfig,
			},
			{
				// Changing type triggers RequiresReplace → destroy old → error.
				Config: provCfg + `
resource "komodo_builder" "test" {
  name = "tf-mock-builder"
  type = "Server"
  server_config {}
}`,
				ExpectError: regexp.MustCompile("(?i)delete"),
			},
		},
	})
}

// TestAccBuilderResource_readClientError covers the Read "GetBuilder returns error" branch.
// Step 1 creates the builder (first GetBuilder call returns valid data). Step 2 refreshes
// state, which triggers a second GetBuilder call that now returns a 500 error.
func TestAccBuilderResource_readClientError(t *testing.T) {
	var mu sync.Mutex
	getBuilderCount := 0

	srv := newStatefulMockServer(t, func(w http.ResponseWriter, reqType string) {
		switch reqType {
		case "GetBuilder":
			mu.Lock()
			getBuilderCount++
			n := getBuilderCount
			mu.Unlock()
			if n > 1 {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`"get failed"`))
			} else {
				_, _ = w.Write([]byte(mockValidBuilderJSON))
			}
		default:
			_, _ = w.Write([]byte(mockValidBuilderJSON))
		}
	})
	defer srv.Close()

	provCfg := mockBuilderProviderConfig(srv.URL)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: provCfg + mockBuilderResourceConfig,
			},
			{
				RefreshState: true,
				ExpectError:  regexp.MustCompile("(?i)error"),
			},
		},
	})
}

// TestAccBuilderResource_updateConfigError covers the Update
// "builderConfigInputFromModel returns error" branch.
// Step 2 keeps type=Url but removes url_config, so builderConfigInputFromModel fails.
func TestAccBuilderResource_updateConfigError(t *testing.T) {
	srv := newTypedMockServer(t, nil)
	defer srv.Close()

	provCfg := mockBuilderProviderConfig(srv.URL)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: provCfg + mockBuilderResourceConfig,
			},
			{
				Config: provCfg + `
resource "komodo_builder" "test" {
  name = "tf-mock-builder"
  type = "Url"
}`,
				ExpectError: regexp.MustCompile("(?i)config error"),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// Unit tests – CRUD req.Plan/State.Get error branches (no API required)
// These cover the defensive `if resp.Diagnostics.HasError() { return }` guards
// that fire when the framework cannot decode a malformed plan or state.
// ---------------------------------------------------------------------------

// wrongRawPlan returns a tfsdk.Plan whose Raw value has a String type instead of
// the Object type the schema expects, causing Plan.Get to add diagnostics.
func wrongRawPlan(t *testing.T, r *BuilderResource) tfsdk.Plan {
	t.Helper()
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schemaResp.Schema,
	}
}

// wrongRawState returns a tfsdk.State with the same type mismatch.
func wrongRawState(t *testing.T, r *BuilderResource) tfsdk.State {
	t.Helper()
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	return tfsdk.State{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schemaResp.Schema,
	}
}

func TestUnitBuilderResource_createPlanGetError(t *testing.T) {
	r := &BuilderResource{client: &client.Client{}}
	req := fwresource.CreateRequest{Plan: wrongRawPlan(t, r)}
	resp := &fwresource.CreateResponse{}
	r.Create(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed plan")
	}
}

func TestUnitBuilderResource_readStateGetError(t *testing.T) {
	r := &BuilderResource{client: &client.Client{}}
	req := fwresource.ReadRequest{State: wrongRawState(t, r)}
	resp := &fwresource.ReadResponse{}
	r.Read(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed state")
	}
}

func TestUnitBuilderResource_updatePlanGetError(t *testing.T) {
	r := &BuilderResource{client: &client.Client{}}
	req := fwresource.UpdateRequest{Plan: wrongRawPlan(t, r)}
	resp := &fwresource.UpdateResponse{}
	r.Update(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed plan")
	}
}

func TestUnitBuilderResource_updateStateGetError(t *testing.T) {
	r := &BuilderResource{client: &client.Client{}}
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	// Build a valid, non-null plan object where every attribute is null of its own type.
	// This makes Plan.Get succeed (all fields zero) so we reach State.Get.
	schemaType := schemaResp.Schema.Type().TerraformType(ctx)
	objType, ok := schemaType.(tftypes.Object)
	if !ok {
		t.Fatal("schema type is not an object")
	}
	attrVals := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name, attrType := range objType.AttributeTypes {
		attrVals[name] = tftypes.NewValue(attrType, nil)
	}
	validRaw := tftypes.NewValue(schemaType, attrVals)
	req := fwresource.UpdateRequest{
		Plan:  tfsdk.Plan{Raw: validRaw, Schema: schemaResp.Schema},
		State: wrongRawState(t, r),
	}
	resp := &fwresource.UpdateResponse{}
	r.Update(ctx, req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed state")
	}
}

func TestUnitBuilderResource_deleteStateGetError(t *testing.T) {
	r := &BuilderResource{client: &client.Client{}}
	req := fwresource.DeleteRequest{State: wrongRawState(t, r)}
	resp := &fwresource.DeleteResponse{}
	r.Delete(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed state")
	}
}
func TestAccBuilderResource_readBuilderToModelError(t *testing.T) {
	var mu sync.Mutex
	getBuilderCount := 0

	srv := newStatefulMockServer(t, func(w http.ResponseWriter, reqType string) {
		switch reqType {
		case "GetBuilder":
			mu.Lock()
			getBuilderCount++
			n := getBuilderCount
			mu.Unlock()
			if n > 1 {
				_, _ = w.Write([]byte(mockBadConfigBuilderJSON))
			} else {
				_, _ = w.Write([]byte(mockValidBuilderJSON))
			}
		default:
			_, _ = w.Write([]byte(mockValidBuilderJSON))
		}
	})
	defer srv.Close()

	provCfg := mockBuilderProviderConfig(srv.URL)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: provCfg + mockBuilderResourceConfig,
			},
			{
				RefreshState: true,
				ExpectError:  regexp.MustCompile("(?i)parse|decode"),
			},
		},
	})
}
