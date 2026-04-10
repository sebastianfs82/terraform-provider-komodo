// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package client

// CreateNetworkRequest is the request payload for CreateNetwork.
type CreateNetworkRequest struct {
	Server string `json:"server"`
	Name   string `json:"name"`
}

// NetworkListItem represents a docker network entry from the ListDockerNetworks response.
type NetworkListItem struct {
	Name        *string `json:"name"`
	ID          *string `json:"id"`
	Created     *string `json:"created"`
	Scope       *string `json:"scope"`
	Driver      *string `json:"driver"`
	EnableIPv6  *bool   `json:"enable_ipv6"`
	IPAMDriver  *string `json:"ipam_driver"`
	IPAMSubnet  *string `json:"ipam_subnet"`
	IPAMGateway *string `json:"ipam_gateway"`
	Internal    *bool   `json:"internal"`
	Attachable  *bool   `json:"attachable"`
	Ingress     *bool   `json:"ingress"`
	InUse       bool    `json:"in_use"`
}
