package client

type VSwitch struct {
	ID            int                   `json:"id"`
	Name          string                `json:"name"`
	VLAN          int                   `json:"vlan"`
	Cancelled     bool                  `json:"cancelled"`
	Servers       []VSwitchServer       `json:"servers,omitempty"`
	Subnets       []VSwitchSubnet       `json:"subnets,omitempty"`
	CloudNetworks []VSwitchCloudNetwork `json:"cloud_networks,omitempty"`
}

type VSwitchServer struct {
	ServerIP      string `json:"server_ip"`
	ServerIPv6Net string `json:"server_ipv6_net"`
	ServerNumber  int    `json:"server_number"`
	Status        string `json:"status"`
}

type VSwitchSubnet struct {
	IP      string `json:"ip"`
	Mask    int    `json:"mask"`
	Gateway string `json:"gateway"`
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
