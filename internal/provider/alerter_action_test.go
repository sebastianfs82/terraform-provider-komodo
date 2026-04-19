// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/action"
)

// ─── AlerterAction tests ──────────────────────────────────────────────────────

func TestUnitAlerterAction_metadata(t *testing.T) {
	a := NewAlerterAction()
	req := action.MetadataRequest{ProviderTypeName: "komodo"}
	resp := &action.MetadataResponse{}
	a.Metadata(context.Background(), req, resp)
	if resp.TypeName != "komodo_alerter" {
		t.Fatalf("expected TypeName komodo_alerter, got %q", resp.TypeName)
	}
}

func TestUnitAlerterAction_schema(t *testing.T) {
	a := &AlerterAction{}
	resp := &action.SchemaResponse{}
	a.Schema(context.Background(), action.SchemaRequest{}, resp)
	if resp.Schema.MarkdownDescription == "" {
		t.Fatal("expected non-empty schema MarkdownDescription")
	}
}

func TestUnitAlerterAction_configure(t *testing.T) {
	ctx := context.Background()

	t.Run("nil_noop", func(t *testing.T) {
		a := &AlerterAction{}
		cfgResp := &action.ConfigureResponse{}
		a.Configure(ctx, action.ConfigureRequest{ProviderData: nil}, cfgResp)
		if cfgResp.Diagnostics.HasError() {
			t.Fatal("expected no error for nil ProviderData")
		}
	})

	t.Run("wrong_type", func(t *testing.T) {
		a := &AlerterAction{}
		cfgResp := &action.ConfigureResponse{}
		a.Configure(ctx, action.ConfigureRequest{ProviderData: "not-a-client"}, cfgResp)
		if !cfgResp.Diagnostics.HasError() {
			t.Fatal("expected error for wrong ProviderData type")
		}
	})

	t.Run("valid_client", func(t *testing.T) {
		a := &AlerterAction{}
		_, c := newActionSuccessMockServer(t)
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

func TestUnitAlerterAction_invoke(t *testing.T) {
	ctx := context.Background()

	getSchema := func() action.SchemaResponse {
		a := &AlerterAction{}
		resp := &action.SchemaResponse{}
		a.Schema(ctx, action.SchemaRequest{}, resp)
		return *resp
	}

	t.Run("test_success", func(t *testing.T) {
		_, c := newActionSuccessMockServer(t)
		a := &AlerterAction{client: c}
		schResp := getSchema()
		invokeResp := &action.InvokeResponse{}
		a.Invoke(ctx, buildSimpleActionInvokeReq(ctx, schResp.Schema, "alerter-id", "test"), invokeResp)
		if invokeResp.Diagnostics.HasError() {
			t.Fatalf("unexpected error: %s", invokeResp.Diagnostics)
		}
	})

	t.Run("test_client_error", func(t *testing.T) {
		_, c := newActionErrorMockServer(t)
		a := &AlerterAction{client: c}
		schResp := getSchema()
		invokeResp := &action.InvokeResponse{}
		a.Invoke(ctx, buildSimpleActionInvokeReq(ctx, schResp.Schema, "alerter-id", "test"), invokeResp)
		if !invokeResp.Diagnostics.HasError() {
			t.Fatal("expected error from client failure")
		}
	})

	t.Run("invalid_action", func(t *testing.T) {
		_, c := newActionSuccessMockServer(t)
		a := &AlerterAction{client: c}
		schResp := getSchema()
		invokeResp := &action.InvokeResponse{}
		a.Invoke(ctx, buildSimpleActionInvokeReq(ctx, schResp.Schema, "alerter-id", "bogus"), invokeResp)
		if !invokeResp.Diagnostics.HasError() {
			t.Fatal("expected error for invalid action")
		}
	})

	t.Run("config_error", func(t *testing.T) {
		_, c := newActionSuccessMockServer(t)
		a := &AlerterAction{client: c}
		schResp := getSchema()
		invokeResp := &action.InvokeResponse{}
		a.Invoke(ctx, buildActionConfigErrorReq(ctx, schResp.Schema), invokeResp)
		if !invokeResp.Diagnostics.HasError() {
			t.Fatal("expected error for invalid config")
		}
	})
}
