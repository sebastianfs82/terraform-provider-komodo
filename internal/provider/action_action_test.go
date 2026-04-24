// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/action"
	actionschema "github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"
)

// ─── shared action test helpers ───────────────────────────────────────────────

// newActionSuccessMockServer returns a pre-configured API-key client pointing at a mock HTTP server
// that returns 200 null for every request.
func newActionSuccessMockServer(t *testing.T) *client.Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`null`))
	}))
	t.Cleanup(srv.Close)
	return client.NewClientWithApiKey(srv.URL, "key", "secret")
}

// newActionErrorMockServer returns a pre-configured API-key client pointing at a mock HTTP server
// that returns 500 for every request.
func newActionErrorMockServer(t *testing.T) *client.Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`"action failed"`))
	}))
	t.Cleanup(srv.Close)
	return client.NewClientWithApiKey(srv.URL, "key", "secret")
}

// buildSimpleActionInvokeReq builds an action.InvokeRequest for an action schema
// that has exactly two string attributes: "id" and "action".
func buildSimpleActionInvokeReq(ctx context.Context, schm actionschema.Schema, id, act string) action.InvokeRequest {
	schemaType := schm.Type().TerraformType(ctx)
	raw := tftypes.NewValue(schemaType, map[string]tftypes.Value{
		"id":     tftypes.NewValue(tftypes.String, id),
		"action": tftypes.NewValue(tftypes.String, act),
	})
	return action.InvokeRequest{Config: tfsdk.Config{Schema: schm, Raw: raw}}
}

// buildActionConfigErrorReq builds an InvokeRequest whose raw value is a String
// (wrong type) instead of an Object, forcing req.Config.Get to produce a diagnostic
// error. This exercises the early-return path present in every action Invoke function.
func buildActionConfigErrorReq(_ context.Context, schm actionschema.Schema) action.InvokeRequest {
	raw := tftypes.NewValue(tftypes.String, "invalid-config")
	return action.InvokeRequest{Config: tfsdk.Config{Schema: schm, Raw: raw}}
}

// ─── KomodoActionAction tests ─────────────────────────────────────────────────

func TestUnitActionAction_metadata(t *testing.T) {
	a := NewKomodoActionAction()
	req := action.MetadataRequest{ProviderTypeName: "komodo"}
	resp := &action.MetadataResponse{}
	a.Metadata(context.Background(), req, resp)
	if resp.TypeName != "komodo_action" {
		t.Fatalf("expected TypeName komodo_action, got %q", resp.TypeName)
	}
}

func TestUnitActionAction_schema(t *testing.T) {
	a := &KomodoActionAction{}
	resp := &action.SchemaResponse{}
	a.Schema(context.Background(), action.SchemaRequest{}, resp)
	if resp.Schema.MarkdownDescription == "" {
		t.Fatal("expected non-empty schema MarkdownDescription")
	}
	if _, ok := resp.Schema.Attributes["id"]; !ok {
		t.Fatal("expected 'id' attribute in schema")
	}
	if _, ok := resp.Schema.Attributes["action"]; !ok {
		t.Fatal("expected 'action' attribute in schema")
	}
}

func TestUnitActionAction_configure(t *testing.T) {
	ctx := context.Background()

	t.Run("nil_noop", func(t *testing.T) {
		a := &KomodoActionAction{}
		cfgResp := &action.ConfigureResponse{}
		a.Configure(ctx, action.ConfigureRequest{ProviderData: nil}, cfgResp)
		if cfgResp.Diagnostics.HasError() {
			t.Fatal("expected no error for nil ProviderData")
		}
		if a.client != nil {
			t.Fatal("expected client to remain nil")
		}
	})

	t.Run("wrong_type", func(t *testing.T) {
		a := &KomodoActionAction{}
		cfgResp := &action.ConfigureResponse{}
		a.Configure(ctx, action.ConfigureRequest{ProviderData: "not-a-client"}, cfgResp)
		if !cfgResp.Diagnostics.HasError() {
			t.Fatal("expected error for wrong ProviderData type")
		}
	})

	t.Run("valid_client", func(t *testing.T) {
		a := &KomodoActionAction{}
		c := newActionSuccessMockServer(t)
		cfgResp := &action.ConfigureResponse{}
		a.Configure(ctx, action.ConfigureRequest{ProviderData: c}, cfgResp)
		if cfgResp.Diagnostics.HasError() {
			t.Fatalf("unexpected error: %s", cfgResp.Diagnostics)
		}
		if a.client == nil {
			t.Fatal("expected client to be set")
		}
	})
}

func TestUnitActionAction_invoke(t *testing.T) {
	ctx := context.Background()

	getSchema := func() actionschema.Schema {
		a := &KomodoActionAction{}
		resp := &action.SchemaResponse{}
		a.Schema(ctx, action.SchemaRequest{}, resp)
		return resp.Schema
	}

	t.Run("run_success", func(t *testing.T) {
		c := newActionSuccessMockServer(t)
		a := &KomodoActionAction{client: c}
		schm := getSchema()
		invokeResp := &action.InvokeResponse{}
		a.Invoke(ctx, buildSimpleActionInvokeReq(ctx, schm, "action-id", "run"), invokeResp)
		if invokeResp.Diagnostics.HasError() {
			t.Fatalf("unexpected error: %s", invokeResp.Diagnostics)
		}
	})

	t.Run("run_client_error", func(t *testing.T) {
		c := newActionErrorMockServer(t)
		a := &KomodoActionAction{client: c}
		schm := getSchema()
		invokeResp := &action.InvokeResponse{}
		a.Invoke(ctx, buildSimpleActionInvokeReq(ctx, schm, "action-id", "run"), invokeResp)
		if !invokeResp.Diagnostics.HasError() {
			t.Fatal("expected error from client failure")
		}
	})

	t.Run("invalid_action", func(t *testing.T) {
		c := newActionSuccessMockServer(t)
		a := &KomodoActionAction{client: c}
		schm := getSchema()
		invokeResp := &action.InvokeResponse{}
		a.Invoke(ctx, buildSimpleActionInvokeReq(ctx, schm, "action-id", "invalid"), invokeResp)
		if !invokeResp.Diagnostics.HasError() {
			t.Fatal("expected error for invalid action")
		}
	})

	t.Run("config_error", func(t *testing.T) {
		c := newActionSuccessMockServer(t)
		a := &KomodoActionAction{client: c}
		schm := getSchema()
		invokeResp := &action.InvokeResponse{}
		a.Invoke(ctx, buildActionConfigErrorReq(ctx, schm), invokeResp)
		if !invokeResp.Diagnostics.HasError() {
			t.Fatal("expected error for invalid config")
		}
	})
}
