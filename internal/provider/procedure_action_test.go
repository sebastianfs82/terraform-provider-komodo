// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/action"
)

// ─── ProcedureAction tests ────────────────────────────────────────────────────

func TestUnitProcedureAction_metadata(t *testing.T) {
	a := NewProcedureAction()
	req := action.MetadataRequest{ProviderTypeName: "komodo"}
	resp := &action.MetadataResponse{}
	a.Metadata(context.Background(), req, resp)
	if resp.TypeName != "komodo_procedure" {
		t.Fatalf("expected TypeName komodo_procedure, got %q", resp.TypeName)
	}
}

func TestUnitProcedureAction_schema(t *testing.T) {
	a := &ProcedureAction{}
	resp := &action.SchemaResponse{}
	a.Schema(context.Background(), action.SchemaRequest{}, resp)
	if resp.Schema.MarkdownDescription == "" {
		t.Fatal("expected non-empty schema MarkdownDescription")
	}
}

func TestUnitProcedureAction_configure(t *testing.T) {
	ctx := context.Background()

	t.Run("nil_noop", func(t *testing.T) {
		a := &ProcedureAction{}
		cfgResp := &action.ConfigureResponse{}
		a.Configure(ctx, action.ConfigureRequest{ProviderData: nil}, cfgResp)
		if cfgResp.Diagnostics.HasError() {
			t.Fatal("expected no error for nil ProviderData")
		}
	})

	t.Run("wrong_type", func(t *testing.T) {
		a := &ProcedureAction{}
		cfgResp := &action.ConfigureResponse{}
		a.Configure(ctx, action.ConfigureRequest{ProviderData: true}, cfgResp)
		if !cfgResp.Diagnostics.HasError() {
			t.Fatal("expected error for wrong ProviderData type")
		}
	})

	t.Run("valid_client", func(t *testing.T) {
		a := &ProcedureAction{}
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

func TestUnitProcedureAction_invoke(t *testing.T) {
	ctx := context.Background()

	getSchema := func() action.SchemaResponse {
		a := &ProcedureAction{}
		resp := &action.SchemaResponse{}
		a.Schema(ctx, action.SchemaRequest{}, resp)
		return *resp
	}

	t.Run("run_success", func(t *testing.T) {
		c := newActionSuccessMockServer(t)
		a := &ProcedureAction{client: c}
		schResp := getSchema()
		invokeResp := &action.InvokeResponse{}
		a.Invoke(ctx, buildSimpleActionInvokeReq(ctx, schResp.Schema, "proc-id", "run"), invokeResp)
		if invokeResp.Diagnostics.HasError() {
			t.Fatalf("unexpected error: %s", invokeResp.Diagnostics)
		}
	})

	t.Run("run_client_error", func(t *testing.T) {
		c := newActionErrorMockServer(t)
		a := &ProcedureAction{client: c}
		schResp := getSchema()
		invokeResp := &action.InvokeResponse{}
		a.Invoke(ctx, buildSimpleActionInvokeReq(ctx, schResp.Schema, "proc-id", "run"), invokeResp)
		if !invokeResp.Diagnostics.HasError() {
			t.Fatal("expected error from client failure")
		}
	})

	t.Run("invalid_action", func(t *testing.T) {
		c := newActionSuccessMockServer(t)
		a := &ProcedureAction{client: c}
		schResp := getSchema()
		invokeResp := &action.InvokeResponse{}
		a.Invoke(ctx, buildSimpleActionInvokeReq(ctx, schResp.Schema, "proc-id", "bogus"), invokeResp)
		if !invokeResp.Diagnostics.HasError() {
			t.Fatal("expected error for invalid action")
		}
	})

	t.Run("config_error", func(t *testing.T) {
		c := newActionSuccessMockServer(t)
		a := &ProcedureAction{client: c}
		schResp := getSchema()
		invokeResp := &action.InvokeResponse{}
		a.Invoke(ctx, buildActionConfigErrorReq(ctx, schResp.Schema), invokeResp)
		if !invokeResp.Diagnostics.HasError() {
			t.Fatal("expected error for invalid config")
		}
	})
}
