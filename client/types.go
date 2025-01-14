package client

import (
	"hcloud-robot-provider/shared"
	"net/http"
)

type NotFoundError struct {
	Message string
}

type VSwitchCloudNetwork struct {
	ID      int    `json:"id"`
	IP      string `json:"ip"`
	Mask    int    `json:"mask"`
	Gateway string `json:"gateway"`
}

type Server struct {
	IP     string `json:"server_ip"`
	Number int    `json:"server_number"`
	Name   string `json:"server_name"`
	Status string `json:"status"`
}

type HetznerRobotClient struct {
	Config *shared.ProviderConfig
	Client *http.Client
}

type VSwitch struct {
	ID        int               `json:"id"`
	Name      string            `json:"name"`
	VLAN      int               `json:"vlan"`
	Cancelled bool              `json:"cancelled"`
	Servers   []VSwitchServer   `json:"servers"`
	Subnets   []VSwitchSubnet   `json:"subnets"`
	CloudNets []VSwitchCloudNet `json:"cloud_networks"`
}

type VSwitchServer struct {
	ServerNumber  int    `json:"server_number,omitempty"`
	ServerIP      string `json:"server_ip,omitempty"`
	ServerIPv6Net string `json:"server_ipv6_net,omitempty"`
	Status        string `json:"status,omitempty"`
}

type VSwitchSubnet struct {
	IP      string `json:"ip"`
	Mask    int    `json:"mask"`
	Gateway string `json:"gateway"`
}

type VSwitchCloudNet struct {
	ID      int    `json:"id"`
	IP      string `json:"ip"`
	Mask    int    `json:"mask"`
	Gateway string `json:"gateway"`
}
