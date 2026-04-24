// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/action"
)

// ─── BuildAction tests ────────────────────────────────────────────────────────

func TestUnitBuildAction_metadata(t *testing.T) {
	a := NewBuildAction()
	req := action.MetadataRequest{ProviderTypeName: "komodo"}
	resp := &action.MetadataResponse{}
	a.Metadata(context.Background(), req, resp)
	if resp.TypeName != "komodo_build" {
		t.Fatalf("expected TypeName komodo_build, got %q", resp.TypeName)
	}
}

func TestUnitBuildAction_schema(t *testing.T) {
	a := &BuildAction{}
	resp := &action.SchemaResponse{}
	a.Schema(context.Background(), action.SchemaRequest{}, resp)
	if resp.Schema.MarkdownDescription == "" {
		t.Fatal("expected non-empty schema MarkdownDescription")
	}
}

func TestUnitBuildAction_configure(t *testing.T) {
	ctx := context.Background()

	t.Run("nil_noop", func(t *testing.T) {
		a := &BuildAction{}
		cfgResp := &action.ConfigureResponse{}
		a.Configure(ctx, action.ConfigureRequest{ProviderData: nil}, cfgResp)
		if cfgResp.Diagnostics.HasError() {
			t.Fatal("expected no error for nil ProviderData")
		}
	})

	t.Run("wrong_type", func(t *testing.T) {
		a := &BuildAction{}
		cfgResp := &action.ConfigureResponse{}
		a.Configure(ctx, action.ConfigureRequest{ProviderData: 42}, cfgResp)
		if !cfgResp.Diagnostics.HasError() {
			t.Fatal("expected error for wrong ProviderData type")
		}
	})

	t.Run("valid_client", func(t *testing.T) {
		a := &BuildAction{}
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

func TestUnitBuildAction_invoke(t *testing.T) {
	ctx := context.Background()

	getSchema := func() action.SchemaResponse {
		a := &BuildAction{}
		resp := &action.SchemaResponse{}
		a.Schema(ctx, action.SchemaRequest{}, resp)
		return *resp
	}

	t.Run("run_success", func(t *testing.T) {
		c := newActionSuccessMockServer(t)
		a := &BuildAction{client: c}
		schResp := getSchema()
		invokeResp := &action.InvokeResponse{}
		a.Invoke(ctx, buildSimpleActionInvokeReq(ctx, schResp.Schema, "build-id", "run"), invokeResp)
		if invokeResp.Diagnostics.HasError() {
			t.Fatalf("unexpected error: %s", invokeResp.Diagnostics)
		}
	})

	t.Run("run_client_error", func(t *testing.T) {
		c := newActionErrorMockServer(t)
		a := &BuildAction{client: c}
		schResp := getSchema()
		invokeResp := &action.InvokeResponse{}
		a.Invoke(ctx, buildSimpleActionInvokeReq(ctx, schResp.Schema, "build-id", "run"), invokeResp)
		if !invokeResp.Diagnostics.HasError() {
			t.Fatal("expected error from client failure")
		}
	})

	t.Run("invalid_action", func(t *testing.T) {
		c := newActionSuccessMockServer(t)
		a := &BuildAction{client: c}
		schResp := getSchema()
		invokeResp := &action.InvokeResponse{}
		a.Invoke(ctx, buildSimpleActionInvokeReq(ctx, schResp.Schema, "build-id", "bogus"), invokeResp)
		if !invokeResp.Diagnostics.HasError() {
			t.Fatal("expected error for invalid action")
		}
	})

	t.Run("config_error", func(t *testing.T) {
		c := newActionSuccessMockServer(t)
		a := &BuildAction{client: c}
		schResp := getSchema()
		invokeResp := &action.InvokeResponse{}
		a.Invoke(ctx, buildActionConfigErrorReq(ctx, schResp.Schema), invokeResp)
		if !invokeResp.Diagnostics.HasError() {
			t.Fatal("expected error for invalid config")
		}
	})
}
