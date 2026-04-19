// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"math/big"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// ─── StackAction tests ────────────────────────────────────────────────────────

func TestUnitStackAction_metadata(t *testing.T) {
	a := NewStackAction()
	req := action.MetadataRequest{ProviderTypeName: "komodo"}
	resp := &action.MetadataResponse{}
	a.Metadata(context.Background(), req, resp)
	if resp.TypeName != "komodo_stack" {
		t.Fatalf("expected TypeName komodo_stack, got %q", resp.TypeName)
	}
}

func TestUnitStackAction_schema(t *testing.T) {
	a := &StackAction{}
	resp := &action.SchemaResponse{}
	a.Schema(context.Background(), action.SchemaRequest{}, resp)
	if resp.Schema.MarkdownDescription == "" {
		t.Fatal("expected non-empty schema MarkdownDescription")
	}
	for _, attr := range []string{"id", "action", "services", "stop_time", "remove_orphans",
		"service", "command", "no_tty", "no_deps", "detach", "service_ports", "env",
		"workdir", "user", "entrypoint", "pull"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Fatalf("expected %q attribute in schema", attr)
		}
	}
}

func TestUnitStackAction_configure(t *testing.T) {
	ctx := context.Background()

	t.Run("nil_noop", func(t *testing.T) {
		a := &StackAction{}
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
		a := &StackAction{}
		cfgResp := &action.ConfigureResponse{}
		a.Configure(ctx, action.ConfigureRequest{ProviderData: "wrong"}, cfgResp)
		if !cfgResp.Diagnostics.HasError() {
			t.Fatal("expected error for wrong ProviderData type")
		}
	})

	t.Run("valid_client", func(t *testing.T) {
		a := &StackAction{}
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

// stackActionSchema returns the schema for StackAction.
func stackActionSchema(ctx context.Context) action.SchemaResponse {
	a := &StackAction{}
	resp := &action.SchemaResponse{}
	a.Schema(ctx, action.SchemaRequest{}, resp)
	return *resp
}

// buildStackActionInvokeReq constructs a StackAction InvokeRequest with all optional fields null.
func buildStackActionInvokeReq(ctx context.Context, schm action.SchemaResponse, stackID, act string) action.InvokeRequest {
	schemaType := schm.Schema.Type().TerraformType(ctx)
	raw := tftypes.NewValue(schemaType, map[string]tftypes.Value{
		"id":             tftypes.NewValue(tftypes.String, stackID),
		"action":         tftypes.NewValue(tftypes.String, act),
		"services":       tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"stop_time":      tftypes.NewValue(tftypes.Number, nil),
		"remove_orphans": tftypes.NewValue(tftypes.Bool, nil),
		"service":        tftypes.NewValue(tftypes.String, nil),
		"command":        tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"no_tty":         tftypes.NewValue(tftypes.Bool, nil),
		"no_deps":        tftypes.NewValue(tftypes.Bool, nil),
		"detach":         tftypes.NewValue(tftypes.Bool, nil),
		"service_ports":  tftypes.NewValue(tftypes.Bool, nil),
		"env":            tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
		"workdir":        tftypes.NewValue(tftypes.String, nil),
		"user":           tftypes.NewValue(tftypes.String, nil),
		"entrypoint":     tftypes.NewValue(tftypes.String, nil),
		"pull":           tftypes.NewValue(tftypes.Bool, nil),
	})
	return action.InvokeRequest{Config: tfsdk.Config{Schema: schm.Schema, Raw: raw}}
}

func TestUnitStackAction_invoke_simple(t *testing.T) {
	ctx := context.Background()

	// Actions that take only id/services/stop_time in their variants —
	// all tested with all optional fields set to null.
	simpleActions := []string{"deploy", "deploy_if_changed", "pause", "unpause", "pull", "restart", "start", "stop"}

	for _, act := range simpleActions {
		t.Run(act+"_success", func(t *testing.T) {
			_, c := newActionSuccessMockServer(t)
			a := &StackAction{client: c}
			schResp := stackActionSchema(ctx)
			invokeResp := &action.InvokeResponse{}
			a.Invoke(ctx, buildStackActionInvokeReq(ctx, schResp, "stack-id", act), invokeResp)
			if invokeResp.Diagnostics.HasError() {
				t.Fatalf("unexpected error for action %q: %s", act, invokeResp.Diagnostics)
			}
		})

		t.Run(act+"_client_error", func(t *testing.T) {
			_, c := newActionErrorMockServer(t)
			a := &StackAction{client: c}
			schResp := stackActionSchema(ctx)
			invokeResp := &action.InvokeResponse{}
			a.Invoke(ctx, buildStackActionInvokeReq(ctx, schResp, "stack-id", act), invokeResp)
			if !invokeResp.Diagnostics.HasError() {
				t.Fatalf("expected error for action %q on client failure", act)
			}
		})
	}

	// deploy with stop_time set
	t.Run("deploy_with_stop_time", func(t *testing.T) {
		_, c := newActionSuccessMockServer(t)
		a := &StackAction{client: c}
		schResp := stackActionSchema(ctx)
		schemaType := schResp.Schema.Type().TerraformType(ctx)
		raw := tftypes.NewValue(schemaType, map[string]tftypes.Value{
			"id":             tftypes.NewValue(tftypes.String, "stack-id"),
			"action":         tftypes.NewValue(tftypes.String, "deploy"),
			"services":       tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
			"stop_time":      tftypes.NewValue(tftypes.Number, new(big.Float).SetInt64(30)),
			"remove_orphans": tftypes.NewValue(tftypes.Bool, nil),
			"service":        tftypes.NewValue(tftypes.String, nil),
			"command":        tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
			"no_tty":         tftypes.NewValue(tftypes.Bool, nil),
			"no_deps":        tftypes.NewValue(tftypes.Bool, nil),
			"detach":         tftypes.NewValue(tftypes.Bool, nil),
			"service_ports":  tftypes.NewValue(tftypes.Bool, nil),
			"env":            tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
			"workdir":        tftypes.NewValue(tftypes.String, nil),
			"user":           tftypes.NewValue(tftypes.String, nil),
			"entrypoint":     tftypes.NewValue(tftypes.String, nil),
			"pull":           tftypes.NewValue(tftypes.Bool, nil),
		})
		invokeResp := &action.InvokeResponse{}
		a.Invoke(ctx, action.InvokeRequest{Config: tfsdk.Config{Schema: schResp.Schema, Raw: raw}}, invokeResp)
		if invokeResp.Diagnostics.HasError() {
			t.Fatalf("unexpected error: %s", invokeResp.Diagnostics)
		}
	})

	// deploy_if_changed with stop_time set
	t.Run("deploy_if_changed_with_stop_time", func(t *testing.T) {
		_, c := newActionSuccessMockServer(t)
		a := &StackAction{client: c}
		schResp := stackActionSchema(ctx)
		schemaType := schResp.Schema.Type().TerraformType(ctx)
		raw := tftypes.NewValue(schemaType, map[string]tftypes.Value{
			"id":             tftypes.NewValue(tftypes.String, "stack-id"),
			"action":         tftypes.NewValue(tftypes.String, "deploy_if_changed"),
			"services":       tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
			"stop_time":      tftypes.NewValue(tftypes.Number, new(big.Float).SetInt64(30)),
			"remove_orphans": tftypes.NewValue(tftypes.Bool, nil),
			"service":        tftypes.NewValue(tftypes.String, nil),
			"command":        tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
			"no_tty":         tftypes.NewValue(tftypes.Bool, nil),
			"no_deps":        tftypes.NewValue(tftypes.Bool, nil),
			"detach":         tftypes.NewValue(tftypes.Bool, nil),
			"service_ports":  tftypes.NewValue(tftypes.Bool, nil),
			"env":            tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
			"workdir":        tftypes.NewValue(tftypes.String, nil),
			"user":           tftypes.NewValue(tftypes.String, nil),
			"entrypoint":     tftypes.NewValue(tftypes.String, nil),
			"pull":           tftypes.NewValue(tftypes.Bool, nil),
		})
		invokeResp := &action.InvokeResponse{}
		a.Invoke(ctx, action.InvokeRequest{Config: tfsdk.Config{Schema: schResp.Schema, Raw: raw}}, invokeResp)
		if invokeResp.Diagnostics.HasError() {
			t.Fatalf("unexpected error: %s", invokeResp.Diagnostics)
		}
	})

	// stop with stop_time set
	t.Run("stop_with_stop_time", func(t *testing.T) {
		_, c := newActionSuccessMockServer(t)
		a := &StackAction{client: c}
		schResp := stackActionSchema(ctx)
		schemaType := schResp.Schema.Type().TerraformType(ctx)
		raw := tftypes.NewValue(schemaType, map[string]tftypes.Value{
			"id":             tftypes.NewValue(tftypes.String, "stack-id"),
			"action":         tftypes.NewValue(tftypes.String, "stop"),
			"services":       tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
			"stop_time":      tftypes.NewValue(tftypes.Number, new(big.Float).SetInt64(30)),
			"remove_orphans": tftypes.NewValue(tftypes.Bool, nil),
			"service":        tftypes.NewValue(tftypes.String, nil),
			"command":        tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
			"no_tty":         tftypes.NewValue(tftypes.Bool, nil),
			"no_deps":        tftypes.NewValue(tftypes.Bool, nil),
			"detach":         tftypes.NewValue(tftypes.Bool, nil),
			"service_ports":  tftypes.NewValue(tftypes.Bool, nil),
			"env":            tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
			"workdir":        tftypes.NewValue(tftypes.String, nil),
			"user":           tftypes.NewValue(tftypes.String, nil),
			"entrypoint":     tftypes.NewValue(tftypes.String, nil),
			"pull":           tftypes.NewValue(tftypes.Bool, nil),
		})
		invokeResp := &action.InvokeResponse{}
		a.Invoke(ctx, action.InvokeRequest{Config: tfsdk.Config{Schema: schResp.Schema, Raw: raw}}, invokeResp)
		if invokeResp.Diagnostics.HasError() {
			t.Fatalf("unexpected error: %s", invokeResp.Diagnostics)
		}
	})
}

func TestUnitStackAction_invoke_destroy(t *testing.T) {
	ctx := context.Background()
	schResp := stackActionSchema(ctx)
	schemaType := schResp.Schema.Type().TerraformType(ctx)

	t.Run("destroy_success", func(t *testing.T) {
		_, c := newActionSuccessMockServer(t)
		a := &StackAction{client: c}
		invokeResp := &action.InvokeResponse{}
		a.Invoke(ctx, buildStackActionInvokeReq(ctx, schResp, "stack-id", "destroy"), invokeResp)
		if invokeResp.Diagnostics.HasError() {
			t.Fatalf("unexpected error: %s", invokeResp.Diagnostics)
		}
	})

	t.Run("destroy_with_remove_orphans", func(t *testing.T) {
		_, c := newActionSuccessMockServer(t)
		a := &StackAction{client: c}
		raw := tftypes.NewValue(schemaType, map[string]tftypes.Value{
			"id":             tftypes.NewValue(tftypes.String, "stack-id"),
			"action":         tftypes.NewValue(tftypes.String, "destroy"),
			"services":       tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
			"stop_time":      tftypes.NewValue(tftypes.Number, nil),
			"remove_orphans": tftypes.NewValue(tftypes.Bool, true),
			"service":        tftypes.NewValue(tftypes.String, nil),
			"command":        tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
			"no_tty":         tftypes.NewValue(tftypes.Bool, nil),
			"no_deps":        tftypes.NewValue(tftypes.Bool, nil),
			"detach":         tftypes.NewValue(tftypes.Bool, nil),
			"service_ports":  tftypes.NewValue(tftypes.Bool, nil),
			"env":            tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
			"workdir":        tftypes.NewValue(tftypes.String, nil),
			"user":           tftypes.NewValue(tftypes.String, nil),
			"entrypoint":     tftypes.NewValue(tftypes.String, nil),
			"pull":           tftypes.NewValue(tftypes.Bool, nil),
		})
		invokeReq := action.InvokeRequest{Config: tfsdk.Config{Schema: schResp.Schema, Raw: raw}}
		invokeResp := &action.InvokeResponse{}
		a.Invoke(ctx, invokeReq, invokeResp)
		if invokeResp.Diagnostics.HasError() {
			t.Fatalf("unexpected error: %s", invokeResp.Diagnostics)
		}
	})

	t.Run("destroy_with_stop_time", func(t *testing.T) {
		_, c := newActionSuccessMockServer(t)
		a := &StackAction{client: c}
		raw := tftypes.NewValue(schemaType, map[string]tftypes.Value{
			"id":             tftypes.NewValue(tftypes.String, "stack-id"),
			"action":         tftypes.NewValue(tftypes.String, "destroy"),
			"services":       tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
			"stop_time":      tftypes.NewValue(tftypes.Number, new(big.Float).SetInt64(30)),
			"remove_orphans": tftypes.NewValue(tftypes.Bool, nil),
			"service":        tftypes.NewValue(tftypes.String, nil),
			"command":        tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
			"no_tty":         tftypes.NewValue(tftypes.Bool, nil),
			"no_deps":        tftypes.NewValue(tftypes.Bool, nil),
			"detach":         tftypes.NewValue(tftypes.Bool, nil),
			"service_ports":  tftypes.NewValue(tftypes.Bool, nil),
			"env":            tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
			"workdir":        tftypes.NewValue(tftypes.String, nil),
			"user":           tftypes.NewValue(tftypes.String, nil),
			"entrypoint":     tftypes.NewValue(tftypes.String, nil),
			"pull":           tftypes.NewValue(tftypes.Bool, nil),
		})
		invokeResp := &action.InvokeResponse{}
		a.Invoke(ctx, action.InvokeRequest{Config: tfsdk.Config{Schema: schResp.Schema, Raw: raw}}, invokeResp)
		if invokeResp.Diagnostics.HasError() {
			t.Fatalf("unexpected error: %s", invokeResp.Diagnostics)
		}
	})

	t.Run("destroy_client_error", func(t *testing.T) {
		_, c := newActionErrorMockServer(t)
		a := &StackAction{client: c}
		invokeResp := &action.InvokeResponse{}
		a.Invoke(ctx, buildStackActionInvokeReq(ctx, schResp, "stack-id", "destroy"), invokeResp)
		if !invokeResp.Diagnostics.HasError() {
			t.Fatal("expected error on client failure")
		}
	})
}

func TestUnitStackAction_invoke_deploy_with_services(t *testing.T) {
	ctx := context.Background()
	schResp := stackActionSchema(ctx)
	schemaType := schResp.Schema.Type().TerraformType(ctx)

	_, c := newActionSuccessMockServer(t)
	a := &StackAction{client: c}

	raw := tftypes.NewValue(schemaType, map[string]tftypes.Value{
		"id":     tftypes.NewValue(tftypes.String, "stack-id"),
		"action": tftypes.NewValue(tftypes.String, "deploy"),
		"services": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
			tftypes.NewValue(tftypes.String, "web"),
			tftypes.NewValue(tftypes.String, "db"),
		}),
		"stop_time":      tftypes.NewValue(tftypes.Number, nil),
		"remove_orphans": tftypes.NewValue(tftypes.Bool, nil),
		"service":        tftypes.NewValue(tftypes.String, nil),
		"command":        tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"no_tty":         tftypes.NewValue(tftypes.Bool, nil),
		"no_deps":        tftypes.NewValue(tftypes.Bool, nil),
		"detach":         tftypes.NewValue(tftypes.Bool, nil),
		"service_ports":  tftypes.NewValue(tftypes.Bool, nil),
		"env":            tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
		"workdir":        tftypes.NewValue(tftypes.String, nil),
		"user":           tftypes.NewValue(tftypes.String, nil),
		"entrypoint":     tftypes.NewValue(tftypes.String, nil),
		"pull":           tftypes.NewValue(tftypes.Bool, nil),
	})

	invokeResp := &action.InvokeResponse{}
	a.Invoke(ctx, action.InvokeRequest{Config: tfsdk.Config{Schema: schResp.Schema, Raw: raw}}, invokeResp)
	if invokeResp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %s", invokeResp.Diagnostics)
	}
}

