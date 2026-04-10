// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

//go:build ignore

// patch_docs sets the subcategory frontmatter field on generated docs so the
// Terraform Registry groups pages by domain rather than by type.
// It is invoked automatically by `make generate` via a //go:generate directive
// in tools/tools.go.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// subcategoryRules maps filename stems (without .md extension and without the
// komodo_ prefix) to their subcategory. Rules are evaluated in order; the
// first match whose prefix matches the stem wins.
var subcategoryRules = []struct {
	prefix   string
	category string
}{
	// Stacks
	{"stack", "Stacks"},
	// Repos
	{"repo", "Repos"},
	// Builds — builder must come before build so "builder" matches first
	{"builder", "Builds"},
	{"run_build", "Builds"},
	{"build", "Builds"},
	// Deployments
	{"deploy_deployment", "Deployments"},
	{"destroy_deployment", "Deployments"},
	{"pause_deployment", "Deployments"},
	{"unpause_deployment", "Deployments"},
	{"pull_deployment", "Deployments"},
	{"restart_deployment", "Deployments"},
	{"start_deployment", "Deployments"},
	{"stop_deployment", "Deployments"},
	{"deployment", "Deployments"},
	// Procedures
	{"run_procedure", "Procedures"},
	{"procedure", "Procedures"},
	// Actions — run_action before action so it matches first
	{"run_action", "Actions"},
	{"action", "Actions"},
	// Servers
	{"server_prune", "Servers"},
	{"server", "Servers"},
	// Networks
	{"network", "Networks"},
	// Alerters
	{"test_alerter", "Alerters"},
	{"alerter", "Alerters"},
	// Resource Syncs
	{"resource_sync", "Resource Syncs"},
	{"run_sync", "Resource Syncs"},
	// Users & Access — user_group before user; service_user before user
	{"user_group", "Users & Access"},
	{"service_user", "Users & Access"},
	{"user", "Users & Access"},
	{"api_key", "Users & Access"},
	{"onboarding_key", "Users & Access"},
	// Configuration
	{"variable", "Configuration"},
	{"tag", "Configuration"},
	{"provider_account", "Configuration"},
	{"registry_account", "Configuration"},
}

var subcategoryRe = regexp.MustCompile(`(?m)^subcategory:\s*"[^"]*"`)

func subcategoryFor(name string) string {
	for _, rule := range subcategoryRules {
		if strings.HasPrefix(name, rule.prefix) {
			return rule.category
		}
	}
	return ""
}

func patchFile(path string) error {
	base := strings.TrimSuffix(filepath.Base(path), ".md")
	category := subcategoryFor(base)
	if category == "" {
		return nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	replacement := fmt.Sprintf(`subcategory: "%s"`, category)
	patched := subcategoryRe.ReplaceAllString(string(content), replacement)
	if patched == string(content) {
		return nil // nothing to do
	}

	return os.WriteFile(path, []byte(patched), 0644)
}

func walkDir(dir string) error {
	return filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".md") {
			if err := patchFile(path); err != nil {
				return fmt.Errorf("patching %s: %w", path, err)
			}
		}
		return nil
	})
}

func main() {
	docsDir := flag.String("docs-dir", "docs", "path to the docs directory")
	flag.Parse()

	if err := walkDir(*docsDir); err != nil {
		fmt.Fprintf(os.Stderr, "patch_docs: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("patch_docs: subcategories applied successfully")
}
