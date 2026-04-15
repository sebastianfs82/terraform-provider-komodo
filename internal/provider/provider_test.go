// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/echoprovider"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"
)

// testAccProtoV6ProviderFactories is used to instantiate a provider during acceptance testing.
// The factory function is called for each Terraform CLI command to create a provider
// server that the CLI can connect to and interact with.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"komodo": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccProtoV6ProviderFactoriesWithEcho includes the echo provider alongside the komodo provider.
// It allows for testing assertions on data returned by an ephemeral resource during Open.
// The echoprovider is used to arrange tests by echoing ephemeral data into the Terraform state.
// This lets the data be referenced in test assertions with state checks.
var _ = map[string]func() (tfprotov6.ProviderServer, error){
	"komodo": providerserver.NewProtocol6WithError(New("test")()),
	"echo":   echoprovider.NewProviderServer(),
}

func testAccPreCheck(t *testing.T) {
	// Verify that required environment variables are set for acceptance tests
	if v := os.Getenv("KOMODO_ENDPOINT"); v == "" {
		t.Fatal("KOMODO_ENDPOINT must be set for acceptance tests")
	}

	if v := os.Getenv("KOMODO_USERNAME"); v == "" {
		t.Fatal("KOMODO_USERNAME must be set for acceptance tests")
	}

	if v := os.Getenv("KOMODO_PASSWORD"); v == "" {
		t.Fatal("KOMODO_PASSWORD must be set for acceptance tests")
	}
}

// testAccLookupServerID returns the ID of the first available server in the Komodo instance.
// Falls back to KOMODO_TEST_SERVER_ID if set. Skips the test if no servers are found.
func testAccLookupServerID(t *testing.T, skipMsg string) string {
	t.Helper()
	if v := os.Getenv("KOMODO_TEST_SERVER_ID"); v != "" {
		return v
	}
	c := client.NewClient(
		os.Getenv("KOMODO_ENDPOINT"),
		os.Getenv("KOMODO_USERNAME"),
		os.Getenv("KOMODO_PASSWORD"),
	)
	servers, err := c.ListServers(context.Background())
	if err != nil {
		t.Skipf("skipping %s: unable to list servers: %v", skipMsg, err)
	}
	if len(servers) == 0 {
		t.Skipf("skipping %s: no servers found in Komodo instance", skipMsg)
	}
	return servers[0].ID.OID
}
