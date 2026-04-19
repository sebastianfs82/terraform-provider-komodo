// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/action"
)

// ─── ServerAction tests ───────────────────────────────────────────────────────

func TestUnitServerAction_metadata(t *testing.T) {
	a := NewServerAction()
	req := action.MetadataRequest{ProviderTypeName: "komodo"}
	resp := &action.MetadataResponse{}
	a.Metadata(context.Background(), req, resp)
	if resp.TypeName != "komodo_server" {
		t.Fatalf("expected TypeName komodo_server, got %q", resp.TypeName)
	}
}

func TestUnitServerAction_schema(t *testing.T) {
	a := &ServerAction{}
	resp := &action.SchemaResponse{}
	a.Schema(context.Background(), action.SchemaRequest{}, resp)
	if resp.Schema.MarkdownDescription == "" {
		t.Fatal("expected non-empty schema MarkdownDescription")
	}
}

func TestUnitServerAction_configure(t *testing.T) {
	ctx := context.Background()

	t.Run("nil_noop", func(t *testing.T) {
		a := &ServerAction{}
		cfgResp := &action.ConfigureResponse{}
		a.Configure(ctx, action.ConfigureRequest{ProviderData: nil}, cfgResp)
		if cfgResp.Diagnostics.HasError() {
			t.Fatal("expected no error for nil ProviderData")
		}
	})

	t.Run("wrong_type", func(t *testing.T) {
		a := &ServerAction{}
		cfgResp := &action.ConfigureResponse{}
		a.Configure(ctx, action.ConfigureRequest{ProviderData: []string{"wrong"}}, cfgResp)
		if !cfgResp.Diagnostics.HasError() {
			t.Fatal("expected error for wrong ProviderData type")
		}
	})

	t.Run("valid_client", func(t *testing.T) {
		a := &ServerAction{}
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

func TestUnitServerAction_invoke(t *testing.T) {
	ctx := context.Background()

	getSchema := func() action.SchemaResponse {
		a := &ServerAction{}
		resp := &action.SchemaResponse{}
		a.Schema(ctx, action.SchemaRequest{}, resp)
		return *resp
	}

	allActions := []string{
		"prune_buildx",
		"prune_containers",
		"prune_builders",
		"prune_images",
		"prune_networks",
		"prune_system",
		"prune_volumes",
	}

	for _, act := range allActions {
		t.Run(act+"_success", func(t *testing.T) {
			_, c := newActionSuccessMockServer(t)
			a := &ServerAction{client: c}
			schResp := getSchema()
			invokeResp := &action.InvokeResponse{}
			a.Invoke(ctx, buildSimpleActionInvokeReq(ctx, schResp.Schema, "server-id", act), invokeResp)
			if invokeResp.Diagnostics.HasError() {
				t.Fatalf("unexpected error for action %q: %s", act, invokeResp.Diagnostics)
			}
		})

		t.Run(act+"_client_error", func(t *testing.T) {
			_, c := newActionErrorMockServer(t)
			a := &ServerAction{client: c}
			schResp := getSchema()
			invokeResp := &action.InvokeResponse{}
			a.Invoke(ctx, buildSimpleActionInvokeReq(ctx, schResp.Schema, "server-id", act), invokeResp)
			if !invokeResp.Diagnostics.HasError() {
				t.Fatalf("expected error for action %q on client failure", act)
			}
		})
	}

	t.Run("invalid_action", func(t *testing.T) {
		_, c := newActionSuccessMockServer(t)
		a := &ServerAction{client: c}
		schResp := getSchema()
		invokeResp := &action.InvokeResponse{}
		a.Invoke(ctx, buildSimpleActionInvokeReq(ctx, schResp.Schema, "server-id", "bogus"), invokeResp)
		if !invokeResp.Diagnostics.HasError() {
			t.Fatal("expected error for invalid action")
		}
	})

	t.Run("config_error", func(t *testing.T) {
		_, c := newActionSuccessMockServer(t)
		a := &ServerAction{client: c}
		schResp := getSchema()
		invokeResp := &action.InvokeResponse{}
		a.Invoke(ctx, buildActionConfigErrorReq(ctx, schResp.Schema), invokeResp)
		if !invokeResp.Diagnostics.HasError() {
			t.Fatal("expected error for invalid config")
		}
	})
}
