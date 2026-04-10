// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package client

type GitProviderAccount struct {
	ID           OID    `json:"_id"`
	Domain       string `json:"domain"`
	HttpsEnabled bool   `json:"https"`
	Username     string `json:"username"`
	Token        string `json:"token"`
}

type CreateGitProviderAccountRequest struct {
	Domain       string `json:"domain"`
	HttpsEnabled bool   `json:"https"`
	Username     string `json:"username"`
	Token        string `json:"token"`
}