func TestUnitStackAction_invoke_run_service(t *testing.T) {
	ctx := context.Background()
	schResp := stackActionSchema(ctx)
	schemaType := schResp.Schema.Type().TerraformType(ctx)

	t.Run("run_service_success", func(t *testing.T) {
		_, c := newActionSuccessMockServer(t)
		a := &StackAction{client: c}
		raw := tftypes.NewValue(schemaType, map[string]tftypes.Value{
			"id":             tftypes.NewValue(tftypes.String, "stack-id"),
			"action":         tftypes.NewValue(tftypes.String, "run_service"),
			"services":       tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
			"stop_time":      tftypes.NewValue(tftypes.Number, nil),
			"remove_orphans": tftypes.NewValue(tftypes.Bool, nil),
			"service":        tftypes.NewValue(tftypes.String, "web"),
			"command": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "echo"),
				tftypes.NewValue(tftypes.String, "hello"),
			}),
			"no_tty":        tftypes.NewValue(tftypes.Bool, true),
			"no_deps":       tftypes.NewValue(tftypes.Bool, false),
			"detach":        tftypes.NewValue(tftypes.Bool, false),
			"service_ports": tftypes.NewValue(tftypes.Bool, false),
			"env": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{
				"MY_VAR": tftypes.NewValue(tftypes.String, "value"),
			}),
			"workdir":    tftypes.NewValue(tftypes.String, "/app"),
			"user":       tftypes.NewValue(tftypes.String, "nobody"),
			"entrypoint": tftypes.NewValue(tftypes.String, "/bin/sh"),
			"pull":       tftypes.NewValue(tftypes.Bool, true),
		})
		invokeResp := &action.InvokeResponse{}
		a.Invoke(ctx, action.InvokeRequest{Config: tfsdk.Config{Schema: schResp.Schema, Raw: raw}}, invokeResp)
		if invokeResp.Diagnostics.HasError() {
			t.Fatalf("unexpected error: %s", invokeResp.Diagnostics)
		}
	})

	t.Run("run_service_missing_service", func(t *testing.T) {
		_, c := newActionSuccessMockServer(t)
		a := &StackAction{client: c}
		// service is null → validation error
		invokeResp := &action.InvokeResponse{}
		a.Invoke(ctx, buildStackActionInvokeReq(ctx, schResp, "stack-id", "run_service"), invokeResp)
		if !invokeResp.Diagnostics.HasError() {
			t.Fatal("expected error when service is not set for run_service")
		}
	})

	t.Run("run_service_client_error", func(t *testing.T) {
		_, c := newActionErrorMockServer(t)
		a := &StackAction{client: c}
		raw := tftypes.NewValue(schemaType, map[string]tftypes.Value{
			"id":             tftypes.NewValue(tftypes.String, "stack-id"),
			"action":         tftypes.NewValue(tftypes.String, "run_service"),
			"services":       tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
			"stop_time":      tftypes.NewValue(tftypes.Number, nil),
			"remove_orphans": tftypes.NewValue(tftypes.Bool, nil),
			"service":        tftypes.NewValue(tftypes.String, "web"),
			"command":        tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
			"no_tty":         tftypes.NewValue(tftypes.Bool, nil),
			"no_deps":        tftypes.NewValue(tftypes.Bool, nil),
			"detach":         tftypes.NewValue(tftypes.Bool, nil),
			"service_ports":  tftypes.NewValue(tftypes.Bool, nil),
			"env":            tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
			"workdir":        tftypes.NewValue(tftypes.String, nil),
			"user":           tftypes.NewValue(tftypes.String, nil),
			"entrypoint":     tftypes.NewValue(tftypes.String, nil),
			"pull":           tftypes.NewValue(tftypes.Bool, nil),
		})
		invokeResp := &action.InvokeResponse{}
		a.Invoke(ctx, action.InvokeRequest{Config: tfsdk.Config{Schema: schResp.Schema, Raw: raw}}, invokeResp)
		if !invokeResp.Diagnostics.HasError() {
			t.Fatal("expected error on client failure")
		}
	})
}

func TestUnitStackAction_invoke_invalid(t *testing.T) {
	ctx := context.Background()
	schResp := stackActionSchema(ctx)
	_, c := newActionSuccessMockServer(t)
	a := &StackAction{client: c}
	invokeResp := &action.InvokeResponse{}
	a.Invoke(ctx, buildStackActionInvokeReq(ctx, schResp, "stack-id", "bogus"), invokeResp)
	if !invokeResp.Diagnostics.HasError() {
		t.Fatal("expected error for invalid action")
	}
}

func TestUnitStackAction_invoke_config_error(t *testing.T) {
	ctx := context.Background()
	_, c := newActionSuccessMockServer(t)
	a := &StackAction{client: c}
	schResp := stackActionSchema(ctx)
	invokeResp := &action.InvokeResponse{}
	a.Invoke(ctx, buildActionConfigErrorReq(ctx, schResp.Schema), invokeResp)
	if !invokeResp.Diagnostics.HasError() {
		t.Fatal("expected error for invalid config")
	}
}
