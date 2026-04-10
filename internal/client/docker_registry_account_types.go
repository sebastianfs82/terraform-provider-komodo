// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package client

type DockerRegistryAccount struct {
	ID       OID    `json:"_id"`
	Domain   string `json:"domain"`
	Username string `json:"username"`
	Token    string `json:"token"`
}

type CreateDockerRegistryAccountRequest struct {
	Domain   string `json:"domain"`
	Username string `json:"username"`
	Token    string `json:"token"`
}
